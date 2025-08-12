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

// ContainerImageIdentifier identifies a container image by registry, repository, and either tag or digest.
//
// It contains four string fields: Registry, Repository, Tag, and Digest.
// The Tag and Digest fields are mutually exclusive; if both are set, the Digest value is used.
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

// Container represents a Kubernetes container with metadata and status information used by the provider package.
//
// It holds identifiers, namespace, node, pod name, runtime details, and preflight results.
// The embedded corev1.Container provides access to spec fields such as image and security context.
// Methods on this type expose helpers for retrieving UID, checking probe configuration,
// determining if the container is an Istio proxy, validating runAsNonRoot settings,
// ensuring a read‑only root filesystem, and formatting its data as strings.
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

// NewContainer creates a new Container instance.
//
// It returns a pointer to a freshly initialized Container struct, ready for use
// in the provider package. The returned container has default field values set,
// but no runtime resources are allocated at this point. This function is intended
// as a constructor helper for internal logic that requires a clean Container
// object.
func NewContainer() *Container {
	return &Container{
		Container: &corev1.Container{}, // initialize the corev1.Container object
	}
}

// GetUID returns the container's UID as a string or an error if extraction fails.
//
// It examines the container’s ID field, which may contain a host identifier
// followed by a colon and the actual UID. The method splits on “:” and,
// depending on the resulting slice length, returns either the full ID or
// just the part after the colon. If the ID cannot be parsed, it logs a
// debug message and returns an empty string with an error.
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

// SetPreflightResults stores the preflight check results for a container.
//
// It receives a map keyed by container name containing PreflightResultsDB
// structures and a pointer to TestEnvironment used during the checks.
// The function runs each registered preflight test, collects their outcomes,
// and writes them into the provided map. It returns an error if any test fails
// or if there is a problem accessing Docker configuration or running the
// preflight checks.
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

// StringLong returns a human‑readable representation of the Container.
// It formats the container's name, image, and status into a single string.
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

// String returns a formatted string representation of the container.
//
// It uses fmt.Sprintf to build a human‑readable description that includes
// key fields such as the container name, image, state, and any other
// relevant metadata stored in the Container struct. The resulting string
// can be used for logging or debugging purposes.
func (c *Container) String() string {
	return fmt.Sprintf("container: %s pod: %s ns: %s",
		c.Name,
		c.Podname,
		c.Namespace,
	)
}

// HasIgnoredContainerName reports whether the container’s name is in the list of ignored names.
//
// It checks if the container matches known special cases such as an Istio proxy,
// and then looks up the container name in a global slice of names to ignore.
// The function returns true when the name should be excluded from standard
// validation logic, otherwise false.
func (c *Container) HasIgnoredContainerName() bool {
	for _, ign := range ignoredContainerNames {
		if c.IsIstioProxy() || strings.Contains(c.Name, ign) {
			return true
		}
	}
	return false
}

// IsIstioProxy reports whether the container is an Istio proxy instance.
//
// It examines the container's name against a known list of Istio
// proxy container names and returns true if a match is found.
// The function has no parameters and returns only a boolean.
func (c *Container) IsIstioProxy() bool {
	return c.Name == IstioProxyContainerName
}

// HasExecProbes reports if the container defines any exec probe.
//
// It checks the container’s readiness, liveness, and startup probe
// configurations for an Exec action and returns true if at least one
// of those probes is present, otherwise false.
func (c *Container) HasExecProbes() bool {
	return c.LivenessProbe != nil && c.LivenessProbe.Exec != nil ||
		c.ReadinessProbe != nil && c.ReadinessProbe.Exec != nil ||
		c.StartupProbe != nil && c.StartupProbe.Exec != nil
}

// IsTagEmpty reports whether the container's image tag is empty.
//
// It examines the Container’s Image field and determines if a tag component
// (the part after the last colon in the image name) has been omitted.
// The function returns true when no tag is present, otherwise false.
func (c *Container) IsTagEmpty() bool {
	return c.ContainerImageIdentifier.Tag == ""
}

// IsReadOnlyRootFilesystem reports whether the container’s root file system is mounted as read‑only.
//
// It takes a logger to record diagnostic information and returns true when the container
// has been configured with a read‑only root file system, otherwise false.
func (c *Container) IsReadOnlyRootFilesystem(logger *log.Logger) bool {
	logger.Info("Testing Container %q", c)
	if c.SecurityContext == nil || c.SecurityContext.ReadOnlyRootFilesystem == nil {
		return false
	}
	return *c.SecurityContext.ReadOnlyRootFilesystem
}

// IsContainerRunAsNonRoot checks if a container is configured to run as non‑root.
//
// It accepts a pointer to a boolean that indicates whether the container’s
// security context specifies runAsNonRoot. The function returns true when
// the flag is set to true, otherwise false. Additionally it returns a string
// describing the result: “runAsNonRoot=true” or “runAsNonRoot=false”. This
// information can be used in audit reports or logging.
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

// IsContainerRunAsNonRootUserID reports whether the container is configured to run with a non‑root user ID.
// It examines the RunAsUser field of the container specification, which may be nil or point to an int64 value.
// If the pointer is nil, it returns false and a message indicating that no RunAsUser was set.
// If the value is zero (root), it returns false with a message that the container runs as root.
// For any other positive user ID, it returns true and a message confirming non‑root execution.
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
