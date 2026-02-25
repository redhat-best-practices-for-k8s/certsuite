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

package performance

import (
	"strings"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/crclient"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func setupCheck() *checksdb.Check {
	var logArchive strings.Builder
	log.SetupLogger(&logArchive, "INFO")
	return checksdb.NewCheck("test-id", nil)
}

// ---- TestGetExecProbesCmds ----

func TestGetExecProbesCmds_NoProbes(t *testing.T) {
	c := &provider.Container{
		Container: &corev1.Container{
			Name: "test-container",
		},
	}
	result := getExecProbesCmds(c)
	assert.Empty(t, result)
}

func TestGetExecProbesCmds_LivenessExec(t *testing.T) {
	c := &provider.Container{
		Container: &corev1.Container{
			Name: "test-container",
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					Exec: &corev1.ExecAction{
						Command: []string{"cat", "/tmp/healthy"},
					},
				},
			},
		},
	}
	result := getExecProbesCmds(c)
	assert.Len(t, result, 1)
	assert.True(t, result["cat/tmp/healthy"])
}

func TestGetExecProbesCmds_ReadinessExec(t *testing.T) {
	c := &provider.Container{
		Container: &corev1.Container{
			Name: "test-container",
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					Exec: &corev1.ExecAction{
						Command: []string{"cat", "/tmp/ready"},
					},
				},
			},
		},
	}
	result := getExecProbesCmds(c)
	assert.Len(t, result, 1)
	assert.True(t, result["cat/tmp/ready"])
}

func TestGetExecProbesCmds_StartupExec(t *testing.T) {
	c := &provider.Container{
		Container: &corev1.Container{
			Name: "test-container",
			StartupProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					Exec: &corev1.ExecAction{
						Command: []string{"cat", "/tmp/started"},
					},
				},
			},
		},
	}
	result := getExecProbesCmds(c)
	assert.Len(t, result, 1)
	assert.True(t, result["cat/tmp/started"])
}

func TestGetExecProbesCmds_AllThreeExec(t *testing.T) {
	c := &provider.Container{
		Container: &corev1.Container{
			Name: "test-container",
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					Exec: &corev1.ExecAction{Command: []string{"cmd1"}},
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					Exec: &corev1.ExecAction{Command: []string{"cmd2"}},
				},
			},
			StartupProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					Exec: &corev1.ExecAction{Command: []string{"cmd3"}},
				},
			},
		},
	}
	result := getExecProbesCmds(c)
	assert.Len(t, result, 3)
	assert.True(t, result["cmd1"])
	assert.True(t, result["cmd2"])
	assert.True(t, result["cmd3"])
}

func TestGetExecProbesCmds_HTTPProbes(t *testing.T) {
	c := &provider.Container{
		Container: &corev1.Container{
			Name: "test-container",
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/healthz",
					},
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{},
				},
			},
		},
	}
	result := getExecProbesCmds(c)
	assert.Empty(t, result)
}

// ---- TestFilterProbeProcesses ----

func TestFilterProbeProcesses_NoProcesses(t *testing.T) {
	c := &provider.Container{
		Container: &corev1.Container{Name: "c1"},
		Namespace: "ns1",
		Podname:   "pod1",
	}
	notProbe, compliant := filterProbeProcesses(nil, c)
	assert.Empty(t, notProbe)
	assert.Empty(t, compliant)
}

func TestFilterProbeProcesses_AllProbeProcesses(t *testing.T) {
	c := &provider.Container{
		Container: &corev1.Container{
			Name: "c1",
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					Exec: &corev1.ExecAction{Command: []string{"cat", "/tmp/healthy"}},
				},
			},
		},
		Namespace: "ns1",
		Podname:   "pod1",
	}
	processes := []*crclient.Process{
		{Pid: 1, PPid: 0, Args: "cat /tmp/healthy"},
		{Pid: 2, PPid: 1, Args: "child-of-probe"},
	}
	notProbe, compliant := filterProbeProcesses(processes, c)
	assert.Empty(t, notProbe)
	assert.Len(t, compliant, 1) // Only the parent matches, child is filtered by PPid
}

func TestFilterProbeProcesses_MixedProcesses(t *testing.T) {
	c := &provider.Container{
		Container: &corev1.Container{
			Name: "c1",
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					Exec: &corev1.ExecAction{Command: []string{"cat", "/tmp/healthy"}},
				},
			},
		},
		Namespace: "ns1",
		Podname:   "pod1",
	}
	processes := []*crclient.Process{
		{Pid: 1, PPid: 0, Args: "cat /tmp/healthy"},
		{Pid: 2, PPid: 1, Args: "child-of-probe"},
		{Pid: 3, PPid: 0, Args: "my-app --serve"},
	}
	notProbe, compliant := filterProbeProcesses(processes, c)
	assert.Len(t, notProbe, 1)
	assert.Equal(t, "my-app --serve", notProbe[0].Args)
	assert.Len(t, compliant, 1)
}

// ---- TestCPUPinningNoExecProbes ----

func TestCPUPinningNoExecProbes_NoExecProbes(t *testing.T) {
	check := setupCheck()
	pods := []*provider.Pod{
		{
			Pod: &corev1.Pod{},
			Containers: []*provider.Container{
				{Container: &corev1.Container{Name: "c1"}},
			},
		},
	}
	testCPUPinningNoExecProbes(check, pods)
	assert.Equal(t, "passed", string(check.Result))
}

func TestCPUPinningNoExecProbes_WithExecProbe(t *testing.T) {
	check := setupCheck()
	pods := []*provider.Pod{
		{
			Pod: &corev1.Pod{},
			Containers: []*provider.Container{
				{
					Container: &corev1.Container{
						Name: "c1",
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								Exec: &corev1.ExecAction{Command: []string{"check"}},
							},
						},
					},
				},
			},
		},
	}
	testCPUPinningNoExecProbes(check, pods)
	assert.Equal(t, "failed", string(check.Result))
}

func TestCPUPinningNoExecProbes_EmptyPodList(t *testing.T) {
	check := setupCheck()
	testCPUPinningNoExecProbes(check, []*provider.Pod{})
	// No pods means no non-compliant objects, both lists empty -> skipped
	assert.Equal(t, "skipped", string(check.Result))
}

func TestCPUPinningNoExecProbes_MultiplePodsMixed(t *testing.T) {
	check := setupCheck()
	pods := []*provider.Pod{
		{
			Pod: &corev1.Pod{},
			Containers: []*provider.Container{
				{Container: &corev1.Container{Name: "clean-container"}},
			},
		},
		{
			Pod: &corev1.Pod{},
			Containers: []*provider.Container{
				{
					Container: &corev1.Container{
						Name: "probe-container",
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								Exec: &corev1.ExecAction{Command: []string{"check"}},
							},
						},
					},
				},
			},
		},
	}
	testCPUPinningNoExecProbes(check, pods)
	assert.Equal(t, "failed", string(check.Result))
}

// ---- TestLimitedUseOfExecProbes ----

func TestLimitedUseOfExecProbes_NoExecProbes(t *testing.T) {
	check := setupCheck()
	testEnv := provider.TestEnvironment{
		Pods: []*provider.Pod{
			{
				Pod: &corev1.Pod{},
				Containers: []*provider.Container{
					{Container: &corev1.Container{Name: "c1"}},
				},
			},
		},
	}
	testLimitedUseOfExecProbes(check, &testEnv)
	assert.Equal(t, "passed", string(check.Result))
}

func TestLimitedUseOfExecProbes_CompliantPeriod(t *testing.T) {
	check := setupCheck()
	testEnv := provider.TestEnvironment{
		Pods: []*provider.Pod{
			{
				Pod: &corev1.Pod{},
				Containers: []*provider.Container{
					{
						Container: &corev1.Container{
							Name: "c1",
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{Command: []string{"check"}},
								},
								PeriodSeconds: 15,
							},
						},
					},
				},
			},
		},
	}
	testLimitedUseOfExecProbes(check, &testEnv)
	assert.Equal(t, "passed", string(check.Result))
}

func TestLimitedUseOfExecProbes_NonCompliantPeriod(t *testing.T) {
	check := setupCheck()
	testEnv := provider.TestEnvironment{
		Pods: []*provider.Pod{
			{
				Pod: &corev1.Pod{},
				Containers: []*provider.Container{
					{
						Container: &corev1.Container{
							Name: "c1",
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{Command: []string{"check"}},
								},
								PeriodSeconds: 5,
							},
						},
					},
				},
			},
		},
	}
	testLimitedUseOfExecProbes(check, &testEnv)
	assert.Equal(t, "failed", string(check.Result))
}

func TestLimitedUseOfExecProbes_TooManyExecProbes(t *testing.T) {
	check := setupCheck()
	// Create 10+ exec probes across multiple containers
	var containers []*provider.Container
	for i := 0; i < 11; i++ {
		containers = append(containers, &provider.Container{
			Container: &corev1.Container{
				Name: "c" + string(rune('0'+i)),
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						Exec: &corev1.ExecAction{Command: []string{"check"}},
					},
					PeriodSeconds: 15, // compliant period
				},
			},
		})
	}

	testEnv := provider.TestEnvironment{
		Pods: []*provider.Pod{
			{
				Pod:        &corev1.Pod{},
				Containers: containers,
			},
		},
	}
	testLimitedUseOfExecProbes(check, &testEnv)
	assert.Equal(t, "failed", string(check.Result))
}

func TestLimitedUseOfExecProbes_ExactlyAtLimit(t *testing.T) {
	check := setupCheck()
	// Create exactly 10 exec probes (at the limit, should still fail with >=10 check)
	var containers []*provider.Container
	for i := 0; i < 10; i++ {
		containers = append(containers, &provider.Container{
			Container: &corev1.Container{
				Name: "c" + string(rune('a'+i)),
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						Exec: &corev1.ExecAction{Command: []string{"check"}},
					},
					PeriodSeconds: 15,
				},
			},
		})
	}

	testEnv := provider.TestEnvironment{
		Pods: []*provider.Pod{
			{
				Pod:        &corev1.Pod{},
				Containers: containers,
			},
		},
	}
	testLimitedUseOfExecProbes(check, &testEnv)
	assert.Equal(t, "failed", string(check.Result))
}

// ---- TestExclusiveCPUPool ----

func TestExclusiveCPUPool_AllExclusive(t *testing.T) {
	check := setupCheck()
	testEnv := provider.TestEnvironment{
		Pods: []*provider.Pod{
			{
				Pod: &corev1.Pod{},
				Containers: []*provider.Container{
					{
						Container: &corev1.Container{
							Name: "c1",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("2"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("2"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
	testExclusiveCPUPool(check, &testEnv)
	assert.Equal(t, "passed", string(check.Result))
}

func TestExclusiveCPUPool_AllShared(t *testing.T) {
	check := setupCheck()
	testEnv := provider.TestEnvironment{
		Pods: []*provider.Pod{
			{
				Pod: &corev1.Pod{},
				Containers: []*provider.Container{
					{
						Container: &corev1.Container{
							Name: "c1",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
	testExclusiveCPUPool(check, &testEnv)
	assert.Equal(t, "passed", string(check.Result))
}

func TestExclusiveCPUPool_MixedInSamePod(t *testing.T) {
	check := setupCheck()
	testEnv := provider.TestEnvironment{
		Pods: []*provider.Pod{
			{
				Pod: &corev1.Pod{},
				Containers: []*provider.Container{
					{
						Container: &corev1.Container{
							Name: "exclusive",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("2"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("2"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
						},
					},
					{
						Container: &corev1.Container{
							Name: "shared",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
	testExclusiveCPUPool(check, &testEnv)
	assert.Equal(t, "failed", string(check.Result))
}
