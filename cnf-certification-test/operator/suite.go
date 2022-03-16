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
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
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

	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestOperatorInstallStatusSucceededIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testOperatorInstallationPhaseSucceeded(&env)
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestOperatorNoPrivileges)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testOperatorInstallationWithoutPrivileges(&env)
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestOperatorIsInstalledViaOLMIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testOperatorOlmSubscription(&env)
	})
})

func testOperatorInstallationPhaseSucceeded(env *provider.TestEnvironment) {
	badCsvs := []string{}
	if len(env.Csvs) == 0 {
		ginkgo.Skip("No CSVs to perform test, skipping.")
	}

	for _, csv := range env.Csvs {
		if csv.Status.Phase != v1alpha1.CSVPhaseSucceeded {
			badCsvs = append(badCsvs, fmt.Sprintf("%s.%s", csv.Namespace, csv.Name))
			tnf.ClaimFilePrintf("CSV %s (ns %s) is in phase %s. Expected phase is %s",
				csv.Name, csv.Namespace, csv.Status.Phase, v1alpha1.CSVPhaseSucceeded)
		}
	}

	if n := len(badCsvs); n > 0 {
		ginkgo.Fail(fmt.Sprintf("Found %d CSVs whose phase is not %s.", n, v1alpha1.CSVPhaseSucceeded))
	}
}

func testOperatorInstallationWithoutPrivileges(env *provider.TestEnvironment) {
	badCsvs := []string{}
	if len(env.Csvs) == 0 {
		ginkgo.Skip("No CSVs to perform test, skipping.")
	}

	for _, csv := range env.Csvs {
		clusterPermissions := csv.Spec.InstallStrategy.StrategySpec.ClusterPermissions
		if len(clusterPermissions) == 0 {
			logrus.Debugf("No clusterPermissions found in csv %s (ns %s)", csv.Name, csv.Namespace)
			continue
		}

		// Fails in case any cluster permission has a rule with any resource name.
		badRuleFound := false
		for i := range clusterPermissions {
			permission := &clusterPermissions[i]
			for ruleIndex := range permission.Rules {
				if n := len(permission.Rules[ruleIndex].ResourceNames); n > 0 {
					tnf.ClaimFilePrintf("CSV %s (ns %s) cluster permission (service account %s) has %d resource names (rule index %d).",
						csv.Name, csv.Namespace, permission.ServiceAccountName, n, ruleIndex)
					// Keep reviewing other permissions' rules so we can log all the failing ones in the claim file.
					badRuleFound = true
				}
			}
		}

		if badRuleFound {
			badCsvs = append(badCsvs, fmt.Sprintf("%s.%s", csv.Namespace, csv.Name))
		}
	}

	if n := len(badCsvs); n > 0 {
		ginkgo.Fail(fmt.Sprintf("Found %d CSVs with priviledges on some resource names.", n))
	}
}

func testOperatorOlmSubscription(env *provider.TestEnvironment) {
	badCsvs := []string{}
	if len(env.Csvs) == 0 {
		ginkgo.Skip("No CSVs to perform test, skipping.")
	}

	ocpClient := clientsholder.GetClientsHolder()
	for _, csv := range env.Csvs {
		ginkgo.By(fmt.Sprintf("Checking OLM subscription for CSV %s (ns %s)", csv.Name, csv.Namespace))
		options := metav1.ListOptions{}
		subscriptions, err := ocpClient.OlmClient.OperatorsV1alpha1().Subscriptions(csv.Namespace).List(context.TODO(), options)
		if err != nil {
			ginkgo.Fail(fmt.Sprintf("Failed to get subscription for CSV %s (ns %s): %s", csv.Name, csv.Namespace, err))
		}

		// Iterate through namespace's subscriptions to get the installed CSV one.
		subscriptionFound := false
		for i := range subscriptions.Items {
			if subscriptions.Items[i].Status.InstalledCSV == csv.Name {
				logrus.Infof("OLM subscription %s found for CSV %s (ns %s)", subscriptions.Items[i].Name, csv.Name, csv.Namespace)
				subscriptionFound = true
				break
			}
		}
		if !subscriptionFound {
			tnf.ClaimFilePrintf("OLM subscription not found for operator csv %s (ns %s)", csv.Name, csv.Namespace)
			badCsvs = append(badCsvs, fmt.Sprintf("%s.%s", csv.Namespace, csv.Name))
		}
	}

	if n := len(badCsvs); n > 0 {
		ginkgo.Fail(fmt.Sprintf("Found %d CSVs not installed by OLM", n))
	}
}
