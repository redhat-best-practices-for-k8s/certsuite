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

package main

import (
	"github.com/test-network-function/cnf-certification-test/pkg/certsuite"
	"github.com/test-network-function/cnf-certification-test/pkg/flags"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/webserver"

	"github.com/test-network-function/cnf-certification-test/internal/log"
)

func main() {
	certsuite.Startup()
	if *flags.ServerModeFlag {
		log.Info("Running CNF Certification Suite in web server mode")
		webserver.StartServer(*flags.OutputDir)
	} else {
		log.Info("Running CNF Certification Suite in stand-alone mode")
		err := certsuite.Run(*flags.LabelsFlag, *flags.OutputDir)
		if err != nil {
			log.Fatal("Failed to run CNF Certification Suite: %v", err)
		}
	}
	certsuite.Shutdown()
}
