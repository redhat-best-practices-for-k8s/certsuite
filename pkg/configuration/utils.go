// Copyright (C) 2020-2024 Red Hat, Inc.
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

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"gopkg.in/yaml.v3"
)

var (
	configuration = TestConfiguration{}
	confLoaded    = false
	parameters    = TestParameters{}
)

// LoadConfiguration loads the test configuration from a file once.
//
// It takes the path to a JSON configuration file, reads its contents,
// unmarshals it into a TestConfiguration struct, and returns that
// configuration along with any error encountered during reading or parsing.
// Subsequent calls will return the same cached configuration without re-reading the file.
func LoadConfiguration(filePath string) (TestConfiguration, error) {
	if confLoaded {
		log.Debug("config file already loaded, return previous element")
		return configuration, nil
	}

	log.Info("Loading config from file: %s", filePath)
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return configuration, err
	}

	err = yaml.Unmarshal(contents, &configuration)
	if err != nil {
		return configuration, err
	}

	// Set default namespace for the probe daemonset pods, in case it was not set.
	if configuration.ProbeDaemonSetNamespace == "" {
		log.Warn("No namespace configured for the probe daemonset. Defaulting to namespace %q", defaultProbeDaemonSetNamespace)
		configuration.ProbeDaemonSetNamespace = defaultProbeDaemonSetNamespace
	} else {
		log.Info("Namespace for probe daemonset: %s", configuration.ProbeDaemonSetNamespace)
	}

	confLoaded = true
	return configuration, nil
}

// GetTestParameters retrieves the current test configuration parameters.
//
// It ensures that the global configuration has been loaded and returns a pointer
// to the TestParameters structure containing all relevant settings for the
// test suite. If the configuration has not yet been initialized, it triggers
// the loading process before returning the parameters.
func GetTestParameters() *TestParameters {
	return &parameters
}
