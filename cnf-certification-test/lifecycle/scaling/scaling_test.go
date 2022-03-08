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

package scaling

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	v1app "k8s.io/api/apps/v1"
)

func TestIsDeploymentRead(t *testing.T) {
	type dpStatus struct {
		condition   v1app.DeploymentConditionType
		replicas    int32
		ready       int32
		available   int32
		unavailable int32
		updated     int32
	}
	m := map[dpStatus]bool{
		{v1app.DeploymentReplicaFailure, 10, 9, 10, 0, 0}: false,
		{v1app.DeploymentAvailable, 10, 9, 9, 0, 10}:      false,
		{v1app.DeploymentAvailable, 10, 10, 10, 1, 10}:    false,
		{v1app.DeploymentAvailable, 10, 1, 10, 0, 10}:     false,
		{v1app.DeploymentAvailable, 10, 10, 10, 0, 9}:     false,
		{v1app.DeploymentAvailable, 10, 10, 10, 0, 10}:    true,
	}
	for key, v := range m {
		dp := v1app.Deployment{}
		dpCondition := v1app.DeploymentCondition{Type: key.condition}
		dp.Status.Conditions = append(dp.Status.Conditions, dpCondition)
		dp.Spec.Replicas = &(key.replicas)
		dp.Status.ReadyReplicas = key.ready
		dp.Status.AvailableReplicas = key.available
		dp.Status.UnavailableReplicas = key.unavailable
		dp.Status.UpdatedReplicas = key.updated
		ready := isDeploymentInstanceReady(&dp)
		assert.Equal(t, v, ready)
	}
}

func TestIsStatefulSetReady(t *testing.T) {
	type stStatus struct {
		replicas  int32
		ready     int32
		available int32
		updated   int32
	}
	m := map[stStatus]bool{
		{10, 9, 10, 0}:   false,
		{10, 9, 9, 10}:   false,
		{10, 10, 10, 11}: false,
		{10, 1, 10, 10}:  false,
		{10, 10, 10, 9}:  false,
		{10, 10, 10, 10}: true,
	}
	for k, v := range m {
		statefulset := v1app.StatefulSet{}
		statefulset.Spec.Replicas = &(k.replicas)
		statefulset.Status.ReadyReplicas = k.ready
		statefulset.Status.AvailableReplicas = k.available
		statefulset.Status.UpdatedReplicas = k.updated
		ready := isStatefulSetReady(&statefulset)
		if ready != v {
			fmt.Println(" k= ", k, " should be ", v, " is ", ready)
		}
		assert.Equal(t, v, ready)
	}
}
