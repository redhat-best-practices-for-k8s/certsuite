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

package networking

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/networking/netutil"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestParseNonTLSPortsAnnotation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		annotations map[string]string
		expected    map[int32]bool
	}{
		{
			name:        "no annotation",
			annotations: nil,
			expected:    map[int32]bool{},
		},
		{
			name:        "empty annotation value",
			annotations: map[string]string{nonTLSPortsAnnotation: ""},
			expected:    map[int32]bool{},
		},
		{
			name:        "single port",
			annotations: map[string]string{nonTLSPortsAnnotation: "8080"},
			expected:    map[int32]bool{8080: true},
		},
		{
			name:        "multiple ports with whitespace",
			annotations: map[string]string{nonTLSPortsAnnotation: "80, 443, 8080"},
			expected:    map[int32]bool{80: true, 443: true, 8080: true},
		},
		{
			name:        "invalid port skipped",
			annotations: map[string]string{nonTLSPortsAnnotation: "80, abc, 443"},
			expected:    map[int32]bool{80: true, 443: true},
		},
		{
			name:        "unrelated annotation ignored",
			annotations: map[string]string{"other-annotation": "8080"},
			expected:    map[int32]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pod := &provider.Pod{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "test-pod",
						Namespace:   "test-ns",
						Annotations: tt.annotations,
					},
				},
			}
			check := checksdb.NewCheck("test-parse-annotation", []string{"test"})
			result := parseNonTLSPortsAnnotation(pod, check)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewPortReportObject(t *testing.T) {
	t.Parallel()

	port := netutil.PortInfo{PortNumber: 8080, Protocol: "TCP"}
	ro := newPortReportObject("test-ns", "test-pod", "port is TLS", true, port)
	assert.NotNil(t, ro)
}
