// Copyright (C) 2020-2023 Red Hat, Inc.
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

package configuration

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	configuration = TestConfiguration{}
	confLoaded    = false
	parameters    = TestParameters{}
)

func init() {
	log.Info("Saving environment variables & parameters.")
	err := envconfig.Process("tnf", &parameters)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Infof("Environment: %+v", parameters)
}

// LoadConfiguration return a function that loads
// the configuration from a file once
func LoadConfiguration(filePath string) (TestConfiguration, error) {
	if confLoaded {
		log.Debug("config file already loaded, return previous element")
		return configuration, nil
	}

	log.Info("Loading config from file: ", filePath)
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return configuration, err
	}

	err = yaml.Unmarshal(contents, &configuration)
	if err != nil {
		return configuration, err
	}

	// Set default namespace for the debug daemonset pods, in case it was not set.
	if configuration.DebugDaemonSetNamespace == "" {
		log.Warnf("No namespace configured for the debug DaemonSet. Defaulting to namespace %s", defaultDebugDaemonSetNamespace)
		configuration.DebugDaemonSetNamespace = defaultDebugDaemonSetNamespace
	} else {
		log.Infof("Namespace for debug DaemonSet: %s", configuration.DebugDaemonSetNamespace)
	}

	confLoaded = true
	return configuration, nil
}

func GetTestParameters() *TestParameters {
	return &parameters
}
