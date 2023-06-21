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
	Namespace                    = "Namespace"
	PodName                      = "Pod Name"
	ContainerName                = "Container Name"
	ProcessID                    = "Process ID"
	ProcessCommandLine           = "Process CommandLine"
	SchedulingPolicy             = "Scheduling Policy"
	SchedulingPriority           = "Scheduling Priority"
	ReasonForNonCompliance       = "Reason For Non Compliance"
	ReasonForCompliance          = "Reason For Compliance"
	Category                     = "Category"
	ProjectedVolumeName          = "Projected Volume Name"
	ProjectedVolumeSAToken       = "Projected Volume SA Token"
	RoleBindingName              = "Role Binding Name"
	RoleBindingNamespace         = "Role Binding Namespace"
	ServiceAccountName           = "Service Account Name"
	ServiceMode                  = "Service Type"
	ServiceName                  = "Service Name"
	DeploymentName               = "Deployment Name"
	StatefulSetName              = "StatefulSet Name"
	PodDisruptionBudgetReference = "Pod Disruption Budget Reference"
	CrdName                      = "Crd Name"
	CrdVersion                   = "Crd Name"

	// ICMP tests
	NetworkName              = "Network Name"
	DestinationNamespace     = "Destination Namespace"
	DestinationPodName       = "Destination Pod Name"
	DestinationContainerName = "Destination Container Name"
	DestinationIP            = "Destination IP"
	SourceIP                 = "Source IP"
)

// When adding new object types, please update the following:

const (
	UndefinedType        = "Undefined Type"
	CnfType              = "Cnf"
	PodType              = "Pod"
	ContainerType        = "Container"
	ContainerProcessType = "ContainerProcess"
	ContainerCategory    = "ContainerCategory"
	ProjectedVolumeType  = "ProjectedVolume"
	ServiceType          = "Service"
	DeploymentType       = "Deployment"
	StatefulSetType      = "StatefulSet"
	ICMPResultType       = "ICMP result"
	NetworkType          = "Network"
	CrdType              = "Crd"
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

func NewPodReportObject(aNamespace, aPodName, aReason string, isCompliant bool) (out *ReportObject) {
	out = NewReportObject(aReason, PodType, isCompliant)
	out.AddField(Namespace, aNamespace)
	out.AddField(PodName, aPodName)
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

func SkipIfEmptyAny(skip func(string, ...int), object ...interface{}) {
	for _, o := range object {
		s := reflect.ValueOf(o)
		if s.Kind() != reflect.Slice && s.Kind() != reflect.Map {
			panic("SkipIfEmpty was given a non slice/map type")
		}

		if s.Len() == 0 {
			skip(fmt.Sprintf("Test skipped because there are no %s to test, please check under test labels", reflect.TypeOf(o)))
		}
	}
}

func SkipIfEmptyAll(skip func(string, ...int), object ...interface{}) {
	countLenZero := 0
	allTypes := ""
	for _, o := range object {
		s := reflect.ValueOf(o)
		if s.Kind() != reflect.Slice && s.Kind() != reflect.Map {
			panic("SkipIfEmpty was given a non slice/map type")
		}

		if s.Len() == 0 {
			countLenZero++
			allTypes = allTypes + reflect.TypeOf(o).String() + ", "
		}
	}
	// all objects have len() of 0
	if countLenZero == len(object) {
		skip(fmt.Sprintf("Test skipped because there are no %s to test, please check under test labels", allTypes))
	}
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
