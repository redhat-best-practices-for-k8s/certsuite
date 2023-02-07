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

type ObjectOut struct {
	ObjectType   string
	ObjectFields map[string]string
}

type FailureReasonOut struct {
	CompliantObjectsOut    []*ObjectOut
	NonCompliantObjectsOut []*ObjectOut
}

const (
	Namespace              = "Namespace"
	PodName                = "PodName"
	ContainerName          = "ContainerName"
	ProcessID              = "ProcessID"
	ProcessCommandLine     = "ProcessCommandLine"
	SchedulingPolicy       = "SchedulingPolicy"
	SchedulingPriority     = "SchedulingPriority"
	ReasonForNonCompliance = "ReasonForNonCompliance"
	ReasonForCompliance    = "ReasonForCompliance"
	Category               = "Category"
)

const (
	UndefinedType        = "Undefined Type"
	CnfType              = "Cnf"
	PodType              = "Pod"
	ContainerType        = "Container"
	ContainerProcessType = "ContainerProcess"
)

func NewProcessObjectOut(aContainer *ObjectOut, aPolicy, aPriority, aCommandLine string) (out *ObjectOut) {
	out = aContainer
	out.ObjectType = ContainerProcessType
	out.ObjectFields[ProcessCommandLine] = aCommandLine
	out.ObjectFields[SchedulingPolicy] = aPolicy
	out.ObjectFields[SchedulingPriority] = aPriority
	return out
}

func NewContainerObjectOut(aContainer *provider.Container, aReason string, isCompliant bool) (out *ObjectOut) {
	return NewContainerObjectOutBase(aContainer.Namespace, aContainer.Podname, aContainer.Name, aReason, isCompliant)
}

func NewContainerObjectOutBase(aNamespace, aPodName, aContainerName, aReason string, isCompliant bool) (out *ObjectOut) {
	out = New(aReason, isCompliant)
	out.ObjectType = ContainerType
	out.ObjectFields[Namespace] = aNamespace
	out.ObjectFields[PodName] = aPodName
	out.ObjectFields[ContainerName] = aContainerName
	return out
}

func NewPodObjectOut(aPod *provider.Pod, aCategory, aReason string, isCompliant bool) (out *ObjectOut) {
	out = New(aReason, isCompliant)
	out.ObjectType = PodType
	out.ObjectFields[Namespace] = aPod.Namespace
	out.ObjectFields[PodName] = aPod.Name
	out.ObjectFields[Category] = aCategory
	return out
}

func New(aReason string, isCompliant bool) (out *ObjectOut) {
	out = &ObjectOut{}
	out.ObjectType = UndefinedType
	out.ObjectFields = make(map[string]string)
	if isCompliant {
		out.ObjectFields[ReasonForCompliance] = aReason
	} else {
		out.ObjectFields[ReasonForNonCompliance] = aReason
	}
	return out
}

func (obj *ObjectOut) AddField(aKey, aString string) (out *ObjectOut) {
	obj.ObjectFields[aKey] = aString
	return obj
}

func (obj *ObjectOut) SetType(aType string) (out *ObjectOut) {
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

func AddTestResultReason(compliantObject, nonCompliantObject []*ObjectOut, fail func(string, ...int)) {
	var aReason FailureReasonOut
	aReason.CompliantObjectsOut = compliantObject
	aReason.NonCompliantObjectsOut = nonCompliantObject
	bytes, err := json.Marshal(aReason)
	if err != nil {
		logrus.Errorf("Could not Marshall FailureReason object, err=%s", err)
	}
	logrus.Info(string(bytes))
	if len(aReason.NonCompliantObjectsOut) > 0 {
		fail(string(bytes))
	}
}
