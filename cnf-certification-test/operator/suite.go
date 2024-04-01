// Copyright (C) 2020-2024 Red Hat, Inc.
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

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common/rbac"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/operator/phasecheck"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/checksdb"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
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

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestOperatorInstallStatusSucceededIdentifier)).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOperatorInstallationPhaseSucceeded(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestOperatorNoPrivileges)).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOperatorInstallationWithoutPrivileges(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestOperatorIsInstalledViaOLMIdentifier)).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOperatorOlmSubscription(c, &env)
			return nil
		}))
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestOperatorSecurityRequiremnents)).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOperatorSecurityRequiremnents(c, &env)
			return nil
		}))	
}

func testOperatorInstallationPhaseSucceeded(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, op := range env.Operators {
		check.LogInfo("Testing Operator %q", op)
		if phasecheck.WaitOperatorReady(op.Csv) {
			check.LogInfo("Operator %q is in Succeeded phase", op)
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(op.Namespace, op.Name,
				"Operator on Succeeded state ", true).AddField(testhelper.OperatorPhase, string(op.Csv.Status.Phase)))
		} else {
			check.LogError("Operator %q is not in Succeeded phase (phase=%q)", op, op.Csv.Status.Phase)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(op.Namespace, op.Name,
				"Operator not in Succeeded state ", false).AddField(testhelper.OperatorPhase, string(op.Csv.Status.Phase)))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testOperatorInstallationWithoutPrivileges(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, op := range env.Operators {
		check.LogInfo("Testing Operator %q", op)
		clusterPermissions := op.Csv.Spec.InstallStrategy.StrategySpec.ClusterPermissions
		if len(clusterPermissions) == 0 {
			check.LogInfo("Operator %q has no privileged on cluster resources. No clusterPermissions found.", op)
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(op.Namespace, op.Name, "Operator has no privileges on cluster resources", true))
			continue
		}

		// Fails in case any cluster permission has a rule with any resource name.
		badRuleFound := false
		for permissionIndex := range clusterPermissions {
			permission := &clusterPermissions[permissionIndex]
			for ruleIndex := range permission.Rules {
				if n := len(permission.Rules[ruleIndex].ResourceNames); n > 0 {
					resources := strings.Join(permission.Rules[ruleIndex].ResourceNames, " ")
					check.LogError("Operator %q has privileges on cluster resources (service account=%q, resources=%q)", op, permission.ServiceAccountName, resources)
					// Keep reviewing other permissions' rules so we can log all the failing ones in the claim file.
					badRuleFound = true
					nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(op.Namespace, op.Name, "Operator has privileges on cluster resources ", false).
						SetType(testhelper.OperatorPermission).AddField(testhelper.ServiceAccountName, permission.ServiceAccountName).AddField(testhelper.ResourceName+"s", resources))
				} else {
					check.LogInfo("Operator %q has no privileges on cluster resources", op)
					compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(op.Namespace, op.Name, "Operator has no privileges on cluster resources", true).
						SetType(testhelper.OperatorPermission).AddField(testhelper.ServiceAccountName, permission.ServiceAccountName).AddField(testhelper.ResourceName+"s", "n/a"))
				}
			}
		}

		if badRuleFound {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(op.Namespace, op.Name, "Operator has privileges on cluster resources ", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(op.Namespace, op.Name, "Operator has no privileges on cluster resources", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testOperatorOlmSubscription(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for i := range env.Operators {
		operator := env.Operators[i]
		check.LogInfo("Testing Operator %q", operator)
		if operator.SubscriptionName == "" {
			check.LogError("OLM subscription not found for Operator %q", operator)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(env.Operators[i].Namespace, env.Operators[i].Name, "OLM subscription not found for operator, so it is not installed via OLM", false).
				AddField(testhelper.SubscriptionName, operator.SubscriptionName))
		} else {
			check.LogInfo("OLM subscription %q found for Operator %q", operator.SubscriptionName, operator)
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(env.Operators[i].Namespace, env.Operators[i].Name, "install-status-no-privilege (subscription found)", true).
				AddField(testhelper.SubscriptionName, operator.SubscriptionName))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testOperatorSecurityRequiremnents(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for i := range env.Operators {
		operator := env.Operators[i]
		check.LogInfo("Testing Operator %q", operator)
		for _, put := range env.Pods {
			check.LogInfo("Testing Pod %q", put)
			if put.IsRunAsUserID(0) {
				check.LogError("Pod %q UserID is 0", put.Name)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has been found with UserID is 0", false))
			}

			if !put.IsRunAsNonRoot() {
				check.LogError("Pod %q is run as root", put.Name)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has been found is run as root", false))
			}
			
			nonCompliantObjects = put.IsReadOnlyRootFilesystem(check, nonCompliantObjects) 

			// Evaluate the pod's automount service tokens and any attached service accounts
			client := clientsholder.GetClientsHolder()
			podPassed, newMsg := rbac.EvaluateAutomountTokens(client.K8sClient.CoreV1(), put.Pod)
			if !podPassed {
				check.LogError(newMsg)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, newMsg, false))
			} else {
				check.LogInfo("Pod %q does not have automount service tokens set to true", put)
				compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod does not have automount service tokens set to true", true))
			}
		}
	
		check.SetResult(compliantObjects, nonCompliantObjects)
	}
}
