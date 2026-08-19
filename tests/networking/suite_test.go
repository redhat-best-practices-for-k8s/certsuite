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
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
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
			name:        "out-of-range ports skipped",
			annotations: map[string]string{nonTLSPortsAnnotation: "0, 80, 70000"},
			expected:    map[int32]bool{80: true},
		},
		{
			name:        "unrelated annotation ignored",
			annotations: map[string]string{"other-annotation": "8080"},
			expected:    map[int32]bool{},
		},
		{
			name:        "min and max valid ports",
			annotations: map[string]string{nonTLSPortsAnnotation: "1, 65535"},
			expected:    map[int32]bool{1: true, 65535: true},
		},
		{
			name:        "trailing comma",
			annotations: map[string]string{nonTLSPortsAnnotation: "80,"},
			expected:    map[int32]bool{80: true},
		},
		{
			name:        "blank token skipped",
			annotations: map[string]string{nonTLSPortsAnnotation: "80, , 443"},
			expected:    map[int32]bool{80: true, 443: true},
		},
		{
			name:        "duplicate ports",
			annotations: map[string]string{nonTLSPortsAnnotation: "80,80"},
			expected:    map[int32]bool{80: true},
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
	assert.Equal(t, testhelper.ListeningPortType, ro.ObjectType)
	assert.Contains(t, ro.ObjectFieldsKeys, testhelper.PortNumber)
	assert.Contains(t, ro.ObjectFieldsKeys, testhelper.PortProtocol)
	assert.Contains(t, ro.ObjectFieldsKeys, testhelper.ReasonForCompliance)
	assert.Contains(t, ro.ObjectFieldsValues, "8080")
	assert.Contains(t, ro.ObjectFieldsValues, "TCP")
	assert.Contains(t, ro.ObjectFieldsValues, "port is TLS")
	assert.Contains(t, ro.ObjectFieldsValues, "test-ns")
	assert.Contains(t, ro.ObjectFieldsValues, "test-pod")

	nonCompliant := newPortReportObject("test-ns", "test-pod", "plaintext", false, port)
	assert.Equal(t, testhelper.ListeningPortType, nonCompliant.ObjectType)
	assert.Contains(t, nonCompliant.ObjectFieldsKeys, testhelper.ReasonForNonCompliance)
	assert.Contains(t, nonCompliant.ObjectFieldsValues, "plaintext")
}

func TestGetFirstProbeContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		probePods map[string]*corev1.Pod
		wantOK    bool
	}{
		{
			name:      "nil map",
			probePods: nil,
			wantOK:    false,
		},
		{
			name:      "empty map",
			probePods: map[string]*corev1.Pod{},
			wantOK:    false,
		},
		{
			name:      "nil probe pod",
			probePods: map[string]*corev1.Pod{"node1": nil},
			wantOK:    false,
		},
		{
			name: "probe pod with no containers",
			probePods: map[string]*corev1.Pod{
				"node1": {
					ObjectMeta: metav1.ObjectMeta{Name: "probe", Namespace: "cnf-certsuite"},
					Spec:       corev1.PodSpec{},
				},
			},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			env := &provider.TestEnvironment{ProbePods: tt.probePods}
			_, _, ok := getFirstProbeContext(env)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

func TestUnsecuredContainerPorts_NoProbePods(t *testing.T) {
	t.Parallel()

	check := checksdb.NewCheck("test-unsecured-ports", []string{"test"})
	env := &provider.TestEnvironment{ProbePods: map[string]*corev1.Pod{}}
	testUnsecuredContainerPorts(check, env)
	assert.Equal(t, checksdb.CheckResultSkipped, check.Result.String())
}
