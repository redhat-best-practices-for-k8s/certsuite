// Copyright (C) 2020-2021 Red Hat, Inc.
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
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/operator/phasecheck"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//
// All actual test code belongs below here.  Utilities belong above.
//
var _ = ginkgo.Describe(common.OperatorTestKey, func() {
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})
	ginkgo.ReportAfterEach(results.RecordResult)

	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestOperatorInstallStatusSucceededIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Operators)
		testOperatorInstallationPhaseSucceeded(&env)
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestOperatorNoPrivileges)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Operators)
		testOperatorInstallationWithoutPrivileges(&env)
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestOperatorIsInstalledViaOLMIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Operators)
		testOperatorOlmSubscription(&env)
	})
})

func testOperatorInstallationPhaseSucceeded(env *provider.TestEnvironment) {
	badOperators := []string{}
	for i := range env.Operators {
		csv := env.Operators[i].Csv
		phase := phasecheck.WaitOperatorReady(csv)
		if phase != v1alpha1.CSVPhaseSucceeded {
			badOperators = append(badOperators, env.Operators[i].String())
			tnf.ClaimFilePrintf("%s is in phase %s. Expected phase is %s",
				&env.Operators[i], csv.Status.Phase, v1alpha1.CSVPhaseSucceeded)
		}
	}

	if n := len(badOperators); n > 0 {
		ginkgo.Fail(fmt.Sprintf("Found %d operators whose CSV's phase is not %s.", n, v1alpha1.CSVPhaseSucceeded))
	}
}

func testOperatorInstallationWithoutPrivileges(env *provider.TestEnvironment) {
	badOperators := []string{}
	for i := range env.Operators {
		csv := env.Operators[i].Csv
		clusterPermissions := csv.Spec.InstallStrategy.StrategySpec.ClusterPermissions
		if len(clusterPermissions) == 0 {
			logrus.Debugf("No clusterPermissions found in %s", &env.Operators[i])
			continue
		}

		// Fails in case any cluster permission has a rule with any resource name.
		badRuleFound := false
		for permissionIndex := range clusterPermissions {
			permission := &clusterPermissions[permissionIndex]
			for ruleIndex := range permission.Rules {
				if n := len(permission.Rules[ruleIndex].ResourceNames); n > 0 {
					tnf.ClaimFilePrintf("%s: cluster permission (service account %s) has %d resource names (rule index %d).",
						&env.Operators[i], permission.ServiceAccountName, n, ruleIndex)
					// Keep reviewing other permissions' rules so we can log all the failing ones in the claim file.
					badRuleFound = true
				}
			}
		}

		if badRuleFound {
			badOperators = append(badOperators, env.Operators[i].String())
		}
	}

	if n := len(badOperators); n > 0 {
		ginkgo.Fail(fmt.Sprintf("Found %d operators with privileges on some resource names.", n))
	}
}

func testOperatorOlmSubscription(env *provider.TestEnvironment) {
	badCsvs := []string{}
	ocpClient := clientsholder.GetClientsHolder()
	for i := range env.Operators {
		csv := env.Operators[i].Csv
		ginkgo.By(fmt.Sprintf("Checking OLM subscription for %s", provider.CsvToString(csv)))
		subscriptions, err := ocpClient.OlmClient.OperatorsV1alpha1().Subscriptions(csv.Namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			ginkgo.Fail(fmt.Sprintf("Failed to get subscription for %s: %s", provider.CsvToString(csv), err))
		}

		// Iterate through namespace's subscriptions to get the installed CSV one.
		subscriptionFound := false
		for i := range subscriptions.Items {
			if subscriptions.Items[i].Status.InstalledCSV == csv.Name {
				logrus.Infof("OLM subscription %s found for %s", subscriptions.Items[i].Name, provider.CsvToString(csv))
				subscriptionFound = true
				break
			}
		}
		if !subscriptionFound {
			tnf.ClaimFilePrintf("OLM subscription not found for operator %s", provider.CsvToString(csv))
			badCsvs = append(badCsvs, provider.CsvToString(csv))
		}
	}

	if n := len(badCsvs); n > 0 {
		ginkgo.Fail(fmt.Sprintf("Found %d CSVs not installed by OLM", n))
	}
}
