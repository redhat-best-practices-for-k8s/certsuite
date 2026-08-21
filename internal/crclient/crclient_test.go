// Copyright (C) 2020-2026 Red Hat, Inc.
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

package crclient

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetNodeProbePodContext(t *testing.T) {
	t.Parallel()

	usable := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "probe-a", Namespace: "cnf-suite"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "container-00"}}},
	}

	tests := []struct {
		name      string
		node      string
		probePods map[string]*corev1.Pod
		wantErr   bool
		wantPod   string
	}{
		{name: "missing node", node: "node-b", probePods: map[string]*corev1.Pod{"node-a": usable}, wantErr: true},
		{name: "empty node name", node: "", probePods: map[string]*corev1.Pod{"node-a": usable}, wantErr: true},
		{name: "nil probe", node: "node-a", probePods: map[string]*corev1.Pod{"node-a": nil}, wantErr: true},
		{name: "no containers", node: "node-a", probePods: map[string]*corev1.Pod{"node-a": {ObjectMeta: metav1.ObjectMeta{Name: "probe"}}}, wantErr: true},
		{name: "usable probe", node: "node-a", probePods: map[string]*corev1.Pod{"node-a": usable}, wantPod: "probe-a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, err := GetNodeProbePodContext(tt.node, &provider.TestEnvironment{ProbePods: tt.probePods})
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantPod, ctx.GetPodName())
			assert.Equal(t, "cnf-suite", ctx.GetNamespace())
			assert.Equal(t, "container-00", ctx.GetContainerName())
		})
	}
}
