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
	TestContainerIsCertifiedDigestIdentifier claim.Identifier
	TestHelmIsCertifiedIdentifier            claim.Identifier
	TestHelmVersionIdentifier                claim.Identifier
	TestOperatorIsCertifiedIdentifier        claim.Identifier
)

//nolint:funlen
func init() {
	TestContainerIsCertifiedDigestIdentifier = AddCatalogEntry(
		"container-is-certified-digest",
		common.AffiliatedCertTestKey,
		`Tests whether container images that are autodiscovered have passed the Red Hat Container Certification Program by their digest(CCP).`,
		ContainerIsCertifiedDigestRemediation,
		AffiliatedCert,
		TestContainerIsCertifiedDigestIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestHelmIsCertifiedIdentifier = AddCatalogEntry(
		"helmchart-is-certified",
		common.AffiliatedCertTestKey,
		`Tests whether helm charts listed in the cluster passed the Red Hat Helm Certification Program.`,
		HelmIsCertifiedRemediation,
		AffiliatedCert,
		TestHelmIsCertifiedIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)
	TestHelmVersionIdentifier = AddCatalogEntry(
		"helm-version",
		common.AffiliatedCertTestKey,
		`Test to check if the helm chart is v3`,
		HelmVersionV3Remediation,
		NoDocumentedProcess,
		TestHelmVersionIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestOperatorIsCertifiedIdentifier = AddCatalogEntry(
		"operator-is-certified",
		common.AffiliatedCertTestKey,
		`Tests whether the workload Operators listed in the configuration file have passed the Red Hat Operator Certification Program (OCP).`,
		OperatorIsCertifiedRemediation,
		AffiliatedCert,
		TestOperatorIsCertifiedIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)
}
