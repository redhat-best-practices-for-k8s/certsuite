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

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"

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
	ginkgo.ReportAfterEach(results.RecordResult)

	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestUnalteredBaseImageIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if !provider.IsOCPCluster() {
			ginkgo.Skip("Non-OCP cluster found, skipping testContainersFsDiff")
		}
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping testContainersFsDiff")
		}
		testContainersFsDiff(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNonTaintedNodeKernelsIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.DebugPods)
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping testTainted")
		}
		testTainted(&env) // Kind tainted kernels are allowed via config
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestIsRedHatReleaseIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		testIsRedHatRelease(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestIsSELinuxEnforcingIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if !provider.IsOCPCluster() {
			ginkgo.Skip("Non-OCP cluster found, skipping testIsSELinuxEnforcing")
		}
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.DebugPods)
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
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.DebugPods)
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
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.DebugPods)
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
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.DebugPods)
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping testSysctlConfigs")
		}
		testSysctlConfigs(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestServiceMeshIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestServiceMesh(&env)
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
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.GetHugepagesPods())
		testPodHugePagesSize(&env, provider.HugePages2Mi)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHugePages1G)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if !provider.IsOCPCluster() {
			ginkgo.Skip("Non-OCP cluster found, skipping testPodHugePagesSize1G")
		}
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.GetHugepagesPods())
		testPodHugePagesSize(&env, provider.HugePages1Gi)
	})

})

func TestServiceMesh(env *provider.TestEnvironment) {
	// check if istio is installed
	if !env.IstioServiceMesh {
		tnf.ClaimFilePrintf("Istio is not installed")
		ginkgo.Skip("No service mesh detected.")
	}
	tnf.ClaimFilePrintf("Istio is installed")

	var badPods []string
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
			badPods = append(badPods, put.String())
			tnf.ClaimFilePrintf("Pod found without service mesh: %s", put.String())
		}
	}
	testhelper.AddTestResultLog("Non-compliant", badPods, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testContainersFsDiff test that all CUT did not install new packages are starting
func testContainersFsDiff(env *provider.TestEnvironment) {
	var badContainers []string
	var errContainers []string
	for _, cut := range env.Containers {
		logrus.Debug(fmt.Sprintf("%s should not install new packages after starting", cut.String()))
		debugPod := env.DebugPods[cut.NodeName]
		if debugPod == nil {
			ginkgo.Fail(fmt.Sprintf("Debug pod not found on Node: %s", cut.NodeName))
		}
		fsDiffTester := cnffsdiff.NewFsDiffTester(clientsholder.GetClientsHolder())
		fsDiffTester.RunTest(clientsholder.NewContext(debugPod.Namespace, debugPod.Name, debugPod.Spec.Containers[0].Name), cut.UID)
		switch fsDiffTester.GetResults() {
		case testhelper.SUCCESS:
			continue
		case testhelper.FAILURE:
			tnf.ClaimFilePrintf("%s - changed folders: %v, deleted folders: %v", cut, fsDiffTester.ChangedFolders, fsDiffTester.DeletedFolders)
			badContainers = append(badContainers, cut.Name)
		case testhelper.ERROR:
			tnf.ClaimFilePrintf("%s - error while running fs-diff: %v: ", cut, fsDiffTester.Error)
			errContainers = append(errContainers, cut.Name)
		}
	}

	if len(badContainers) > 0 {
		tnf.ClaimFilePrintf("Containers were found with changed or deleted folders: %v", badContainers)
		ginkgo.Fail("Containers were found with changed or deleted folders.")
	}

	if len(errContainers) > 0 {
		tnf.ClaimFilePrintf("Containers were unable to run fs-diff: %v", errContainers)
		ginkgo.Fail("Containers were unable to run fs-diff.")
	}
}

//nolint:funlen
func testTainted(env *provider.TestEnvironment) {
	// errNodes has nodes that failed some operation while checking kernel taints.
	errNodes := []string{}
	// badModules maps node names to list of "bad"/offending modules.

	type badModuleTaints struct {
		name   string
		taints []string
	}

	badModules := map[string][]badModuleTaints{}
	// otherTaints maps a node to a list of taint bits that haven't been set by any module.
	otherTaints := map[string][]int{}

	logrus.Infof("Modules whitelist: %+v", env.Config.AcceptedKernelTaints)
	// helper map to make the checks easier.
	whiteListedModules := map[string]bool{}
	for _, module := range env.Config.AcceptedKernelTaints {
		whiteListedModules[module.Module] = true
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
			errNodes = append(errNodes, nodeName)
			continue
		}

		if taintsMask == 0 {
			tnf.ClaimFilePrintf("Node %s has no kernel taints.", nodeName)
			continue
		}

		tnf.ClaimFilePrintf("Node %s kernel is tainted. Taints mask=%d - Decoded taints: %v",
			nodeName, taintsMask, nodetainted.DecodeKernelTaintsFromBitMask(taintsMask))

		// Check the white list. If empty, mark this node as failed.
		if len(whiteListedModules) == 0 {
			errNodes = append(errNodes, nodeName)
			continue
		}

		// White list check.
		// Get the list of modules (tainters) that have set a taint bit.
		//   1. Each module should appear in the white list.
		//   2. All kernel taint bits (one bit <-> one letter) should have been set by at least
		//      one tainter module.
		tainters, taintBitsByAllModules, err := tf.GetTainterModules(whiteListedModules)
		if err != nil {
			tnf.ClaimFilePrintf("failed to get tainter modules from node %s: %v", nodeName, err)
			errNodes = append(errNodes, nodeName)
			continue
		}

		// Save modules' names only.
		for moduleName, taintsLetters := range tainters {
			moduleTaints := nodetainted.DecodeKernelTaintsFromLetters(taintsLetters)
			badModules[nodeName] = append(badModules[nodeName], badModuleTaints{name: moduleName, taints: moduleTaints})

			tnf.ClaimFilePrintf("Node %s - module %s taints kernel: %s", nodeName, moduleName, moduleTaints)
		}

		// Lastly, check that all kernel taint bits come from modules.
		otherKernelTaints := nodetainted.GetOtherTaintedBits(taintsMask, taintBitsByAllModules)
		for _, taintedBit := range otherKernelTaints {
			tnf.ClaimFilePrintf("Node %s - taint bit %d is set but it's not caused by any module.", nodeName, taintedBit)
			otherTaints[nodeName] = append(otherTaints[nodeName], taintedBit)
		}
	}

	logrus.Infof("Nodes with errors: %+v", errNodes)
	logrus.Infof("Bad Modules: %+v", badModules)
	logrus.Infof("Taints not related to any module: %+v", otherTaints)

	if len(errNodes) > 0 {
		ginkgo.Fail(fmt.Sprintf("Failed to get kernel taints from some nodes: %+v", errNodes))
	}

	if len(badModules) > 0 || len(otherTaints) > 0 {
		ginkgo.Fail("Nodes have been found to be tainted. Check claim log for more details.")
	}
}

func testIsRedHatRelease(env *provider.TestEnvironment) {
	ginkgo.By("should report a proper Red Hat version")
	failedContainers := []string{}
	for _, cut := range env.Containers {
		ginkgo.By(fmt.Sprintf("%s is checked for Red Hat version", cut))
		baseImageTester := isredhat.NewBaseImageTester(clientsholder.GetClientsHolder(), clientsholder.NewContext(cut.Namespace, cut.Podname, cut.Name))

		result, err := baseImageTester.TestContainerIsRedHatRelease()
		if err != nil {
			logrus.Error("failed to collect release information from container: ", err)
		}
		if !result {
			failedContainers = append(failedContainers, cut.Namespace+"/"+cut.Podname+"/"+cut.Name)
			tnf.ClaimFilePrintf("%s has failed the RHEL release check", cut)
		}
	}

	if len(failedContainers) > 0 {
		tnf.ClaimFilePrintf("Containers have been found without a proper Red Hat version: %v", failedContainers)
		ginkgo.Fail("Containers have been found without a proper Red Hat version.")
	}
}

func testIsSELinuxEnforcing(env *provider.TestEnvironment) {
	const (
		getenforceCommand = `chroot /host getenforce`
		enforcingString   = "Enforcing\n"
	)
	o := clientsholder.GetClientsHolder()
	nodesFailed := 0
	nodesError := 0
	for _, debugPod := range env.DebugPods {
		ctx := clientsholder.NewContext(debugPod.Namespace, debugPod.Name, debugPod.Spec.Containers[0].Name)
		outStr, errStr, err := o.ExecCommandContainer(ctx, getenforceCommand)
		if err != nil || errStr != "" {
			logrus.Errorf("Failed to execute command %s in debug %s, errStr: %s, err: %v", getenforceCommand, debugPod.String(), errStr, err)
			nodesError++
			continue
		}
		if outStr != enforcingString {
			tnf.ClaimFilePrintf(fmt.Sprintf("Node %s is not running selinux, %s command returned: %s", debugPod.Spec.NodeName, getenforceCommand, outStr))
			nodesFailed++
		}
	}
	if nodesError > 0 {
		ginkgo.Fail(fmt.Sprintf("Failed because could not run %s command on %d nodes", getenforceCommand, nodesError))
	}
	if nodesFailed > 0 {
		ginkgo.Fail(fmt.Sprintf("Failed because %d nodes are not running selinux", nodesFailed))
	}
}

func testHugepages(env *provider.TestEnvironment) {
	var badNodes []string
	for i := range env.Nodes {
		node := env.Nodes[i]
		if !node.IsWorkerNode() {
			continue
		}

		hpTester, err := hugepages.NewTester(&node, env.DebugPods[node.Data.Name], clientsholder.GetClientsHolder())
		if err != nil {
			tnf.ClaimFilePrintf("Unable to get node hugepages tester for node %s, err: %v", node.Data.Name, err)
			badNodes = append(badNodes, node.Data.Name)
		}

		if err := hpTester.Run(); err != nil {
			tnf.ClaimFilePrintf("Node %s: %v", node.Data.Name, err)
			badNodes = append(badNodes, node.Data.Name)
		}
	}

	if n := len(badNodes); n > 0 {
		ginkgo.Fail(fmt.Sprintf("Found %d failing nodes: %v", n, badNodes))
	}
}

func testUnalteredBootParams(env *provider.TestEnvironment) {
	failedNodes := []string{}
	alreadyCheckedNodes := map[string]bool{}
	for _, cut := range env.Containers {
		if alreadyCheckedNodes[cut.NodeName] {
			logrus.Debugf("Skipping node %s: already checked.", cut.NodeName)
			continue
		}
		alreadyCheckedNodes[cut.NodeName] = true

		claimsLog, err := bootparams.TestBootParamsHelper(env, cut)

		if err != nil || len(claimsLog.GetLogLines()) != 0 {
			failedNodes = append(failedNodes, fmt.Sprintf("node %s (%s)", cut.NodeName, cut.String()))
			tnf.ClaimFilePrintf("%s", claimsLog.GetLogLines())
		}
	}

	if len(failedNodes) > 0 {
		tnf.ClaimFilePrintf("Nodes have been found with altered boot params: %v", failedNodes)
		ginkgo.Fail("Nodes have been found with altered boot params.")
	}
}

func testSysctlConfigs(env *provider.TestEnvironment) {
	badContainers := []string{}

	alreadyCheckedNodes := map[string]bool{}
	for _, cut := range env.Containers {
		if alreadyCheckedNodes[cut.NodeName] {
			continue
		}
		alreadyCheckedNodes[cut.NodeName] = true

		debugPod := env.DebugPods[cut.NodeName]
		if debugPod == nil {
			ginkgo.Fail(fmt.Sprintf("Debug pod not found on Node: %s", cut.NodeName))
		}

		sysctlSettings, err := sysctlconfig.GetSysctlSettings(env, cut.NodeName)
		if err != nil {
			tnf.ClaimFilePrintf("Could not get sysctl settings for node %s, error: %v", cut.NodeName, err)
			badContainers = append(badContainers, cut.String())
			continue
		}

		mcKernelArgumentsMap := bootparams.GetMcKernelArguments(env, cut.NodeName)

		for key, sysctlConfigVal := range sysctlSettings {
			if mcVal, ok := mcKernelArgumentsMap[key]; ok {
				if mcVal != sysctlConfigVal {
					tnf.ClaimFilePrintf(fmt.Sprintf("Kernel config mismatch in node %s for %s (sysctl value: %s, machine config value: %s)",
						cut.NodeName, key, sysctlConfigVal, mcVal))
					badContainers = append(badContainers, cut.String())
				}
			}
		}
	}

	testhelper.AddTestResultLog("Non-compliant", badContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
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

//nolint:funlen,gocyclo
func testNodeOperatingSystemStatus(env *provider.TestEnvironment) {
	ginkgo.By("Testing the control-plane and workers in the cluster for Operating System compatibility")

	logrus.Debug(fmt.Sprintf("There are %d nodes to process for Operating System compatibility.", len(env.Nodes)))

	failedControlPlaneNodes := []string{}
	failedWorkerNodes := []string{}
	for _, node := range env.Nodes {
		// Get the OSImage which should tell us what version of operating system the node is running.
		logrus.Debug(fmt.Sprintf("Node %s is running operating system: %s", node.Data.Name, node.Data.Status.NodeInfo.OSImage))

		// Control plane nodes must be RHCOS.
		// Per the release notes from OCP documentation:
		// "You must use RHCOS machines for the control plane, and you can use either RHCOS or RHEL for compute machines."
		if node.IsMasterNode() && !node.IsRHCOS() {
			tnf.ClaimFilePrintf("Master Node %s has been found to be running an incompatible operating system: %s", node.Data.Name, node.Data.Status.NodeInfo.OSImage)
			failedControlPlaneNodes = append(failedControlPlaneNodes, node.Data.Name)
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
					continue
				}
			} else if node.IsRHEL() {
				// Get the short version from the node
				shortVersion, err := node.GetRHELVersion()
				if err != nil {
					tnf.ClaimFilePrintf("Node %s failed to gather RHEL version. Error: %v", node.Data.Name, err)
					failedWorkerNodes = append(failedWorkerNodes, node.Data.Name)
					continue
				}

				// If the node's RHEL version and the OpenShift version are not compatible, the node fails.
				logrus.Debugf("Comparing RHEL shortVersion: %s to openshiftVersion: %s", shortVersion, env.OpenshiftVersion)
				if !compatibility.IsRHELCompatible(shortVersion, env.OpenshiftVersion) {
					tnf.ClaimFilePrintf("Node %s has been found to be running an incompatible version of RHEL: %s", node.Data.Name, shortVersion)
					failedWorkerNodes = append(failedWorkerNodes, node.Data.Name)
					continue
				}
			} else {
				tnf.ClaimFilePrintf("Node %s has been found to be running an incompatible operating system", node.Data.Name)
				failedWorkerNodes = append(failedWorkerNodes, node.Data.Name)
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

	// Write the combined failure string if there are any failures
	if len(failedControlPlaneNodes) > 0 || len(failedWorkerNodes) > 0 {
		ginkgo.Fail(b.String())
	}
}

func testPodHugePagesSize(env *provider.TestEnvironment, size string) {
	var badPods []*provider.Pod
	for _, put := range env.GetHugepagesPods() {
		result := put.CheckResourceHugePagesSize(size)
		if !result {
			badPods = append(badPods, put)
		}
	}
	testhelper.AddTestResultLog("Non-compliant", badPods, tnf.ClaimFilePrintf, ginkgo.Fail)
}
