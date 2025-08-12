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

package ownerreference

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	corev1 "k8s.io/api/core/v1"
)

const (
	// statefulSet variable
	statefulSet = "StatefulSet"
	// replicaSet variable
	replicaSet = "ReplicaSet"
)

// OwnerReference holds a Kubernetes pod and test results.
//
// OwnerReference holds a reference to a pod and tracks the outcome of its lifecycle tests.
// It stores the pod to be tested in the put field and records an integer result after running tests.
// The RunTest method executes the test suite on the pod, logging progress and errors,
// and updates the result field. GetResults returns the stored result value.
type OwnerReference struct {
	put    *corev1.Pod
	result int
}

// NewOwnerReference creates an OwnerReference object from a Kubernetes pod.
//
// It takes a pointer to a corev1.Pod and constructs an OwnerReference that
// references the pod as its owner. The returned *OwnerReference can be used
// to set ownership relationships in other Kubernetes objects, ensuring
// proper garbage collection and lifecycle management.
func NewOwnerReference(put *corev1.Pod) *OwnerReference {
	o := OwnerReference{
		put:    put,
		result: testhelper.ERROR,
	}
	return &o
}

// RunTest executes the owner reference tests and stores the results.
//
// It accepts a logger to record progress and errors during execution.
// The function runs all configured tests, logs relevant information,
// and updates the OwnerReference's result field with the outcome.
func (o *OwnerReference) RunTest(logger *log.Logger) {
	for _, k := range o.put.OwnerReferences {
		if k.Kind == statefulSet || k.Kind == replicaSet {
			logger.Info("Pod %q owner reference kind is %q", o.put, k.Kind)
			o.result = testhelper.SUCCESS
		} else {
			logger.Error("Pod %q has owner of type %q (%q or %q expected)", o.put, k.Kind, replicaSet, statefulSet)
			o.result = testhelper.FAILURE
			return
		}
	}
}

// GetResults retrieves the current result count for an OwnerReference instance.
//
// It returns an integer representing how many times a particular event or check has been recorded
// within this OwnerReference. The value reflects internal state maintained by the instance and
// can be used to assert correctness in tests or monitor progress during execution.
func (o *OwnerReference) GetResults() int {
	return o.result
}
