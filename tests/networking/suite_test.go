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
	"fmt"
	"strings"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
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

func testProbePod(name string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "cnf-suite"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "container-00"}},
		},
	}
}

func TestUsableProbeContexts(t *testing.T) {
	t.Parallel()

	usable := testProbePod("probe-a")
	tests := []struct {
		name      string
		probePods map[string]*corev1.Pod
		wantPods  map[string]string
	}{
		{name: "nil map", probePods: nil, wantPods: map[string]string{}},
		{name: "empty map", probePods: map[string]*corev1.Pod{}, wantPods: map[string]string{}},
		{name: "nil probe", probePods: map[string]*corev1.Pod{"node-a": nil}, wantPods: map[string]string{}},
		{name: "no containers", probePods: map[string]*corev1.Pod{"node-a": {ObjectMeta: metav1.ObjectMeta{Name: "probe"}}}, wantPods: map[string]string{}},
		{name: "usable probe", probePods: map[string]*corev1.Pod{"node-a": usable}, wantPods: map[string]string{"node-a": "probe-a"}},
		{name: "mixed nil and usable", probePods: map[string]*corev1.Pod{"node-a": nil, "node-b": usable}, wantPods: map[string]string{"node-b": "probe-a"}},
		{
			name: "two nodes",
			probePods: map[string]*corev1.Pod{
				"node-a": testProbePod("probe-a"),
				"node-b": testProbePod("probe-b"),
			},
			wantPods: map[string]string{"node-a": "probe-a", "node-b": "probe-b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := usableProbeContexts(&provider.TestEnvironment{ProbePods: tt.probePods})
			assert.Len(t, got, len(tt.wantPods))
			for node, wantPod := range tt.wantPods {
				ctx, ok := got[node]
				assert.True(t, ok, "missing context for node %s", node)
				assert.Equal(t, wantPod, ctx.GetPodName())
				assert.Equal(t, "cnf-suite", ctx.GetNamespace())
				assert.Equal(t, "container-00", ctx.GetContainerName())
			}
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

func TestUnsecuredContainerPorts_MissingNodeLocalProbe(t *testing.T) {
	t.Parallel()

	check := checksdb.NewCheck("test-unsecured-ports", []string{"test"})
	env := &provider.TestEnvironment{
		ProbePods: map[string]*corev1.Pod{
			"node-a": testProbePod("probe-a"),
		},
		Pods: []*provider.Pod{
			{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "workload", Namespace: "app-ns"},
					Spec:       corev1.PodSpec{NodeName: "node-b"},
				},
			},
		},
	}
	testUnsecuredContainerPorts(check, env)
	assert.Equal(t, checksdb.CheckResultFailed, check.Result.String())
	assert.Contains(t, check.GetLogs(), `No probe pod available on node "node-b"`)
}

func joinedReportValues(objs []*testhelper.ReportObject) string {
	var parts []string
	for _, o := range objs {
		parts = append(parts, o.ObjectFieldsValues...)
	}
	return strings.Join(parts, " ")
}

func TestCheckPodPortTLS(t *testing.T) {
	orig := getListeningPorts
	t.Cleanup(func() { getListeningPorts = orig })

	probeCtx := clientsholder.NewContext("cnf-suite", "probe-a", "container-00")
	app := &provider.Container{Container: &corev1.Container{Name: "app"}}
	istio := &provider.Container{Container: &corev1.Container{Name: provider.IstioProxyContainerName}}
	tcp8080 := netutil.PortInfo{PortNumber: 8080, Protocol: "TCP"}
	udp53 := netutil.PortInfo{PortNumber: 53, Protocol: "UDP"}
	istioPort := netutil.PortInfo{PortNumber: 15001, Protocol: "TCP"}

	newPod := func(ip string, anns map[string]string, containers ...*provider.Container) *provider.Pod {
		return &provider.Pod{
			Pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "workload", Namespace: "app-ns", Annotations: anns},
				Status:     corev1.PodStatus{PodIP: ip},
			},
			Containers: containers,
		}
	}

	tests := []struct {
		name                     string
		pod                      *provider.Pod
		ports                    map[netutil.PortInfo]bool
		portsErr                 error
		opensslStdout            string
		opensslErr               error
		wantCompliant            int
		wantNonCompliant         int
		wantCompliantContains    string
		wantNonCompliantContains string
		wantCalls                int
	}{
		{
			name:                     "listening ports error",
			pod:                      newPod("10.0.0.1", nil, app),
			portsErr:                 fmt.Errorf("ss failed"),
			wantNonCompliant:         1,
			wantNonCompliantContains: "Failed to get listening ports",
		},
		{
			name:                  "no listening ports",
			pod:                   newPod("10.0.0.1", nil, app),
			ports:                 map[netutil.PortInfo]bool{},
			wantCompliant:         1,
			wantCompliantContains: "No listening ports",
		},
		{
			name:                     "no containers",
			pod:                      newPod("10.0.0.1", nil),
			wantNonCompliant:         1,
			wantNonCompliantContains: "no containers",
		},
		{
			name:                     "empty pod IP",
			pod:                      newPod("", nil, app),
			ports:                    map[netutil.PortInfo]bool{tcp8080: true},
			wantNonCompliant:         1,
			wantNonCompliantContains: "no PodIP",
		},
		{
			name:  "udp only is ignored",
			pod:   newPod("10.0.0.1", nil, app),
			ports: map[netutil.PortInfo]bool{udp53: true},
		},
		{
			name:                  "exempt annotation",
			pod:                   newPod("10.0.0.1", map[string]string{nonTLSPortsAnnotation: "8080"}, app),
			ports:                 map[netutil.PortInfo]bool{tcp8080: true},
			wantCompliant:         1,
			wantCompliantContains: "exempt via annotation",
		},
		{
			name:                  "tls port",
			pod:                   newPod("10.0.0.1", nil, app),
			ports:                 map[netutil.PortInfo]bool{tcp8080: true},
			opensslStdout:         "CONNECTED(00000003)\n---\nProtocol  : TLSv1.3\nCipher    : TLS_AES_256_GCM_SHA384\n---",
			wantCompliant:         1,
			wantCompliantContains: "uses TLS",
			wantCalls:             1,
		},
		{
			name:                     "plaintext port",
			pod:                      newPod("10.0.0.1", nil, app),
			ports:                    map[netutil.PortInfo]bool{tcp8080: true},
			opensslStdout:            "CONNECTED(00000003)\n---\nCipher is (NONE)\npacket length too long\n---",
			wantNonCompliant:         1,
			wantNonCompliantContains: "plaintext",
			wantCalls:                1,
		},
		{
			name:                  "unreachable port",
			pod:                   newPod("10.0.0.1", nil, app),
			ports:                 map[netutil.PortInfo]bool{tcp8080: true},
			opensslStdout:         "connect:errno=111\nConnection refused",
			wantCompliant:         1,
			wantCompliantContains: "unreachable",
			wantCalls:             1,
		},
		{
			name:                  "exec probe failed treated as unreachable",
			pod:                   newPod("10.0.0.1", nil, app),
			ports:                 map[netutil.PortInfo]bool{tcp8080: true},
			opensslErr:            fmt.Errorf("command not found"),
			wantCompliant:         1,
			wantCompliantContains: "unreachable",
			wantCalls:             1,
		},
		{
			name:  "istio reserved port skipped",
			pod:   newPod("10.0.0.1", nil, app, istio),
			ports: map[netutil.PortInfo]bool{istioPort: true},
		},
		{
			name:                     "istio reserved port ignored, app port classified",
			pod:                      newPod("10.0.0.1", nil, app, istio),
			ports:                    map[netutil.PortInfo]bool{istioPort: true, tcp8080: true},
			opensslStdout:            "CONNECTED(00000003)\n---\nCipher is (NONE)\npacket length too long\n---",
			wantNonCompliant:         1,
			wantNonCompliantContains: "plaintext",
			wantCalls:                1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getListeningPorts = func(*provider.Container) (map[netutil.PortInfo]bool, error) {
				return tt.ports, tt.portsErr
			}
			mock := &clientsholder.MockCommand{
				ExecFunc: func(_ clientsholder.Context, _ string) (string, string, error) {
					return tt.opensslStdout, "", tt.opensslErr
				},
			}
			result := &checksdb.ParallelResult{}
			checkPodPortTLS(checksdb.NewCheck("test-unsecured-ports", []string{"test"}), tt.pod, mock, probeCtx, result)
			compliant, nonCompliant := result.Results()
			assert.Len(t, compliant, tt.wantCompliant)
			assert.Len(t, nonCompliant, tt.wantNonCompliant)
			if tt.wantCompliantContains != "" {
				assert.Contains(t, joinedReportValues(compliant), tt.wantCompliantContains)
			}
			if tt.wantNonCompliantContains != "" {
				assert.Contains(t, joinedReportValues(nonCompliant), tt.wantNonCompliantContains)
			}
			assert.Equal(t, tt.wantCalls, mock.CallCount())
		})
	}
}
