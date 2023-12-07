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

package ownerreference

import (
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	corev1 "k8s.io/api/core/v1"
)

const (
	// statefulSet variable
	statefulSet = "StatefulSet"
	// replicaSet variable
	replicaSet = "ReplicaSet"
)

type OwnerReference struct {
	put    *corev1.Pod
	result int
}

func NewOwnerReference(put *corev1.Pod) *OwnerReference {
	o := OwnerReference{
		put:    put,
		result: testhelper.ERROR,
	}
	return &o
}

// func (o *OwnerReference)  run the tests and store results in
// o.result
func (o *OwnerReference) RunTest() {
	for _, k := range o.put.OwnerReferences {
		log.Debug("kind is %s", k.Kind)
		if k.Kind == statefulSet || k.Kind == replicaSet {
			o.result = testhelper.SUCCESS
		} else {
			log.Error("Pod %s has owner of type %s", o.put.Name, k.Kind)
			o.result = testhelper.FAILURE
			return
		}
	}
}

// GetResults return result of the OwnerReference type
func (o *OwnerReference) GetResults() int {
	return o.result
}
