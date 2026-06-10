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

var (
	env provider.TestEnvironment

	beforeEachFn = func(check *checksdb.Check) error {
		env = provider.GetTestEnvironment()
		return nil
	}
)

func LoadChecks() {
	log.Debug("Loading %s suite checks", common.OperatorTestKey)

	checksGroup := checksdb.NewChecksGroup(common.OperatorTestKey).
		WithBeforeEachFn(beforeEachFn)

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("operator-install-status-succeeded")).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckOperatorInstallStatusSucceeded).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("operator-install-status-no-privileges")).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckOperatorNoSCCAccess).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("operator-install-source")).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckOperatorInstalledViaOLM).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("operator-semantic-versioning")).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckOperatorSemanticVersioning).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("operator-crd-versioning")).
		WithSkipCheckFn(testhelper.GetNoOperatorCrdsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckCrdVersioning).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("operator-crd-openapi-schema")).
		WithSkipCheckFn(testhelper.GetNoOperatorCrdsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckCrdOpenAPISchema).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("operator-single-crd-owner")).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckSingleCrdOwner).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("operator-pods-no-hugepages")).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env), testhelper.GetNoOperatorPodsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckOperatorPodsNoHugepages).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("operator-olm-skip-range")).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckOperatorOlmSkipRange).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("operator-multiple-same-operators")).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckMultipleSameOperators).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("operator-catalogsource-bundle-count")).
		WithSkipCheckFn(testhelper.GetNoCatalogSourcesSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckCatalogSourceBundleCount).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("operator-single-or-multi-namespaced-allowed-in-tenant-namespaces")).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckSingleOrMultiNamespacedOperators).MakeCheckFn(&env)))
}
