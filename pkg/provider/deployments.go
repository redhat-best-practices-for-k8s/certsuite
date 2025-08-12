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

// Deployment represents a Kubernetes deployment resource within the provider package.
//
// It embeds *appsv1.Deployment to provide direct access to all fields of the underlying
// Kubernetes Deployment object. The struct is used by the provider to manage and inspect
// deployments during certificate suite execution. Methods such as IsDeploymentReady
// determine if the deployment has reached a ready state, and ToString returns a human‑readable
// representation of the deployment's name, namespace, and status. This struct serves as
// the primary abstraction for interacting with Kubernetes deployments in the context
// of the certsuite provider.
type Deployment struct {
	*appsv1.Deployment
}

// IsDeploymentReady reports whether the Deployment has reached a ready state.
//
// It inspects the Deployment's status conditions and replica counts to determine
// if all desired replicas are available and not in progress of scaling or
// undergoing disruptions. The function returns true when the deployment is
// fully operational, otherwise false.
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

// ToString returns a human readable description of the Deployment.
//
// It formats the deployment's name, namespace and labels into a single string
// suitable for logging or debugging purposes. The returned value is the result
// of calling fmt.Sprintf on the deployment's fields.
func (d *Deployment) ToString() string {
	return fmt.Sprintf("deployment: %s ns: %s",
		d.Name,
		d.Namespace,
	)
}

// GetUpdatedDeployment retrieves the current Deployment object from the cluster and returns a deep copy that can be modified.
//
// It takes a client interface for the AppsV1 API, the name of the deployment, and its namespace.
// The function uses FindDeploymentByNameByNamespace to fetch the existing Deployment,
// then creates and returns a new Deployment instance containing the same spec and metadata.
// Errors from the lookup are returned unchanged.
func GetUpdatedDeployment(ac appv1client.AppsV1Interface, namespace, name string) (*Deployment, error) {
	result, err := autodiscover.FindDeploymentByNameByNamespace(ac, namespace, name)
	return &Deployment{
		result,
	}, err
}
