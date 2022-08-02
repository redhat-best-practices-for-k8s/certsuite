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

const (
	istioContainerName = "istio-proxy"
)

//
// All actual test code belongs below here.  Utilities belong above.
//
var _ = ginkgo.Describe(common.PlatformAlterationTestKey, func() {
	logrus.Debugf("Entering %s suite", common.PlatformAlterationTestKey)
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})
	ginkgo.ReportAfterEach(results.RecordResult)

	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestUnalteredBaseImageIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		if provider.IsOCPCluster() {
			testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
			testContainersFsDiff(&env)
		} else {
			ginkgo.Skip(" non ocp cluster ")
		}
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestNonTaintedNodeKernelsIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.DebugPods)
		testTainted(&env, nodetainted.NewNodeTaintedTester(clientsholder.GetClientsHolder())) // minikube tainted kernels are allowed via config
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestIsRedHatReleaseIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		testIsRedHatRelease(&env)
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestIsSELinuxEnforcingIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		if provider.IsOCPCluster() {
			testhelper.SkipIfEmptyAny(ginkgo.Skip, env.DebugPods)
			testIsSELinuxEnforcing(&env)
		} else {
			ginkgo.Skip(" non ocp cluster ")
		}
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestHugepagesNotManuallyManipulated)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		if provider.IsOCPCluster() {
			testhelper.SkipIfEmptyAny(ginkgo.Skip, env.DebugPods)
			testHugepages(&env)
		} else {
			ginkgo.Skip(" non ocp cluster ")
		}
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestUnalteredStartupBootParamsIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		if provider.IsOCPCluster() {
			testhelper.SkipIfEmptyAny(ginkgo.Skip, env.DebugPods)
			testUnalteredBootParams(&env)
		} else {
			ginkgo.Skip(" non ocp cluster ")
		}
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestSysctlConfigsIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		if provider.IsOCPCluster() {
			testhelper.SkipIfEmptyAny(ginkgo.Skip, env.DebugPods)
			testSysctlConfigs(&env)
		} else {
			ginkgo.Skip(" non ocp cluster ")
		}
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestServiceMeshIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestServiceMesh(&env)
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestOCPLifecycleIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		if provider.IsOCPCluster() {
			testOCPStatus(&env)
		}
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestNodeOperatingSystemIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		if provider.IsOCPCluster() {
			testNodeOperatingSystemStatus(&env)
		}
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
			if cut.Status.Name == istioContainerName {
				tnf.ClaimFilePrintf("Istio proxy container found on %s", put)
				istioProxyFound = true
				break
			}
		}
		if !istioProxyFound {
			badPods = append(badPods, put.String())
		}
	}
	if n := len(badPods); n > 0 {
		tnf.ClaimFilePrintf("Pods not using service mesh: %v", badPods)
		ginkgo.Fail(fmt.Sprintf("Found %d pods that do not use service mesh.", n))
	}
}

// testContainersFsDiff test that all CUT didn't install new packages are starting
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
		fsDiffTester.RunTest(clientsholder.Context{
			Namespace:     debugPod.Namespace,
			Podname:       debugPod.Name,
			Containername: debugPod.Spec.Containers[0].Name,
		}, cut.UID)
		switch fsDiffTester.GetResults() {
		case testhelper.SUCCESS:
			continue
		case testhelper.FAILURE:
			tnf.ClaimFilePrintf("%s - changed folders: %v, deleted folders: %v", cut, fsDiffTester.ChangedFolders, fsDiffTester.DeletedFolders)
			badContainers = append(badContainers, cut.Data.Name)
		case testhelper.ERROR:
			tnf.ClaimFilePrintf("%s - error while running fs-diff: %v: ", cut, fsDiffTester.Error)
			errContainers = append(errContainers, cut.Data.Name)
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
func testTainted(env *provider.TestEnvironment, testerFuncs nodetainted.TaintedFuncs) {
	var taintedNodes []string
	var errNodes []string

	// Loop through the debug pods that are tied to each node.
	for _, dp := range env.DebugPods {
		// Create OCP context to pass around
		ocpContext := clientsholder.Context{
			Namespace:     dp.Namespace,
			Podname:       dp.Name,
			Containername: dp.Spec.Containers[0].Name,
		}

		// Get the taint information from the node kernel
		taintInfo, err := testerFuncs.GetKernelTaintInfo(ocpContext)
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

		// Count how many taints come from `module was loaded` taints versus `other`
		logrus.Debug("Checking for 'module was loaded' taints")
		logrus.Debug("individualTaints", individualTaints)
		moduleTaintsFound := false
		otherTaintsFound := false

		otherTaints := []string{}
		for _, it := range individualTaints {
			if strings.Contains(it, `module was loaded`) {
				moduleTaintsFound = true
			} else {
				otherTaintsFound = true
				otherTaints = append(otherTaints, it)
			}
		}

		if otherTaintsFound {
			nodeTaintsAccepted = false

			// Surface more information about tainted kernel failures that have nothing to do with modules.
			tnf.ClaimFilePrintf("Please note that taints other than 'module was loaded' were found on node %s.", dp.Spec.NodeName)
			logrus.Debugf("Please note that taints other than 'module was loaded' were found on node %s.", dp.Spec.NodeName)
			for _, ot := range otherTaints {
				tnf.ClaimFilePrintf("Taint causing failure: %s on node: %s", ot, dp.Spec.NodeName)
				logrus.Debugf("Taint causing failure: %s on node: %s", ot, dp.Spec.NodeName)
			}
		} else if moduleTaintsFound {
			// Retrieve the modules from the node (via the debug pod)
			modules := testerFuncs.GetModulesFromNode(ocpContext)
			logrus.Debugf("Got the modules from node %s: %v", dp.Name, modules)

			// Retrieve all of the out of tree modules.
			taintedModules := testerFuncs.GetOutOfTreeModules(modules, ocpContext)
			logrus.Debug("Collected all of the tainted modules: ", taintedModules)
			logrus.Debug("Modules allowed via configuration: ", env.Config.AcceptedKernelTaints)

			// Looks through the accepted taints listed in the tnf-config file.
			// If all of the tainted modules show up in the configuration file, do not fail the test.
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
	if len(taintedNodes) > 0 {
		tnf.ClaimFilePrintf("Nodes have been found to be tainted: %v", taintedNodes)
		ginkgo.Fail("Nodes have been found to be tainted.")
	}

	if len(errNodes) > 0 {
		tnf.ClaimFilePrintf("Nodes have been found to be tainted: %v", taintedNodes)
		ginkgo.Fail("Nodes have been found to be tainted.")
	}
}

func testIsRedHatRelease(env *provider.TestEnvironment) {
	ginkgo.By("should report a proper Red Hat version")
	failedContainers := []string{}
	for _, cut := range env.Containers {
		ginkgo.By(fmt.Sprintf("%s is checked for Red Hat version", cut))
		baseImageTester := isredhat.NewBaseImageTester(clientsholder.GetClientsHolder(), clientsholder.Context{
			Namespace:     cut.Namespace,
			Podname:       cut.Podname,
			Containername: cut.Data.Name,
		})

		result, err := baseImageTester.TestContainerIsRedHatRelease()
		if err != nil {
			logrus.Error("failed to collect release information from container: ", err)
		}
		if !result {
			failedContainers = append(failedContainers, cut.Namespace+"/"+cut.Podname+"/"+cut.Data.Name)
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
		ctx := clientsholder.Context{Namespace: debugPod.Namespace, Podname: debugPod.Name, Containername: debugPod.Spec.Containers[0].Name}
		outStr, errStr, err := o.ExecCommandContainer(ctx, getenforceCommand)
		if err != nil || errStr != "" {
			logrus.Errorf("Failed to execute command %s in debug %s, errStr: %s, err: %s", getenforceCommand, debugPod.String(), errStr, err)
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

		hpTester, err := hugepages.NewTester(&node, env.DebugPods[node.Data.Name])
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

//nolint:funlen
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
			tnf.ClaimFilePrintf("Could not get sysctl settings for node %s, error: %s", cut.NodeName, err)
			badContainers = append(badContainers, cut.String())
			continue
		}

		mcKernelArgumentsMap, err := bootparams.GetMcKernelArguments(env, cut.NodeName)
		if err != nil {
			tnf.ClaimFilePrintf("Failed to get the machine config kernel arguments for node %s, error: %s", cut.NodeName, err)
			badContainers = append(badContainers, cut.String())
			continue
		}

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

	if n := len(badContainers); n > 0 {
		errMsg := fmt.Sprintf("Number of containers running of faulty nodes: %d", n)
		tnf.ClaimFilePrintf(errMsg)
		ginkgo.Fail(errMsg)
	}
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
					tnf.ClaimFilePrintf("Node %s failed to gather RHCOS version. Error: %s", node.Data.Name, err.Error())
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
					tnf.ClaimFilePrintf("Node %s failed to gather RHEL version. Error: %s", node.Data.Name, err.Error())
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
