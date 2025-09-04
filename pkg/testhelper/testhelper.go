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

// ReportObject Represents a structured report entry with type and key/value attributes
//
// This structure holds the kind of object being reported, along with parallel
// slices that store field names and corresponding values. The fields are
// populated via methods such as AddField, SetContainerProcessValues, or
// SetType, allowing callers to build descriptive reports for compliance checks.
// It serves as a lightweight container used throughout the test helper package
// to aggregate and serialize results.
type ReportObject struct {
	ObjectType         string
	ObjectFieldsKeys   []string
	ObjectFieldsValues []string
}

// FailureReasonOut Represents collections of compliant and non-compliant report objects
//
// This structure stores two separate lists of report objects, one for items
// that meet the compliance criteria and another for those that do not. Each
// list holds pointers to ReportObject instances, allowing callers to access
// detailed information about each item. The struct provides an Equal method to
// compare two instances by checking both slices for identical contents.
type FailureReasonOut struct {
	CompliantObjectsOut    []*ReportObject
	NonCompliantObjectsOut []*ReportObject
}

// Equal Compares two slices of ReportObject pointers for deep equality
//
// The function first verifies that both slices have the same length. It then
// iterates through each index, treating nil entries as equal only when both are
// nil; a mismatch in nil status causes an immediate false result. For non-nil
// elements, it uses reflect.DeepEqual on the dereferenced values to determine
// equality, returning true only if all corresponding pairs match.
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

// FailureReasonOutTestString Formats a FailureReasonOut as a readable string
//
// This function takes a FailureReasonOut value and builds a formatted string
// that includes the compliant and non‑compliant object lists. It uses helper
// formatting to produce a concise representation of each list, then
// concatenates them into a single string for debugging or test output.
func FailureReasonOutTestString(p FailureReasonOut) (out string) {
	out = "testhelper.FailureReasonOut{"
	out += fmt.Sprintf("CompliantObjectsOut: %s,", ReportObjectTestStringPointer(p.CompliantObjectsOut))
	out += fmt.Sprintf("NonCompliantObjectsOut: %s,", ReportObjectTestStringPointer(p.NonCompliantObjectsOut))
	out += "}"
	return out
}

// ReportObjectTestStringPointer Formats a slice of ReportObject pointers into a readable string
//
// It receives a list of pointers to ReportObject, iterates over each element,
// and appends a formatted representation of the dereferenced object to an
// output string. The resulting string starts with "[]*testhelper.ReportObject"
// and ends with "", enclosing all items separated by commas. This string is
// used primarily for debugging or test failure messages.
func ReportObjectTestStringPointer(p []*ReportObject) (out string) {
	out = "[]*testhelper.ReportObject{"
	for _, p := range p {
		out += fmt.Sprintf("&%#v,", *p)
	}
	out += "}"
	return out
}

// ReportObjectTestString Creates a formatted string of ReportObject values
//
// The function takes a slice of pointers to ReportObject and builds a single
// string that lists each element in the same order as the input. Each object is
// rendered with the %#v format specifier, appended with a comma, and the entire
// list is wrapped in brackets prefixed by "[]testhelper.ReportObject". The
// resulting string is returned for use in test output or debugging.
func ReportObjectTestString(p []*ReportObject) (out string) {
	out = "[]testhelper.ReportObject{"
	for _, p := range p {
		out += fmt.Sprintf("%#v,", *p)
	}
	out += "}"
	return out
}

// FailureReasonOut.Equal determines equality of two FailureReasonOut instances
//
// It compares the CompliantObjectsOut and NonCompliantObjectsOut fields of both
// structs, returning true only if all corresponding values match. The
// comparison is performed using the generic Equal function for each field. If
// any field differs, it returns false.
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

// ReportObject.SetContainerProcessValues Stores container process details in the report object
//
// It records the command line, scheduling policy, and priority of a container
// process by adding these fields to the report. The function also tags the
// report with a type indicating it represents a container process. The updated
// report object is returned for further chaining.
func (obj *ReportObject) SetContainerProcessValues(aPolicy, aPriority, aCommandLine string) *ReportObject {
	obj.AddField(ProcessCommandLine, aCommandLine)
	obj.AddField(SchedulingPolicy, aPolicy)
	obj.AddField(SchedulingPriority, aPriority)
	obj.ObjectType = ContainerProcessType
	return obj
}

// NewContainerReportObject Creates a report object for a container
//
// It builds a ReportObject with type ContainerType, attaching the provided
// namespace, pod name, container name, and compliance reason as fields. The
// function uses NewReportObject to set the compliance status and then adds
// additional identifying fields before returning the pointer.
func NewContainerReportObject(aNamespace, aPodName, aContainerName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, ContainerType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(PodName, aPodName)
	out.AddField(ContainerName, aContainerName)
	return out
}

// NewCertifiedContainerReportObject Creates a report object for a container image
//
// This function receives an image identifier, a compliance reason string, and a
// flag indicating whether the image meets compliance requirements. It
// constructs a new report object of type ContainerImageType, annotating it with
// the provided reason as either compliant or non‑compliant. The resulting
// object includes fields for digest, repository, tag, and registry derived from
// the identifier.
func NewCertifiedContainerReportObject(cii provider.ContainerImageIdentifier, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, ContainerImageType, isCompliant)
	out.AddField(ImageDigest, cii.Digest)
	out.AddField(ImageRepo, cii.Repository)
	out.AddField(ImageTag, cii.Tag)
	out.AddField(ImageRegistry, cii.Registry)
	return out
}

// NewNodeReportObject Creates a node-specific report object
//
// The function builds a ReportObject for a node by calling the generic
// constructor with the provided reason, type identifier, and compliance flag.
// It then attaches the node name as an additional field before returning the
// fully populated object.
func NewNodeReportObject(aNodeName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, NodeType, isCompliant)
	out.AddField(Name, aNodeName)
	return out
}

// NewClusterVersionReportObject Creates a report object containing cluster version information
//
// The function takes a version string, a reason for compliance or
// non‑compliance, and a boolean indicating compliance status. It constructs a
// new ReportObject with the provided reason and type, then adds the version as
// an additional field before returning the object.
func NewClusterVersionReportObject(version, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, OCPClusterType, isCompliant)
	out.AddField(OCPClusterVersionType, version)
	return out
}

// NewTaintReportObject Creates a taint report object with node details
//
// This function builds a ReportObject that records a specific taint bit on a
// given node. It initializes the object with the reason for compliance or
// non‑compliance, sets its type to a predefined taint category, and then adds
// fields for the node name and the taint bit value. The resulting pointer is
// returned for further use in testing or reporting.
func NewTaintReportObject(taintBit, nodeName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, TaintType, isCompliant)
	out.AddField(NodeType, nodeName)
	out.AddField(TaintBit, taintBit)
	return out
}

// NewPodReportObject Creates a report object for a pod
//
// The function builds a ReportObject by calling NewReportObject with the given
// reason, type set to PodType, and compliance flag. It then attaches the
// namespace and pod name as fields on the resulting object. Finally, it returns
// a pointer to this populated ReportObject.
func NewPodReportObject(aNamespace, aPodName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, PodType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(PodName, aPodName)
	return out
}

// NewHelmChartReportObject Creates a report object for a Helm chart
//
// It constructs a new report object with the provided namespace, chart name,
// reason, and compliance status. The function first creates a base report
// object using the supplied reason and compliance flag, then adds fields for
// the namespace and chart name to that object. The completed report object is
// returned for use in testing or reporting.
func NewHelmChartReportObject(aNamespace, aHelmChartName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, HelmType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(Name, aHelmChartName)
	return out
}

// NewOperatorReportObject Creates a report object for an operator
//
// The function builds a new ReportObject using the provided namespace, operator
// name, reason, and compliance flag. It initializes the base object with type
// information, then adds fields for namespace and operator name before
// returning it.
func NewOperatorReportObject(aNamespace, aOperatorName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, OperatorType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(Name, aOperatorName)
	return out
}

// NewClusterOperatorReportObject Creates a report object for a cluster operator
//
// This function builds a ReportObject by calling the generic constructor with a
// reason, type label, and compliance flag. It then adds the operator name as an
// additional field before returning the populated object. The returned pointer
// represents a structured report entry that can be used in test results.
func NewClusterOperatorReportObject(aClusterOperatorName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, ClusterOperatorType, isCompliant)
	out.AddField(Name, aClusterOperatorName)
	return out
}

// NewCatalogSourceReportObject Creates a report object for a catalog source
//
// The function builds a new report object using the provided namespace, catalog
// source name, reason, and compliance flag. It delegates creation to an
// internal helper that sets the type and records whether the item is compliant.
// Finally, it adds namespace and name fields before returning the populated
// report.
func NewCatalogSourceReportObject(aNamespace, aCatalogSourceName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, CatalogSourceType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(Name, aCatalogSourceName)
	return out
}

// NewDeploymentReportObject Creates a deployment report object with namespace, name, reason, and compliance status
//
// This function builds a new ReportObject by first invoking the generic
// constructor with the provided reason, type identifier for deployments, and
// compliance flag. It then adds fields for the namespace and deployment name to
// the object's key/value store. The resulting pointer is returned for further
// use or inspection.
func NewDeploymentReportObject(aNamespace, aDeploymentName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, DeploymentType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(DeploymentName, aDeploymentName)
	return out
}

// NewStatefulSetReportObject Creates a report object for a StatefulSet
//
// It builds a ReportObject with the type set to a constant representing
// StatefulSet, attaches compliance or non‑compliance reason, then adds
// namespace and name fields. The function returns the fully populated
// ReportObject for use in tests.
func NewStatefulSetReportObject(aNamespace, aStatefulSetName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, StatefulSetType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(StatefulSetName, aStatefulSetName)
	return out
}

// NewCrdReportObject Creates a report object for a custom resource definition
//
// This function takes the name, version, reason, and compliance status of a
// CRD. It builds a ReportObject by delegating to NewReportObject, then adds
// fields for the CRD's name and version before returning the constructed
// object.
func NewCrdReportObject(aName, aVersion, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, CustomResourceDefinitionType, isCompliant)
	out.AddField(CustomResourceDefinitionName, aName)
	out.AddField(CustomResourceDefinitionVersion, aVersion)
	return out
}

// NewReportObject Creates a report object with reason and type
//
// This function initializes an empty ReportObject, sets its type field, and
// adds the provided reason as either a compliance or non‑compliance note
// depending on the boolean flag. The resulting pointer is returned for further
// augmentation by caller functions.
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

// ReportObject.AddField Adds a key-value pair to the report
//
// The method appends the supplied key to an internal slice of keys and the
// corresponding value to a parallel slice of values, maintaining order. It
// returns the same ReportObject pointer so calls can be chained. This enables
// constructing structured reports by sequentially adding fields.
func (obj *ReportObject) AddField(aKey, aValue string) (out *ReportObject) {
	obj.ObjectFieldsKeys = append(obj.ObjectFieldsKeys, aKey)
	obj.ObjectFieldsValues = append(obj.ObjectFieldsValues, aValue)
	return obj
}

// NewNamespacedReportObject Creates a ReportObject that includes namespace information
//
// The function constructs a new report object with the provided reason, type,
// and compliance status, then appends an additional field for the namespace. It
// returns the resulting report object. This allows callers to generate reports
// that are scoped to a specific Kubernetes namespace.
func NewNamespacedReportObject(aReason, aType string, isCompliant bool, aNamespace string) (out *ReportObject) {
	return NewReportObject(aReason, aType, isCompliant).AddField(Namespace, aNamespace)
}

// NewNamespacedNamedReportObject Creates a report object with namespace and name fields
//
// It builds a new ReportObject using the reason, type, and compliance flag,
// then appends the specified namespace and name as additional fields. The
// resulting pointer is returned for further use.
func NewNamespacedNamedReportObject(aReason, aType string, isCompliant bool, aNamespace, aName string) (out *ReportObject) {
	return NewReportObject(aReason, aType, isCompliant).AddField(Namespace, aNamespace).AddField(Name, aName)
}

// ReportObject.SetType Assigns a new type to the report object
//
// The method receives a string that represents the desired type and stores it
// in the ObjectType field of the ReportObject instance. It then returns the
// same instance, allowing callers to chain further configuration calls.
func (obj *ReportObject) SetType(aType string) (out *ReportObject) {
	obj.ObjectType = aType
	return obj
}

// ResultToString Translates a result code into its textual form
//
// The function receives an integer representing a status code and returns the
// matching string: "SUCCESS", "FAILURE" or "ERROR". If the input does not match
// any known code, it yields an empty string.
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

// GetNonOCPClusterSkipFn provides a test skip function for non‑OCP clusters
//
// This helper creates and returns a zero‑argument function that, when called,
// checks whether the current environment is an OpenShift cluster. If it is not,
// the returned function signals to skip the test by returning true along with a
// descriptive message; otherwise it indicates no skip with false and an empty
// string.
func GetNonOCPClusterSkipFn() func() (bool, string) {
	return func() (bool, string) {
		if !provider.IsOCPCluster() {
			return true, "non-OCP cluster detected"
		}
		return false, ""
	}
}

// GetNoServicesUnderTestSkipFn Checks whether the test environment has any services defined
//
// The function produces a closure that inspects the provided test environment's
// service list. If the list is empty it signals to skip the test with an
// explanatory message; otherwise it indicates the test should proceed.
func GetNoServicesUnderTestSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Services) == 0 {
			return true, "no services to check found"
		}

		return false, ""
	}
}

// GetDaemonSetFailedToSpawnSkipFn returns a closure that skips tests when the probe daemonset fails to spawn
//
// The function takes a test environment and produces a zero‑argument function
// returning a boolean and a message. When called, the inner function checks
// whether the environment records a failed daemonset launch; if so it signals
// the test should be skipped with an explanatory string. Otherwise it indicates
// no skip is needed.
func GetDaemonSetFailedToSpawnSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if env.DaemonsetFailedToSpawn {
			return true, "probe daemonset failed to spawn. please check the logs."
		}

		return false, ""
	}
}

// GetNoCPUPinningPodsSkipFn Checks for the presence of CPU pinning pods before running a test
//
// This function receives an environment object and returns a closure that
// indicates if a test should be skipped. The inner function counts
// CPU‑pinning pods with DPDK; if none are found it signals to skip with an
// explanatory message, otherwise it allows the test to proceed.
func GetNoCPUPinningPodsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.GetCPUPinningPodsWithDpdk()) == 0 {
			return true, "no CPU pinning pods to check found"
		}

		return false, ""
	}
}

// GetNoSRIOVPodsSkipFn Provides a skip function for tests when no SRIOV pods are present
//
// This returns a closure that checks the test environment for SRIOV-enabled
// pods. If retrieving the list fails or the list is empty, it signals to skip
// the test with an explanatory message; otherwise the test proceeds normally.
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

// GetNoContainersUnderTestSkipFn skips tests when there are no containers to evaluate
//
// This function receives a test environment and returns another function that
// determines whether the current test should be skipped. It checks if the
// container list in the environment is empty; if so, it signals to skip with an
// explanatory message. Otherwise, it allows the test to proceed.
func GetNoContainersUnderTestSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Containers) == 0 {
			return true, "no containers to check found"
		}

		return false, ""
	}
}

// GetNoPodsUnderTestSkipFn skips the test when there are no pods to check
//
// This function creates a closure that examines the supplied test environment's
// pod list. If the list is empty, it signals that the test should be skipped by
// returning true and an explanatory message; otherwise, it indicates the test
// should run.
func GetNoPodsUnderTestSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Pods) == 0 {
			return true, "no pods to check found"
		}

		return false, ""
	}
}

// GetNoDeploymentsUnderTestSkipFn Determines whether tests should be skipped due to absence of deployments
//
// The function returns a closure that checks the length of the Deployments
// slice in a test environment. If no deployments are present, it signals that
// the test should skip with an explanatory message. Otherwise, it indicates
// that testing can proceed.
func GetNoDeploymentsUnderTestSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Deployments) == 0 {
			return true, "no deployments to check found"
		}

		return false, ""
	}
}

// GetNoStatefulSetsUnderTestSkipFn Skips tests when there are no StatefulSets in the environment
//
// This function receives a test environment and produces a callback used by
// test frameworks to decide whether to skip a particular check. The returned
// closure inspects the number of StatefulSet objects present; if none exist, it
// signals that the test should be skipped with an explanatory message.
// Otherwise it indicates the test can proceed.
func GetNoStatefulSetsUnderTestSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.StatefulSets) == 0 {
			return true, "no statefulSets to check found"
		}

		return false, ""
	}
}

// GetNoCrdsUnderTestSkipFn Provides a skip function for tests when no CRDs are present
//
// It returns an anonymous function that checks the TestEnvironment's Crds
// slice. If the slice is empty, the inner function signals to skip the test
// with a message indicating there are no roles to check. Otherwise it allows
// the test to proceed.
func GetNoCrdsUnderTestSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Crds) == 0 {
			return true, "no roles to check"
		}

		return false, ""
	}
}

// GetNoNamespacesSkipFn Determines whether tests should be skipped due to lack of namespaces
//
// The function returns a closure that checks the provided test environment for
// configured namespaces. If no namespaces are present, it signals that tests
// should be skipped and supplies an explanatory message. Otherwise, it
// indicates that testing can proceed normally.
func GetNoNamespacesSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Namespaces) == 0 {
			return true, "There are no namespaces to check. Please check config."
		}

		return false, ""
	}
}

// GetNoRolesSkipFn Determines whether tests should be skipped due to missing roles
//
// The returned function checks the Roles slice in the test environment. If no
// roles are present, it signals a skip by returning true along with an
// explanatory message. Otherwise, it indicates that testing can proceed.
func GetNoRolesSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Roles) == 0 {
			return true, "There are no roles to check. Please check config."
		}

		return false, ""
	}
}

// GetSharedProcessNamespacePodsSkipFn Determines whether to skip tests based on shared process namespace pod presence
//
// It examines the test environment for pods that share a process namespace. If
// none are present, it signals that the condition required for the test is not
// met and returns true along with an explanatory message. Otherwise, it
// indicates the test should proceed.
func GetSharedProcessNamespacePodsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.GetShareProcessNamespacePods()) == 0 {
			return true, "Shared process namespace pods found."
		}

		return false, ""
	}
}

// GetNotIntrusiveSkipFn Provides a skip function for non‑intrusive tests
//
// The returned closure checks whether the test environment is marked as
// intrusive. If it is not, the function signals that the test should be skipped
// by returning true along with an explanatory message. Otherwise, it indicates
// the test should run normally.
func GetNotIntrusiveSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if !env.IsIntrusive() {
			return true, "not intrusive test"
		}

		return false, ""
	}
}

// GetNoPersistentVolumesSkipFn skips tests when no persistent volumes exist
//
// It produces a function that inspects the test environment’s list of
// persistent volumes. If the list is empty, it signals to skip the related
// tests and provides an explanatory message; otherwise it allows the tests to
// run.
func GetNoPersistentVolumesSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.PersistentVolumes) == 0 {
			return true, "no persistent volumes to check found"
		}

		return false, ""
	}
}

// GetNotEnoughWorkersSkipFn Creates a test skip function based on worker count
//
// This returns a closure that checks whether the current environment has fewer
// workers than the required minimum. If the condition is met, it signals to
// skip the test by returning true along with an explanatory message; otherwise
// it indicates the test should proceed.
func GetNotEnoughWorkersSkipFn(env *provider.TestEnvironment, minWorkerNodes int) func() (bool, string) {
	return func() (bool, string) {
		if env.GetWorkerCount() < minWorkerNodes {
			return true, "not enough nodes to check found"
		}

		return false, ""
	}
}

// GetPodsWithoutAffinityRequiredLabelSkipFn Creates a skip function for tests that require pods with an affinity label
//
// It receives the test environment and returns a closure that checks whether
// any pods lack the required affinity label. If none are found, the closure
// signals to skip the test with an explanatory message; otherwise it allows the
// test to proceed.
func GetPodsWithoutAffinityRequiredLabelSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.GetPodsWithoutAffinityRequiredLabel()) == 0 {
			return true, "no pods with required affinity label found"
		}

		return false, ""
	}
}

// GetNoGuaranteedPodsWithExclusiveCPUsSkipFn skips test when there are no pods using exclusive CPUs
//
// The returned closure examines the test environment for pods that have been
// assigned exclusive CPU resources. If none are found, it signals to skip the
// test by returning true and a descriptive message. Otherwise, it allows the
// test to proceed.
func GetNoGuaranteedPodsWithExclusiveCPUsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.GetGuaranteedPodsWithExclusiveCPUs()) == 0 {
			return true, "no pods with exclusive CPUs found"
		}

		return false, ""
	}
}

// GetNoAffinityRequiredPodsSkipFn Determines if a test should be skipped due to absence of affinity-required pods
//
// The function returns a closure that checks the test environment for any pods
// marked with required node affinity. If none are found, it signals that the
// test should be skipped and provides an explanatory message. Otherwise, it
// indicates the test can proceed.
func GetNoAffinityRequiredPodsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.GetAffinityRequiredPods()) == 0 {
			return true, "no pods with required affinity found"
		}

		return false, ""
	}
}

// GetNoStorageClassesSkipFn Skips tests when no storage classes are present
//
// This function returns a closure that checks the length of the environment's
// storage class list. If the list is empty, it signals to skip the test with an
// explanatory message; otherwise, it allows the test to proceed normally.
func GetNoStorageClassesSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.StorageClassList) == 0 {
			return true, "no storage classes found"
		}
		return false, ""
	}
}

// GetNoPersistentVolumeClaimsSkipFn Determines if tests should be skipped due to absence of persistent volume claims
//
// The function receives a test environment and produces a closure used by the
// testing framework. When invoked, the closure checks whether the environment
// contains any persistent volume claim objects. If none are present, it signals
// that the test should be skipped and supplies an explanatory message;
// otherwise it allows the test to proceed.
func GetNoPersistentVolumeClaimsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.PersistentVolumeClaims) == 0 {
			return true, "no persistent volume claims found"
		}
		return false, ""
	}
}

// GetNoBareMetalNodesSkipFn skips tests when no bare-metal nodes exist
//
// The returned function checks the test environment for bare-metal nodes by
// calling GetBaremetalNodes. If none are found, it signals that the current
// test should be skipped with a descriptive message. Otherwise, it allows the
// test to proceed normally.
func GetNoBareMetalNodesSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.GetBaremetalNodes()) == 0 {
			return true, "no baremetal nodes found"
		}
		return false, ""
	}
}

// GetNoIstioSkipFn Decides if tests should be skipped due to missing Istio
//
// The function creates and returns a closure that inspects the test environment
// for an Istio service mesh flag. If the flag indicates no Istio is present, it
// signals to skip with a descriptive message; otherwise it allows the test to
// proceed.
func GetNoIstioSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if !env.IstioServiceMeshFound {
			return true, "no istio service mesh found"
		}
		return false, ""
	}
}

// GetNoHugepagesPodsSkipFn Determines if a test should be skipped due to lack of hugepage pods
//
// This function receives a testing environment and returns another function
// that, when called, checks whether any pods are requesting hugepages. If none
// exist, it signals the test framework to skip with an explanatory message.
// Otherwise, it allows the test to proceed normally.
func GetNoHugepagesPodsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.GetHugepagesPods()) == 0 {
			return true, "no pods requesting hugepages found"
		}
		return false, ""
	}
}

// GetNoCatalogSourcesSkipFn Determines whether to skip tests due to missing catalog sources
//
// The function returns a closure that checks the test environment for catalog
// source entries. If no catalog sources are present, it signals that the
// associated tests should be skipped with an explanatory message. Otherwise, it
// indicates that testing can proceed normally.
func GetNoCatalogSourcesSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.AllCatalogSources) == 0 {
			return true, "no catalog sources found"
		}
		return false, ""
	}
}

// GetNoOperatorsSkipFn Decides if a test should be skipped because no operators are present
//
// The function generates a closure that inspects the provided environment's
// operator list. If the list is empty, it signals to skip the test and supplies
// an explanatory message; otherwise it indicates the test can proceed.
func GetNoOperatorsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Operators) == 0 {
			return true, "no operators found"
		}
		return false, ""
	}
}

// GetNoOperatorPodsSkipFn Determines whether to skip tests due to missing operator pods
//
// The returned function checks the TestEnvironment's mapping of CSVs to pod
// lists. If no entries exist, it signals that tests should be skipped by
// returning true along with a message explaining that no operator pods were
// found. Otherwise, it indicates tests can proceed.
func GetNoOperatorPodsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.CSVToPodListMap) == 0 {
			return true, "no operator pods found"
		}

		return false, ""
	}
}

// GetNoOperatorCrdsSkipFn Skips tests when no operator CRDs are present
//
// The function takes a test environment and returns a closure used to decide
// whether a test should be skipped. The closure checks the length of the Crds
// slice in the environment; if it is empty, it signals to skip the test with an
// explanatory message. Otherwise, it indicates that the test should proceed.
func GetNoOperatorCrdsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Crds) == 0 {
			return true, "no operator crds found"
		}
		return false, ""
	}
}

// GetNoNodesWithRealtimeKernelSkipFn Skips tests when no node uses a realtime kernel
//
// This helper returns a function that checks all nodes in the test environment
// for a realtime kernel type. If any node is found to use such a kernel, the
// returned function signals not to skip; otherwise it indicates a skip with an
// explanatory message.
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

// ResultObjectsToString Serializes compliant and non‑compliant report objects into a JSON string
//
// The function receives two slices of ReportObject values, one for compliant
// items and another for non‑compliant ones. It constructs a FailureReasonOut
// structure containing these slices, marshals the structure to JSON, and
// returns the resulting string. If the marshalling fails, an error is returned
// with context.
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
