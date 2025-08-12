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
	corev1 "k8s.io/api/core/v1"
)

func getPVCFromSlice(pvcs []corev1.PersistentVolumeClaim, pvcName string) *corev1.PersistentVolumeClaim {
	for i := range pvcs {
		if pvcs[i].Name == pvcName {
			return &pvcs[i]
		}
	}
	return nil
}

func IsPodVolumeReclaimPolicyDelete(vol *corev1.Volume, pvs []corev1.PersistentVolume, pvcs []corev1.PersistentVolumeClaim) bool {
	// Check if the Volume is bound to a PVC.
	if putPVC := getPVCFromSlice(pvcs, vol.PersistentVolumeClaim.ClaimName); putPVC != nil {
		// Loop through the PersistentVolumes in the cluster, looking for bound PV/PVCs.
		for pvIndex := range pvs {
			// Check to make sure its reclaim policy is DELETE.
			if putPVC.Spec.VolumeName == pvs[pvIndex].Name && pvs[pvIndex].Spec.PersistentVolumeReclaimPolicy == corev1.PersistentVolumeReclaimDelete {
				return true
			}
		}
	}

	return false
}
