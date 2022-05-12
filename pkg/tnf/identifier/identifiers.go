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

package identifier

import (
	"github.com/test-network-function/cnf-certification-test/pkg/tnf/dependencies"
)

const (
	commandIdentifierURL                  = urlTests + "/command"
	nodeselectorIdentifierURL             = urlTests + "/nodeselector"
	hostnameIdentifierURL                 = urlTests + "/hostname"
	ipAddrIdentifierURL                   = urlTests + "/ipaddr"
	nodesIdentifierURL                    = urlTests + "/nodes"
	operatorIdentifierURL                 = urlTests + "/operator"
	pingIdentifierURL                     = urlTests + "/ping"
	podIdentifierURL                      = urlTests + "/container/pod"
	versionIdentifierURL                  = urlTests + "/generic/version"
	containerIDURL                        = urlTests + "/generic/containerId"
	serviceAccountIdentifierURL           = urlTests + "/serviceaccount"
	roleBindingIdentifierURL              = urlTests + "/rolebinding"
	clusterRoleBindingIdentifierURL       = urlTests + "/clusterrolebinding"
	nodePortIdentifierURL                 = urlTests + "/nodeport"
	ImagePullPolicyIdentifierURL          = urlTests + "/imagepullpolicy"
	nodeNamesIdentifierURL                = urlTests + "/nodenames"
	nodeTaintedIdentifierURL              = urlTests + "/nodetainted"
	gracePeriodIdentifierURL              = urlTests + "/gracePeriod"
	hugepagesIdentifierURL                = urlTests + "/hugepages"
	nodehugepagesIdentifierURL            = urlTests + "/nodehugepages"
	podsetsIdentifierURL                  = urlTests + "/podsets"
	deploymentsnodesIdentifierURL         = urlTests + "/deploymentsnodes"
	deploymentsdrainIdentifierURL         = urlTests + "/deploymentsdrain"
	ownersIdentifierURL                   = urlTests + "/owners"
	cnfFsDiffURL                          = urlTests + "/generic/cnf_fs_diff"
	podnodenameIdentifierURL              = urlTests + "/podnodename"
	nodemcnameIdentifierURL               = urlTests + "/nodemcname"
	mckernelargumentsIdentifierURL        = urlTests + "/mckernelarguments"
	currentKernelCmdlineArgsIdentifierURL = urlTests + "/currentKernelCmdlineArgs"
	grubKernelCmdlineArgsIdentifierURL    = urlTests + "/grubKernelCmdlineArgs"
	sysctlConfigFilesListIdentifierURL    = urlTests + "/sysctlConfigFilesList"
	sysctlAllConfigsArgsURL               = urlTests + "/sysctlAllConfigsArgs"
	readRemoteFileIdentifierURL           = urlTests + "/readRemoteFile"
	uncordonNodeIdentifierURL             = urlTests + "/node/uncordon"
	checkSubscriptionIdentifierURL        = urlTests + "/operator/check-subscription"
	nodeDebugIdentifierURL                = urlTests + "/nodedebug"
	loggingIdentifierURL                  = urlTests + "/logging"
	podantiaffinityIdentifierURL          = urlTests + "/testPodHighAvailability"
	shutdownIdentifierURL                 = urlTests + "/shutdown"
	scalingIdentifierURL                  = urlTests + "/scaling"
	csiDriverIdentifierURL                = urlTests + "/csiDriver"
	clusterVersionIdentifierURL           = urlTests + "/clusterVersion"
	crdStatusExistenceIdentifierURL       = urlTests + "/crdStatusExistence"
	daemonSetIdentifierURL                = urlTests + "/daemonset"
	automountserviceIdentifierURL         = urlTests + "/automountservice"
	versionOne                            = "v1.0.0"
)

const (
	// Normative is the test type used for a test that returns normative results.
	Normative = "normative"
	// Informative is the test type used for a test that returns informative results.
	Informative = "informative"
)

// TestCatalogEntry is a container for required test facets.
type TestCatalogEntry struct {

	// Identifier is the unique test identifier.
	Identifier Identifier `json:"identifier" yaml:"identifier"`

	// Description is a helpful description of the purpose of the test.
	Description string `json:"description" yaml:"description"`

	// Type is the type of the test (i.e., normative).
	Type string `json:"type" yaml:"type"`

	// IntrusionSettings is used to specify test intrusion behavior into a target system.
	IntrusionSettings IntrusionSettings `json:"intrusionSettings" yaml:"intrusionSettings"`

	// BinaryDependencies tracks the needed binaries to complete tests, such as `ping`.
	BinaryDependencies []string `json:"binaryDependencies" yaml:"binaryDependencies"`
}

// IntrusionSettings is used to specify test intrusion behavior into a target system.
type IntrusionSettings struct {
	// ModifiesSystem records whether the test makes changes to target systems.
	ModifiesSystem bool `json:"modifiesSystem" yaml:"modifiesSystem"`

	// ModificationIsPersistent records whether the test makes a modification to the system that persists after the test
	// completes.  This is not always negative, and could involve something like setting up a tunnel that is used in
	// future tests.
	ModificationIsPersistent bool `json:"modificationIsPersistent" yaml:"modificationIsPersistent"`
}

// Catalog is the test catalog.
var Catalog = map[string]TestCatalogEntry{
	commandIdentifierURL: {
		Identifier:  CommandIdentifier,
		Description: "A generic test used with any command and would match any output. The caller is responsible for interpreting the output and extracting data from it.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{},
	},
	hostnameIdentifierURL: {
		Identifier:  HostnameIdentifier,
		Description: "A generic test used to check the hostname of a target machine/container.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.HostnameBinaryName,
		},
	},
	ipAddrIdentifierURL: {
		Identifier:  IPAddrIdentifier,
		Description: "A generic test used to derive the default network interface IP address of a target container.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.IPBinaryName,
		},
	},
	nodesIdentifierURL: {
		Identifier:  NodesIdentifier,
		Description: "Polls the state of the OpenShift cluster nodes using \"oc get nodes -o json\".",
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
		},
	},
	operatorIdentifierURL: {
		Identifier: OperatorIdentifier,
		Description: "An operator-specific test used to exercise the behavior of a given operator.  In the current " +
			"offering, we check if the operator ClusterServiceVersion (CSV) is installed properly.  A CSV is a YAML " +
			"manifest created from Operator metadata that assists the Operator Lifecycle Manager (OLM) in running " +
			"the Operator.",
		Type: Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.JqBinaryName,
			dependencies.OcBinaryName,
		},
	},
	pingIdentifierURL: {
		Identifier:  PingIdentifier,
		Description: "A generic test used to test ICMP connectivity from a source machine/container to a target destination.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.PingBinaryName,
		},
	},
	podIdentifierURL: {
		Identifier:  PodIdentifier,
		Description: "A container-specific test suite used to verify various aspects of the underlying container.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.JqBinaryName,
			dependencies.OcBinaryName,
		},
	},
	versionIdentifierURL: {
		Identifier:  VersionIdentifier,
		Description: "A generic test used to determine if a target container/machine is based on RHEL.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.CatBinaryName,
		},
	},
	cnfFsDiffURL: {
		Identifier:  CnfFsDiffIdentifier,
		Description: "A test used to check if there were no installation during container runtime",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.GrepBinaryName,
			dependencies.CutBinaryName,
		},
	},
	serviceAccountIdentifierURL: {
		Identifier:  ServiceAccountIdentifier,
		Description: "A generic test used to extract the CNF pod's ServiceAccount name.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.GrepBinaryName,
			dependencies.CutBinaryName,
		},
	},
	containerIDURL: {
		Identifier:  ContainerIDIdentifier,
		Description: "A test used to check what is the id of the crio generated container this command is run from",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.CatBinaryName,
		},
	},
	roleBindingIdentifierURL: {
		Identifier:  RoleBindingIdentifier,
		Description: "A generic test used to test RoleBindings of CNF pod's ServiceAccount.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.CatBinaryName,
			dependencies.OcBinaryName,
		},
	},
	clusterRoleBindingIdentifierURL: {
		Identifier:  ClusterRoleBindingIdentifier,
		Description: "A generic test used to test ClusterRoleBindings of CNF pod's ServiceAccount.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
		},
	},
	nodePortIdentifierURL: {
		Identifier:  NodePortIdentifier,
		Description: "A generic test used to test services of CNF pod's namespace.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
			dependencies.GrepBinaryName,
		},
	},
	ImagePullPolicyIdentifierURL: {
		Identifier:  ImagePullPolicyIdentifier,
		Description: "A generic test used to get Image Pull Policy type.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
		},
	},
	nodeNamesIdentifierURL: {
		Identifier:  NodeNamesIdentifier,
		Description: "A generic test used to get node names",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
		},
	},
	nodeTaintedIdentifierURL: {
		Identifier:  NodeTaintedIdentifier,
		Description: "A generic test used to test whether node is tainted",
		Type:        Informative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
			dependencies.CatBinaryName,
			dependencies.EchoBinaryName,
		},
	},
	gracePeriodIdentifierURL: {
		Identifier:  GracePeriodIdentifier,
		Description: "A generic test used to extract the CNF pod's terminationGracePeriod.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.GrepBinaryName,
			dependencies.CutBinaryName,
		},
	},
	hugepagesIdentifierURL: {
		Identifier:  HugepagesIdentifier,
		Description: "A generic test used to read cluster's hugepages configuration",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.GrepBinaryName,
			dependencies.CutBinaryName,
			dependencies.OcBinaryName,
			dependencies.GrepBinaryName,
		},
	},
	nodehugepagesIdentifierURL: {
		Identifier:  NodeHugepagesIdentifier,
		Description: "A generic test used to verify a node's hugepages configuration",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
			dependencies.GrepBinaryName,
		},
	},
	podsetsIdentifierURL: {
		Identifier:  PodSetsIdentifier,
		Description: "A generic test used to read namespace's deployments/statefulsets",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
		},
	},
	podnodenameIdentifierURL: {
		Identifier:  PodNodeNameIdentifier,
		Description: "A generic test used to get a pod's node",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
		},
	},
	deploymentsnodesIdentifierURL: {
		Identifier:  DeploymentsNodesIdentifier,
		Description: "A generic test used to read node names of pods owned by deployments in namespace",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
			dependencies.GrepBinaryName,
		},
	},
	nodeselectorIdentifierURL: {
		Identifier:  NodeSelectorIdentifier,
		Description: "A generic test used to verify a pod's nodeSelector and nodeAffinity configuration",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
			dependencies.GrepBinaryName,
		},
	},
	nodemcnameIdentifierURL: {
		Identifier:  NodeMcNameIdentifier,
		Description: "A generic test used to get a node's current mc",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
			dependencies.GrepBinaryName,
		},
	},
	deploymentsdrainIdentifierURL: {
		Identifier:  DeploymentsNodesIdentifier,
		Description: "A generic test used to drain node from its deployment pods",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           true,
			ModificationIsPersistent: true,
		},
		BinaryDependencies: []string{
			dependencies.JqBinaryName,
			dependencies.EchoBinaryName,
		},
	},
	mckernelargumentsIdentifierURL: {
		Identifier:  McKernelArgumentsIdentifier,
		Description: "A generic test used to get an mc's kernel arguments",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
			dependencies.JqBinaryName,
			dependencies.EchoBinaryName,
		},
	},
	ownersIdentifierURL: {
		Identifier:  OwnersIdentifier,
		Description: "A generic test used to verify pod is managed by a ReplicaSet/StatefulSet",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.CatBinaryName,
		},
	},
	currentKernelCmdlineArgsIdentifierURL: {
		Identifier:  CurrentKernelCmdlineArgsURLIdentifier,
		Description: "A generic test used to get node's /proc/cmdline",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.CatBinaryName,
		},
	},
	grubKernelCmdlineArgsIdentifierURL: {
		Identifier:  GrubKernelCmdlineArgsURLIdentifier,
		Description: "A generic test used to get node's next boot kernel args",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.LsBinaryName,
			dependencies.SortBinaryName,
			dependencies.HeadBinaryName,
			dependencies.CutBinaryName,
			dependencies.OcBinaryName,
		},
	},
	sysctlConfigFilesListIdentifierURL: {
		Identifier:  SysctlConfigFilesListURLIdentifier,
		Description: "A generic test used to get node's list of sysctl config files",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.CatBinaryName,
		},
	},
	readRemoteFileIdentifierURL: {
		Identifier:  ReadRemoteFileURLIdentifier,
		Description: "A generic test used to read a specified file at a specified node",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.EchoBinaryName,
		},
	},
	uncordonNodeIdentifierURL: {
		Identifier:  UncordonNodeURLIdentifier,
		Description: "A generic test used to uncordon a node",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           true,
			ModificationIsPersistent: true,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
		},
	},
	checkSubscriptionIdentifierURL: {
		Identifier:  CheckSubscriptionURLIdentifier,
		Description: "A test used to check the subscription of a given operator",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
		},
	},
	nodeDebugIdentifierURL: {
		Identifier:  NodeDebugIdentifier,
		Description: "A generic test used to execute a command in a node",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
			dependencies.EchoBinaryName,
		},
	},
	loggingIdentifierURL: {
		Identifier:  LoggingURLIdentifier,
		Description: "A test used to check logs are redirected to stderr/stdout",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
			dependencies.WcBinaryName,
		},
	},
	podantiaffinityIdentifierURL: {
		Identifier:  PodAntiAffinityIdentifier,
		Description: "A generic test used to check pod's replica and podAntiAffinity configuration in high availability mode",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
		},
	},
	shutdownIdentifierURL: {
		Identifier:  ShutdownURLIdentifier,
		Description: "A test used to check pre-stop lifecycle is defined",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
		},
	},
	sysctlAllConfigsArgsURL: {
		Identifier:  SysctlAllConfigsArgsIdentifier,
		Description: "A test used to find all sysctl configuration args",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.SysctlBinaryName,
		},
	},
	scalingIdentifierURL: {
		Identifier: ScalingIdentifier,
		Description: "A test to check the deployments scale in/out. The tests issues the oc scale " +
			"command on a deployment for a given number of replicas and checks whether the command output " +
			"is valid.",
		Type: Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           true,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
		},
	},
	csiDriverIdentifierURL: {
		Identifier:  CSIDriverIdentifier,
		Description: "extracts the csi driver info in the cluster",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
		},
	},
	clusterVersionIdentifierURL: {
		Identifier:  ClusterVersionIdentifier,
		Description: "Extracts OCP versions from the cluster",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
		},
	},
	crdStatusExistenceIdentifierURL: {
		Identifier:  CrdStatusExistenceIdentifier,
		Description: "Checks whether a give CRD has status subresource specification.",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
			dependencies.JqBinaryName,
		},
	},
	daemonSetIdentifierURL: {
		Identifier:  DaemonSetIdentifier,
		Description: "check whether a given daemonset was deployed successfully",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
		},
	},
	automountserviceIdentifierURL: {
		Identifier:  AutomountServiceIdentifier,
		Description: "check if automount service account token is set to false",
		Type:        Normative,
		IntrusionSettings: IntrusionSettings{
			ModifiesSystem:           false,
			ModificationIsPersistent: false,
		},
		BinaryDependencies: []string{
			dependencies.OcBinaryName,
		},
	},
}

// TestIDBaseDomain is the BaseDomain for the IDs of test cases building blocks
var TestIDBaseDomain = urlTests

// CommandIdentifier is  the Identifier used to represent the generic command test case.
var CommandIdentifier = Identifier{
	URL:             commandIdentifierURL,
	SemanticVersion: versionOne,
}

// HostnameIdentifier is the Identifier used to represent the generic hostname test case.
var HostnameIdentifier = Identifier{
	URL:             hostnameIdentifierURL,
	SemanticVersion: versionOne,
}

// IPAddrIdentifier is the Identifier used to represent the generic IP Addr test case.
var IPAddrIdentifier = Identifier{
	URL:             ipAddrIdentifierURL,
	SemanticVersion: versionOne,
}

// NodesIdentifier is the Identifier used to represent the nodes test case.
var NodesIdentifier = Identifier{
	URL:             nodesIdentifierURL,
	SemanticVersion: versionOne,
}

// OperatorIdentifier is the Identifier used to represent the operator-specific test suite.
var OperatorIdentifier = Identifier{
	URL:             operatorIdentifierURL,
	SemanticVersion: versionOne,
}

// PingIdentifier is the Identifier used to represent the generic Ping test.
var PingIdentifier = Identifier{
	URL:             pingIdentifierURL,
	SemanticVersion: versionOne,
}

// PodIdentifier is the Identifier used to represent the container-specific test suite.
var PodIdentifier = Identifier{
	URL:             podIdentifierURL,
	SemanticVersion: versionOne,
}

// VersionIdentifier is the Identifier used to represent the generic container base image test.
var VersionIdentifier = Identifier{
	URL:             versionIdentifierURL,
	SemanticVersion: versionOne,
}

// CnfFsDiffIdentifier is the Identifier used to represent the generic cnf_fs_diff test.
var CnfFsDiffIdentifier = Identifier{
	URL:             cnfFsDiffURL,
	SemanticVersion: versionOne,
}

// ContainerIDIdentifier is the Identifier used to represent the generic cnf_fs_diff test.
var ContainerIDIdentifier = Identifier{
	URL:             containerIDURL,
	SemanticVersion: versionOne,
}

// ServiceAccountIdentifier is the Identifier used to represent the generic serviceAccount test.
var ServiceAccountIdentifier = Identifier{
	URL:             serviceAccountIdentifierURL,
	SemanticVersion: versionOne,
}

// RoleBindingIdentifier is the Identifier used to represent the generic roleBinding test.
var RoleBindingIdentifier = Identifier{
	URL:             roleBindingIdentifierURL,
	SemanticVersion: versionOne,
}

// ClusterRoleBindingIdentifier is the Identifier used to represent the generic clusterRoleBinding test.
var ClusterRoleBindingIdentifier = Identifier{
	URL:             clusterRoleBindingIdentifierURL,
	SemanticVersion: versionOne,
}

// NodePortIdentifier is the Identifier used to represent the generic NodePort test.
var NodePortIdentifier = Identifier{
	URL:             nodePortIdentifierURL,
	SemanticVersion: versionOne,
}

var ImagePullPolicyIdentifier = Identifier{
	URL:             ImagePullPolicyIdentifierURL,
	SemanticVersion: versionOne,
}

// NodeNamesIdentifier is the Identifier used to represent the generic NodeNames test.
var NodeNamesIdentifier = Identifier{
	URL:             nodeNamesIdentifierURL,
	SemanticVersion: versionOne,
}

// NodeTaintedIdentifier is the Identifier used to represent the generic NodeTainted test.
var NodeTaintedIdentifier = Identifier{
	URL:             nodeTaintedIdentifierURL,
	SemanticVersion: versionOne,
}

// GracePeriodIdentifier is the Identifier used to represent the generic GracePeriod test.
var GracePeriodIdentifier = Identifier{
	URL:             gracePeriodIdentifierURL,
	SemanticVersion: versionOne,
}

// HugepagesIdentifier is the Identifier used to represent the generic Hugepages test.
var HugepagesIdentifier = Identifier{
	URL:             hugepagesIdentifierURL,
	SemanticVersion: versionOne,
}

// NodeHugepagesIdentifier is the Identifier used to represent the generic NodeHugepages test.
var NodeHugepagesIdentifier = Identifier{
	URL:             nodehugepagesIdentifierURL,
	SemanticVersion: versionOne,
}

// PodSetsIdentifier is the Identifier used to represent the generic PodSets test.
var PodSetsIdentifier = Identifier{
	URL:             podsetsIdentifierURL,
	SemanticVersion: versionOne,
}

// DeploymentsNodesIdentifier is the Identifier used to represent the generic DeploymentsNodes test.
var DeploymentsNodesIdentifier = Identifier{
	URL:             deploymentsnodesIdentifierURL,
	SemanticVersion: versionOne,
}

// DeploymentsDrainIdentifier is the Identifier used to represent the generic DeploymentsDrain test.
var DeploymentsDrainIdentifier = Identifier{
	URL:             deploymentsdrainIdentifierURL,
	SemanticVersion: versionOne,
}

// OwnersIdentifier is the Identifier used to represent the generic Owners test.
var OwnersIdentifier = Identifier{
	URL:             ownersIdentifierURL,
	SemanticVersion: versionOne,
}

// NodeSelectorIdentifier is the Identifier used to represent the generic NodeSelector test.
var NodeSelectorIdentifier = Identifier{
	URL:             nodeselectorIdentifierURL,
	SemanticVersion: versionOne,
}

// PodNodeNameIdentifier is the Identifier used to represent the generic PodNodeName test.
var PodNodeNameIdentifier = Identifier{
	URL:             podnodenameIdentifierURL,
	SemanticVersion: versionOne,
}

// NodeMcNameIdentifier is the Identifier used to represent the generic NodeMcName test.
var NodeMcNameIdentifier = Identifier{
	URL:             nodemcnameIdentifierURL,
	SemanticVersion: versionOne,
}

// McKernelArgumentsIdentifier is the Identifier used to represent the generic McKernelArguments test.
var McKernelArgumentsIdentifier = Identifier{
	URL:             mckernelargumentsIdentifierURL,
	SemanticVersion: versionOne,
}

// CurrentKernelCmdlineArgsURLIdentifier is the Identifier used to represent the generic getCurrentKernelCmdlineArgs test.
var CurrentKernelCmdlineArgsURLIdentifier = Identifier{
	URL:             currentKernelCmdlineArgsIdentifierURL,
	SemanticVersion: versionOne,
}

// GrubKernelCmdlineArgsURLIdentifier is the Identifier used to represent the generic getCurrentKernelCmdlineArgs test.
var GrubKernelCmdlineArgsURLIdentifier = Identifier{
	URL:             grubKernelCmdlineArgsIdentifierURL,
	SemanticVersion: versionOne,
}

// SysctlConfigFilesListURLIdentifier is the Identifier used to represent the generic getCurrentKernelCmdlineArgs test.
var SysctlConfigFilesListURLIdentifier = Identifier{
	URL:             sysctlConfigFilesListIdentifierURL,
	SemanticVersion: versionOne,
}

// ReadRemoteFileURLIdentifier is the Identifier used to represent the generic getCurrentKernelCmdlineArgs test.
var ReadRemoteFileURLIdentifier = Identifier{
	URL:             readRemoteFileIdentifierURL,
	SemanticVersion: versionOne,
}

// UncordonNodeURLIdentifier is the Identifier used to represent a test that uncordons a node.
var UncordonNodeURLIdentifier = Identifier{
	URL:             uncordonNodeIdentifierURL,
	SemanticVersion: versionOne,
}

// CheckSubscriptionURLIdentifier is the Identifier used to represent a test that checks the subscription of an operator.
var CheckSubscriptionURLIdentifier = Identifier{
	URL:             checkSubscriptionIdentifierURL,
	SemanticVersion: versionOne,
}

// NodeDebugIdentifier is the Identifier used to represent the generic NodeDebug test.
var NodeDebugIdentifier = Identifier{
	URL:             nodeDebugIdentifierURL,
	SemanticVersion: versionOne,
}

// LoggingURLIdentifier is the Identifier used to represent a test that checks if the stdout/stderr is used
var LoggingURLIdentifier = Identifier{
	URL:             loggingIdentifierURL,
	SemanticVersion: versionOne,
}

// PodAntiAffinityIdentifier is the Identifier used to represent the generic podAffinity test.
var PodAntiAffinityIdentifier = Identifier{
	URL:             podantiaffinityIdentifierURL,
	SemanticVersion: versionOne,
}

// ShutdownURLIdentifier is the Identifier used to represent a test that checks if pre-stop lifecyle is defined
var ShutdownURLIdentifier = Identifier{
	URL:             shutdownIdentifierURL,
	SemanticVersion: versionOne,
}

// SysctlAllConfigsArgsIdentifier is the Identifier used to represent a test that checks all args in all sysctl conf files ordered
// in the same way as they are loaded by the os
var SysctlAllConfigsArgsIdentifier = Identifier{
	URL:             sysctlAllConfigsArgsURL,
	SemanticVersion: versionOne,
}

// ScalingIdentifier is the Identifier used to represent a test that checks deployments scale in/out
var ScalingIdentifier = Identifier{
	URL:             scalingIdentifierURL,
	SemanticVersion: versionOne,
}

// CSIDriverIdentifier is the Identifier used to represent the CSI driver test case.
var CSIDriverIdentifier = Identifier{
	URL:             csiDriverIdentifierURL,
	SemanticVersion: versionOne,
}

// ClusterVersionIdentifier is the Identifier used to represent the OCP versions test case.
var ClusterVersionIdentifier = Identifier{
	URL:             clusterVersionIdentifierURL,
	SemanticVersion: versionOne,
}

// CrdStatusExistenceIdentifier is the Identifier used to represent the generic test for CRD status spec existence.
var CrdStatusExistenceIdentifier = Identifier{
	URL:             crdStatusExistenceIdentifierURL,
	SemanticVersion: versionOne,
}

var DaemonSetIdentifier = Identifier{
	URL:             daemonSetIdentifierURL,
	SemanticVersion: versionOne,
}
var AutomountServiceIdentifier = Identifier{
	URL:             automountserviceIdentifierURL,
	SemanticVersion: versionOne,
}
