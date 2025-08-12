// Copyright (C) 2020-2024 Red Hat, Inc.
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

package testhelper

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
)

const (
	SUCCESS = iota
	FAILURE
	ERROR
)

// ReportObject holds key-value pairs and a type for reporting purposes.
//
// It stores field keys and corresponding values in separate slices,
// allowing incremental construction of structured data.
// The ObjectType string indicates the kind of entity being reported, such as
// a container process or a cluster operator. Methods on the struct enable
// adding fields and setting the type, facilitating fluent configuration of
// report objects used throughout the test helper package.
type ReportObject struct {
	ObjectType         string
	ObjectFieldsKeys   []string
	ObjectFieldsValues []string
}

// FailureReasonOut represents the outcome of a failure reason check.
//
// It contains slices of report objects that either complied with or failed to meet
// the specified criteria. The CompliantObjectsOut slice holds pointers to
// ReportObject instances that passed the check, while NonCompliantObjectsOut
// holds those that did not. This struct is used by test helpers to summarize
// which objects satisfied a condition and which did not.
type FailureReasonOut struct {
	CompliantObjectsOut    []*ReportObject
	NonCompliantObjectsOut []*ReportObject
}

// Equal compares two slices of ReportObject pointers for equality.
//
// It first checks that the slices have the same length. If they do not, it
// returns false. Then it iterates over the elements and uses reflect.DeepEqual
// to compare each corresponding pair. If any comparison fails, the function
// immediately returns false. If all pairs are deeply equal, the function
// returns true.
func Equal(p, other []*ReportObject) bool {
	if len(p) != len(other) {
		return false
	}
	for i := 0; i < len(p); i++ {
		if p[i] == nil && other[i] == nil {
			continue
		}
		if p[i] == nil || other[i] == nil {
			return false
		}
		if !reflect.DeepEqual(*p[i], *other[i]) {
			return false
		}
	}
	return true
}

// FailureReasonOutTestString returns a string representation of a FailureReasonOut struct.
//
// It formats the struct fields into a readable string, calling ReportObjectTestStringPointer on any nested pointers to produce their test strings and using Sprintf to assemble the final output.
func FailureReasonOutTestString(p FailureReasonOut) (out string) {
	out = "testhelper.FailureReasonOut{"
	out += fmt.Sprintf("CompliantObjectsOut: %s,", ReportObjectTestStringPointer(p.CompliantObjectsOut))
	out += fmt.Sprintf("NonCompliantObjectsOut: %s,", ReportObjectTestStringPointer(p.NonCompliantObjectsOut))
	out += "}"
	return out
}

// ReportObjectTestStringPointer converts a slice of *ReportObject to its string representation.
//
// It takes a slice of pointers to ReportObject and returns a formatted string
// showing the slice contents in the form
//   []*testhelper.ReportObject{&{...}, &{...}, ...}
//
// The function uses fmt.Sprintf to build the output.
func ReportObjectTestStringPointer(p []*ReportObject) (out string) {
	out = "[]*testhelper.ReportObject{"
	for _, p := range p {
		out += fmt.Sprintf("&%#v,", *p)
	}
	out += "}"
	return out
}

// ReportObjectTestString returns a string representation of the given slice of ReportObject.
//
// It takes a slice of pointers to ReportObject and formats each element using the %#v verb,
// concatenating them into a single string. The output is wrapped in square brackets
// and prefixed with "[]testhelper.ReportObject{", producing a readable summary
// suitable for debugging or test output. The function returns this formatted string.
func ReportObjectTestString(p []*ReportObject) (out string) {
	out = "[]testhelper.ReportObject{"
	for _, p := range p {
		out += fmt.Sprintf("%#v,", *p)
	}
	out += "}"
	return out
}

// Equal checks if two FailureReasonOut values are identical.
//
// It compares the CompliantObjectsOut and NonCompliantObjectsOut fields of both
// structs and returns true only when all corresponding elements match. If any
// difference is found, it returns false.
func (p FailureReasonOut) Equal(other FailureReasonOut) bool {
	return Equal(p.CompliantObjectsOut, other.CompliantObjectsOut) &&
		Equal(p.NonCompliantObjectsOut, other.NonCompliantObjectsOut)
}

// When adding new field types, please update the following:

const (
	Namespace                       = "Namespace"
	Name                            = "Name"
	PodName                         = "Pod Name"
	ContainerName                   = "Container Name"
	ProcessID                       = "Process ID"
	ProcessCommandLine              = "Process CommandLine"
	SchedulingPolicy                = "Scheduling Policy"
	SchedulingPriority              = "Scheduling Priority"
	ReasonForNonCompliance          = "Reason For Non Compliance"
	ReasonForCompliance             = "Reason For Compliance"
	Category                        = "Category"
	RoleBindingName                 = "Role Binding Name"
	ClusterRoleName                 = "Cluster Role Reference Name"
	RoleBindingNamespace            = "Role Binding Namespace"
	ServiceAccountName              = "Service Account Name"
	ServiceMode                     = "Service Type"
	ServiceName                     = "Service Name"
	ServiceIPVersion                = "Service IP Version"
	DeploymentName                  = "Deployment Name"
	StatefulSetName                 = "StatefulSet Name"
	PodDisruptionBudgetReference    = "Pod Disruption Budget Reference"
	CustomResourceDefinitionName    = "Custom Resource Definition Name"
	CustomResourceDefinitionVersion = "Custom Resource Definition Version"
	SCCCapability                   = "SCC Capability"
	Path                            = "Path"
	Repository                      = "Repository"
	ImageName                       = "Image Name"
	Version                         = "Version"
	OpenAPIV3Schema                 = "OpenAPIV3Schema"
	CrdVersion                      = "Operator CRD Version"
	OCPVersion                      = "OCP Version"
	OCPChannel                      = "OCP Channel"
	NodeSelector                    = "Node Selector"
	PersistentVolumeName            = "Persistent Volume Name"
	PersistentVolumeClaimName       = "Persistent Volume Claim Name"
	TolerationKey                   = "Toleration Key"
	TolerationEffect                = "Toleration Effect"
	StorageClassName                = "Storage Class Name"
	StorageClassProvisioner         = "Storage Class Provisioner"
	ChangedFolders                  = "Changed Folders"
	DeletedFolders                  = "Deleted Folders"
	TaintBit                        = "Taint Bit"
	TaintBitDescription             = "Taint Bit Description"
	TaintMask                       = "Taint Mask"
	ModuleName                      = "Module Name"
	Taints                          = "Taints"
	SysctlKey                       = "Sysctl Key"
	SysctlValue                     = "Sysctl Value"
	OSImage                         = "OS Image"
	ProbePodName                    = "Probe Pod Name"

	// ICMP tests
	NetworkName              = "Network Name"
	DestinationNamespace     = "Destination Namespace"
	DestinationPodName       = "Destination Pod Name"
	DestinationContainerName = "Destination Container Name"
	DestinationIP            = "Destination IP"
	SourceIP                 = "Source IP"

	// Rbac roles
	RoleName     = "Role Name"
	Group        = "Group"
	ResourceName = "Resource Name"
	Verb         = "Verb"

	// Listening ports
	PortNumber   = "Port Number"
	PortProtocol = "Port Protocol"

	// OLM
	SubscriptionName = "Subscription Name"
	OperatorPhase    = "Operator Phase"
	OperatorName     = "Operator Name"

	// Lists
	OperatorList = "Operator List"
)

// When adding new object types, please update the following:

const (
	UndefinedType                = "Undefined Type"
	CnfType                      = "Cnf"
	PodType                      = "Pod"
	HelmType                     = "Helm"
	OperatorType                 = "Operator"
	ClusterOperatorType          = "Cluster Operator"
	ContainerType                = "Container"
	CatalogSourceType            = "Catalog Source"
	ContainerImageType           = "Container Image"
	NodeType                     = "Node"
	OCPClusterType               = "OCP Cluster"
	OCPClusterVersionType        = "OCP Cluster Version"
	ContainerProcessType         = "ContainerProcess"
	ContainerCategory            = "ContainerCategory"
	ServiceType                  = "Service"
	DeploymentType               = "Deployment"
	StatefulSetType              = "StatefulSet"
	ICMPResultType               = "ICMP result"
	NetworkType                  = "Network"
	CustomResourceDefinitionType = "Custom Resource Definition"
	RoleRuleType                 = "Role Rule"
	RoleType                     = "Role"
	ListeningPortType            = "Listening Port"
	DeclaredPortType             = "Declared Port"
	ContainerPort                = "Container Port"
	HostPortType                 = "Host Port"
	HostPathType                 = "Host Path"
	HelmVersionType              = "Helm Version"
	Error                        = "Error"
	OperatorPermission           = "Operator Cluster Permission"
	TaintType                    = "Taint"
	ImageDigest                  = "Image Digest"
	ImageRepo                    = "Image Repo"
	ImageTag                     = "Image Tag"
	ImageRegistry                = "Image Registry"
	PodRoleBinding               = "Pods with RoleBindings details"
)

// SetContainerProcessValues adds container process information to the report object.
//
// It records the scheduling policy, scheduling priority and command line for a container
// process by adding corresponding fields to the report. The function also sets the object type to ContainerProcessType.
func (obj *ReportObject) SetContainerProcessValues(aPolicy, aPriority, aCommandLine string) *ReportObject {
	obj.AddField(ProcessCommandLine, aCommandLine)
	obj.AddField(SchedulingPolicy, aPolicy)
	obj.AddField(SchedulingPriority, aPriority)
	obj.ObjectType = ContainerProcessType
	return obj
}

// NewContainerReportObject creates a new ReportObject for a container.
//
// It takes the namespace, pod name, container name, reason, and compliance status as parameters.
// The function constructs a ReportObject with fields Namespace, PodName, ContainerName, ReasonForCompliance or ReasonForNonCompliance set based on the compliance flag, and returns a pointer to it.
func NewContainerReportObject(aNamespace, aPodName, aContainerName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, ContainerType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(PodName, aPodName)
	out.AddField(ContainerName, aContainerName)
	return out
}

// NewCertifiedContainerReportObject creates a ReportObject for a certified container.
//
// It takes a ContainerImageIdentifier, a reason string, and a boolean indicating whether the container is compliant.
// The function returns a pointer to the created ReportObject.
func NewCertifiedContainerReportObject(cii provider.ContainerImageIdentifier, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, ContainerImageType, isCompliant)
	out.AddField(ImageDigest, cii.Digest)
	out.AddField(ImageRepo, cii.Repository)
	out.AddField(ImageTag, cii.Tag)
	out.AddField(ImageRegistry, cii.Registry)
	return out
}

// NewNodeReportObject creates a report object for a node.
//
// It takes the node name, a reason string, and a boolean indicating
// compliance status. The function constructs a ReportObject using
// NewReportObject, adds fields for the node name and reason,
// and sets the compliance flag before returning the pointer.
func NewNodeReportObject(aNodeName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, NodeType, isCompliant)
	out.AddField(Name, aNodeName)
	return out
}

// NewClusterVersionReportObject creates a ReportObject that records the status of a cluster version check.
//
// It accepts three parameters: the version string, an explanatory reason, and a boolean indicating compliance.
// The function builds a new ReportObject with these values set as fields and returns it for further use in reporting.
func NewClusterVersionReportObject(version, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, OCPClusterType, isCompliant)
	out.AddField(OCPClusterVersionType, version)
	return out
}

// NewTaintReportObject creates a report object containing taint information.
//
// It constructs a ReportObject with fields for the taint bit, node name,
// reason for compliance or non‑compliance, and whether the node is compliant.
// The returned pointer can be used to add further details before reporting.
func NewTaintReportObject(taintBit, nodeName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, TaintType, isCompliant)
	out.AddField(NodeType, nodeName)
	out.AddField(TaintBit, taintBit)
	return out
}

// NewPodReportObject creates a new report object for a pod.
//
// It takes the namespace, pod name, reason, and compliance status as input,
// constructs a ReportObject using NewReportObject, adds the PodName, Namespace,
// ReasonForCompliance or ReasonForNonCompliance field based on the status,
// and returns a pointer to the created ReportObject.
func NewPodReportObject(aNamespace, aPodName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, PodType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(PodName, aPodName)
	return out
}

// NewHelmChartReportObject creates a ReportObject for a Helm chart.
//
// It takes a namespace, the Helm chart name, a reason string, and a boolean indicating
// whether the chart is compliant. The function constructs a new ReportObject,
// adds fields for the chart type, name, namespace, reason, and compliance status,
// and then returns the populated object.
func NewHelmChartReportObject(aNamespace, aHelmChartName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, HelmType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(Name, aHelmChartName)
	return out
}

// NewOperatorReportObject creates a ReportObject for an operator.
//
// It builds a report object with the given namespace, operator name,
// reason string and compliance flag. The returned object contains
// fields populated from these inputs and can be used in test reports.
func NewOperatorReportObject(aNamespace, aOperatorName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, OperatorType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(Name, aOperatorName)
	return out
}

// NewClusterOperatorReportObject creates a ReportObject for a cluster operator test.
//
// It takes the name of the operator, its version string, and a boolean indicating
// whether the operator is expected to be compliant. The function initializes a new
// report object with these values and adds fields for the operator name,
// version, compliance expectation, and an empty list of reasons.
// The returned *ReportObject can then be populated with test results.
func NewClusterOperatorReportObject(aClusterOperatorName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, ClusterOperatorType, isCompliant)
	out.AddField(Name, aClusterOperatorName)
	return out
}

// NewCatalogSourceReportObject creates a ReportObject for a catalog source.
//
// It accepts the namespace, catalog source name, reason string and compliance boolean.
// The function constructs a new ReportObject, adds fields for namespace, name, reason and compliance status,
// and returns the resulting object.
func NewCatalogSourceReportObject(aNamespace, aCatalogSourceName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, CatalogSourceType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(Name, aCatalogSourceName)
	return out
}

// NewDeploymentReportObject creates a deployment report object.
//
// It takes a namespace, deployment name, reason string, and a compliance flag,
// constructs a ReportObject with these fields set, and returns a pointer to it.
func NewDeploymentReportObject(aNamespace, aDeploymentName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, DeploymentType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(DeploymentName, aDeploymentName)
	return out
}

// NewStatefulSetReportObject creates a new ReportObject for a StatefulSet.
//
// It takes a namespace, the name of the StatefulSet, a reason string,
// and a boolean indicating compliance status.
// The function returns a pointer to the created ReportObject with fields
// populated for reporting purposes.
func NewStatefulSetReportObject(aNamespace, aStatefulSetName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, StatefulSetType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(StatefulSetName, aStatefulSetName)
	return out
}

// NewCrdReportObject creates a ReportObject for a custom resource definition (CRD).
//
// It accepts the CRD name, its version, a reason string and a compliance flag.
// The returned ReportObject contains fields for Name, Version, Reason, and Status set to
// "Success" or "Failure" based on the boolean.
func NewCrdReportObject(aName, aVersion, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, CustomResourceDefinitionType, isCompliant)
	out.AddField(CustomResourceDefinitionName, aName)
	out.AddField(CustomResourceDefinitionVersion, aVersion)
	return out
}

// NewReportObject creates a new ReportObject with the specified reason, type, and compliance status.
//
// It takes a reason string, a type string, and a boolean indicating compliance.
// If isCompliant is true, the reason is stored under the key ReasonForCompliance;
// otherwise it is stored under ReasonForNonCompliance. The function returns a pointer to the newly created ReportObject.
func NewReportObject(aReason, aType string, isCompliant bool) (out *ReportObject) {
	out = &ReportObject{}
	out.ObjectType = aType
	if isCompliant {
		out.AddField(ReasonForCompliance, aReason)
	} else {
		out.AddField(ReasonForNonCompliance, aReason)
	}
	return out
}

// AddField adds a key-value pair to the ReportObject.
//
// It appends the given key to ObjectFieldsKeys and the value to ObjectFieldsValues,
// then returns the modified ReportObject instance for chaining.
func (obj *ReportObject) AddField(aKey, aValue string) (out *ReportObject) {
	obj.ObjectFieldsKeys = append(obj.ObjectFieldsKeys, aKey)
	obj.ObjectFieldsValues = append(obj.ObjectFieldsValues, aValue)
	return obj
}

// NewNamespacedReportObject creates a new ReportObject with the specified reason, type, compliance status, and namespace.
//
// It calls NewReportObject to initialise the object and then adds a
// "namespace" field using AddField. The returned *ReportObject can be used
// directly in tests or reporting logic.
func NewNamespacedReportObject(aReason, aType string, isCompliant bool, aNamespace string) (out *ReportObject) {
	return NewReportObject(aReason, aType, isCompliant).AddField(Namespace, aNamespace)
}

// NewNamespacedNamedReportObject creates a ReportObject with namespace and name.
//
// It accepts a reason string, type string, compliance status boolean,
// a namespace string, and a name string. The function constructs
// a new report object using NewReportObject, then adds fields for
// reason, type, compliance, namespace, and name before returning the pointer.
func NewNamespacedNamedReportObject(aReason, aType string, isCompliant bool, aNamespace, aName string) (out *ReportObject) {
	return NewReportObject(aReason, aType, isCompliant).AddField(Namespace, aNamespace).AddField(Name, aName)
}

// SetType sets the type of a ReportObject.
//
// SetType assigns the provided string value to the ObjectType field
// of the ReportObject and returns a pointer to the modified object.
// It is intended to be used in fluent chains when building or updating
// report objects for test scenarios.
func (obj *ReportObject) SetType(aType string) (out *ReportObject) {
	obj.ObjectType = aType
	return obj
}

// ResultToString converts an integer result code into a string.
//
// 
// It takes an int representing a result and returns the corresponding
// string constant: SUCCESS, FAILURE or ERROR. If the value does not match
// any known code, it returns an empty string.
func ResultToString(result int) (str string) {
	switch result {
	case SUCCESS:
		return "SUCCESS"
	case FAILURE:
		return "FAILURE"
	case ERROR:
		return "ERROR"
	}
	return ""
}

// GetNonOCPClusterSkipFn returns a closure that determines whether the current test should be skipped because it runs on an OpenShift cluster.
//
// GetNonOCPClusterSkipFn returns a closure that inspects the environment to see if the
// tests are running on an OpenShift (OCP) cluster. The returned function returns a
// boolean indicating whether the skip condition is met and a string explaining the reason.
// If the cluster is OCP, it will return true with a message stating that non-OCP clusters
// are required for this test; otherwise it returns false with an empty message.
func GetNonOCPClusterSkipFn() func() (bool, string) {
	return func() (bool, string) {
		if !provider.IsOCPCluster() {
			return true, "non-OCP cluster detected"
		}
		return false, ""
	}
}

// GetNoServicesUnderTestSkipFn returns a function that can be used to skip tests when no services are under test.
//
// The returned function checks the provided TestEnvironment for an empty Services list.
// If the list is empty, it returns true and a message indicating that there are no services to test,
// causing the calling test to be skipped. Otherwise, it returns false with an empty string.
func GetNoServicesUnderTestSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Services) == 0 {
			return true, "no services to check found"
		}

		return false, ""
	}
}

// GetDaemonSetFailedToSpawnSkipFn returns a function that determines whether to skip tests when a DaemonSet fails to spawn.
// The returned closure checks the test environment state and may trigger an abort if certain conditions are met,
// returning a boolean indicating skip status and a message explaining the reason.
func GetDaemonSetFailedToSpawnSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if env.DaemonsetFailedToSpawn {
			return true, "probe daemonset failed to spawn. please check the logs."
		}

		return false, ""
	}
}

// GetNoCPUPinningPodsSkipFn returns a predicate used to skip tests when no CPU‑pinned pods are present.
//
// The returned function examines the current test environment and determines
// whether any CPU‑pinning pods with DPDK support exist. If the list is empty,
// it signals that the test should be skipped, providing an explanatory string.
// Otherwise, the test proceeds normally. This helper allows tests to be
// conditionally bypassed in environments lacking the required pod setup.
func GetNoCPUPinningPodsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.GetCPUPinningPodsWithDpdk()) == 0 {
			return true, "no CPU pinning pods to check found"
		}

		return false, ""
	}
}

// GetNoSRIOVPodsSkipFn returns a function that can be used to skip tests when no SR‑I/O‑V pods are present.
//
// It accepts a test environment and returns a closure. When the closure is called it checks how many
// pods in the environment use SR‑I/O‑V networking. If none do, the closure returns true along with a message
// indicating that the test is being skipped because no SR‑I/O‑V pods were found.
// Otherwise it returns false and an empty string. The returned function can be passed to test frameworks
// that support skip callbacks.
func GetNoSRIOVPodsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		pods, err := env.GetPodsUsingSRIOV()
		if err != nil {
			return true, fmt.Sprintf("failed to get SRIOV pods: %v", err)
		}

		if len(pods) == 0 {
			return true, "no SRIOV pods to check found"
		}

		return false, ""
	}
}

// GetNoContainersUnderTestSkipFn returns a skip function for tests that require at least one container under test.
//
// It examines the provided TestEnvironment and determines whether any containers are marked for testing.
// If none are found, it returns true along with an explanatory message; otherwise it returns false.
func GetNoContainersUnderTestSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Containers) == 0 {
			return true, "no containers to check found"
		}

		return false, ""
	}
}

// GetNoPodsUnderTestSkipFn returns a closure that indicates whether to skip tests when no pods are present.
//
// The returned function examines the TestEnvironment passed to GetNoPodsUnderTestSkipFn and
// determines if the number of pods under test is zero. If so, it returns true along with a
// message explaining that the tests are skipped because there are no pods to evaluate.
// Otherwise, it returns false and an empty string. This helper is used in test suites to
// conditionally skip test execution when the required pod resources are absent.
func GetNoPodsUnderTestSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Pods) == 0 {
			return true, "no pods to check found"
		}

		return false, ""
	}
}

// GetNoDeploymentsUnderTestSkipFn returns a function that can be used to conditionally skip tests when there are no deployments under test.
//
// It accepts a pointer to a TestEnvironment and produces a zero‑argument function.
// The returned function evaluates the environment’s deployment list; if it is empty,
// it returns true with an explanatory message, otherwise false. This allows callers
// to defer the decision until runtime while keeping the skip logic encapsulated.
func GetNoDeploymentsUnderTestSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Deployments) == 0 {
			return true, "no deployments to check found"
		}

		return false, ""
	}
}

// GetNoStatefulSetsUnderTestSkipFn returns a function that can be used as a test skip predicate.  
// It examines the supplied TestEnvironment and determines whether any StatefulSet resources are present under test.  
// If the count of StatefulSets is zero, the returned function signals to skip the test by returning true along with an explanatory message; otherwise it indicates that the test should run.
func GetNoStatefulSetsUnderTestSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.StatefulSets) == 0 {
			return true, "no statefulSets to check found"
		}

		return false, ""
	}
}

// GetNoCrdsUnderTestSkipFn returns a function that determines whether to skip tests
// when no custom resource definitions (CRDs) are present in the test environment.
//
// It accepts a pointer to a TestEnvironment and produces a closure.
// The closure checks if the number of CRDs under test is zero; if so, it returns
// true along with a message explaining that tests are skipped because no CRDs were found.
// Otherwise, it returns false and an empty string.
func GetNoCrdsUnderTestSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Crds) == 0 {
			return true, "no roles to check"
		}

		return false, ""
	}
}

// GetNoNamespacesSkipFn returns a skip function based on namespace support.
//
// It inspects the provided TestEnvironment; if the environment reports no
// namespaces (len==0), the returned closure will signal that a test should be
// skipped and supply a message explaining the lack of namespace support.
// Otherwise the closure signals that the test may proceed normally. This is
// intended for use in test suites where tests must be conditionally bypassed
// when namespaces are unavailable.
func GetNoNamespacesSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Namespaces) == 0 {
			return true, "There are no namespaces to check. Please check config."
		}

		return false, ""
	}
}

// GetNoRolesSkipFn returns a test skip predicate based on the TestEnvironment’s NoRoles list.
//
// GetNoRolesSkipFn returns a function that indicates whether a test should be skipped
// because no roles are defined in the current TestEnvironment.
//
// The returned function examines the TestEnvironment passed to GetNoRolesSkipFn.
// If the environment has an empty NoRoles slice, it returns true along with a message
// explaining that the test is being skipped due to lack of roles. Otherwise it returns
// false and an empty string, indicating the test should proceed.
func GetNoRolesSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Roles) == 0 {
			return true, "There are no roles to check. Please check config."
		}

		return false, ""
	}
}

// GetSharedProcessNamespacePodsSkipFn returns a function that determines whether to skip tests for pods using shared process namespace.
//
// The returned closure checks the test environment for any pods that have
// shareProcessNamespace set. If such pods exist, it signals to skip by
// returning true and an explanatory message; otherwise it indicates no skip
// with false and an empty string. It uses GetShareProcessNamespacePods to
// gather the relevant pods and len to decide if skipping is needed.
func GetSharedProcessNamespacePodsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.GetShareProcessNamespacePods()) == 0 {
			return true, "Shared process namespace pods found."
		}

		return false, ""
	}
}

// GetNotIntrusiveSkipFn returns a function that can be used to skip tests when the environment is marked as intrusive.
//
// The returned closure checks whether the given TestEnvironment is flagged as intrusive
// by calling IsIntrusive. If it is, the closure signals that the test should be skipped,
// returning true along with a message explaining why. Otherwise it indicates that
// the test should proceed by returning false and an empty string. This helper is
// useful for tests that must not run in environments where intrusive checks are enabled.
func GetNotIntrusiveSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if !env.IsIntrusive() {
			return true, "not intrusive test"
		}

		return false, ""
	}
}

// GetNoPersistentVolumesSkipFn returns a function that decides whether to skip tests requiring persistent volumes based on the test environment.
//
// The returned closure examines the provided TestEnvironment and checks the number of persistent volume claims it contains.
// If no persistent volumes are present, it returns true along with an explanatory message; otherwise it returns false and an empty string.
func GetNoPersistentVolumesSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.PersistentVolumes) == 0 {
			return true, "no persistent volumes to check found"
		}

		return false, ""
	}
}

// GetNotEnoughWorkersSkipFn creates a test skip function based on the required number of workers.
//
// It accepts a TestEnvironment and an integer indicating how many workers are needed.
// The returned function queries the environment for its current worker count
// using GetWorkerCount. If the count is below the requested amount, it returns
// true along with a message explaining that there are not enough workers;
// otherwise it returns false and an empty string.
func GetNotEnoughWorkersSkipFn(env *provider.TestEnvironment, minWorkerNodes int) func() (bool, string) {
	return func() (bool, string) {
		if env.GetWorkerCount() < minWorkerNodes {
			return true, "not enough nodes to check found"
		}

		return false, ""
	}
}

// GetPodsWithoutAffinityRequiredLabelSkipFn returns a skip function that can be used to
// determine whether a test should be skipped because the current environment has no pods
// lacking the required affinity label.
//
// It takes a TestEnvironment pointer, retrieves all pods missing the affinity-required
// label by calling GetPodsWithoutAffinityRequiredLabel, and returns a closure. The
// closure returns true and an explanatory message if any such pods exist; otherwise it
// returns false with an empty string.
func GetPodsWithoutAffinityRequiredLabelSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.GetPodsWithoutAffinityRequiredLabel()) == 0 {
			return true, "no pods with required affinity label found"
		}

		return false, ""
	}
}

// GetNoGuaranteedPodsWithExclusiveCPUsSkipFn returns a test skip function.
//
// It creates a closure that inspects the test environment for any
// guaranteed pods that have exclusive CPU allocations. If such pods are
// found, the returned function will signal to skip the test by returning
// true along with an explanatory message; otherwise it signals not to
// skip by returning false and an empty string. The function relies on
// GetGuaranteedPodsWithExclusiveCPUs to retrieve the relevant pod list
// and uses len to determine if any entries exist.
func GetNoGuaranteedPodsWithExclusiveCPUsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.GetGuaranteedPodsWithExclusiveCPUs()) == 0 {
			return true, "no pods with exclusive CPUs found"
		}

		return false, ""
	}
}

// GetNoAffinityRequiredPodsSkipFn returns a skip function that checks whether there are any pods requiring affinity in the test environment.
//
// It retrieves the list of affinity-required pods via GetAffinityRequiredPods, then returns true with an explanatory message if that list is empty,
// indicating that no such pods exist and the test can be skipped. If affinity-required pods are present, it returns false to continue testing.
func GetNoAffinityRequiredPodsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.GetAffinityRequiredPods()) == 0 {
			return true, "no pods with required affinity found"
		}

		return false, ""
	}
}

// GetNoStorageClassesSkipFn returns a closure used to decide whether to skip tests when the cluster provides no StorageClass resources.
//
// It accepts a *provider.TestEnvironment and yields a function that, when called,
// evaluates the number of storage classes available in the test environment.
// If none are present, it returns true along with a message explaining that
// the test is skipped due to the absence of StorageClasses. Otherwise it
// returns false and an empty string indicating the test should proceed.
func GetNoStorageClassesSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.StorageClassList) == 0 {
			return true, "no storage classes found"
		}
		return false, ""
	}
}

// GetNoPersistentVolumeClaimsSkipFn returns a test skip predicate based on persistent volume claim usage.
//
// It takes a pointer to a TestEnvironment and produces a closure that, when called,
// determines whether the environment has zero PersistentVolumeClaim resources.
// If there are no claims, it signals that the test should be skipped by returning true
// along with an explanatory message. The function uses len on the environment's
// claim list to perform this check.
func GetNoPersistentVolumeClaimsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.PersistentVolumeClaims) == 0 {
			return true, "no persistent volume claims found"
		}
		return false, ""
	}
}

// GetNoBareMetalNodesSkipFn returns a skip function that indicates whether tests should be skipped when no bare metal nodes are available.
//
// It takes a TestEnvironment pointer and examines the number of bare metal nodes
// in the environment. If zero, it returns a function that always signals to skip,
// providing an explanatory message. Otherwise, it returns a function that never skips.
func GetNoBareMetalNodesSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.GetBaremetalNodes()) == 0 {
			return true, "no baremetal nodes found"
		}
		return false, ""
	}
}

// GetNoIstioSkipFn returns a function that determines whether Istio tests should be skipped based on the test environment.
//
// It accepts a pointer to TestEnvironment and returns a closure.
// The returned function evaluates the environment configuration and
// indicates if Istio is not present, returning (true, "reason") or
// (false, "") accordingly. This allows callers to conditionally skip
// tests when Istio is absent from the cluster.
func GetNoIstioSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if !env.IstioServiceMeshFound {
			return true, "no istio service mesh found"
		}
		return false, ""
	}
}

// GetNoHugepagesPodsSkipFn returns a function that determines if a test should be skipped when no hugepages pods are present.
//
// It accepts a TestEnvironment pointer and produces a closure that, when called,
// checks the number of hugepages pods available in the environment using
// GetHugepagesPods. If none are found, it signals to skip by returning true
// along with a descriptive message; otherwise it returns false and an empty string.
func GetNoHugepagesPodsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.GetHugepagesPods()) == 0 {
			return true, "no pods requesting hugepages found"
		}
		return false, ""
	}
}

// GetNoCatalogSourcesSkipFn returns a function that determines whether tests should be skipped due to the absence of CatalogSource objects.
//
// It accepts a *provider.TestEnvironment and produces a closure that, when called,
// checks the test environment for any catalog sources. If none are found,
// it returns true along with a message explaining that no catalog sources
// were detected; otherwise it returns false and an empty string.
func GetNoCatalogSourcesSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.AllCatalogSources) == 0 {
			return true, "no catalog sources found"
		}
		return false, ""
	}
}

// GetNoOperatorsSkipFn returns a function that determines whether to skip tests when the test environment has no operators configured.
//
// It examines the provided TestEnvironment and, if the NoOperators slice is empty,
// signals that tests requiring operators should be skipped by returning true
// along with an explanatory message. If operators are present, it returns false
// and an empty string. The returned function can be used as a skip condition in
// test suites to conditionally bypass operator‑dependent checks.
func GetNoOperatorsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Operators) == 0 {
			return true, "no operators found"
		}
		return false, ""
	}
}

// GetNoOperatorPodsSkipFn returns a function that decides whether to skip tests when no operator pods are present.
//
// The returned closure examines the TestEnvironment passed to GetNoOperatorPodsSkipFn and
// determines if any operator pods exist in the current test cluster.
// If zero operator pods are found, it returns true along with an explanatory message,
// otherwise it returns false. This helper is used by test suites that should be bypassed
// when operators are not running.
func GetNoOperatorPodsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.CSVToPodListMap) == 0 {
			return true, "no operator pods found"
		}

		return false, ""
	}
}

// GetNoOperatorCrdsSkipFn returns a closure used to determine whether tests should be skipped when no Operator Custom Resource Definitions (CRDs) are present in the test environment.
//
// It accepts a pointer to a TestEnvironment and produces a function that, when called,
// returns a boolean indicating if the skip condition is met and an accompanying message.
// The inner function typically checks the number of operator CRDs available; if none are found
// it signals that related tests should be skipped. The returned message explains the reason for skipping.
func GetNoOperatorCrdsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Crds) == 0 {
			return true, "no operator crds found"
		}
		return false, ""
	}
}

// GetNoNodesWithRealtimeKernelSkipFn returns a function that indicates whether tests should be skipped due to lack of realtime kernel nodes.
//
// It accepts a test environment and produces a closure. The closure checks the node list for any node running a realtime kernel by calling IsRTKernel. If no such node is found, it returns true along with a message explaining the skip condition. Otherwise, it returns false, allowing tests to proceed.
func GetNoNodesWithRealtimeKernelSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		for i := range env.Nodes {
			node := env.Nodes[i]

			if node.IsRTKernel() {
				return false, ""
			}
		}

		return true, "no nodes with realtime kernel type found"
	}
}

// ResultObjectsToString converts two slices of ReportObject pointers into a JSON string representation.
//
// It takes two slices: the first contains objects that passed checks,
// and the second contains objects that failed checks.
// The function marshals both slices into a single JSON object with
// "passed" and "failed" fields. If marshalling fails, it returns an error
// wrapped with context about the conversion failure. On success it returns
// the resulting JSON string and nil error.
func ResultObjectsToString(compliantObject, nonCompliantObject []*ReportObject) (string, error) {
	reason := FailureReasonOut{
		CompliantObjectsOut:    compliantObject,
		NonCompliantObjectsOut: nonCompliantObject,
	}

	bytes, err := json.Marshal(reason)
	if err != nil {
		return "", fmt.Errorf("could not marshall FailureReasonOut object: %v", err)
	}

	return string(bytes), nil
}

var AbortTrigger string
