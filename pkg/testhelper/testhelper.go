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

package testhelper

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"

	"github.com/test-network-function/cnf-certification-test/pkg/provider"
)

const (
	SUCCESS = iota
	FAILURE
	ERROR
)

type ReportObject struct {
	ObjectType         string
	ObjectFieldsKeys   []string
	ObjectFieldsValues []string
}

type FailureReasonOut struct {
	CompliantObjectsOut    []*ReportObject
	NonCompliantObjectsOut []*ReportObject
}

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

func FailureReasonOutTestString(p FailureReasonOut) (out string) {
	out = "testhelper.FailureReasonOut{"
	out += fmt.Sprintf("CompliantObjectsOut: %s,", ReportObjectTestStringPointer(p.CompliantObjectsOut))
	out += fmt.Sprintf("NonCompliantObjectsOut: %s,", ReportObjectTestStringPointer(p.NonCompliantObjectsOut))
	out += "}"
	return out
}

func ReportObjectTestStringPointer(p []*ReportObject) (out string) {
	out = "[]*testhelper.ReportObject{"
	for _, p := range p {
		out += fmt.Sprintf("&%#v,", *p)
	}
	out += "}"
	return out
}

func ReportObjectTestString(p []*ReportObject) (out string) {
	out = "[]testhelper.ReportObject{"
	for _, p := range p {
		out += fmt.Sprintf("%#v,", *p)
	}
	out += "}"
	return out
}

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
	DebugPodName                    = "Debug Pod Name"

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

	//
	SubscriptionName = "Subscription Name"
	OperatorPhase    = "Operator Phase"
)

// When adding new object types, please update the following:

const (
	UndefinedType                = "Undefined Type"
	CnfType                      = "Cnf"
	PodType                      = "Pod"
	HelmType                     = "Helm"
	OperatorType                 = "Operator"
	ContainerType                = "Container"
	ContainerImageType           = "Container Image"
	NodeType                     = "Node"
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

func (obj *ReportObject) SetContainerProcessValues(aPolicy, aPriority, aCommandLine string) *ReportObject {
	obj.AddField(ProcessCommandLine, aCommandLine)
	obj.AddField(SchedulingPolicy, aPolicy)
	obj.AddField(SchedulingPriority, aPriority)
	obj.ObjectType = ContainerProcessType
	return obj
}

func NewContainerReportObject(aNamespace, aPodName, aContainerName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, ContainerType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(PodName, aPodName)
	out.AddField(ContainerName, aContainerName)
	return out
}

func NewCertifiedContainerReportObject(cii provider.ContainerImageIdentifier, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, ContainerImageType, isCompliant)
	out.AddField(ImageDigest, cii.Digest)
	out.AddField(ImageRepo, cii.Repository)
	out.AddField(ImageTag, cii.Tag)
	out.AddField(ImageRegistry, cii.Registry)
	return out
}

func NewNodeReportObject(aNodeName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, NodeType, isCompliant)
	out.AddField(Name, aNodeName)
	return out
}

func NewTaintReportObject(taintBit, nodeName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, TaintType, isCompliant)
	out.AddField(NodeType, nodeName)
	out.AddField(TaintBit, taintBit)
	return out
}

func NewPodReportObject(aNamespace, aPodName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, PodType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(PodName, aPodName)
	return out
}

func NewHelmChartReportObject(aNamespace, aHelmChartName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, HelmType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(Name, aHelmChartName)
	return out
}

func NewOperatorReportObject(aNamespace, aOperatorName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, OperatorType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(Name, aOperatorName)
	return out
}

func NewDeploymentReportObject(aNamespace, aDeploymentName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, DeploymentType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(DeploymentName, aDeploymentName)
	return out
}

func NewStatefulSetReportObject(aNamespace, aStatefulSetName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, StatefulSetType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(StatefulSetName, aStatefulSetName)
	return out
}

func NewCrdReportObject(aName, aVersion, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, CustomResourceDefinitionType, isCompliant)
	out.AddField(CustomResourceDefinitionName, aName)
	out.AddField(CustomResourceDefinitionVersion, aVersion)
	return out
}

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

func (obj *ReportObject) AddField(aKey, aValue string) (out *ReportObject) {
	obj.ObjectFieldsKeys = append(obj.ObjectFieldsKeys, aKey)
	obj.ObjectFieldsValues = append(obj.ObjectFieldsValues, aValue)
	return obj
}

func NewNamespacedReportObject(aReason, aType string, isCompliant bool, aNamespace string) (out *ReportObject) {
	return NewReportObject(aReason, aType, isCompliant).AddField(Namespace, aNamespace)
}

func NewNamespacedNamedReportObject(aReason, aType string, isCompliant bool, aNamespace, aName string) (out *ReportObject) {
	return NewReportObject(aReason, aType, isCompliant).AddField(Namespace, aNamespace).AddField(Name, aName)
}

func (obj *ReportObject) SetType(aType string) (out *ReportObject) {
	obj.ObjectType = aType
	return obj
}

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

func GetNoContainersUnderTestSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Containers) == 0 {
			return true, "There are no containers to check. Please check under test labels."
		}

		return false, ""
	}
}

func GetNoPodsUnderTestSkipFn(env *provider.TestEnvironment) func() (bool, string) {
	return func() (bool, string) {
		if len(env.Pods) == 0 {
			return true, "There are no pods to check. Please check under test labels."
		}

		return false, ""
	}
}

func SkipIfEmptyAny(skip func(string, ...int), object ...[2]interface{}) {
	for _, o := range object {
		s := reflect.ValueOf(o[0])
		if s.Kind() != reflect.Slice && s.Kind() != reflect.Map {
			panic("SkipIfEmpty was given a non slice/map type")
		}
		if str, ok := o[1].(string); ok {
			if s.Len() == 0 {
				skip(fmt.Sprintf("Test skipped because there are no %s (%s) to test, please check under test labels", reflect.TypeOf(o[0]), str))
			}
		} else {
			panic("Value is not a string")
		}

		s = reflect.ValueOf(o[1])
		if s.Kind() != reflect.String {
			panic("SkipIfEmpty object name is not a string")
		}
	}
}

func SkipIfEmptyAll(skip func(string, ...int), object ...[2]interface{}) {
	countLenZero := 0
	allTypes := ""
	for _, o := range object {
		s := reflect.ValueOf(o[0])
		if s.Kind() != reflect.Slice && s.Kind() != reflect.Map {
			panic("SkipIfEmpty was given a non slice/map type")
		}

		if s.Len() == 0 {
			countLenZero++
			if str, ok := o[1].(string); ok {
				allTypes = allTypes + reflect.TypeOf(o[0]).String() + " (" + str + ")" + ", "
			} else {
				panic("Value is not a string")
			}
		}

		s = reflect.ValueOf(o[1])
		if s.Kind() != reflect.String {
			panic("SkipIfEmpty object name is not a string")
		}
	}
	// all objects have len() of 0
	if countLenZero == len(object) {
		skip(fmt.Sprintf("Test skipped because there are no %s to test, please check under test labels", allTypes))
	}
}

func NewSkipObject(object interface{}, name string) (skipObject [2]interface{}) {
	skipObject[0] = object
	skipObject[1] = name
	return skipObject
}

func AddTestResultLog(prefix string, object interface{}, log func(string, ...interface{}), fail func(string, ...int)) {
	s := reflect.ValueOf(object)
	if s.Kind() != reflect.Slice && s.Kind() != reflect.Map {
		panic("AddTestResultLog object param is a non slice/map type")
	}
	if s.Len() > 0 {
		log(fmt.Sprintf("%s %s: %v", prefix, reflect.TypeOf(object), object))
		fail(fmt.Sprintf("Number of %s %s = %d", prefix, reflect.TypeOf(object), s.Len()))
	}
}

func AddTestResultReason(compliantObject, nonCompliantObject []*ReportObject, log func(string, ...interface{}), fail func(string, ...int)) {
	var aReason FailureReasonOut
	aReason.CompliantObjectsOut = compliantObject
	aReason.NonCompliantObjectsOut = nonCompliantObject
	bytes, err := json.Marshal(aReason)
	if err != nil {
		logrus.Errorf("Could not Marshall FailureReason object, err=%s", err)
	}
	log(string(bytes))
	if len(aReason.NonCompliantObjectsOut) > 0 {
		fail(string(bytes))
	}
}

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
