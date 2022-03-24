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
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	v1 "k8s.io/api/core/v1"
)

//
// All actual test code belongs below here.  Utilities belong above.
//
var _ = ginkgo.Describe(common.ObservabilityTestKey, func() {
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})

	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestLoggingIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testContainersLogging(&env)
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestCrdsStatusSubresourceIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testCrds(&env)
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestTerminationMessagePolicyIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testTerminationMessagePolicy(&env)
	})
})

// containerHasLoggingOutput helper function to get the last line of logging output from
// a container. Returns true in case some output was found, false otherwise.
func containerHasLoggingOutput(cut *provider.Container) (bool, error) {
	ocpClient := clientsholder.GetClientsHolder()

	numLogLines := int64(1)
	podLogOptions := v1.PodLogOptions{TailLines: &numLogLines, Container: cut.Data.Name}
	req := ocpClient.Coreclient.Pods(cut.Namespace).GetLogs(cut.Podname, &podLogOptions)

	podLogsReaderCloser, err := req.Stream(context.TODO())
	if err != nil {
		return false, fmt.Errorf("unable to get log streamer, err: %s", err)
	}

	defer podLogsReaderCloser.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogsReaderCloser)
	if err != nil {
		return false, fmt.Errorf("unable to get log data, err: %s", err)
	}

	return buf.String() != "", nil
}

func testContainersLogging(env *provider.TestEnvironment) {
	if len(env.Containers) == 0 {
		ginkgo.Skip("No containers to run test, skipping")
	}

	// Iterate through all the CUTs to get their log output. The TC checks that at least
	// one log line is found.
	badContainers := []string{}
	for _, cut := range env.Containers {
		ginkgo.By(fmt.Sprintf("Checking %s has some logging output", cut.StringShort()))
		hasLoggingOutput, err := containerHasLoggingOutput(cut)
		if err != nil {
			tnf.ClaimFilePrintf("Failed to get %s log output: %s", cut.StringShort(), err)
			badContainers = append(badContainers, cut.StringShort())
		}

		if !hasLoggingOutput {
			tnf.ClaimFilePrintf("%s does not have any line of log to stderr/stdout", cut.StringShort())
			badContainers = append(badContainers, cut.StringShort())
		}
	}

	if n := len(badContainers); n > 0 {
		logrus.Debugf("Containers without logging: %+v", badContainers)
		ginkgo.Fail(fmt.Sprintf("%d containers don't have any log to stdout/stderr.", n))
	}
}

// testCrds testing if crds have a status sub resource set
func testCrds(env *provider.TestEnvironment) {
	failedCrds := []string{}
	for _, crd := range env.Crds {
		ginkgo.By("Testing CRD " + crd.Name)

		for _, ver := range crd.Spec.Versions {
			if _, ok := ver.Schema.OpenAPIV3Schema.Properties["status"]; !ok {
				tnf.ClaimFilePrintf("FAILURE: CRD %s, version: %s does not have a status subresource.", crd.Name, ver.Name)
				failedCrds = append(failedCrds, crd.Name+"."+ver.Name)
			}
		}
	}

	if n := len(failedCrds); n > 0 {
		logrus.Debugf("CRD.version without status subresource: %+v", failedCrds)
		ginkgo.Fail(fmt.Sprintf("%d CRDs don't have status subresource", n))
	}
}

// testTerminationMessagePolicy tests to make sure that pods
func testTerminationMessagePolicy(env *provider.TestEnvironment) {
	failedContainers := []string{}
	for _, cut := range env.Containers {
		ginkgo.By("Testing for terminationMessagePolicy: " + cut.StringShort())
		if cut.Data.TerminationMessagePolicy != v1.TerminationMessageFallbackToLogsOnError {
			tnf.ClaimFilePrintf("FAILURE: %s does not have a TerminationMessagePolicy: FallbackToLogsOnError", cut.StringShort())
			failedContainers = append(failedContainers, cut.Data.Name)
		}
	}
	if n := len(failedContainers); n > 0 {
		ginkgo.Fail("Containers were found to not have a termination message policy set to FallbackToLogsOnError")
	}
}
