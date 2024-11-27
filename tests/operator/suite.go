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

	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/operator/phasecheck"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/versions"
)

var (
	env provider.TestEnvironment

	beforeEachFn = func(check *checksdb.Check) error {
		env = provider.GetTestEnvironment()
		return nil
	}
)

//nolint:funlen
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

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestOperatorNoSCCAccess)).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOperatorInstallationAccessToSCC(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestOperatorIsInstalledViaOLMIdentifier)).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOperatorOlmSubscription(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestOperatorHasSemanticVersioningIdentifier)).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOperatorSemanticVersioning(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestOperatorCrdVersioningIdentifier)).
		WithSkipCheckFn(testhelper.GetNoOperatorCrdsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOperatorCrdVersioning(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestOperatorCrdSchemaIdentifier)).
		WithSkipCheckFn(testhelper.GetNoOperatorCrdsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOperatorCrdOpenAPISpec(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestOperatorSingleCrdOwnerIdentifier)).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOperatorSingleCrdOwner(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestOperatorPodsNoHugepages)).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env), testhelper.GetNoOperatorPodsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOperatorPodsNoHugepages(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestOperatorOlmSkipRange)).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOperatorOlmSkipRange(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestMultipleSameOperatorsIdentifier)).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testMultipleSameOperators(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestOperatorInstallationInTenantNamespace)).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOperatorInstallationInTenantNamespace(c, &env)
			return nil
		}))
}

/*
Checks :

 1. Operators whose InstallTypeMode is not SingleNamespace must not be installed in the namespaces
    specified by targetNamespace in the OperatorGroup of the operators

 2. Operators that are SingleNamespace must have CRs in only tenant namespace
*/
func testOperatorInstallationInTenantNamespace(check *checksdb.Check, env *provider.TestEnvironment) {
	check.LogInfo("Starting testInstalledSingleNamespaceOperatorInTenanttNamespace")

	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	check.LogInfo("Total operators found %d ", len(env.Operators))

	for _, operator := range env.Operators {
		check.LogInfo("Checking operator %s in namespace %s ", operator.Name, operator.Namespace)

		csv := operator.Csv
		isSingleNamespaceInstallModeSupported := hasOperatorInstallModeSingleNamespace(csv.Spec.InstallModes)

		csvNamespace := csv.Namespace
		operatorNamespace := csv.Annotations["olm.operatorNamespace"]
		targetNamespaces := operator.TargetNamespaces

		check.LogInfo("operatorNamespace=%s, csvNamespace=%s, targetNamespaces=%v, singleNamespace=%v", operatorNamespace,
			csvNamespace, targetNamespaces, isSingleNamespaceInstallModeSupported)
		isCompliant := checkOperatorInstallationCompliance(
			operatorNamespace, csvNamespace, targetNamespaces,
			isSingleNamespaceInstallModeSupported,
		)
		check.LogInfo("Operator is installation Compliant %v", isCompliant)

		if isCompliant {
			check.LogInfo("Operator %s has valid installation in tenant namespace %s ", operator.Name, targetNamespaces[0])
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(operator.Namespace, operator.Name,
				"Operator has valid installation in tenant namespace ", true).AddField(testhelper.OperatorName, operator.Name))
		} else {
			check.LogInfo("Operator %s has invalid installation in tenant namespace ", operator.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(operator.Namespace, operator.Name,
				"Operator has invalid installation in tenant namespace ", false).AddField(testhelper.OperatorName, operator.Name))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// This function checks if the Operator CRD version follows K8s versioning
func testOperatorCrdVersioning(check *checksdb.Check, env *provider.TestEnvironment) {
	check.LogInfo("Starting testOperatorCrdVersioning")
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, crd := range env.Crds {
		doesUseK8sVersioning := true
		nonCompliantVersion := ""

		for _, crdVersion := range crd.Spec.Versions {
			versionName := crdVersion.Name
			check.LogDebug("Checking for Operator CRD %s with version %s", crd.Name, versionName)

			if !versions.IsValidK8sVersion(versionName) {
				doesUseK8sVersioning = false
				nonCompliantVersion = versionName
				break
			}
		}

		if doesUseK8sVersioning {
			check.LogInfo("Operator CRD %s has valid K8s versioning ", crd.Name)
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(crd.Namespace, crd.Name,
				"Operator CRD has valid K8s versioning ", true).AddField(testhelper.CrdVersion, crd.Name))
		} else {
			check.LogError("Operator CRD %s has invalid K8s versioning %s ", crd.Name, nonCompliantVersion)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(crd.Namespace, crd.Name,
				"Operator CRD has invalid K8s versioning ", false).AddField(testhelper.CrdVersion, crd.Name))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// This function checks if the operator CRD is defined with OpenAPI 3 specification
func testOperatorCrdOpenAPISpec(check *checksdb.Check, env *provider.TestEnvironment) {
	check.LogInfo("Starting testOperatorCrdOpenAPISpec")
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, crd := range env.Crds {
		isCrdDefinedWithOpenAPI3Schema := false

		for _, version := range crd.Spec.Versions {
			crdSchema := version.Schema.String()

			containsOpenAPIV3SchemaSubstr := strings.Contains(strings.ToLower(crdSchema),
				strings.ToLower(testhelper.OpenAPIV3Schema))

			if containsOpenAPIV3SchemaSubstr {
				isCrdDefinedWithOpenAPI3Schema = true
				break
			}
		}

		if isCrdDefinedWithOpenAPI3Schema {
			check.LogInfo("Operator CRD %s is defined with OpenAPIV3 schema ", crd.Name)
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(crd.Namespace, crd.Name,
				"Operator CRD is defined with OpenAPIV3 schema ", true).AddField(testhelper.OpenAPIV3Schema, crd.Name))
		} else {
			check.LogInfo("Operator CRD %s is not defined with OpenAPIV3 schema ", crd.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(crd.Namespace, crd.Name,
				"Operator CRD is not defined with OpenAPIV3 schema ", false).AddField(testhelper.OpenAPIV3Schema, crd.Name))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// This function checks for semantic versioning of the installed operators
func testOperatorSemanticVersioning(check *checksdb.Check, env *provider.TestEnvironment) {
	check.LogInfo("Starting testOperatorSemanticVersioning")
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, operator := range env.Operators {
		operatorVersion := operator.Version
		check.LogInfo("Testing Operator %q for version %s", operator, operatorVersion)

		if versions.IsValidSemanticVersion(operatorVersion) {
			check.LogInfo("Operator %q has a valid semantic version %s", operator, operatorVersion)
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(operator.Namespace, operator.Name,
				"Operator has a valid semantic version ", true).AddField(testhelper.Version, operatorVersion))
		} else {
			check.LogError("Operator %q has an invalid semantic version %s", operator, operatorVersion)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(operator.Namespace, operator.Name,
				"Operator has an invalid semantic version ", false).AddField(testhelper.Version, operatorVersion))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
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

func testOperatorInstallationAccessToSCC(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for i := range env.Operators {
		operator := env.Operators[i]
		csv := operator.Csv
		check.LogDebug("Checking operator %s", operator)
		clusterPermissions := csv.Spec.InstallStrategy.StrategySpec.ClusterPermissions
		if len(clusterPermissions) == 0 {
			check.LogInfo("No clusterPermissions found in %s's CSV", operator)
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(operator.Namespace, operator.Name,
				"No RBAC rules for Security Context Constraints found in CSV (no clusterPermissions found)", true))
			continue
		}

		// Fails in case any cluster permission has a rule that refers to securitycontextconstraints.
		badRuleFound := false
		for permissionIndex := range clusterPermissions {
			permission := &clusterPermissions[permissionIndex]
			for ruleIndex := range permission.Rules {
				rule := &permission.Rules[ruleIndex]

				// Check whether the rule is for the security api group.
				securityGroupFound := false
				for _, group := range rule.APIGroups {
					if group == "*" || group == "security.openshift.io" {
						securityGroupFound = true
						break
					}
				}

				if !securityGroupFound {
					continue
				}

				// Now check whether it grants some access to securitycontextconstraint resources.
				for _, resource := range rule.Resources {
					if resource == "*" || resource == "securitycontextconstraints" {
						check.LogInfo("Operator %s has a rule (index %d) for service account %s to access cluster SCCs",
							operator, ruleIndex, permission.ServiceAccountName)
						// Keep reviewing other permissions' rules so we can log all the failing ones in the claim file.
						badRuleFound = true
						break
					}
				}
			}
		}

		if badRuleFound {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(operator.Namespace, operator.Name, "One or more RBAC rules for Security Context Constraints found in CSV", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(operator.Namespace, operator.Name, "No RBAC rules for Security Context Constraints found in CSV", true))
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

func testOperatorSingleCrdOwner(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	// Map each CRD to a list of operators that own it
	crdOwners := map[string][]string{}
	for i := range env.Operators {
		operator := env.Operators[i]
		ownedCrds := operator.Csv.Spec.CustomResourceDefinitions.Owned
		for j := range ownedCrds {
			crdOwners[ownedCrds[j].Name] = append(crdOwners[ownedCrds[j].Name], operator.Name)
		}
	}

	// Flag those that are owned by more than one operator
	for crd, opList := range crdOwners {
		if len(opList) > 1 {
			check.LogError("CRD %q is owned by more than one operator (owners: %v)", crd, opList)
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewCrdReportObject(crd, "", "CRD is owned by more than one operator", false).
					AddField(testhelper.OperatorList, strings.Join(opList, ", ")))
		} else {
			check.LogDebug("CRD %q is owned by a single operator (%v)", crd, opList[0])
			compliantObjects = append(compliantObjects,
				testhelper.NewCrdReportObject(crd, "", "CRD is owned by a single operator", true).
					AddField(testhelper.OperatorName, opList[0]))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testOperatorPodsNoHugepages(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for csv, pods := range env.CSVToPodListMap {
		CsvResult := SplitCsv(csv)
		check.LogInfo("Name of csv: %q in namespaces: %q", CsvResult.NameCsv, CsvResult.Namespace)
		for _, pod := range pods {
			check.LogInfo("Testing Pod %q in namespace %q", pod.Name, pod.Namespace)
			if pod.HasHugepages() {
				check.LogError("Pod %q in namespace %q has hugepages", pod.Name, pod.Namespace)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, "Pod has hugepages", false))
			} else {
				check.LogInfo("Pod %q in namespace %q has no hugepages", pod.Name, pod.Namespace)
				compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, "Pod has no hugepages", true))
			}
		}
		check.SetResult(compliantObjects, nonCompliantObjects)
	}
}

func testOperatorOlmSkipRange(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for i := range env.Operators {
		operator := env.Operators[i]
		check.LogInfo("Testing Operator %q", operator)

		if operator.Csv.Annotations["olm.skipRange"] == "" {
			check.LogError("OLM skipRange not found for Operator %q", operator)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(env.Operators[i].Namespace, env.Operators[i].Name, "OLM skipRange not found for operator", false))
		} else {
			check.LogInfo("OLM skipRange %q found for Operator %q", operator.Csv.Annotations["olm.skipRange"], operator)
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(env.Operators[i].Namespace, env.Operators[i].Name, "OLM skipRange found for operator", true).
				AddField("olm.SkipRange", operator.Csv.Annotations["olm.skipRange"]))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testMultipleSameOperators(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	// Ensure the CSV name is unique and not installed more than once.
	// CSV Names are unique and OLM installs them with name.version format.
	// So, we can check if the CSV name is installed more than once.

	check.LogInfo("Checking if the operator is installed more than once")

	for _, op := range env.AllOperators {
		check.LogDebug("Checking operator %q", op.Name)
		check.LogDebug("Number of operators to check %s against: %d", op.Name, len(env.AllOperators))
		for _, op2 := range env.AllOperators {
			check.LogDebug("Comparing operator %q with operator %q", op.Name, op2.Name)

			// Retrieve the version from each CSV
			csv1Version := op.Csv.Spec.Version.String()
			csv2Version := op2.Csv.Spec.Version.String()

			log.Debug("CSV1 Version: %s", csv1Version)
			log.Debug("CSV2 Version: %s", csv2Version)

			// Strip the version from the CSV name by removing the suffix (which should be the version)
			csv1Name := strings.TrimSuffix(op.Csv.Name, ".v"+csv1Version)
			csv2Name := strings.TrimSuffix(op2.Csv.Name, ".v"+csv2Version)

			check.LogDebug("Comparing CSV names %q and %q", csv1Name, csv2Name)

			// The CSV name should be the same, but the version should be different
			// if the operator is installed more than once.
			if op.Csv != nil && op2.Csv != nil &&
				csv1Name == csv2Name &&
				csv1Version != csv2Version {
				check.LogError("Operator %q is installed more than once", op.Name)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(
					op.Namespace, op.Name, "Operator is installed more than once", false))
				break
			}
		}

		compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(
			op.Namespace, op.Name, "Operator is installed only once", true))
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}
