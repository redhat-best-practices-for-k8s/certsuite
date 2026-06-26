// Copyright (C) 2021-2026 Red Hat, Inc.
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

import (
	"github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
)

var (
	TestContainerPortNameFormat claim.Identifier
	TestContainersImageTag      claim.Identifier
)

func init() {
	TestContainerPortNameFormat = AddCatalogEntry(
		"container-port-name-format",
		common.ManageabilityTestKey,
		"Check that the container's ports name follow the naming conventions. Name field in ContainerPort section must be of form `<protocol>[-<suffix>]`. More naming convention requirements may be released in future",
		ContainerPortNameFormatRemediation,
		NoExceptionProcessForExtendedTests,
		TestContainerPortNameFormatDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagExtended)
	TestContainersImageTag = AddCatalogEntry(
		"containers-image-tag",
		common.ManageabilityTestKey,
		`Check that image tag exists on containers.`,
		ContainersImageTagRemediation,
		NoExceptionProcessForExtendedTests,
		TestContainersImageTagDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagExtended)
}
