// Copyright (C) 2022-2024 Red Hat, Inc.
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

// Deployment Represents a Kubernetes deployment with helper methods
//
// This type wraps the standard appsv1.Deployment object to provide convenient
// operations such as checking readiness and generating a string representation.
// It exposes the embedded Deployment fields directly while adding methods that
// evaluate status conditions and replica counts for quick health checks.
type Deployment struct {
	*appsv1.Deployment
}

// Deployment.IsDeploymentReady Determines whether a deployment has reached the desired state
//
// It inspects the deployment’s status conditions to see if an available
// condition is present, then compares replica counts from the spec with various
// status fields such as unavailable, ready, available, and updated replicas. If
// any of these checks fail, it returns false; otherwise true.
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

// Deployment.ToString Formats deployment details into a human‑readable string
//
// This method creates a concise representation of a Deployment by combining its
// name and namespace. It uses standard formatting to return the result as a
// single string, which can be printed or logged for debugging purposes.
func (d *Deployment) ToString() string {
	return fmt.Sprintf("deployment: %s ns: %s",
		d.Name,
		d.Namespace,
	)
}

// GetUpdatedDeployment Retrieves the latest state of a Kubernetes deployment
//
// The function queries the cluster for a specific deployment in a given
// namespace, then wraps the result in a custom Deployment type that exposes
// helper methods. It returns a pointer to this wrapper and an error if the
// lookup fails or the API call encounters an issue.
func GetUpdatedDeployment(ac appv1client.AppsV1Interface, namespace, name string) (*Deployment, error) {
	result, err := autodiscover.FindDeploymentByNameByNamespace(ac, namespace, name)
	return &Deployment{
		result,
	}, err
}
