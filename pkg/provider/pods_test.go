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

package provider

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/api/resource"
)

func TestPod_CheckResourceOnly2MiHugePages(t *testing.T) {
	tests := []struct {
		name string
		aPod Pod
		want bool
	}{
		{
			name: "pass",
			aPod: *generatePod(10, 10, 0, 0),
			want: true,
		},
		{
			name: "fail",
			aPod: *generatePod(10, 10, 1, 1),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.aPod
			got := p.CheckResourceOnly2MiHugePages()
			if got != tt.want {
				t.Errorf("Pod.CheckResourceOnly2MiHugePages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func generatePod(requestsValue2M, limitsValue2M, requestsValue1G, limitsValue1G int64) *Pod {
	aPod := Pod{
		Containers: []*Container{
			{
				Container: &corev1.Container{
					Name: "test1",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{},
						Limits:   corev1.ResourceList{}}},
			},
		},
	}
	var aQuantity v1.Quantity
	if requestsValue2M != 0 {
		aQuantity.Set(requestsValue2M)
		aPod.Containers[0].Resources.Requests[hugePages2Mi] = aQuantity
	}
	if limitsValue2M != 0 {
		aQuantity.Set(limitsValue2M)
		aPod.Containers[0].Resources.Limits[hugePages2Mi] = aQuantity
	}

	if requestsValue1G != 0 {
		aQuantity.Set(requestsValue1G)
		aPod.Containers[0].Resources.Requests[hugePages1Gi] = aQuantity
	}
	if limitsValue1G != 0 {
		aQuantity.Set(limitsValue1G)
		aPod.Containers[0].Resources.Limits[hugePages1Gi] = aQuantity
	}
	return &aPod
}
