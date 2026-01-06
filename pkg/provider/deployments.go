// Copyright (C) 2022-2026 Red Hat, Inc.
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

package provider

import (
	"fmt"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/autodiscover"
	appsv1 "k8s.io/api/apps/v1"
	appv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
)

type Deployment struct {
	*appsv1.Deployment
}

func (d *Deployment) IsDeploymentReady() bool {
	notReady := true

	// Check the deployment's conditions for deploymentAvailable.
	for _, condition := range d.Status.Conditions {
		if condition.Type == appsv1.DeploymentAvailable {
			notReady = false // Deployment is ready
			break
		}
	}

	// Find the number of expected replicas
	var replicas int32
	if d.Spec.Replicas != nil {
		replicas = *(d.Spec.Replicas)
	} else {
		replicas = 1
	}

	// If condition says that the deployment is not ready or replicas do not match totals specified in spec.replicas.
	if notReady ||
		d.Status.UnavailableReplicas != 0 || //
		d.Status.ReadyReplicas != replicas || // eg. 10 ready replicas == 10 total replicas
		d.Status.AvailableReplicas != replicas ||
		d.Status.UpdatedReplicas != replicas {
		return false
	}
	return true
}

func (d *Deployment) ToString() string {
	return fmt.Sprintf("deployment: %s ns: %s",
		d.Name,
		d.Namespace,
	)
}

func GetUpdatedDeployment(ac appv1client.AppsV1Interface, namespace, name string) (*Deployment, error) {
	result, err := autodiscover.FindDeploymentByNameByNamespace(ac, namespace, name)
	return &Deployment{
		result,
	}, err
}
