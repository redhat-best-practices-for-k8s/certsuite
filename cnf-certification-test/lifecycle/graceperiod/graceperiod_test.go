// Copyright (C) 2020-2021 Red Hat, Inc.
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

package graceperiod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/loghelper"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	v1app "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_getTerminationGracePeriodConfiguredInYaml(t *testing.T) {
	type args struct {
		lastAppliedConfigString string
	}
	tests := []struct {
		name                              string
		args                              args
		wantTerminationGracePeriodSeconds int
		wantErr                           bool
	}{
		{
			name:                              "ok",
			args:                              args{lastAppliedConfigString: "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"annotations\":{},\"name\":\"test\",\"namespace\":\"tnf\"},\"spec\":{\"replicas\":2,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"template\":{\"metadata\":{\"annotations\":{\"k8s.v1.cni.cncf.io/networks\":\"[ { \\\"name\\\" : \\\"mynet-ipv4-0\\\" },{ \\\"name\\\" : \\\"mynet-ipv4-1\\\" },{ \\\"name\\\" : \\\"mynet-ipv6-0\\\" },{ \\\"name\\\" : \\\"mynet-ipv6-1\\\" } ]\",\"test-network-function.com/container_tests\":\"[\\\"PRIVILEGED_POD\\\",\\\"PRIVILEGED_ROLE\\\"]\",\"test-network-function.com/defaultnetworkinterface\":\"\\\"eth0\\\"\"},\"labels\":{\"app\":\"test\",\"test-network-function.com/container\":\"target\",\"test-network-function.com/generic\":\"target\"},\"name\":\"test\"},\"spec\":{\"affinity\":{\"podAntiAffinity\":{\"requiredDuringSchedulingIgnoredDuringExecution\":[{\"labelSelector\":{\"matchExpressions\":[{\"key\":\"app\",\"operator\":\"In\",\"values\":[\"test\"]}]},\"topologyKey\":\"kubernetes.io/hostname\"}]}},\"automountServiceAccountToken\":false,\"containers\":[{\"command\":[\"./bin/app\"],\"image\":\"quay.io/testnetworkfunction/cnf-test-partner:latest\",\"imagePullPolicy\":\"IfNotPresent\",\"lifecycle\":{\"preStop\":{\"exec\":{\"command\":[\"/bin/sh\",\"-c\",\"killall -0 tail\"]}}},\"livenessProbe\":{\"httpGet\":{\"httpHeaders\":[{\"name\":\"health-check\",\"value\":\"liveness\"}],\"path\":\"/health\",\"port\":8080},\"initialDelaySeconds\":10,\"periodSeconds\":5},\"name\":\"test\",\"ports\":[{\"containerPort\":8080,\"name\":\"http-probe\"}],\"readinessProbe\":{\"httpGet\":{\"httpHeaders\":[{\"name\":\"health-check\",\"value\":\"readiness\"}],\"path\":\"/ready\",\"port\":8080},\"initialDelaySeconds\":10,\"periodSeconds\":5},\"resources\":{\"limits\":{\"cpu\":0.25,\"memory\":\"512Mi\"}}}],\"terminationGracePeriodSeconds\":30}}}}\n"}, //nolint:lll
			wantTerminationGracePeriodSeconds: 30,
			wantErr:                           false,
		},
		{
			name:                              "ok",
			args:                              args{lastAppliedConfigString: "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"annotations\":{},\"name\":\"test\",\"namespace\":\"tnf\"},\"spec\":{\"replicas\":2,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"template\":{\"metadata\":{\"annotations\":{\"k8s.v1.cni.cncf.io/networks\":\"[ { \\\"name\\\" : \\\"mynet-ipv4-0\\\" },{ \\\"name\\\" : \\\"mynet-ipv4-1\\\" },{ \\\"name\\\" : \\\"mynet-ipv6-0\\\" },{ \\\"name\\\" : \\\"mynet-ipv6-1\\\" } ]\",\"test-network-function.com/container_tests\":\"[\\\"PRIVILEGED_POD\\\",\\\"PRIVILEGED_ROLE\\\"]\",\"test-network-function.com/defaultnetworkinterface\":\"\\\"eth0\\\"\"},\"labels\":{\"app\":\"test\",\"test-network-function.com/container\":\"target\",\"test-network-function.com/generic\":\"target\"},\"name\":\"test\"},\"spec\":{\"affinity\":{\"podAntiAffinity\":{\"requiredDuringSchedulingIgnoredDuringExecution\":[{\"labelSelector\":{\"matchExpressions\":[{\"key\":\"app\",\"operator\":\"In\",\"values\":[\"test\"]}]},\"topologyKey\":\"kubernetes.io/hostname\"}]}},\"automountServiceAccountToken\":false,\"containers\":[{\"command\":[\"./bin/app\"],\"image\":\"quay.io/testnetworkfunction/cnf-test-partner:latest\",\"imagePullPolicy\":\"IfNotPresent\",\"lifecycle\":{\"preStop\":{\"exec\":{\"command\":[\"/bin/sh\",\"-c\",\"killall -0 tail\"]}}},\"livenessProbe\":{\"httpGet\":{\"httpHeaders\":[{\"name\":\"health-check\",\"value\":\"liveness\"}],\"path\":\"/health\",\"port\":8080},\"initialDelaySeconds\":10,\"periodSeconds\":5},\"name\":\"test\",\"ports\":[{\"containerPort\":8080,\"name\":\"http-probe\"}],\"readinessProbe\":{\"httpGet\":{\"httpHeaders\":[{\"name\":\"health-check\",\"value\":\"readiness\"}],\"path\":\"/ready\",\"port\":8080},\"initialDelaySeconds\":10,\"periodSeconds\":5},\"resources\":{\"limits\":{\"cpu\":0.25,\"memory\":\"512Mi\"}}}]}}}}\n"}, //nolint:lll
			wantTerminationGracePeriodSeconds: -1,
			wantErr:                           false,
		},
		{
			name:                              "ok",
			args:                              args{lastAppliedConfigString: ""},
			wantTerminationGracePeriodSeconds: -1,
			wantErr:                           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTerminationGracePeriodSeconds, err := getTerminationGracePeriodConfiguredInYaml(tt.args.lastAppliedConfigString)
			if (err != nil) != tt.wantErr {
				t.Errorf("getTerminationGracePeriodConfiguredInYaml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotTerminationGracePeriodSeconds != tt.wantTerminationGracePeriodSeconds {
				t.Errorf("getTerminationGracePeriodConfiguredInYaml() = %v, want %v", gotTerminationGracePeriodSeconds, tt.wantTerminationGracePeriodSeconds)
			}
		})
	}
}

//nolint:funlen
func TestTestTerminationGracePeriodOnPods(t *testing.T) {
	generateEnv := func(podName, annotation string, gracePeriod int64) *provider.TestEnvironment {
		return &provider.TestEnvironment{
			Pods: []*corev1.Pod{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: podName,
						Annotations: map[string]string{
							"kubectl.kubernetes.io/last-applied-configuration": annotation,
						},
					},
					Spec: corev1.PodSpec{
						TerminationGracePeriodSeconds: &gracePeriod,
					},
				},
			},
		}
	}

	testCases := []struct {
		podName              string
		annotation           string
		expectedBadPods      int
		expectedCuratedLines []string
		gracePeriodSecs      int64
	}{
		{ // Test Case #1 - No terminationGracePeriodSeconds value set, default to 30 seconds
			podName:    "nginx-deployment",
			annotation: "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"annotations\":{},\"name\":\"test\",\"namespace\":\"tnf\"},\"spec\":{\"replicas\":2,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"template\":{\"metadata\":{\"annotations\":{\"k8s.v1.cni.cncf.io/networks\":\"[ { \\\"name\\\" : \\\"mynet-ipv4-0\\\" },{ \\\"name\\\" : \\\"mynet-ipv4-1\\\" },{ \\\"name\\\" : \\\"mynet-ipv6-0\\\" },{ \\\"name\\\" : \\\"mynet-ipv6-1\\\" } ]\",\"test-network-function.com/container_tests\":\"[\\\"PRIVILEGED_POD\\\",\\\"PRIVILEGED_ROLE\\\"]\",\"test-network-function.com/defaultnetworkinterface\":\"\\\"eth0\\\"\"},\"labels\":{\"app\":\"test\",\"test-network-function.com/container\":\"target\",\"test-network-function.com/generic\":\"target\"},\"name\":\"test\"},\"spec\":{\"affinity\":{\"podAntiAffinity\":{\"requiredDuringSchedulingIgnoredDuringExecution\":[{\"labelSelector\":{\"matchExpressions\":[{\"key\":\"app\",\"operator\":\"In\",\"values\":[\"test\"]}]},\"topologyKey\":\"kubernetes.io/hostname\"}]}},\"automountServiceAccountToken\":false,\"containers\":[{\"command\":[\"./bin/app\"],\"image\":\"quay.io/testnetworkfunction/cnf-test-partner:latest\",\"imagePullPolicy\":\"IfNotPresent\",\"lifecycle\":{\"preStop\":{\"exec\":{\"command\":[\"/bin/sh\",\"-c\",\"killall -0 tail\"]}}},\"livenessProbe\":{\"httpGet\":{\"httpHeaders\":[{\"name\":\"health-check\",\"value\":\"liveness\"}],\"path\":\"/health\",\"port\":8080},\"initialDelaySeconds\":10,\"periodSeconds\":5},\"name\":\"test\",\"ports\":[{\"containerPort\":8080,\"name\":\"http-probe\"}],\"readinessProbe\":{\"httpGet\":{\"httpHeaders\":[{\"name\":\"health-check\",\"value\":\"readiness\"}],\"path\":\"/ready\",\"port\":8080},\"initialDelaySeconds\":10,\"periodSeconds\":5},\"resources\":{\"limits\":{\"cpu\":0.25,\"memory\":\"512Mi\"}}}]}}}}\n", //nolint:lll
			expectedCuratedLines: []string{
				"pod: nginx-deployment ns:  does not have a terminationGracePeriodSeconds value set. Default value (30) is used.",
			},
			gracePeriodSecs: 30,
			expectedBadPods: 1,
		},
		{ // Test Case #2 - terminationGracePeriodSeconds is set to 30
			podName:              "nginx-deployment",
			annotation:           "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"annotations\":{},\"name\":\"test\",\"namespace\":\"tnf\"},\"spec\":{\"replicas\":2,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"template\":{\"metadata\":{\"annotations\":{\"k8s.v1.cni.cncf.io/networks\":\"[ { \\\"name\\\" : \\\"mynet-ipv4-0\\\" },{ \\\"name\\\" : \\\"mynet-ipv4-1\\\" },{ \\\"name\\\" : \\\"mynet-ipv6-0\\\" },{ \\\"name\\\" : \\\"mynet-ipv6-1\\\" } ]\",\"test-network-function.com/container_tests\":\"[\\\"PRIVILEGED_POD\\\",\\\"PRIVILEGED_ROLE\\\"]\",\"test-network-function.com/defaultnetworkinterface\":\"\\\"eth0\\\"\"},\"labels\":{\"app\":\"test\",\"test-network-function.com/container\":\"target\",\"test-network-function.com/generic\":\"target\"},\"name\":\"test\"},\"spec\":{\"affinity\":{\"podAntiAffinity\":{\"requiredDuringSchedulingIgnoredDuringExecution\":[{\"labelSelector\":{\"matchExpressions\":[{\"key\":\"app\",\"operator\":\"In\",\"values\":[\"test\"]}]},\"topologyKey\":\"kubernetes.io/hostname\"}]}},\"automountServiceAccountToken\":false,\"containers\":[{\"command\":[\"./bin/app\"],\"image\":\"quay.io/testnetworkfunction/cnf-test-partner:latest\",\"imagePullPolicy\":\"IfNotPresent\",\"lifecycle\":{\"preStop\":{\"exec\":{\"command\":[\"/bin/sh\",\"-c\",\"killall -0 tail\"]}}},\"livenessProbe\":{\"httpGet\":{\"httpHeaders\":[{\"name\":\"health-check\",\"value\":\"liveness\"}],\"path\":\"/health\",\"port\":8080},\"initialDelaySeconds\":10,\"periodSeconds\":5},\"name\":\"test\",\"ports\":[{\"containerPort\":8080,\"name\":\"http-probe\"}],\"readinessProbe\":{\"httpGet\":{\"httpHeaders\":[{\"name\":\"health-check\",\"value\":\"readiness\"}],\"path\":\"/ready\",\"port\":8080},\"initialDelaySeconds\":10,\"periodSeconds\":5},\"resources\":{\"limits\":{\"cpu\":0.25,\"memory\":\"512Mi\"}}}],\"terminationGracePeriodSeconds\":30}}}}\n", //nolint:lll
			expectedCuratedLines: []string{},
			gracePeriodSecs:      30,
			expectedBadPods:      0,
		},
	}

	for _, tc := range testCases {
		cLogs := loghelper.CuratedLogLines{}
		for _, ll := range tc.expectedCuratedLines {
			cLogs = cLogs.AddLogLine(ll)
		}
		badPods, curatedLogs := TestTerminationGracePeriodOnPods(generateEnv(tc.podName, tc.annotation, tc.gracePeriodSecs))
		assert.Equal(t, cLogs.GetLogLines(), curatedLogs.GetLogLines())
		assert.Equal(t, tc.expectedBadPods, len(badPods))
	}
}

//nolint:funlen,dupl
func TestTestTerminationGracePeriodOnDeployments(t *testing.T) {
	generateEnv := func(deploymentName, annotation string, gracePeriod int64) *provider.TestEnvironment {
		return &provider.TestEnvironment{
			Deployments: []*v1app.Deployment{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: deploymentName,
						Annotations: map[string]string{
							"kubectl.kubernetes.io/last-applied-configuration": annotation,
						},
					},
					Spec: v1app.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								TerminationGracePeriodSeconds: &gracePeriod,
							},
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		podName                string
		annotation             string
		expectedBadDeployments int
		expectedCuratedLines   []string
		gracePeriodSecs        int64
	}{
		{ // Test Case #1 - No terminationGracePeriodSeconds value set, default to 30 seconds
			podName:    "nginx-deployment",
			annotation: "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"annotations\":{},\"name\":\"test\",\"namespace\":\"tnf\"},\"spec\":{\"replicas\":2,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"template\":{\"metadata\":{\"annotations\":{\"k8s.v1.cni.cncf.io/networks\":\"[ { \\\"name\\\" : \\\"mynet-ipv4-0\\\" },{ \\\"name\\\" : \\\"mynet-ipv4-1\\\" },{ \\\"name\\\" : \\\"mynet-ipv6-0\\\" },{ \\\"name\\\" : \\\"mynet-ipv6-1\\\" } ]\",\"test-network-function.com/container_tests\":\"[\\\"PRIVILEGED_POD\\\",\\\"PRIVILEGED_ROLE\\\"]\",\"test-network-function.com/defaultnetworkinterface\":\"\\\"eth0\\\"\"},\"labels\":{\"app\":\"test\",\"test-network-function.com/container\":\"target\",\"test-network-function.com/generic\":\"target\"},\"name\":\"test\"},\"spec\":{\"affinity\":{\"podAntiAffinity\":{\"requiredDuringSchedulingIgnoredDuringExecution\":[{\"labelSelector\":{\"matchExpressions\":[{\"key\":\"app\",\"operator\":\"In\",\"values\":[\"test\"]}]},\"topologyKey\":\"kubernetes.io/hostname\"}]}},\"automountServiceAccountToken\":false,\"containers\":[{\"command\":[\"./bin/app\"],\"image\":\"quay.io/testnetworkfunction/cnf-test-partner:latest\",\"imagePullPolicy\":\"IfNotPresent\",\"lifecycle\":{\"preStop\":{\"exec\":{\"command\":[\"/bin/sh\",\"-c\",\"killall -0 tail\"]}}},\"livenessProbe\":{\"httpGet\":{\"httpHeaders\":[{\"name\":\"health-check\",\"value\":\"liveness\"}],\"path\":\"/health\",\"port\":8080},\"initialDelaySeconds\":10,\"periodSeconds\":5},\"name\":\"test\",\"ports\":[{\"containerPort\":8080,\"name\":\"http-probe\"}],\"readinessProbe\":{\"httpGet\":{\"httpHeaders\":[{\"name\":\"health-check\",\"value\":\"readiness\"}],\"path\":\"/ready\",\"port\":8080},\"initialDelaySeconds\":10,\"periodSeconds\":5},\"resources\":{\"limits\":{\"cpu\":0.25,\"memory\":\"512Mi\"}}}]}}}}\n", //nolint:lll
			expectedCuratedLines: []string{
				"Deployment deployment: nginx-deployment ns:  does not have a terminationGracePeriodSeconds value set. Default value (30) is used.",
			},
			gracePeriodSecs: 30,
		},
		{ // Test Case #2 - terminationGracePeriodSeconds is set to 30
			podName:              "nginx-deployment",
			annotation:           "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"annotations\":{},\"name\":\"test\",\"namespace\":\"tnf\"},\"spec\":{\"replicas\":2,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"template\":{\"metadata\":{\"annotations\":{\"k8s.v1.cni.cncf.io/networks\":\"[ { \\\"name\\\" : \\\"mynet-ipv4-0\\\" },{ \\\"name\\\" : \\\"mynet-ipv4-1\\\" },{ \\\"name\\\" : \\\"mynet-ipv6-0\\\" },{ \\\"name\\\" : \\\"mynet-ipv6-1\\\" } ]\",\"test-network-function.com/container_tests\":\"[\\\"PRIVILEGED_POD\\\",\\\"PRIVILEGED_ROLE\\\"]\",\"test-network-function.com/defaultnetworkinterface\":\"\\\"eth0\\\"\"},\"labels\":{\"app\":\"test\",\"test-network-function.com/container\":\"target\",\"test-network-function.com/generic\":\"target\"},\"name\":\"test\"},\"spec\":{\"affinity\":{\"podAntiAffinity\":{\"requiredDuringSchedulingIgnoredDuringExecution\":[{\"labelSelector\":{\"matchExpressions\":[{\"key\":\"app\",\"operator\":\"In\",\"values\":[\"test\"]}]},\"topologyKey\":\"kubernetes.io/hostname\"}]}},\"automountServiceAccountToken\":false,\"containers\":[{\"command\":[\"./bin/app\"],\"image\":\"quay.io/testnetworkfunction/cnf-test-partner:latest\",\"imagePullPolicy\":\"IfNotPresent\",\"lifecycle\":{\"preStop\":{\"exec\":{\"command\":[\"/bin/sh\",\"-c\",\"killall -0 tail\"]}}},\"livenessProbe\":{\"httpGet\":{\"httpHeaders\":[{\"name\":\"health-check\",\"value\":\"liveness\"}],\"path\":\"/health\",\"port\":8080},\"initialDelaySeconds\":10,\"periodSeconds\":5},\"name\":\"test\",\"ports\":[{\"containerPort\":8080,\"name\":\"http-probe\"}],\"readinessProbe\":{\"httpGet\":{\"httpHeaders\":[{\"name\":\"health-check\",\"value\":\"readiness\"}],\"path\":\"/ready\",\"port\":8080},\"initialDelaySeconds\":10,\"periodSeconds\":5},\"resources\":{\"limits\":{\"cpu\":0.25,\"memory\":\"512Mi\"}}}],\"terminationGracePeriodSeconds\":30}}}}\n", //nolint:lll
			expectedCuratedLines: []string{},
			gracePeriodSecs:      29,
		},
	}

	for _, tc := range testCases {
		cLogs := loghelper.CuratedLogLines{}
		for _, ll := range tc.expectedCuratedLines {
			cLogs = cLogs.AddLogLine(ll)
		}
		_, curatedLogs := TestTerminationGracePeriodOnDeployments(generateEnv(tc.podName, tc.annotation, tc.gracePeriodSecs))
		assert.Equal(t, cLogs.GetLogLines(), curatedLogs.GetLogLines())
	}
}

//nolint:funlen,dupl
func TestTestTerminationGracePeriodOnStatefulsets(t *testing.T) {
	generateEnv := func(ssName, annotation string, gracePeriod int64) *provider.TestEnvironment {
		return &provider.TestEnvironment{
			StatetfulSets: []*v1app.StatefulSet{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: ssName,
						Annotations: map[string]string{
							"kubectl.kubernetes.io/last-applied-configuration": annotation,
						},
					},
					Spec: v1app.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								TerminationGracePeriodSeconds: &gracePeriod,
							},
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		podName                string
		annotation             string
		expectedBadDeployments int
		expectedCuratedLines   []string
		gracePeriodSecs        int64
	}{
		{ // Test Case #1 - No terminationGracePeriodSeconds value set, default to 30 seconds
			podName:    "nginx-deployment",
			annotation: "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"annotations\":{},\"name\":\"test\",\"namespace\":\"tnf\"},\"spec\":{\"replicas\":2,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"template\":{\"metadata\":{\"annotations\":{\"k8s.v1.cni.cncf.io/networks\":\"[ { \\\"name\\\" : \\\"mynet-ipv4-0\\\" },{ \\\"name\\\" : \\\"mynet-ipv4-1\\\" },{ \\\"name\\\" : \\\"mynet-ipv6-0\\\" },{ \\\"name\\\" : \\\"mynet-ipv6-1\\\" } ]\",\"test-network-function.com/container_tests\":\"[\\\"PRIVILEGED_POD\\\",\\\"PRIVILEGED_ROLE\\\"]\",\"test-network-function.com/defaultnetworkinterface\":\"\\\"eth0\\\"\"},\"labels\":{\"app\":\"test\",\"test-network-function.com/container\":\"target\",\"test-network-function.com/generic\":\"target\"},\"name\":\"test\"},\"spec\":{\"affinity\":{\"podAntiAffinity\":{\"requiredDuringSchedulingIgnoredDuringExecution\":[{\"labelSelector\":{\"matchExpressions\":[{\"key\":\"app\",\"operator\":\"In\",\"values\":[\"test\"]}]},\"topologyKey\":\"kubernetes.io/hostname\"}]}},\"automountServiceAccountToken\":false,\"containers\":[{\"command\":[\"./bin/app\"],\"image\":\"quay.io/testnetworkfunction/cnf-test-partner:latest\",\"imagePullPolicy\":\"IfNotPresent\",\"lifecycle\":{\"preStop\":{\"exec\":{\"command\":[\"/bin/sh\",\"-c\",\"killall -0 tail\"]}}},\"livenessProbe\":{\"httpGet\":{\"httpHeaders\":[{\"name\":\"health-check\",\"value\":\"liveness\"}],\"path\":\"/health\",\"port\":8080},\"initialDelaySeconds\":10,\"periodSeconds\":5},\"name\":\"test\",\"ports\":[{\"containerPort\":8080,\"name\":\"http-probe\"}],\"readinessProbe\":{\"httpGet\":{\"httpHeaders\":[{\"name\":\"health-check\",\"value\":\"readiness\"}],\"path\":\"/ready\",\"port\":8080},\"initialDelaySeconds\":10,\"periodSeconds\":5},\"resources\":{\"limits\":{\"cpu\":0.25,\"memory\":\"512Mi\"}}}]}}}}\n", //nolint:lll
			expectedCuratedLines: []string{
				"statefulset: nginx-deployment ns:  does not have a terminationGracePeriodSeconds value set. Default value (30) is used.",
			},
			gracePeriodSecs: 30,
		},
		{ // Test Case #2 - terminationGracePeriodSeconds is set to 30
			podName:              "nginx-deployment",
			annotation:           "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"annotations\":{},\"name\":\"test\",\"namespace\":\"tnf\"},\"spec\":{\"replicas\":2,\"selector\":{\"matchLabels\":{\"app\":\"test\"}},\"template\":{\"metadata\":{\"annotations\":{\"k8s.v1.cni.cncf.io/networks\":\"[ { \\\"name\\\" : \\\"mynet-ipv4-0\\\" },{ \\\"name\\\" : \\\"mynet-ipv4-1\\\" },{ \\\"name\\\" : \\\"mynet-ipv6-0\\\" },{ \\\"name\\\" : \\\"mynet-ipv6-1\\\" } ]\",\"test-network-function.com/container_tests\":\"[\\\"PRIVILEGED_POD\\\",\\\"PRIVILEGED_ROLE\\\"]\",\"test-network-function.com/defaultnetworkinterface\":\"\\\"eth0\\\"\"},\"labels\":{\"app\":\"test\",\"test-network-function.com/container\":\"target\",\"test-network-function.com/generic\":\"target\"},\"name\":\"test\"},\"spec\":{\"affinity\":{\"podAntiAffinity\":{\"requiredDuringSchedulingIgnoredDuringExecution\":[{\"labelSelector\":{\"matchExpressions\":[{\"key\":\"app\",\"operator\":\"In\",\"values\":[\"test\"]}]},\"topologyKey\":\"kubernetes.io/hostname\"}]}},\"automountServiceAccountToken\":false,\"containers\":[{\"command\":[\"./bin/app\"],\"image\":\"quay.io/testnetworkfunction/cnf-test-partner:latest\",\"imagePullPolicy\":\"IfNotPresent\",\"lifecycle\":{\"preStop\":{\"exec\":{\"command\":[\"/bin/sh\",\"-c\",\"killall -0 tail\"]}}},\"livenessProbe\":{\"httpGet\":{\"httpHeaders\":[{\"name\":\"health-check\",\"value\":\"liveness\"}],\"path\":\"/health\",\"port\":8080},\"initialDelaySeconds\":10,\"periodSeconds\":5},\"name\":\"test\",\"ports\":[{\"containerPort\":8080,\"name\":\"http-probe\"}],\"readinessProbe\":{\"httpGet\":{\"httpHeaders\":[{\"name\":\"health-check\",\"value\":\"readiness\"}],\"path\":\"/ready\",\"port\":8080},\"initialDelaySeconds\":10,\"periodSeconds\":5},\"resources\":{\"limits\":{\"cpu\":0.25,\"memory\":\"512Mi\"}}}],\"terminationGracePeriodSeconds\":30}}}}\n", //nolint:lll
			expectedCuratedLines: []string{},
			gracePeriodSecs:      29,
		},
	}

	for _, tc := range testCases {
		cLogs := loghelper.CuratedLogLines{}
		for _, ll := range tc.expectedCuratedLines {
			cLogs = cLogs.AddLogLine(ll)
		}
		_, curatedLogs := TestTerminationGracePeriodOnStatefulsets(generateEnv(tc.podName, tc.annotation, tc.gracePeriodSecs))
		assert.Equal(t, cLogs.GetLogLines(), curatedLogs.GetLogLines())
	}
}
