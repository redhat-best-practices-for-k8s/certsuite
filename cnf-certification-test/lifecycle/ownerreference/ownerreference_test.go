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

package ownerreference_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/ownerreference"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRunTest(t *testing.T) {
	// instantiate test pod
	pod := v1.Pod{}
	// test when pods owner reference is not specified
	o := ownerreference.NewOwnerReference(&pod)
	o.RunTest()
	assert.Equal(t, testhelper.ERROR, o.GetResults())
	// test when pods owner is not StatefulSet Or ReplicaSet
	pod.OwnerReferences = append(pod.OwnerReferences, metav1.OwnerReference{Kind: "bad_kind"})
	o.RunTest()
	assert.Equal(t, testhelper.FAILURE, o.GetResults())
	// test when pods owner is StatefulSet Or ReplicaSet
	pod.OwnerReferences = pod.OwnerReferences[1:]
	pod.OwnerReferences = append(pod.OwnerReferences, metav1.OwnerReference{Kind: "StatefulSet"}, metav1.OwnerReference{Kind: "ReplicaSet"})
	o.RunTest()
	assert.Equal(t, testhelper.SUCCESS, o.GetResults())

	// test when pods owner is StatefulSet And bad kind
	pod.OwnerReferences = append(pod.OwnerReferences, metav1.OwnerReference{Kind: "bad_kind"})
	o.RunTest()
	assert.Equal(t, testhelper.SUCCESS, o.GetResults())
}
