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

// OwnerReference Tracks a pod's ownership status
//
// This structure stores a reference to a pod and an integer indicating the test
// outcome. The RunTest method examines each owner reference of the pod, logging
// information or errors based on whether the kind matches expected values such
// as StatefulSet or ReplicaSet. If any mismatches are found, it records a
// failure; otherwise, it marks success. GetResults simply returns the stored
// result value.
type OwnerReference struct {
	put    *corev1.Pod
	result int
}

// NewOwnerReference Creates a new owner reference checker for a Pod
//
// The function accepts a pointer to a Pod object and constructs an
// OwnerReference instance configured to evaluate the pod's owner references. It
// sets the initial result status to an error state, indicating that validation
// has not yet succeeded. The constructed instance is returned as a pointer so
// it can be used for further testing or result retrieval.
func NewOwnerReference(put *corev1.Pod) *OwnerReference {
	o := OwnerReference{
		put:    put,
		result: testhelper.ERROR,
	}
	return &o
}

// OwnerReference.RunTest verifies a pod’s owner references are either stateful set or replica set
//
// The method iterates over all owner references attached to the pod. For each
// reference it logs the kind and marks the test as successful if the kind
// matches one of the expected types; otherwise it logs an error, records
// failure, and stops further checks.
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

// OwnerReference.GetResults retrieves the stored result value
//
// The method returns the integer stored in the OwnerReference instance’s
// result field. It takes no arguments and simply accesses the private field to
// provide its current value.
func (o *OwnerReference) GetResults() int {
	return o.result
}
