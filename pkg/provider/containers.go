// Copyright (C) 2022-2024 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package provider

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	defaultLog "log"
	"strings"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/stringhelper"
	corev1 "k8s.io/api/core/v1"

	"github.com/redhat-openshift-ecosystem/openshift-preflight/artifacts"
	plibRuntime "github.com/redhat-openshift-ecosystem/openshift-preflight/certification"
	plibContainer "github.com/redhat-openshift-ecosystem/openshift-preflight/container"
)

var (
	// Certain tests that have been known to fail because of injected containers (such as Istio) that fail certain tests.
	ignoredContainerNames = []string{"istio-proxy"}
)

// ContainerImageIdentifier Represents a container image reference with optional tag or digest
//
// This structure holds the components of a container image: registry,
// repository name, an optional tag, and an optional digest. When both tag and
// digest are provided, the digest is used to uniquely identify the image,
// overriding the tag. The fields map directly to YAML and JSON keys for easy
// serialization.
type ContainerImageIdentifier struct {
	// Repository is the name of the image that you want to check if exists in the RedHat catalog
	Repository string `yaml:"repository" json:"repository"`

	// Registry is the name of the registry `docker.io` of the container
	// This is valid for container only and required field
	Registry string `yaml:"registry" json:"registry"`

	// Tag is the optional image tag. "latest" is implied if not specified
	Tag string `yaml:"tag" json:"tag"`

	// Digest is the image digest following the "@" in a URL, e.g. image@sha256:45b23dee08af5e43a7fea6c4cf9c25ccf269ee113168c19722f87876677c5cb2
	Digest string `yaml:"digest" json:"digest"`
}

// Container Represents a Kubernetes container with its status and metadata
//
// This structure holds information about a container running in a pod,
// including the core container spec, runtime details, node assignment, and
// namespace. It tracks the container’s current state through the status field
// and stores a unique identifier for the container instance. The struct also
// keeps an image identifier and any preflight test results that have been run
// against the container.
type Container struct {
	*corev1.Container
	Status                   corev1.ContainerStatus
	Namespace                string
	Podname                  string
	NodeName                 string
	Runtime                  string
	UID                      string
	ContainerImageIdentifier ContainerImageIdentifier
	PreflightResults         PreflightResultsDB
}

// NewContainer Creates an empty Container instance
//
// The function returns a pointer to a new Container struct with its embedded
// corev1.Container field initialized to an empty object. No parameters are
// required, and the returned value can be used as a starting point for building
// a container configuration.
func NewContainer() *Container {
	return &Container{
		Container: &corev1.Container{}, // initialize the corev1.Container object
	}
}

// Container.GetUID Retrieves the unique identifier of a container
//
// The method splits the container’s ID string on "://" and uses the last
// segment as the UID, handling empty results with an error. It logs debug
// messages indicating success or failure and returns the UID along with any
// error encountered.
func (c *Container) GetUID() (string, error) {
	split := strings.Split(c.Status.ContainerID, "://")
	uid := ""
	if len(split) > 0 {
		uid = split[len(split)-1]
	}
	if uid == "" {
		log.Debug("could not find uid of %s/%s/%s\n", c.Namespace, c.Podname, c.Name)
		return "", errors.New("cannot determine container UID")
	}
	log.Debug("uid of %s/%s/%s=%s\n", c.Namespace, c.Podname, c.Name, uid)
	return uid, nil
}

// Container.SetPreflightResults Stores preflight test results for a container image
//
// This method runs the OpenShift Preflight container checks on the image
// associated with the receiver, capturing logs and test outcomes. If the image
// has been processed before, it reuses cached results; otherwise it configures
// Docker credentials and optional insecure connections, executes the check,
// converts raw results into a structured database format, and caches them for
// future use. The function returns an error if any part of the execution fails.
func (c *Container) SetPreflightResults(preflightImageCache map[string]PreflightResultsDB, env *TestEnvironment) error {
	log.Info("Running Preflight container test for container %q with image %q", c, c.Image)

	// Short circuit if the image already exists in the cache
	if _, exists := preflightImageCache[c.Image]; exists {
		log.Info("Container image %q exists in the cache. Skipping this run.", c.Image)
		c.PreflightResults = preflightImageCache[c.Image]
		return nil
	}

	opts := []plibContainer.Option{}
	opts = append(opts, plibContainer.WithDockerConfigJSONFromFile(env.GetDockerConfigFile()))
	if env.IsPreflightInsecureAllowed() {
		log.Info("Insecure connections are being allowed to Preflight")
		opts = append(opts, plibContainer.WithInsecureConnection())
	}

	// Create artifacts handler
	artifactsWriter, err := artifacts.NewMapWriter()
	if err != nil {
		return err
	}
	ctx := artifacts.ContextWithWriter(context.TODO(), artifactsWriter)

	// Add logger output to the context
	logbytes := bytes.NewBuffer([]byte{})
	checklogger := defaultLog.Default()
	checklogger.SetOutput(logbytes)
	logger := stdr.New(checklogger)
	ctx = logr.NewContext(ctx, logger)

	check := plibContainer.NewCheck(c.Image, opts...)

	results, runtimeErr := check.Run(ctx)
	if runtimeErr != nil {
		_, checks, err := check.List(ctx)
		if err != nil {
			return fmt.Errorf("could not get preflight container test list")
		}

		results.TestedImage = c.Image
		for _, c := range checks {
			results.PassedOverall = false
			result := plibRuntime.Result{Check: c, ElapsedTime: 0}
			results.Errors = append(results.Errors, *result.WithError(runtimeErr))
		}
	}

	// Take all of the Preflight logs and stick them into our log.
	log.Info("%s", logbytes.String())

	// Store the Preflight test results into the container's PreflightResults var and into the cache.
	resultsDB := GetPreflightResultsDB(&results)
	c.PreflightResults = resultsDB
	preflightImageCache[c.Image] = resultsDB
	return nil
}

// Container.StringLong Formats container details into a readable string
//
// This method assembles key fields from the container such as node name,
// namespace, pod name, container name, UID, and runtime into a single formatted
// line. It uses standard string formatting to produce a concise representation
// of the container’s identity. The resulting text is returned for logging or
// display purposes.
func (c *Container) StringLong() string {
	return fmt.Sprintf("node: %s ns: %s podName: %s containerName: %s containerUID: %s containerRuntime: %s",
		c.NodeName,
		c.Namespace,
		c.Podname,
		c.Name,
		c.Status.ContainerID,
		c.Runtime,
	)
}

// Container.String Formats container details into a readable string
//
// This method returns a string that describes the container by combining its
// name, pod name, and namespace in a single line. It uses standard formatting
// to create a clear human-readable representation of the container's identity
// within the cluster.
func (c *Container) String() string {
	return fmt.Sprintf("container: %s pod: %s ns: %s",
		c.Name,
		c.Podname,
		c.Namespace,
	)
}

// Container.HasIgnoredContainerName Determines if the container should be excluded from processing
//
// This method checks each name in a predefined ignore list against the
// container’s name, also treating any Istio proxy container as ignored. If a
// match is found it returns true; otherwise false. The result guides callers to
// skip containers that are not relevant for certain operations.
func (c *Container) HasIgnoredContainerName() bool {
	for _, ign := range ignoredContainerNames {
		if c.IsIstioProxy() || strings.Contains(c.Name, ign) {
			return true
		}
	}
	return false
}

// Container.IsIstioProxy Determines if the container is an Istio proxy
//
// It checks whether the container’s name matches the predefined Istio proxy
// name. If it does, the function returns true; otherwise, it returns false.
// This simple check is used to identify and potentially ignore Istio-related
// containers in other logic.
func (c *Container) IsIstioProxy() bool {
	return c.Name == IstioProxyContainerName
}

// Container.HasExecProbes Checks if any probe uses an exec command
//
// The method inspects the container's liveness, readiness, and startup probes
// for non-nil Exec fields. It returns true if at least one of these probes has
// an Exec configuration defined; otherwise it returns false.
func (c *Container) HasExecProbes() bool {
	return c.LivenessProbe != nil && c.LivenessProbe.Exec != nil ||
		c.ReadinessProbe != nil && c.ReadinessProbe.Exec != nil ||
		c.StartupProbe != nil && c.StartupProbe.Exec != nil
}

// Container.IsTagEmpty Checks whether the container image tag is unset
//
// This method inspects the container's image identifier and compares its Tag
// field to an empty string. It returns true when no tag has been specified,
// indicating a default or unspecified tag. The result helps callers determine
// if they need to supply a tag value.
func (c *Container) IsTagEmpty() bool {
	return c.ContainerImageIdentifier.Tag == ""
}

// Container.IsReadOnlyRootFilesystem Determines if the container’s root filesystem is read‑only
//
// It logs a message indicating the container being tested, then checks whether
// the security context and its ReadOnlyRootFilesystem field are defined. If
// either is missing it returns false; otherwise it returns the value of that
// field.
func (c *Container) IsReadOnlyRootFilesystem(logger *log.Logger) bool {
	logger.Info("Testing Container %q", c)
	if c.SecurityContext == nil || c.SecurityContext.ReadOnlyRootFilesystem == nil {
		return false
	}
	return *c.SecurityContext.ReadOnlyRootFilesystem
}

// Container.IsContainerRunAsNonRoot Determines if a container should run as non-root
//
// The method checks the container’s security context for a RunAsNonRoot
// setting, falling back to an optional pod-level value if the container does
// not specify one. It returns a boolean indicating whether the container will
// run as non‑root and a descriptive reason explaining which level provided
// the decision. If neither level supplies a value, it reports that both are
// unset.
func (c *Container) IsContainerRunAsNonRoot(podRunAsNonRoot *bool) (isContainerRunAsNonRoot bool, reason string) {
	if c.SecurityContext != nil && c.SecurityContext.RunAsNonRoot != nil {
		return *c.SecurityContext.RunAsNonRoot, fmt.Sprintf("RunAsNonRoot is set to %t at the container level, overriding a %v value defined at pod level",
			*c.SecurityContext.RunAsNonRoot, stringhelper.PointerToString(podRunAsNonRoot))
	}

	if podRunAsNonRoot != nil {
		return *podRunAsNonRoot, fmt.Sprintf("RunAsNonRoot is set to nil at container level and inheriting a %t value from the pod level RunAsNonRoot setting", *podRunAsNonRoot)
	}

	return false, "RunAsNonRoot is set to nil at pod and container level"
}

// Container.IsContainerRunAsNonRootUserID checks whether the container is running as a non-root user
//
// The function evaluates the container’s security context to determine if it
// has a RunAsUser value different from zero, indicating a non‑root user ID.
// It also considers any pod-level RunAsUser setting that may be inherited when
// the container does not specify its own. The result is a boolean flag and a
// descriptive reason explaining which level provided the decision.
func (c *Container) IsContainerRunAsNonRootUserID(podRunAsNonRootUserID *int64) (isContainerRunAsNonRootUserID bool, reason string) {
	if c.SecurityContext != nil && c.SecurityContext.RunAsUser != nil {
		return *c.SecurityContext.RunAsUser != 0, fmt.Sprintf("RunAsUser is set to %v at the container level, overriding a %s value defined at pod level",
			*c.SecurityContext.RunAsUser, stringhelper.PointerToString(podRunAsNonRootUserID))
	}

	if podRunAsNonRootUserID != nil {
		return *podRunAsNonRootUserID != 0, fmt.Sprintf("RunAsUser is set to nil at container level and inheriting a %v value from the pod level RunAsUser setting", *podRunAsNonRootUserID)
	}

	return false, "RunAsUser is set to nil at pod and container level"
}
