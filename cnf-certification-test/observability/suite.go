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

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	pdbv1 "github.com/test-network-function/cnf-certification-test/cnf-certification-test/observability/pdb"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	corev1 "k8s.io/api/core/v1"
)

// All actual test code belongs below here.  Utilities belong above.
var _ = ginkgo.Describe(common.ObservabilityTestKey, func() {
	logrus.Debugf("Entering %s suite", common.ObservabilityTestKey)
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})
	ginkgo.ReportAfterEach(results.RecordResult)

	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestLoggingIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testContainersLogging(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestCrdsStatusSubresourceIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Crds, "env.Crds"))
		testCrds(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestTerminationMessagePolicyIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testTerminationMessagePolicy(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodDisruptionBudgetIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, testhelper.NewSkipObject(env.Deployments, "env.Deployments"), testhelper.NewSkipObject(env.StatefulSets, "env.StatefulSets"))
		testPodDisruptionBudgets(&env)
	})
})

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

func testContainersLogging(env *provider.TestEnvironment) {
	// Iterate through all the CUTs to get their log output. The TC checks that at least
	// one log line is found.
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		ginkgo.By(fmt.Sprintf("Checking %s has some logging output", cut))
		hasLoggingOutput, err := containerHasLoggingOutput(cut)
		if err != nil {
			tnf.ClaimFilePrintf("Failed to get %s log output: %s", cut, err)
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Could not get log output", false))
			continue
		}

		if !hasLoggingOutput {
			tnf.ClaimFilePrintf("%s does not have any line of log to stderr/stdout", cut)
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "No log line to stderr/stdout found", false))
		} else {
			compliantObjects = append(compliantObjects,
				testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Found log line to stderr/stdout", true))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testCrds testing if crds have a status sub resource set
func testCrds(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, crd := range env.Crds {
		ginkgo.By("Testing CRD " + crd.Name)

		for _, ver := range crd.Spec.Versions {
			if _, ok := ver.Schema.OpenAPIV3Schema.Properties["status"]; !ok {
				tnf.ClaimFilePrintf("FAILURE: CRD %s, version: %s does not have a status subresource.", crd.Name, ver.Name)
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

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testTerminationMessagePolicy tests to make sure that pods
func testTerminationMessagePolicy(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		ginkgo.By("Testing for terminationMessagePolicy: " + cut.String())
		if cut.TerminationMessagePolicy != corev1.TerminationMessageFallbackToLogsOnError {
			tnf.ClaimFilePrintf("FAILURE: %s does not have a TerminationMessagePolicy: FallbackToLogsOnError", cut)
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "TerminationMessagePolicy is not FallbackToLogsOnError", false))
		} else {
			compliantObjects = append(compliantObjects,
				testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "TerminationMessagePolicy is FallbackToLogsOnError", true))
		}
	}
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

//nolint:funlen
func testPodDisruptionBudgets(env *provider.TestEnvironment) {
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

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}
