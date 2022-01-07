// Copyright (C) 2020-2021 Red Hat, Inc.
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
	"gopkg.in/yaml.v2"
)

// LoadConfiguration return a function that loads
// the configuration from a file once
func LoadConfiguration() func(filePath string) (TestConfiguration, error) {
	env := TestConfiguration{}
	loaded := false
	return func(filePath string) (TestConfiguration, error) {
		if loaded {
			log.Debug("config file already loaded, return previous element")
			return env, nil
		}
		loaded = true
		log.Info("Loading config from file: ", filePath)
		contents, err := os.ReadFile(filePath)
		if err != nil {
			return env, err
		}

		err = yaml.Unmarshal(contents, &env)
		if err != nil {
			return env, err
		}
		return env, nil
	}
}

// LoadEnvironmentVariables return a function
// that loads the environment variables
func LoadEnvironmentVariables() func() (TestParameters, error) {
	s := TestParameters{}
	b := false
	return func() (TestParameters, error) {
		if b {
			log.Debug("environment variables already processed, return previous element")
			return s, nil
		}
		err := envconfig.Process("tnf", &s)
		if err != nil {
			log.Fatal(err.Error())
		}
		b = true
		return s, nil
	}
}
