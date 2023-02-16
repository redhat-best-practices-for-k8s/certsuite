// Copyright (C) 2020-2022 Red Hat, Inc.
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
	"github.com/onsi/ginkgo/v2"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
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
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Operators)
		testOperatorInstallationPhaseSucceeded(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestOperatorNoPrivileges)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Operators)
		testOperatorInstallationWithoutPrivileges(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestOperatorIsInstalledViaOLMIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Operators)
		testOperatorOlmSubscription(&env)
	})
})

func testOperatorInstallationPhaseSucceeded(env *provider.TestEnvironment) {
	badOperators := []string{}
	for i := range env.Operators {
		csv := env.Operators[i].Csv
		if phasecheck.IsOperatorPhaseSucceeded(csv) {
			continue
		}

		// Operator is not ready, but we need to take into account that its pods
		// could have been deleted by some of the lifecycle test cases, so they
		// could be restarting. Let's give it some time before declaring it failed.
		phase := phasecheck.WaitOperatorReady(csv)
		if phase != v1alpha1.CSVPhaseSucceeded {
			badOperators = append(badOperators, env.Operators[i].String())
			tnf.ClaimFilePrintf("%s is in phase %s. Expected phase is %s",
				&env.Operators[i], csv.Status.Phase, v1alpha1.CSVPhaseSucceeded)
		}
	}

	testhelper.AddTestResultLog("Non-compliant", badOperators, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testOperatorInstallationWithoutPrivileges(env *provider.TestEnvironment) {
	badOperators := []string{}
	for i := range env.Operators {
		csv := env.Operators[i].Csv
		clusterPermissions := csv.Spec.InstallStrategy.StrategySpec.ClusterPermissions
		if len(clusterPermissions) == 0 {
			logrus.Debugf("No clusterPermissions found in %s", env.Operators[i])
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
				}
			}
		}

		if badRuleFound {
			badOperators = append(badOperators, env.Operators[i].String())
		}
	}

	testhelper.AddTestResultLog("Non-compliant", badOperators, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testOperatorOlmSubscription(env *provider.TestEnvironment) {
	nonCompliantCsvs := []string{}

	for i := range env.Operators {
		operator := env.Operators[i]
		if operator.SubscriptionName == "" {
			tnf.ClaimFilePrintf("OLM subscription not found for operator from csv %s", provider.CsvToString(operator.Csv))
			nonCompliantCsvs = append(nonCompliantCsvs, provider.CsvToString(operator.Csv))
		}
	}

	testhelper.AddTestResultLog("Non-compliant", nonCompliantCsvs, tnf.ClaimFilePrintf, ginkgo.Fail)
}
