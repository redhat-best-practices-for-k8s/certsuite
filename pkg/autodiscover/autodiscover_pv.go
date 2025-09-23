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

package autodiscover

import (
	"context"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	storagev1typed "k8s.io/client-go/kubernetes/typed/storage/v1"
)

// getPersistentVolumes Retrieves all persistent volumes in the cluster
//
// The function calls the core V1 client to list PersistentVolume resources,
// returning a slice of those objects or an error if the API call fails.
func getPersistentVolumes(oc corev1client.CoreV1Interface) ([]corev1.PersistentVolume, error) {
	pvs, err := oc.PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pvs.Items, nil
}

// getPersistentVolumeClaims Retrieves all PersistentVolumeClaim objects from the cluster
//
// This function queries the Kubernetes API for every PersistentVolumeClaim
// across all namespaces, returning a slice of claim objects or an error if the
// request fails. It performs a List operation with no namespace filter and uses
// a context placeholder. The resulting claims are extracted from the response
// items field.
func getPersistentVolumeClaims(oc corev1client.CoreV1Interface) ([]corev1.PersistentVolumeClaim, error) {
	pvcs, err := oc.PersistentVolumeClaims("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pvcs.Items, nil
}

// getAllStorageClasses Retrieves all storage classes from the cluster
//
// The function queries the Kubernetes API for a list of StorageClass objects
// using the provided client interface. It returns the slice of discovered
// storage classes or an error if the list operation fails, logging any errors
// encountered.
func getAllStorageClasses(client storagev1typed.StorageV1Interface) ([]storagev1.StorageClass, error) {
	storageclasslist, err := client.StorageClasses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error("Error when listing storage classes, err: %v", err)
		return nil, err
	}
	return storageclasslist.Items, nil
}
