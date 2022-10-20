// Copyright (C) 2022 Red Hat, Inc.
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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	corev1 "k8s.io/api/core/v1"

	plibRuntime "github.com/sebrandon1/openshift-preflight/certification/runtime"
	plib "github.com/sebrandon1/openshift-preflight/lib"
)

var (
	// Certain tests that have been known to fail because of injected containers (such as Istio) that fail certain tests.
	ignoredContainerNames = []string{"istio-proxy"}
)

type Container struct {
	*corev1.Container
	Status                   corev1.ContainerStatus
	Namespace                string
	Podname                  string
	NodeName                 string
	Runtime                  string
	UID                      string
	ContainerImageIdentifier configuration.ContainerImageIdentifier
	PreflightResults         plibRuntime.Results
}

func GetContainer() *Container {
	return &Container{}
}

func (c *Container) GetUID() (string, error) {
	split := strings.Split(c.Status.ContainerID, "://")
	uid := ""
	if len(split) > 0 {
		uid = split[len(split)-1]
	}
	if uid == "" {
		logrus.Debugln(fmt.Sprintf("could not find uid of %s/%s/%s\n", c.Namespace, c.Podname, c.Name))
		return "", errors.New("cannot determine container UID")
	}
	logrus.Debugln(fmt.Sprintf("uid of %s/%s/%s=%s\n", c.Namespace, c.Podname, c.Name, uid))
	return uid, nil
}

func (c *Container) SetPreflightResults() error {
	if _, err := os.Stat(fmt.Sprintf("artifacts/containers/%s", c.Image)); os.IsNotExist(err) {
		logrus.Infof("Directory artifacts/%s does not exist. Running preflight.", c.Image)

		preflightConfig := plibRuntime.NewManualContainerConfig(c.Image, "json", fmt.Sprintf("artifacts/containers/%s", c.Image), false, true)

		runner, err := plib.NewCheckContainerRunner(context.TODO(), preflightConfig, false)
		if err != nil {
			return err
		}

		err = plib.PreflightCheck(context.TODO(), runner.Cfg, runner.Pc, runner.Eng, runner.Formatter, runner.Rw, runner.Rs)
		if err != nil {
			return err
		}
	}

	// Read the JSON file
	f, err := os.ReadFile(fmt.Sprintf("artifacts/containers/%s/results.json", c.Image))
	if err != nil {
		return err
	}

	// Unmarshal the JSON blob into the preflight results struct
	var tempPreflightResults plibRuntime.Results
	err = json.Unmarshal(f, &tempPreflightResults)
	if err != nil {
		panic(err)
	}

	logrus.Infof("Storing container preflight results into object for %s", c.Image)
	c.PreflightResults = tempPreflightResults
	return nil
}

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
func (c *Container) String() string {
	return fmt.Sprintf("container: %s pod: %s ns: %s",
		c.Name,
		c.Podname,
		c.Namespace,
	)
}

func (c *Container) HasIgnoredContainerName() bool {
	for _, ign := range ignoredContainerNames {
		if strings.Contains(c.Name, ign) {
			return true
		}
	}
	return false
}
