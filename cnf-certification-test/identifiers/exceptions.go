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

package identifiers

/*
	Use this file to store the strings for the various exception processes for each test in our suite.
	There are various tests that have a level of flexibility to their results depending on the situation and some that do not have
	an exception process.
*/

const (

	// Tests with exception processes
	// TODO: Add more exception processes if/when we encounter more opportunities with partners
	AutomountServiceTokenExceptionProcess = `Identify which Kubernetes APIs are required if you need to utilize automount service tokens.  Depending on
												which APIs are utilized, Red Hat possibly might make those APIs available to use via OpenShift.`

	ShutdownExceptionProcess = `Identify which pod is not conforming to the process and submit information as to why it cannot 
									use a prestop shutdown specification.`

	IsRedHatReleaseExceptionProcess = `Document which containers are not able to meet the RHEL-based container 
											requirement and if/when the base image can be updated.`

	SecConNonRootUserExceptionProcess = `If your application needs root user access, please document why your application cannot be ran as
											non-root and supply the reasoning for exception.`

	SecConCapabilitiesExceptionProcess = `Identify the pod that is needing special capabilities and document why  `

	// Tests that do not have an exception process but have additional insight
	UnalteredBaseImageExceptionProcess = `Images should not be changed during runtime.  There is no exception process for this.`

	// Generic Exception Process Message
	NoDocumentedProcess = `There is no documented exception process for this.`
)
