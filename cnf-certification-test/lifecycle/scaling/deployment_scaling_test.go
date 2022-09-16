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

package scaling

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/podsets"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	v1app "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

//nolint:funlen
func TestScaleDeploymentFunc(t *testing.T) {
	generateDeployment := func(name string, replicas *int32) *v1app.Deployment {
		return &v1app.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "namespace1",
			},
			Spec: v1app.DeploymentSpec{
				Replicas: replicas,
			},
		}
	}

	testCases := []struct {
		deploymentName string
		replicaCount   int
	}{
		{
			deploymentName: "dp1",
			replicaCount:   3,
		},
		{
			deploymentName: "dp2",
			replicaCount:   0,
		},
	}

	// Always return that the Deployment is Ready
	origFunc := podsets.WaitForDeploymentSetReady
	defer func() {
		podsets.WaitForDeploymentSetReady = origFunc
	}()
	podsets.WaitForDeploymentSetReady = func(ns, name string, timeout time.Duration) bool {
		return true
	}

	for _, tc := range testCases {
		var runtimeObjects []runtime.Object
		intVar := new(int32)
		*intVar = int32(tc.replicaCount)
		tempDP := generateDeployment(tc.deploymentName, intVar)
		runtimeObjects = append(runtimeObjects, tempDP)
		c := clientsholder.GetTestClientsHolder(runtimeObjects)

		// Run the function
		TestScaleDeployment(tempDP, 10*time.Second)

		// Get the deployment from the fake API
		dp, err := c.K8sClient.AppsV1().Deployments("namespace1").Get(context.TODO(), tc.deploymentName, metav1.GetOptions{})
		assert.Nil(t, err)
		assert.Equal(t, int32(tc.replicaCount), *dp.Spec.Replicas)
	}
}
