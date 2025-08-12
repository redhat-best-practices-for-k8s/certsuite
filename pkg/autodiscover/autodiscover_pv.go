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

// getPersistentVolumes retrieves all PersistentVolume objects from the cluster.
//
// It takes a CoreV1Interface client and uses it to list persistent volumes via
// the List method on the PersistentVolumes resource. The function returns a slice
// of corev1.PersistentVolume objects and an error if the operation fails.
func getPersistentVolumes(oc corev1client.CoreV1Interface) ([]corev1.PersistentVolume, error) {
	pvs, err := oc.PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pvs.Items, nil
}

// getPersistentVolumeClaims retrieves all PersistentVolumeClaim objects in the cluster.
//
// It takes a Kubernetes CoreV1Interface client and returns a slice of PersistentVolumeClaim
// structs along with an error if the operation fails. The function lists all PVCs across
// namespaces using the client's List method on the PersistentVolumeClaims resource.
// ```
func getPersistentVolumeClaims(oc corev1client.CoreV1Interface) ([]corev1.PersistentVolumeClaim, error) {
	pvcs, err := oc.PersistentVolumeClaims("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pvcs.Items, nil
}

// getAllStorageClasses retrieves all StorageClass objects from the cluster.
//
// It accepts a client interface for the StorageV1 API and returns a slice of
// storagev1.StorageClass along with an error if the list operation fails.
// The function internally calls List on the StorageClasses resource to fetch
// the data.
func getAllStorageClasses(client storagev1typed.StorageV1Interface) ([]storagev1.StorageClass, error) {
	storageclasslist, err := client.StorageClasses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error("Error when listing storage classes, err: %v", err)
		return nil, err
	}
	return storageclasslist.Items, nil
}
