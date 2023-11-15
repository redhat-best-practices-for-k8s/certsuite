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

package platform

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	clientsholder "github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/compatibility"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform/bootparams"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform/cnffsdiff"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform/hugepages"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform/isredhat"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform/operatingsystem"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform/sysctlconfig"

	"github.com/onsi/ginkgo/v2"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform/nodetainted"
)

// All actual test code belongs below here.  Utilities belong above.
var _ = ginkgo.Describe(common.PlatformAlterationTestKey, func() {
	logrus.Debugf("Entering %s suite", common.PlatformAlterationTestKey)
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})

	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestHyperThreadEnable)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.GetBaremetalNodes(), "env.GetBaremetalNodes()"))
		testHyperThreadingEnabled(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestUnalteredBaseImageIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if !provider.IsOCPCluster() {
			ginkgo.Skip("Non-OCP cluster found, skipping testContainersFsDiff")
		}
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping testContainersFsDiff")
		}
		testContainersFsDiff(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNonTaintedNodeKernelsIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.DebugPods, "env.DebugPods"))
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping testTainted")
		}
		testTainted(&env) // Kind tainted kernels are allowed via config
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestIsRedHatReleaseIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testIsRedHatRelease(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestIsSELinuxEnforcingIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if !provider.IsOCPCluster() {
			ginkgo.Skip("Non-OCP cluster found, skipping testIsSELinuxEnforcing")
		}
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.DebugPods, "env.DebugPods"))
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping testIsSELinuxEnforcing")
		}
		testIsSELinuxEnforcing(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestHugepagesNotManuallyManipulated)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if !provider.IsOCPCluster() {
			ginkgo.Skip("Non-OCP cluster found, skipping testHugepages")
		}
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.DebugPods, "env.DebugPods"))
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping testHugepages")
		}
		testHugepages(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestUnalteredStartupBootParamsIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if !provider.IsOCPCluster() {
			ginkgo.Skip("Non-OCP cluster found, skipping testUnalteredBootParams")
		}
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.DebugPods, "env.DebugPods"))
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping testUnalteredBootParams")
		}
		testUnalteredBootParams(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSysctlConfigsIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if !provider.IsOCPCluster() {
			ginkgo.Skip("Non-OCP cluster found, skipping testSysctlConfigs")
		}
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.DebugPods, "env.DebugPods"))
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping testSysctlConfigs")
		}
		testSysctlConfigs(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestServiceMeshIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		testServiceMesh(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestOCPLifecycleIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if !provider.IsOCPCluster() {
			ginkgo.Skip("Non-OCP cluster found, skipping testOCPStatus")
		}
		testOCPStatus(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNodeOperatingSystemIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if !provider.IsOCPCluster() {
			ginkgo.Skip("Non-OCP cluster found, skipping testNodeOperatingSystemStatus")
		}
		testNodeOperatingSystemStatus(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHugePages2M)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if !provider.IsOCPCluster() {
			ginkgo.Skip("Non-OCP cluster found, skipping testPodHugePagesSize2M")
		}
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.GetHugepagesPods(), "env.GetHugepagesPods()"))
		testPodHugePagesSize(&env, provider.HugePages2Mi)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHugePages1G)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if !provider.IsOCPCluster() {
			ginkgo.Skip("Non-OCP cluster found, skipping testPodHugePagesSize1G")
		}
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.GetHugepagesPods(), "env.GetHugepagesPods()"))
		testPodHugePagesSize(&env, provider.HugePages1Gi)
	})

})

func testHyperThreadingEnabled(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	baremetalNodes := env.GetBaremetalNodes()
	for _, node := range baremetalNodes {
		nodeName := node.Data.Name
		enable, err := node.IsHyperThreadNode(env)
		//nolint:gocritic
		if enable {
			compliantObjects = append(compliantObjects, testhelper.NewNodeReportObject(nodeName, "Node has hyperthreading enabled", true))
		} else if err != nil {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(nodeName, "Error with executing the checke for hyperthreading: "+err.Error(), false))
		} else {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(nodeName, "Node has hyperthreading disabled ", false))
		}
	}
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testServiceMesh(env *provider.TestEnvironment) {
	// check if istio is installed
	if !env.IstioServiceMeshFound {
		tnf.ClaimFilePrintf("Istio is not installed")
		ginkgo.Skip("No service mesh detected.")
	}
	tnf.ClaimFilePrintf("Istio is installed")

	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		istioProxyFound := false
		for _, cut := range put.Containers {
			if cut.IsIstioProxy() {
				tnf.ClaimFilePrintf("Istio proxy container found on %s", put)
				istioProxyFound = true
				break
			}
		}
		if !istioProxyFound {
			tnf.ClaimFilePrintf("Pod found without service mesh: %s", put.String())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod found without service mesh container", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod found with service mesh container", true))
		}
	}
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testContainersFsDiff test that all CUT did not install new packages are starting
func testContainersFsDiff(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		logrus.Debug(fmt.Sprintf("%s should not install new packages after starting", cut.String()))
		debugPod := env.DebugPods[cut.NodeName]
		if debugPod == nil {
			ginkgo.Fail(fmt.Sprintf("Debug pod not found on Node: %s", cut.NodeName))
		}
		ctxt := clientsholder.NewContext(debugPod.Namespace, debugPod.Name, debugPod.Spec.Containers[0].Name)
		fsDiffTester := cnffsdiff.NewFsDiffTester(clientsholder.GetClientsHolder(), ctxt)
		fsDiffTester.RunTest(cut.UID)
		switch fsDiffTester.GetResults() {
		case testhelper.SUCCESS:
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is not modified", true))
			continue
		case testhelper.FAILURE:
			tnf.ClaimFilePrintf("%s - changed folders: %v, deleted folders: %v", cut, fsDiffTester.ChangedFolders, fsDiffTester.DeletedFolders)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is modified", false).
				AddField("ChangedFolders", strings.Join(fsDiffTester.ChangedFolders, ",")).
				AddField("DeletedFolders", strings.Join(fsDiffTester.DeletedFolders, ",")))

		case testhelper.ERROR:
			tnf.ClaimFilePrintf("%s - error while running fs-diff: %v", cut, fsDiffTester.Error)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Error while running fs-diff", false).AddField(testhelper.Error, fsDiffTester.Error.Error()))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

//nolint:funlen
func testTainted(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	// errNodes has nodes that failed some operation while checking kernel taints.
	errNodes := []string{}

	type badModuleTaints struct {
		name   string
		taints []string
	}

	// badModules maps node names to list of "bad"/offending modules.
	badModules := map[string][]badModuleTaints{}
	// otherTaints maps a node to a list of taint bits that haven't been set by any module.
	otherTaints := map[string][]int{}

	logrus.Infof("Modules allowlist: %+v", env.Config.AcceptedKernelTaints)
	// helper map to make the checks easier.
	allowListedModules := map[string]bool{}
	for _, module := range env.Config.AcceptedKernelTaints {
		allowListedModules[module.Module] = true
	}

	// Loop through the debug pods that are tied to each node.
	for _, dp := range env.DebugPods {
		nodeName := dp.Spec.NodeName

		ginkgo.By(fmt.Sprintf("Checking kernel taints of node %s", nodeName))

		ocpContext := clientsholder.NewContext(dp.Namespace, dp.Name, dp.Spec.Containers[0].Name)
		tf := nodetainted.NewNodeTaintedTester(&ocpContext, nodeName)

		// Get the taints mask from the node kernel
		taintsMask, err := tf.GetKernelTaintsMask()
		if err != nil {
			tnf.ClaimFilePrintf("Failed to retrieve kernel taint information from node %s: %v", nodeName, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(nodeName, "Failed to retrieve kernel taint information from node", false).
				AddField(testhelper.Error, err.Error()))
			continue
		}

		if taintsMask == 0 {
			tnf.ClaimFilePrintf("Node %s has no non-approved kernel taints.", nodeName)
			compliantObjects = append(compliantObjects, testhelper.NewNodeReportObject(nodeName, "Node has no non-approved kernel taints", true))
			continue
		}

		tnf.ClaimFilePrintf("Node %s kernel is tainted. Taints mask=%d - Decoded taints: %v",
			nodeName, taintsMask, nodetainted.DecodeKernelTaintsFromBitMask(taintsMask))

		// Check the allow list. If empty, mark this node as failed.
		if len(allowListedModules) == 0 {
			taintsMaskStr := strconv.FormatUint(taintsMask, 10)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(nodeName, "Node contains taints not covered by module allowlist", false).
				AddField(testhelper.TaintMask, taintsMaskStr).
				AddField(testhelper.Taints, strings.Join(nodetainted.DecodeKernelTaintsFromBitMask(taintsMask), ",")))
			continue
		}

		// allow list check.
		// Get the list of modules (tainters) that have set a taint bit.
		//   1. Each module should appear in the allow list.
		//   2. All kernel taint bits (one bit <-> one letter) should have been set by at least
		//      one tainter module.
		tainters, taintBitsByAllModules, err := tf.GetTainterModules(allowListedModules)
		if err != nil {
			tnf.ClaimFilePrintf("failed to get tainter modules from node %s: %v", nodeName, err)
			errNodes = append(errNodes, nodeName)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(nodeName, "Failed to get tainter modules", false).
				AddField(testhelper.Error, err.Error()))
			continue
		}

		// Keep track of whether or not this node is compliant with module allow list.
		compliantNode := true

		// Save modules' names only.
		for moduleName, taintsLetters := range tainters {
			moduleTaints := nodetainted.DecodeKernelTaintsFromLetters(taintsLetters)
			badModules[nodeName] = append(badModules[nodeName], badModuleTaints{name: moduleName, taints: moduleTaints})

			// Create non-compliant taint objects for each of the taints
			for _, taint := range moduleTaints {
				tnf.ClaimFilePrintf("Node %s - module %s taints kernel: %s", nodeName, moduleName, taint)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewTaintReportObject(nodetainted.RemoveAllExceptNumbers(taint), nodeName, taint, false).AddField(testhelper.ModuleName, moduleName))

				// Set the node as non-compliant for future reporting
				compliantNode = false
			}
		}

		// Lastly, check that all kernel taint bits come from modules.
		otherKernelTaints := nodetainted.GetOtherTaintedBits(taintsMask, taintBitsByAllModules)
		for _, taintedBit := range otherKernelTaints {
			tnf.ClaimFilePrintf("Node %s - taint bit %d is set but it is not caused by any module.", nodeName, taintedBit)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewTaintReportObject(strconv.Itoa(taintedBit), nodeName, nodetainted.GetTaintMsg(taintedBit), false).
				AddField(testhelper.ModuleName, "N/A"))
			otherTaints[nodeName] = append(otherTaints[nodeName], taintedBit)

			// Set the node as non-compliant for future reporting
			compliantNode = false
		}

		if compliantNode {
			compliantObjects = append(compliantObjects, testhelper.NewNodeReportObject(nodeName, "Passed the tainted kernel check", true))
		}
	}

	logrus.Infof("Nodes with errors: %+v", errNodes)
	logrus.Infof("Bad Modules: %+v", badModules)
	logrus.Infof("Taints not related to any module: %+v", otherTaints)

	if len(errNodes) > 0 {
		logrus.Infof("Failed to get kernel taints from some nodes: %+v", errNodes)
	}

	if len(badModules) > 0 || len(otherTaints) > 0 {
		logrus.Info("Nodes have been found to be tainted. Check claim log for more details.")
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testIsRedHatRelease(env *provider.TestEnvironment) {
	ginkgo.By("should report a proper Red Hat version")
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		ginkgo.By(fmt.Sprintf("%s is checked for Red Hat version", cut))
		baseImageTester := isredhat.NewBaseImageTester(clientsholder.GetClientsHolder(), clientsholder.NewContext(cut.Namespace, cut.Podname, cut.Name))

		result, err := baseImageTester.TestContainerIsRedHatRelease()
		if err != nil {
			logrus.Error("failed to collect release information from container: ", err)
		}
		if !result {
			tnf.ClaimFilePrintf("%s has failed the RHEL release check", cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Failed the RHEL release check", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Passed the RHEL release check", true))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testIsSELinuxEnforcing(env *provider.TestEnvironment) {
	const (
		getenforceCommand = `chroot /host getenforce`
		enforcingString   = "Enforcing\n"
	)
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	o := clientsholder.GetClientsHolder()
	nodesFailed := 0
	nodesError := 0
	for _, debugPod := range env.DebugPods {
		ctx := clientsholder.NewContext(debugPod.Namespace, debugPod.Name, debugPod.Spec.Containers[0].Name)
		outStr, errStr, err := o.ExecCommandContainer(ctx, getenforceCommand)
		if err != nil || errStr != "" {
			logrus.Errorf("Failed to execute command %s in debug %s, errStr: %s, err: %v", getenforceCommand, debugPod.String(), errStr, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(debugPod.Namespace, debugPod.Name, "Failed to execute command", false))
			nodesError++
			continue
		}
		if outStr != enforcingString {
			tnf.ClaimFilePrintf(fmt.Sprintf("Node %s is not running selinux, %s command returned: %s", debugPod.Spec.NodeName, getenforceCommand, outStr))
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(debugPod.Spec.NodeName, "SELinux is not enforced", false))
			nodesFailed++
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewNodeReportObject(debugPod.Spec.NodeName, "SELinux is enforced", true))
		}
	}
	if nodesError > 0 {
		logrus.Infof("Failed because could not run %s command on %d nodes", getenforceCommand, nodesError)
	}
	if nodesFailed > 0 {
		logrus.Infof(fmt.Sprintf("Failed because %d nodes are not running selinux", nodesFailed))
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testHugepages(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for i := range env.Nodes {
		node := env.Nodes[i]
		if !node.IsWorkerNode() {
			compliantObjects = append(compliantObjects, testhelper.NewNodeReportObject(node.Data.Name, "Not a worker node", true))
			continue
		}

		debugPod, exist := env.DebugPods[node.Data.Name]
		if !exist {
			tnf.ClaimFilePrintf("Node %s: tnf debug pod not found.", node.Data.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(node.Data.Name, "tnf debug pod not found", false))
			continue
		}

		hpTester, err := hugepages.NewTester(&node, debugPod, clientsholder.GetClientsHolder())
		if err != nil {
			tnf.ClaimFilePrintf("Unable to get node hugepages tester for node %s, err: %v", node.Data.Name, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(node.Data.Name, "Unable to get node hugepages tester", false))
		}

		if err := hpTester.Run(); err != nil {
			tnf.ClaimFilePrintf("Node %s: %v", node.Data.Name, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(node.Data.Name, err.Error(), false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewNodeReportObject(node.Data.Name, "Passed the hugepages check", true))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testUnalteredBootParams(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	alreadyCheckedNodes := map[string]bool{}
	for _, cut := range env.Containers {
		if alreadyCheckedNodes[cut.NodeName] {
			logrus.Debugf("Skipping node %s: already checked.", cut.NodeName)
			continue
		}
		alreadyCheckedNodes[cut.NodeName] = true

		claimsLog, err := bootparams.TestBootParamsHelper(env, cut)

		if err != nil || len(claimsLog.GetLogLines()) != 0 {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(cut.NodeName, "Failed the boot params check", false).
				AddField(testhelper.DebugPodName, env.DebugPods[cut.NodeName].Name))
			tnf.ClaimFilePrintf("%s", claimsLog.GetLogLines())
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewNodeReportObject(cut.NodeName, "Passed the boot params check", true).
				AddField(testhelper.DebugPodName, env.DebugPods[cut.NodeName].Name))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testSysctlConfigs(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	alreadyCheckedNodes := map[string]bool{}
	for _, cut := range env.Containers {
		if alreadyCheckedNodes[cut.NodeName] {
			continue
		}
		alreadyCheckedNodes[cut.NodeName] = true
		debugPod := env.DebugPods[cut.NodeName]
		if debugPod == nil {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(cut.NodeName, "tnf debug pod not found", false))
			continue
		}

		sysctlSettings, err := sysctlconfig.GetSysctlSettings(env, cut.NodeName)
		if err != nil {
			tnf.ClaimFilePrintf("Could not get sysctl settings for node %s, error: %v", cut.NodeName, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(cut.NodeName, "Could not get sysctl settings", false))
			continue
		}

		mcKernelArgumentsMap := bootparams.GetMcKernelArguments(env, cut.NodeName)
		validSettings := true
		for key, sysctlConfigVal := range sysctlSettings {
			if mcVal, ok := mcKernelArgumentsMap[key]; ok {
				if mcVal != sysctlConfigVal {
					tnf.ClaimFilePrintf(fmt.Sprintf("Kernel config mismatch in node %s for %s (sysctl value: %s, machine config value: %s)",
						cut.NodeName, key, sysctlConfigVal, mcVal))
					nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(cut.NodeName, fmt.Sprintf("Kernel config mismatch for %s", key), false))
					validSettings = false
				}
			}
		}
		if validSettings {
			compliantObjects = append(compliantObjects, testhelper.NewNodeReportObject(cut.NodeName, "Passed the sysctl config check", true))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testOCPStatus(env *provider.TestEnvironment) {
	ginkgo.By("Testing the OCP Version for lifecycle status")

	switch env.OCPStatus {
	case compatibility.OCPStatusEOL:
		msg := fmt.Sprintf("OCP Version %s has been found to be in end of life", env.OpenshiftVersion)
		tnf.ClaimFilePrintf(msg)
		ginkgo.Fail(msg)
	case compatibility.OCPStatusMS:
		msg := fmt.Sprintf("OCP Version %s has been found to be in maintenance support", env.OpenshiftVersion)
		tnf.ClaimFilePrintf(msg)
	case compatibility.OCPStatusGA:
		msg := fmt.Sprintf("OCP Version %s has been found to be in general availability", env.OpenshiftVersion)
		tnf.ClaimFilePrintf(msg)
	case compatibility.OCPStatusPreGA:
		msg := fmt.Sprintf("OCP Version %s has been found to be in pre-general availability", env.OpenshiftVersion)
		tnf.ClaimFilePrintf(msg)
	default:
		msg := fmt.Sprintf("OCP Version %s was unable to be found in the lifecycle compatibility matrix", env.OpenshiftVersion)
		tnf.ClaimFilePrintf(msg)
	}
}

//nolint:funlen
func testNodeOperatingSystemStatus(env *provider.TestEnvironment) {
	ginkgo.By("Testing the control-plane and workers in the cluster for Operating System compatibility")

	logrus.Debug(fmt.Sprintf("There are %d nodes to process for Operating System compatibility.", len(env.Nodes)))

	failedControlPlaneNodes := []string{}
	failedWorkerNodes := []string{}
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, node := range env.Nodes {
		// Get the OSImage which should tell us what version of operating system the node is running.
		logrus.Debug(fmt.Sprintf("Node %s is running operating system: %s", node.Data.Name, node.Data.Status.NodeInfo.OSImage))

		// Control plane nodes must be RHCOS (also CentOS Stream starting in OCP 4.13)
		// Per the release notes from OCP documentation:
		// "You must use RHCOS machines for the control plane, and you can use either RHCOS or RHEL for compute machines."
		if node.IsMasterNode() && !node.IsRHCOS() && !node.IsCSCOS() {
			tnf.ClaimFilePrintf("Master node %s has been found to be running an incompatible operating system: %s", node.Data.Name, node.Data.Status.NodeInfo.OSImage)
			failedControlPlaneNodes = append(failedControlPlaneNodes, node.Data.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(node.Data.Name, "Master node has been found to be running an incompatible OS", false).AddField(testhelper.OSImage, node.Data.Status.NodeInfo.OSImage))
			continue
		}

		// Worker nodes can either be RHEL or RHCOS
		if node.IsWorkerNode() {
			//nolint:gocritic
			if node.IsRHCOS() {
				// Get the short version from the node
				shortVersion, err := node.GetRHCOSVersion()
				if err != nil {
					tnf.ClaimFilePrintf("Node %s failed to gather RHCOS version. Error: %v", node.Data.Name, err)
					failedWorkerNodes = append(failedWorkerNodes, node.Data.Name)
					nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(node.Data.Name, "Failed to gather RHCOS version", false))
					continue
				}

				if shortVersion == operatingsystem.NotFoundStr {
					tnf.ClaimFilePrintf("Node %s has an RHCOS operating system that is not found in our internal database.  Skipping as to not cause failures due to database mismatch.", node.Data.Name)
					continue
				}

				// If the node's RHCOS version and the OpenShift version are not compatible, the node fails.
				logrus.Debugf("Comparing RHCOS shortVersion: %s to openshiftVersion: %s", shortVersion, env.OpenshiftVersion)
				if !compatibility.IsRHCOSCompatible(shortVersion, env.OpenshiftVersion) {
					tnf.ClaimFilePrintf("Node %s has been found to be running an incompatible version of RHCOS: %s", node.Data.Name, shortVersion)
					failedWorkerNodes = append(failedWorkerNodes, node.Data.Name)
					nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(node.Data.Name, "Worker node has been found to be running an incompatible OS", false).
						AddField(testhelper.OSImage, node.Data.Status.NodeInfo.OSImage))
					continue
				}
				compliantObjects = append(compliantObjects, testhelper.NewNodeReportObject(node.Data.Name, "Worker node has been found to be running a compatible OS", true).
					AddField(testhelper.OSImage, node.Data.Status.NodeInfo.OSImage))
			} else if node.IsCSCOS() {
				// Get the short version from the node
				shortVersion, err := node.GetCSCOSVersion()
				if err != nil {
					tnf.ClaimFilePrintf("Node %s failed to gather CentOS Stream CoreOS version. Error: %v", node.Data.Name, err)
					failedWorkerNodes = append(failedWorkerNodes, node.Data.Name)
					nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(node.Data.Name, "Failed to gather CentOS Stream CoreOS version", false))
					continue
				}

				// Warning: CentOS Stream CoreOS has not been released yet in any
				// OCP RC/GA versions, so for the moment, we cannot compare the
				// version with the OCP one, or retrieve it on the internal database
				msg := `
					Node %s is using CentOS Stream CoreOS %s, which is not being used yet in any
					OCP RC/GA version. Relaxing the conditions to check the OS as a result.
					`
				tnf.ClaimFilePrintf(msg, node.Data.Name, shortVersion)
			} else if node.IsRHEL() {
				// Get the short version from the node
				shortVersion, err := node.GetRHELVersion()
				if err != nil {
					tnf.ClaimFilePrintf("Node %s failed to gather RHEL version. Error: %v", node.Data.Name, err)
					failedWorkerNodes = append(failedWorkerNodes, node.Data.Name)
					nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(node.Data.Name, "Failed to gather RHEL version", false))
					continue
				}

				// If the node's RHEL version and the OpenShift version are not compatible, the node fails.
				logrus.Debugf("Comparing RHEL shortVersion: %s to openshiftVersion: %s", shortVersion, env.OpenshiftVersion)
				if !compatibility.IsRHELCompatible(shortVersion, env.OpenshiftVersion) {
					tnf.ClaimFilePrintf("Node %s has been found to be running an incompatible version of RHEL: %s", node.Data.Name, shortVersion)
					failedWorkerNodes = append(failedWorkerNodes, node.Data.Name)
					nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(node.Data.Name, "Worker node has been found to be running an incompatible OS", false).AddField(testhelper.OSImage, node.Data.Status.NodeInfo.OSImage))
				} else {
					compliantObjects = append(compliantObjects, testhelper.NewNodeReportObject(node.Data.Name, "Worker node has been found to be running a compatible OS", true).AddField(testhelper.OSImage, node.Data.Status.NodeInfo.OSImage))
				}
			} else {
				tnf.ClaimFilePrintf("Node %s has been found to be running an incompatible operating system", node.Data.Name)
				failedWorkerNodes = append(failedWorkerNodes, node.Data.Name)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(node.Data.Name, "Worker node has been found to be running an incompatible OS", false).AddField(testhelper.OSImage, node.Data.Status.NodeInfo.OSImage))
			}
		}
	}

	var b strings.Builder
	if n := len(failedControlPlaneNodes); n > 0 {
		errMsg := fmt.Sprintf("Number of control plane nodes running non-RHCOS based operating systems: %d", n)
		b.WriteString(errMsg)
		tnf.ClaimFilePrintf(errMsg)
	}

	if n := len(failedWorkerNodes); n > 0 {
		errMsg := fmt.Sprintf("Number of worker nodes running non-RHCOS or non-RHEL based operating systems: %d", n)
		b.WriteString(errMsg)
		tnf.ClaimFilePrintf(errMsg)
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testPodHugePagesSize(env *provider.TestEnvironment, size string) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.GetHugepagesPods() {
		result := put.CheckResourceHugePagesSize(size)
		if !result {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has been found to be running with an incorrect hugepages size", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has been found to be running with a correct hugepages size", true))
		}
	}
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}
