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

package platform

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	clientsholder "github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform/cnffsdiff"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform/isredhat"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform/nodetainted"
)

//
// All actual test code belongs below here.  Utilities belong above.
//
var _ = ginkgo.Describe(common.PlatformAlterationTestKey, func() {
	logrus.Debugf("Entering %s suite", common.PlatformAlterationTestKey)
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
		provider.WaitDebugPodReady()
	})

	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestUnalteredBaseImageIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		if provider.IsOCPCluster() {
			testContainersFsDiff(&env)
		} else {
			ginkgo.Skip(" non ocp cluster ")
		}
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestNonTaintedNodeKernelsIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testTainted(&env) // minikube tainted kernels are allowed via config
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestIsRedHatReleaseIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testIsRedHatRelease(&env)
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

		fsdiff.RunTest(clientsholder.NewClientsHolder(), clientsholder.Context{Namespace: debugPod.Namespace,
			Podname: debugPod.Name, Containername: debugPod.Spec.Containers[0].Name})
		switch fsdiff.GetResults() {
		case testhelper.SUCCESS:
			continue
		case testhelper.FAILURE:
			badContainers = append(badContainers, cut.Data.Name)
		case testhelper.ERROR:
			errContainers = append(errContainers, cut.Data.Name)
		}
	}
	logrus.Println("bad containers ", badContainers)
	logrus.Println("err containers ", errContainers)
	gomega.Expect(badContainers).To(gomega.BeNil())
	gomega.Expect(errContainers).To(gomega.BeNil())
}

//nolint:funlen
func testTainted(env *provider.TestEnvironment) {
	var taintedNodes []string
	var errNodes []string

	// Loop through the debug pods that are tied to each node.
	for _, dp := range env.DebugPods {
		// Build a NodeTainted tester object.
		c := clientsholder.NewClientsHolder()
		tester := nodetainted.NewNodeTaintedTester(common.DefaultTimeout, c, clientsholder.Context{
			Namespace:     dp.Namespace,
			Podname:       dp.Name,
			Containername: dp.Spec.Containers[0].Name,
		})
		taintInfo, err := tester.GetKernelTaintInfo()
		if err != nil {
			logrus.Error("failed to retrieve kernel taint information from debug pod/host")
			tnf.ClaimFilePrintf("failed to retrieve kernel taint information from debug pod/host")
			errNodes = append(errNodes, dp.Name)
			break
		}
		tnf.ClaimFilePrintf(fmt.Sprintf("Namespace: %s Pod: %s taintInfo retrieved: %s", dp.Namespace, dp.Name, taintInfo))

		var taintedBitmap uint64
		nodeTaintsAccepted := true
		taintedBitmap, err = strconv.ParseUint(taintInfo, 10, 64) //nolint:gomnd // base 10 and uint64

		if err != nil {
			logrus.Errorf("failed to parse uint with: %s", err)
			tnf.ClaimFilePrintf("Could not decode tainted kernel causes (code=%d) for node %s\n", taintedBitmap, dp.Name)
			errNodes = append(errNodes, dp.Name)
			break
		}
		taintMsg, individualTaints := nodetainted.DecodeKernelTaints(taintedBitmap)

		// We only will fail the tainted kernel check if the reason for the taint
		// only pertains to `module was loaded`.
		logrus.Debug("Checking for 'module was loaded' taints")
		logrus.Debug("individualTaints", individualTaints)
		moduleCheck := false
		for _, it := range individualTaints {
			if strings.Contains(it, `module was loaded`) {
				moduleCheck = true
				break
			}
		}

		if moduleCheck {
			// Retrieve the modules from the node (via the debug pod)
			modules := tester.GetModulesFromNode(dp.Name)
			logrus.Debugf("Got the modules from node %s: %v", dp.Name, modules)

			// Retrieve all of the out of tree modules.
			taintedModules := nodetainted.GetOutOfTreeModules(modules, dp.Name, tester)
			logrus.Debug("Collected all of the tainted modules: ", taintedModules)
			logrus.Debug("Modules allowed via configuration: ", env.Config.AcceptedKernelTaints)

			// Looks through the accepted taints listed in the tnf-config file.
			// If all of the tainted modules show up in the configuration file, don't fail the test.
			nodeTaintsAccepted = nodetainted.TaintsAccepted(env.Config.AcceptedKernelTaints, taintedModules)
		}

		// Only add the tainted node to the slice if the taint is acceptable.
		if !nodeTaintsAccepted {
			taintedNodes = append(taintedNodes, dp.Name)
		}

		// Only print the message if there is something to report failure or tainted node wise.
		if len(taintedNodes) != 0 || len(errNodes) != 0 {
			tnf.ClaimFilePrintf("Decoded tainted kernel causes (code=%d) for node %s : %s\n", taintedBitmap, dp.Name, taintMsg)
		}
	}

	// We are expecting tainted nodes to be Nil, but only if:
	// 1) The reason for the tainted node is contains(`module was loaded`)
	// 2) The modules loaded are all whitelisted.
	gomega.Expect(taintedNodes).To(gomega.BeNil())
	gomega.Expect(errNodes).To(gomega.BeNil())
}

func testIsRedHatRelease(env *provider.TestEnvironment) {
	ginkgo.By("should report a proper Red Hat version")
	failedContainers := []string{}
	logrus.Infof("looping through %d containers", len(env.Containers))
	for _, cut := range env.Containers {
		ginkgo.By(fmt.Sprintf("%s(%s) is checked for Red Hat version", cut.Podname, cut.Data.Name))
		baseImageTester := isredhat.NewBaseImageTester(common.DefaultTimeout, clientsholder.NewClientsHolder(), clientsholder.Context{
			Namespace:     cut.Namespace,
			Podname:       cut.Podname,
			Containername: cut.Data.Name,
		})

		result, err := baseImageTester.TestContainerIsRedHatRelease()
		if err != nil {
			logrus.Debug("failed to collect release information from container: ")
		}
		if !result {
			failedContainers = append(failedContainers, cut.Namespace+"/"+cut.Podname+"/"+cut.Data.Name)
			tnf.ClaimFilePrintf("Container: %s/%s (ns: %s) has failed the RHEL release check", cut.Podname, cut.Data.Name, cut.Namespace)
		}
	}

	gomega.Expect(failedContainers).To(gomega.BeEmpty())
}
