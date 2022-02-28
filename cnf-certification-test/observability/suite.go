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
	"github.com/test-network-function/cnf-certification-test/internal/ocpclient"
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
		provider.BuildTestEnvironment()
		env = provider.GetTestEnvironment()
	})

	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestLoggingIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testContainersLogging(&env)
	})
})

func testContainersLogging(env *provider.TestEnvironment) {
	if len(env.Containers) == 0 {
		ginkgo.Skip("No containers to run test, skipping")
	}

	ocpClient := ocpclient.NewOcpClient()

	// Iterate through all the CUTs to get their log output. The TC checks that at least
	// one log line is found.
	numLogLines := int64(1)
	badContainers := []string{}
	for _, cut := range env.Containers {
		podLogOptions := v1.PodLogOptions{TailLines: &numLogLines, Container: cut.Data.Name}
		req := ocpClient.Coreclient.Pods(cut.Namespace).GetLogs(cut.Podname, &podLogOptions)
		podLogsReaderCloser, err := req.Stream(context.TODO())
		if err != nil {
			tnf.ClaimFilePrintf("Unable to get log streamer for pod %s, container %s (ns %s), err: %s",
				cut.Podname, cut.Data.Name, cut.Namespace, err)
			badContainers = append(badContainers, fmt.Sprintf("%s.%s.%s", cut.Namespace, cut.Podname, cut.Data.Name))
			continue
		}

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogsReaderCloser)
		if err != nil {
			tnf.ClaimFilePrintf("Unable to get log data from pod %s, container %s (ns %s), err :%s",
				cut.Podname, cut.Data.Name, cut.Namespace, err)
			badContainers = append(badContainers, fmt.Sprintf("%s.%s.%s", cut.Namespace, cut.Podname, cut.Data.Name))
			continue
		}

		if buf.String() == "" {
			tnf.ClaimFilePrintf("Container: %s (Pod %s ns %s) does not have any line of log to stderr/stdout",
				cut.Data.Name, cut.Podname, cut.Namespace)
			badContainers = append(badContainers, fmt.Sprintf("%s.%s.%s", cut.Namespace, cut.Podname, cut.Data.Name))
			podLogsReaderCloser.Close()
		}
	}

	if n := len(badContainers); n > 0 {
		logrus.Debugf("Containers without logging: %+v", badContainers)
		ginkgo.Fail(fmt.Sprintf("%d containers don't have any log to stdout/stderr.", n))
	}
}
