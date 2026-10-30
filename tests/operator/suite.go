// Copyright (C) 2020-2026 Red Hat, Inc.
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

package operator

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksadapter"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	checksfn "github.com/redhat-best-practices-for-k8s/checks/operator"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
)

var env provider.TestEnvironment

func LoadChecks() {
	log.Debug("Loading %s suite checks", common.OperatorTestKey)

	checksGroup := checksdb.NewChecksGroup(common.OperatorTestKey).
		WithBeforeEachFn(checksdb.DefaultBeforeEachFn(func() { env = provider.GetTestEnvironment() }))

	checksadapter.AddCheck(checksGroup, "operator-install-status-succeeded", checksfn.CheckOperatorInstallStatusSucceeded, &env,
		testhelper.GetNoOperatorsSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "operator-install-status-no-privileges", checksfn.CheckOperatorNoSCCAccess, &env,
		testhelper.GetNoOperatorsSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "operator-install-source", checksfn.CheckOperatorInstalledViaOLM, &env,
		testhelper.GetNoOperatorsSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "operator-semantic-versioning", checksfn.CheckOperatorSemanticVersioning, &env,
		testhelper.GetNoOperatorsSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "operator-crd-versioning", checksfn.CheckCrdVersioning, &env,
		testhelper.GetNoOperatorCrdsSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "operator-crd-openapi-schema", checksfn.CheckCrdOpenAPISchema, &env,
		testhelper.GetNoOperatorCrdsSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "operator-single-crd-owner", checksfn.CheckSingleCrdOwner, &env,
		testhelper.GetNoOperatorsSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "operator-pods-no-hugepages", checksfn.CheckOperatorPodsNoHugepages, &env,
		testhelper.GetNoOperatorsSkipFn(&env), testhelper.GetNoOperatorPodsSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "operator-olm-skip-range", checksfn.CheckOperatorOlmSkipRange, &env,
		testhelper.GetNoOperatorsSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "operator-multiple-same-operators", checksfn.CheckMultipleSameOperators, &env,
		testhelper.GetNoOperatorsSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "operator-catalogsource-bundle-count", checksfn.CheckCatalogSourceBundleCount, &env,
		testhelper.GetNoCatalogSourcesSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "operator-single-or-multi-namespaced-allowed-in-tenant-namespaces", checksfn.CheckSingleOrMultiNamespacedOperators, &env,
		testhelper.GetNoOperatorsSkipFn(&env))
}
