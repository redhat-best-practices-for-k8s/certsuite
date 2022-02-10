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

package platform

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/ocpclient"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform/cnffsdiff"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

//
// All actual test code belongs below here.  Utilities belong above.
//
var _ = ginkgo.Describe(common.PlatformAlterationTestKey, func() {
	logrus.Debug(common.PlatformAlterationTestKey, " not moved yet to new framework")
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		provider.BuildTestEnvironment()
		env = provider.GetTestEnvironment()
	})
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestUnalteredBaseImageIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		if provider.IsOCPCluster() {
			testContainersFsDiff(&env)
		} else {
			ginkgo.Skip(" non ocp cluster ")
		}
	})
})

// testContainersFsDiff test that all CUT didn't install new packages are starting
func testContainersFsDiff(env *provider.TestEnvironment) {
	var badContainers []string
	var errContainers []string
	for _, cut := range env.Containers {
		logrus.Debug(fmt.Sprintf("%s(%s) should not install new packages after starting", cut.Podname, cut.Data.Name))
		fsdiff, err := cnffsdiff.NewFsDiff(cut)
		if err != nil {
			logrus.Error("can't create FsDiff instance")
			errContainers = append(errContainers, cut.Data.Name)
			continue
		}
		nodeName := cut.NodeName
		debugPod := env.DebugPods[nodeName]
		fsdiff.RunTest(ocpclient.NewOcpClient(), ocpclient.Context{Namespace: debugPod.Namespace,
			Podname: debugPod.Name, Containername: debugPod.Spec.Containers[0].Name})
		switch fsdiff.GetResults() {
		case tnf.SUCCESS:
			continue
		case tnf.FAILURE:
			badContainers = append(badContainers, cut.Data.Name)
		case tnf.ERROR:
			errContainers = append(errContainers, cut.Data.Name)
		}
	}
	logrus.Println("bad containers ", badContainers)
	logrus.Println("err containers ", errContainers)
	gomega.Expect(badContainers).To(gomega.BeNil())
	gomega.Expect(errContainers).To(gomega.BeNil())
}
