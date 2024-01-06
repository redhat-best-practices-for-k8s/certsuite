// Copyright (C) 2022 Red Hat, Inc.
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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetPersistentVolumes(t *testing.T) {
	generatePersistentVolume := func(name string) *corev1.PersistentVolume {
		return &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: corev1.PersistentVolumeSpec{},
		}
	}

	testCases := []struct {
		rqName      string
		expectedRQs []corev1.PersistentVolume
	}{
		{
			rqName: "test1",
			expectedRQs: []corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		var testRuntimeObjects []runtime.Object
		testRuntimeObjects = append(testRuntimeObjects, generatePersistentVolume(tc.rqName))
		oc := clientsholder.GetTestClientsHolder(testRuntimeObjects)
		PersistentVolumes, err := getPersistentVolumes(oc.K8sClient.CoreV1())
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedRQs[0].Name, PersistentVolumes[0].Name)
	}
}

func TestGetPersistentVolumeClaims(t *testing.T) {
	generatePersistentVolumeClaims := func(name string) *corev1.PersistentVolumeClaim {
		return &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: corev1.PersistentVolumeClaimSpec{},
		}
	}

	testCases := []struct {
		rqName      string
		expectedRQs []corev1.PersistentVolumeClaim
	}{
		{
			rqName: "test1",
			expectedRQs: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		var testRuntimeObjects []runtime.Object
		testRuntimeObjects = append(testRuntimeObjects, generatePersistentVolumeClaims(tc.rqName))
		oc := clientsholder.GetTestClientsHolder(testRuntimeObjects)
		PersistentVolumesClaims, err := getPersistentVolumeClaims(oc.K8sClient.CoreV1())
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedRQs[0].Name, PersistentVolumesClaims[0].Name)
	}
}

func TestGetAllStorageClasses(t *testing.T) {
	generateStorageClasses := func(name string) *storagev1.StorageClass {
		return &storagev1.StorageClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Provisioner: name,
		}
	}

	testCases := []struct {
		scName      string
		expectedRQs []storagev1.StorageClass
	}{
		{
			scName: "test1",
			expectedRQs: []storagev1.StorageClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
					Provisioner: "test1",
				},
			},
		},
	}

	for _, tc := range testCases {
		var testRuntimeObjects []runtime.Object
		testRuntimeObjects = append(testRuntimeObjects, generateStorageClasses(tc.scName))
		oc := clientsholder.GetTestClientsHolder(testRuntimeObjects)
		StorageClasses, err := getAllStorageClasses(oc.K8sClient.StorageV1())
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedRQs[0].Name, StorageClasses[0].Name)
	}
}
