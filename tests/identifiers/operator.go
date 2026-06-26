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
	TestMultipleSameOperatorsIdentifier                              claim.Identifier
	TestOperatorAutomountTokens                                      claim.Identifier
	TestOperatorCatalogSourceBundleCountIdentifier                   claim.Identifier
	TestOperatorCrdSchemaIdentifier                                  claim.Identifier
	TestOperatorCrdVersioningIdentifier                              claim.Identifier
	TestOperatorHasSemanticVersioningIdentifier                      claim.Identifier
	TestOperatorInstallStatusSucceededIdentifier                     claim.Identifier
	TestOperatorIsInstalledViaOLMIdentifier                          claim.Identifier
	TestOperatorNoSCCAccess                                          claim.Identifier
	TestOperatorOlmSkipRange                                         claim.Identifier
	TestOperatorPodsNoHugepages                                      claim.Identifier
	TestOperatorRunAsNonRoot                                         claim.Identifier
	TestOperatorSingleCrdOwnerIdentifier                             claim.Identifier
	TestSingleOrMultiNamespacedOperatorInstallationInTenantNamespace claim.Identifier
)

//nolint:funlen
func init() {
	TestMultipleSameOperatorsIdentifier = AddCatalogEntry(
		"multiple-same-operators",
		common.OperatorTestKey,
		`Tests whether multiple instances of the same Operator CSV are installed.`,
		MultipleSameOperatorsRemediation,
		NoExceptions,
		TestMultipleSameOperatorsIdentifierDocLink,
		false,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)
	TestOperatorCatalogSourceBundleCountIdentifier = AddCatalogEntry(
		"catalogsource-bundle-count",
		common.OperatorTestKey,
		`Tests operator catalog source bundle count is less than 1000`,
		OperatorCatalogSourceBundleCountRemediation,
		NoExceptions,
		TestOperatorCatalogSourceBundleCountIdentifierDocLink,
		false,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestOperatorCrdSchemaIdentifier = AddCatalogEntry(
		"crd-openapi-schema",
		common.OperatorTestKey,
		`Tests whether an application Operator CRD is defined with OpenAPI spec.`,
		OperatorCrdSchemaIdentifierRemediation,
		NoExceptions,
		TestOperatorCrdSchemaIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestOperatorCrdVersioningIdentifier = AddCatalogEntry(
		"crd-versioning",
		common.OperatorTestKey,
		`Tests whether the Operator CRD has a valid versioning.`,
		OperatorCrdVersioningRemediation,
		NoExceptions,
		TestOperatorCrdVersioningIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestOperatorHasSemanticVersioningIdentifier = AddCatalogEntry(
		"semantic-versioning",
		common.OperatorTestKey,
		`Tests whether an application Operator has a valid semantic versioning.`,
		OperatorHasSemanticVersioningRemediation,
		NoExceptions,
		TestOperatorHasSemanticVersioningIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestOperatorInstallStatusSucceededIdentifier = AddCatalogEntry(
		"install-status-succeeded",
		common.OperatorTestKey,
		`Ensures that the target workload operators report "Succeeded" as their installation status.`,
		OperatorInstallStatusSucceededRemediation,
		NoExceptions,
		TestOperatorInstallStatusSucceededIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestOperatorIsInstalledViaOLMIdentifier = AddCatalogEntry(
		"install-source",
		common.OperatorTestKey,
		`Tests whether a workload Operator is installed via OLM.`,
		OperatorIsInstalledViaOLMRemediation,
		NoExceptions,
		TestOperatorIsInstalledViaOLMIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestOperatorNoSCCAccess = AddCatalogEntry(
		"install-status-no-privileges",
		common.OperatorTestKey,
		`Checks whether the operator needs access to Security Context Constraints. Test passes if clusterPermissions is not present in the CSV manifest or is present with no RBAC rules related to SCCs.`,
		OperatorNoPrivilegesRemediation,
		NoExceptions,
		TestOperatorNoPrivilegesDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestOperatorOlmSkipRange = AddCatalogEntry(
		"olm-skip-range",
		common.OperatorTestKey,
		`Test that checks the operator has a valid olm skip range.`,
		OperatorOlmSkipRangeRemediation,
		OperatorSkipRangeExceptionProcess,
		TestOperatorOlmSkipRangeDocLink,
		false,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagCommon)

	TestOperatorPodsNoHugepages = AddCatalogEntry(
		"pods-no-hugepages",
		common.OperatorTestKey,
		`Tests that the pods do not have hugepages enabled.`,
		OperatorPodsNoHugepagesRemediation,
		NoExceptions,
		TestOperatorPodsNoHugepagesDocLink,
		false,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagCommon)

	TestOperatorSingleCrdOwnerIdentifier = AddCatalogEntry(
		"single-crd-owner",
		common.OperatorTestKey,
		`Tests whether a CRD is owned by a single Operator.`,
		OperatorSingleCrdOwnerRemediation,
		NoExceptions,
		TestOperatorSingleCrdOwnerIdentifierDocLink,
		false,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestSingleOrMultiNamespacedOperatorInstallationInTenantNamespace = AddCatalogEntry(
		"single-or-multi-namespaced-allowed-in-tenant-namespaces",
		common.OperatorTestKey,
		`Verifies that only single/multi namespaced operators are installed in a tenant-dedicated namespace. The test fails if this namespace contains any installed operator with Own/All-namespaced install mode, unlabeled operators, operands of any operator installed elsewhere, or pods unrelated to any operator.`, //nolint:lll
		SingleOrMultiNamespacedOperatorInstallationInTenantNamespaceRemediation,
		NoExceptions,
		TestSingleOrMultiNamespacedOperatorInstallationInTenantNamespaceDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagExtended)
}
