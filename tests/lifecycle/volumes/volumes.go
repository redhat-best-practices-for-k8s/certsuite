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

// getPVCFromSlice retrieves a PersistentVolumeClaim by name from a list
//
// This function iterates over the provided slice of claims, comparing each
// claim's name to the target name. If a match is found, it returns a pointer to
// that claim; otherwise, it returns nil to indicate no matching claim was
// present.
func getPVCFromSlice(pvcs []corev1.PersistentVolumeClaim, pvcName string) *corev1.PersistentVolumeClaim {
	for i := range pvcs {
		if pvcs[i].Name == pvcName {
			return &pvcs[i]
		}
	}
	return nil
}

// IsPodVolumeReclaimPolicyDelete Verifies that a pod volume’s reclaim policy is DELETE
//
// The function receives a pod volume, the cluster's persistent volumes, and
// persistent volume claims. It first finds the claim referenced by the volume,
// then checks if the corresponding persistent volume has a delete reclaim
// policy. If both conditions are satisfied, it returns true; otherwise false.
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
