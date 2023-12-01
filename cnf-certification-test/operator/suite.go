// Copyright (C) 2020-2023 Red Hat, Inc.
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
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/operator/phasecheck"
	"github.com/test-network-function/cnf-certification-test/pkg/checksdb"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

var (
	env provider.TestEnvironment

	beforeEachFn = func(check *checksdb.Check) error {
		logrus.Infof("Check %s: getting test environment.", check.ID)
		env = provider.GetTestEnvironment()
		return nil
	}
)

func LoadChecks() {
	logrus.Debugf("Entering %s suite", common.OperatorTestKey)

	checksGroup := checksdb.NewChecksGroup(common.OperatorTestKey).
		WithBeforeEachFn(beforeEachFn)

	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestOperatorInstallStatusSucceededIdentifier)
	check := checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOperatorInstallationPhaseSucceeded(c, &env)
			return nil
		})

	checksGroup.Add(check)

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestOperatorNoPrivileges)
	check = checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOperatorInstallationWithoutPrivileges(c, &env)
			return nil
		})

	checksGroup.Add(check)

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestOperatorIsInstalledViaOLMIdentifier)
	check = checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOperatorOlmSubscription(c, &env)
			return nil
		})

	checksGroup.Add(check)
}

func testOperatorInstallationPhaseSucceeded(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for i := range env.Operators {
		csv := env.Operators[i].Csv
		if phasecheck.WaitOperatorReady(csv) {
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(env.Operators[i].Namespace, env.Operators[i].Name,
				"Operator on Succeeded state ", true).AddField(testhelper.OperatorPhase, string(csv.Status.Phase)))
		} else {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(env.Operators[i].Namespace, env.Operators[i].Name,
				"Operator not in Succeeded state ", false).AddField(testhelper.OperatorPhase, string(csv.Status.Phase)))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testOperatorInstallationWithoutPrivileges(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for i := range env.Operators {
		csv := env.Operators[i].Csv
		clusterPermissions := csv.Spec.InstallStrategy.StrategySpec.ClusterPermissions
		if len(clusterPermissions) == 0 {
			logrus.Debugf("No clusterPermissions found in %s", env.Operators[i])
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(env.Operators[i].Namespace, env.Operators[i].Name, "Operator has no privileges on cluster resources", true))
			continue
		}

		// Fails in case any cluster permission has a rule with any resource name.
		badRuleFound := false
		for permissionIndex := range clusterPermissions {
			permission := &clusterPermissions[permissionIndex]
			for ruleIndex := range permission.Rules {
				if n := len(permission.Rules[ruleIndex].ResourceNames); n > 0 {
					tnf.ClaimFilePrintf("%s: cluster permission (service account %s) has %d resource names (rule index %d).",
						env.Operators[i], permission.ServiceAccountName, n, ruleIndex)
					// Keep reviewing other permissions' rules so we can log all the failing ones in the claim file.
					badRuleFound = true
					nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(env.Operators[i].Namespace, env.Operators[i].Name, "Operator has privileges on cluster resources ", false).
						SetType(testhelper.OperatorPermission).AddField(testhelper.ServiceAccountName, permission.ServiceAccountName).AddField(testhelper.ResourceName+"s", strings.Join(permission.Rules[ruleIndex].ResourceNames, "")))
				} else {
					compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(env.Operators[i].Namespace, env.Operators[i].Name, "Operator has no privileges on cluster resources", true).
						SetType(testhelper.OperatorPermission).AddField(testhelper.ServiceAccountName, permission.ServiceAccountName).AddField(testhelper.ResourceName+"s", "n/a"))
				}
			}
		}

		if badRuleFound {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(env.Operators[i].Namespace, env.Operators[i].Name, "Operator has privileges on cluster resources ", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(env.Operators[i].Namespace, env.Operators[i].Name, "Operator has no privileges on cluster resources", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testOperatorOlmSubscription(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for i := range env.Operators {
		operator := env.Operators[i]
		if operator.SubscriptionName == "" {
			tnf.ClaimFilePrintf("OLM subscription not found for operator from csv %s", provider.CsvToString(operator.Csv))
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(env.Operators[i].Namespace, env.Operators[i].Name, "OLM subscription not found for operator, so it is not installed via OLM", false).
				AddField(testhelper.SubscriptionName, operator.SubscriptionName))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(env.Operators[i].Namespace, env.Operators[i].Name, "install-status-no-privilege (subscription found)", true).
				AddField(testhelper.SubscriptionName, operator.SubscriptionName))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}
