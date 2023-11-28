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

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/operator/phasecheck"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

// All actual test code belongs below here.  Utilities belong above.
var _ = ginkgo.Describe(common.OperatorTestKey, func() {
	logrus.Debugf("Entering %s suite", common.OperatorTestKey)
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})
	ginkgo.ReportAfterEach(results.RecordResult)

	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestOperatorInstallStatusSucceededIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Operators, "env.Operators"))
		testOperatorInstallationPhaseSucceeded(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestOperatorNoPrivileges)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Operators, "env.Operators"))
		testOperatorInstallationWithoutPrivileges(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestOperatorIsInstalledViaOLMIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Operators, "env.Operators"))
		testOperatorOlmSubscription(&env)
	})
})

func testOperatorInstallationPhaseSucceeded(env *provider.TestEnvironment) {
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

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testOperatorInstallationWithoutPrivileges(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for i := range env.Operators {
		operator := env.Operators[i]
		csv := operator.Csv
		clusterPermissions := csv.Spec.InstallStrategy.StrategySpec.ClusterPermissions
		if len(clusterPermissions) == 0 {
			logrus.Debugf("No clusterPermissions found in %s", operator)
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(operator.Namespace, operator.Name, "Operator has no privileges on cluster resources", true))
			continue
		}

		if operator.IsClusterWide {
			logrus.Debugf("Operator %s has clusterPermissions (%d) but it is cluster-wide type.", operator, len(clusterPermissions))
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(operator.Namespace, operator.Name, "Operator has clusterPermissions config in the CSV, but it was installed as cluster-wide", true))
			continue
		}

		// Fails in case any cluster permission has a rule with any resource name.
		badRuleFound := false
		for permissionIndex := range clusterPermissions {
			permission := &clusterPermissions[permissionIndex]
			for ruleIndex := range permission.Rules {
				if n := len(permission.Rules[ruleIndex].ResourceNames); n > 0 {
					tnf.ClaimFilePrintf("%s: cluster permission (service account %s) has %d resource names (rule index %d).",
						operator, permission.ServiceAccountName, n, ruleIndex)
					// Keep reviewing other permissions' rules so we can log all the failing ones in the claim file.
					badRuleFound = true
					nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(operator.Namespace, operator.Name, "Operator has privileges on cluster resources ", false).
						SetType(testhelper.OperatorPermission).AddField(testhelper.ServiceAccountName, permission.ServiceAccountName).AddField(testhelper.ResourceName+"s", strings.Join(permission.Rules[ruleIndex].ResourceNames, "")))
				} else {
					compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(operator.Namespace, operator.Name, "Operator has no privileges on cluster resources", true).
						SetType(testhelper.OperatorPermission).AddField(testhelper.ServiceAccountName, permission.ServiceAccountName).AddField(testhelper.ResourceName+"s", "n/a"))
				}
			}
		}

		if badRuleFound {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(operator.Namespace, operator.Name, "Operator has privileges on cluster resources ", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(operator.Namespace, operator.Name, "Operator has no privileges on cluster resources", true))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testOperatorOlmSubscription(env *provider.TestEnvironment) {
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

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}
