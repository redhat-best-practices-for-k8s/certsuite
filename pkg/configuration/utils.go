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

// LoadConfiguration Loads and parses a configuration file once
//
// The function reads the specified YAML file, unmarshals its contents into a
// TestConfiguration structure, and caches the result for subsequent calls. It
// logs progress and warns if the probe daemonset namespace is missing,
// defaulting it to a predefined value. Errors during reading or unmarshalling
// are returned alongside the configuration.
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

// GetTestParameters Retrieves the current global test configuration
//
// This function returns a pointer to the singleton TestParameters instance that
// holds all runtime settings for the certification suite. The parameters are
// initialized once at program start and can be modified through command‑line
// flags or environment variables before use. Subsequent calls return the same
// instance, allowing different parts of the application to read shared
// configuration values.
func GetTestParameters() *TestParameters {
	return &parameters
}
