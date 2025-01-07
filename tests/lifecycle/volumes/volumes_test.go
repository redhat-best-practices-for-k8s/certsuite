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

package volumes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsPodVolumeReclaimPolicyDelete(t *testing.T) {
	testCases := []struct {
		testVolume     corev1.Volume
		testPVs        []corev1.PersistentVolume
		testPVCs       []corev1.PersistentVolumeClaim
		expectedResult bool
	}{
		{ // Test Case #1 - Happy Path, "test1-volume" maps all the way back to a PV that has a DELETE reclaim policy
			testVolume: corev1.Volume{
				Name: "test1-volume",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: "test1-volume1-claim",
					},
				},
			},
			testPVs: []corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pv-1",
					},
					Spec: corev1.PersistentVolumeSpec{
						PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
					},
				},
			},
			testPVCs: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1-volume1-claim",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "pv-1",
					},
				},
			},
			expectedResult: true,
		},
		{ // Test Case #2 - "test1-volume1-claim" does not exist, fail
			testVolume: corev1.Volume{
				Name: "test1-volume",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: "test1-volume1-claim", // does not exist
					},
				},
			},
			testPVs: []corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pv-1",
					},
					Spec: corev1.PersistentVolumeSpec{
						PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
					},
				},
			},
			testPVCs:       []corev1.PersistentVolumeClaim{},
			expectedResult: false,
		},
		{ // Test Case #3 - All resources exist, however the reclaim policy is NOT delete
			testVolume: corev1.Volume{
				Name: "test1-volume",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: "test1-volume1-claim",
					},
				},
			},
			testPVs: []corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pv-1",
					},
					Spec: corev1.PersistentVolumeSpec{
						PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain, // set to retain
					},
				},
			},
			testPVCs: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1-volume1-claim",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "pv-1",
					},
				},
			},
			expectedResult: false,
		},
	}

	for index, tc := range testCases {
		assert.Equal(t, tc.expectedResult, IsPodVolumeReclaimPolicyDelete(&testCases[index].testVolume, tc.testPVs, tc.testPVCs))
	}
}
