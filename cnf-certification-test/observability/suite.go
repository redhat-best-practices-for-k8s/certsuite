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

package observability

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	pdbv1 "github.com/test-network-function/cnf-certification-test/cnf-certification-test/observability/pdb"

	// "github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/checksdb"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	corev1 "k8s.io/api/core/v1"
)

var (
	env provider.TestEnvironment

	beforeAll = func([]*checksdb.Check) error {
		logrus.Infof("OBSERVABILITY GROUP BEFORE-ALL")
		return nil
	}

	afterAllFn = func(checks []*checksdb.Check) error {
		logrus.Infof("OBSERVABILITY GROUP AFTER-ALL")
		return nil // errors.New("crappy afterAll func!")
	}

	beforeEachFn = func(check *checksdb.Check) error {
		logrus.Infof("Getting test environment for check %s", check.ID)
		env = provider.GetTestEnvironment()
		// panic("beforeEach panickism")
		return nil
	}

	afterEachFn = func(check *checksdb.Check) error {
		logrus.Infof("Fake afterEachFn for check %s", check.ID)
		return nil
	}

	skipIfNoContainersFn = func() (bool, string) {
		if len(env.Containers) == 0 {
			logrus.Warnf("No containers to check...")
			return true, "There are no containers to check. Please check under test labels."
		}

		return false, ""
	}

	skipIfNoCrdsFn = func() (bool, string) {
		if len(env.Crds) == 0 {
			logrus.Warn("No CRDs to check.")
			return true, "There are no CRDs to check."
		}

		return false, ""
	}

	skipIfNoDeploymentsNorStatefulSets = func() (bool, string) {
		if len(env.Deployments) == 0 && len(env.StatefulSets) == 0 {
			logrus.Warn("No deployments nor statefulsets to check.")
			return true, "There are no deployments nor statefulsets to check."
		}
		return false, ""
	}
)

func init() {
	logrus.Debugf("Entering %s suite", common.ObservabilityTestKey)

	checksGroup := checksdb.NewChecksGroup(common.ObservabilityTestKey).
		WithBeforeAllFn(beforeAll).
		WithAfterAllFn(afterAllFn).
		WithBeforeEachFn(beforeEachFn).
		WithAfterEachFn(afterEachFn)

	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestLoggingIdentifier)
	check := checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(skipIfNoContainersFn).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainersLogging(c, &env)
			return nil
		})

	checksGroup.Add(check)

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestCrdsStatusSubresourceIdentifier)
	check = checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(skipIfNoCrdsFn).
		WithCheckFn(func(c *checksdb.Check) error {
			testCrds(c, &env)
			return nil
		})

	checksGroup.Add(check)

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestTerminationMessagePolicyIdentifier)
	check = checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(skipIfNoContainersFn).
		WithCheckFn(func(c *checksdb.Check) error {
			testTerminationMessagePolicy(c, &env)
			return nil
		})

	checksGroup.Add(check)

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodDisruptionBudgetIdentifier)
	check = checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(skipIfNoDeploymentsNorStatefulSets).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodDisruptionBudgets(c, &env)
			return nil
		})

	checksGroup.Add(check)
}

// containerHasLoggingOutput helper function to get the last line of logging output from
// a container. Returns true in case some output was found, false otherwise.
func containerHasLoggingOutput(cut *provider.Container) (bool, error) {
	ocpClient := clientsholder.GetClientsHolder()

	// K8s' API won't return lines that do not have the newline termination char, so
	// We need to ask for the last two lines.
	const tailLogLines = 2
	numLogLines := int64(tailLogLines)
	podLogOptions := corev1.PodLogOptions{TailLines: &numLogLines, Container: cut.Name}
	req := ocpClient.K8sClient.CoreV1().Pods(cut.Namespace).GetLogs(cut.Podname, &podLogOptions)

	podLogsReaderCloser, err := req.Stream(context.TODO())
	if err != nil {
		return false, fmt.Errorf("unable to get log streamer, err: %v", err)
	}

	defer podLogsReaderCloser.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogsReaderCloser)
	if err != nil {
		return false, fmt.Errorf("unable to get log data, err: %v", err)
	}

	return buf.String() != "", nil
}

func testContainersLogging(check *checksdb.Check, env *provider.TestEnvironment) {
	// Iterate through all the CUTs to get their log output. The TC checks that at least
	// one log line is found.
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		logrus.Info(fmt.Sprintf("Checking %s has some logging output", cut))
		hasLoggingOutput, err := containerHasLoggingOutput(cut)
		if err != nil {
			tnf.Logf(logrus.ErrorLevel, "Failed to get %s log output: %s", cut, err)
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Could not get log output", false))
			continue
		}

		if !hasLoggingOutput {
			tnf.Logf(logrus.ErrorLevel, "%s does not have any line of log to stderr/stdout", cut)
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "No log line to stderr/stdout found", false))
		} else {
			compliantObjects = append(compliantObjects,
				testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Found log line to stderr/stdout", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testCrds testing if crds have a status sub resource set
func testCrds(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, crd := range env.Crds {
		logrus.Info("Testing CRD " + crd.Name)
		// v := []string{}
		// fmt.Printf("%s", v[1])
		for _, ver := range crd.Spec.Versions {
			if _, ok := ver.Schema.OpenAPIV3Schema.Properties["status"]; !ok {
				tnf.Logf(logrus.ErrorLevel, "FAILURE: CRD %s, version: %s does not have a status subresource.", crd.Name, ver.Name)
				nonCompliantObjects = append(nonCompliantObjects,
					testhelper.NewReportObject("Crd does not have a status sub resource set", testhelper.CustomResourceDefinitionType, false).
						AddField(testhelper.CustomResourceDefinitionName, crd.Name).
						AddField(testhelper.CustomResourceDefinitionVersion, ver.Name))
			} else {
				compliantObjects = append(compliantObjects,
					testhelper.NewReportObject("Crd has a status sub resource set", testhelper.CustomResourceDefinitionType, true).
						AddField(testhelper.CustomResourceDefinitionName, crd.Name).
						AddField(testhelper.CustomResourceDefinitionVersion, ver.Name))
			}
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testTerminationMessagePolicy tests to make sure that pods
func testTerminationMessagePolicy(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, cut := range env.Containers {
		logrus.Info("Testing for terminationMessagePolicy: " + cut.String())
		if cut.TerminationMessagePolicy != corev1.TerminationMessageFallbackToLogsOnError {
			tnf.ClaimFilePrintf("FAILURE: %s does not have a TerminationMessagePolicy: FallbackToLogsOnError", cut)
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "TerminationMessagePolicy is not FallbackToLogsOnError", false))
		} else {
			compliantObjects = append(compliantObjects,
				testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "TerminationMessagePolicy is FallbackToLogsOnError", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

//nolint:funlen
func testPodDisruptionBudgets(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	// Loop through all of the of Deployments and StatefulSets and check if the PDBs are valid
	for _, d := range env.Deployments {
		pdbFound := false
		for pdbIndex := range env.PodDisruptionBudgets {
			for k, v := range d.Spec.Template.Labels {
				if env.PodDisruptionBudgets[pdbIndex].Spec.Selector.MatchLabels[k] == v {
					pdbFound = true
					if ok, err := pdbv1.CheckPDBIsValid(&env.PodDisruptionBudgets[pdbIndex], d.Spec.Replicas); !ok {
						nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject(fmt.Sprintf("Invalid PodDisruptionBudget config: %v", err), testhelper.DeploymentType, false).
							AddField(testhelper.DeploymentName, d.Name).
							AddField(testhelper.Namespace, d.Namespace).
							AddField(testhelper.PodDisruptionBudgetReference, env.PodDisruptionBudgets[pdbIndex].Name))
						tnf.ClaimFilePrintf("PDB %s is not valid for Deployment %s, err: %v", env.PodDisruptionBudgets[pdbIndex].Name, d.Name, err)
					} else {
						logrus.Infof("PDB %s is valid for Deployment: %s", env.PodDisruptionBudgets[pdbIndex].Name, d.Name)
						compliantObjects = append(compliantObjects, testhelper.NewReportObject("Deployment: references PodDisruptionBudget", testhelper.DeploymentType, true).
							AddField(testhelper.DeploymentName, d.Name).
							AddField(testhelper.Namespace, d.Namespace).
							AddField(testhelper.PodDisruptionBudgetReference, env.PodDisruptionBudgets[pdbIndex].Name))
					}
				}
			}
		}
		if !pdbFound {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject("Deployment is missing a corresponding PodDisruptionBudget", testhelper.DeploymentType, false).
				AddField(testhelper.DeploymentName, d.Name).
				AddField(testhelper.Namespace, d.Namespace))
		}
	}

	for _, s := range env.StatefulSets {
		pdbFound := false
		for pdbIndex := range env.PodDisruptionBudgets {
			for k, v := range s.Spec.Template.Labels {
				if env.PodDisruptionBudgets[pdbIndex].Spec.Selector.MatchLabels[k] == v {
					pdbFound = true
					if ok, err := pdbv1.CheckPDBIsValid(&env.PodDisruptionBudgets[pdbIndex], s.Spec.Replicas); !ok {
						nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject(fmt.Sprintf("Invalid PodDisruptionBudget config: %v", err), testhelper.StatefulSetType, false).
							AddField(testhelper.StatefulSetName, s.Name).
							AddField(testhelper.Namespace, s.Namespace).
							AddField(testhelper.PodDisruptionBudgetReference, env.PodDisruptionBudgets[pdbIndex].Name))
						tnf.ClaimFilePrintf("PDB %s is not valid for StatefulSet %s, err: %v", env.PodDisruptionBudgets[pdbIndex].Name, s.Name, err)
					} else {
						logrus.Infof("PDB %s is valid for StatefulSet: %s", env.PodDisruptionBudgets[pdbIndex].Name, s.Name)
						compliantObjects = append(compliantObjects, testhelper.NewReportObject("StatefulSet: references PodDisruptionBudget", testhelper.StatefulSetType, true).
							AddField(testhelper.StatefulSetName, s.Name).
							AddField(testhelper.Namespace, s.Namespace).
							AddField(testhelper.PodDisruptionBudgetReference, env.PodDisruptionBudgets[pdbIndex].Name))
					}
				}
			}
		}
		if !pdbFound {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject("StatefulSet is missing a corresponding PodDisruptionBudget", testhelper.StatefulSetType, false).
				AddField(testhelper.StatefulSetName, s.Name).
				AddField(testhelper.Namespace, s.Namespace))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}
