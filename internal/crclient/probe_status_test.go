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
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func stubProbePodGetter(t *testing.T, fn func(namespace, name string) (*corev1.Pod, error)) {
	t.Helper()
	orig := getProbePod
	getProbePod = fn
	resetProbeStatusDumpState()
	t.Cleanup(func() {
		getProbePod = orig
		resetProbeStatusDumpState()
	})
}

func stubProbeCommand(t *testing.T, cmd clientsholder.Command) {
	t.Helper()
	orig := probeCommand
	probeCommand = cmd
	t.Cleanup(func() {
		probeCommand = orig
	})
}

func stubTestEnvironment(t *testing.T, env *provider.TestEnvironment) {
	t.Helper()
	orig := getTestEnvironment
	getTestEnvironment = func() provider.TestEnvironment { return *env }
	t.Cleanup(func() {
		getTestEnvironment = orig
	})
}

func testCNFContainer() *provider.Container {
	return &provider.Container{
		Container: &corev1.Container{Name: "fm"},
		Namespace: "samsung",
		Podname:   "nf-alert-s2rns",
		NodeName:  "worker-0",
		Runtime:   "cri-o",
		UID:       "abc123",
	}
}

func TestIsProbeExecFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "unrelated", err: errors.New("connection refused"), want: false},
		{name: "container not found", err: errors.New(`unable to upgrade connection: container not found ("container-00")`), want: true},
		{name: "mixed case container not found", err: errors.New(`Unable to Upgrade Connection: Container Not Found ("container-00")`), want: true},
		{name: "unable to upgrade connection", err: errors.New("unable to upgrade connection: error dialing backend"), want: true},
		{name: "context deadline exceeded", err: context.DeadlineExceeded, want: true},
		{name: "wrapped deadline exceeded", err: fmt.Errorf("exec failed: %w", context.DeadlineExceeded), want: true},
		{name: "wrapped container not found", err: fmt.Errorf("cannot execute command: %w", errors.New(`container not found ("container-00")`)), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, IsProbeExecFailure(tt.err))
		})
	}
}

func TestFormatProbePodStatus_Running(t *testing.T) {
	t.Parallel()

	started := metav1.NewTime(time.Date(2026, 8, 18, 20, 0, 0, 0, time.UTC))
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "certsuite-probe-bf2nm", Namespace: "cnf-suite"},
		Spec:       corev1.PodSpec{NodeName: "worker-0"},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			PodIP: "10.0.0.12",
			ContainerStatuses: []corev1.ContainerStatus{{
				Name:         "container-00",
				Ready:        true,
				RestartCount: 0,
				State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{
					StartedAt: started,
				}},
			}},
		},
	}

	got := formatProbePodStatus(pod)
	assert.Contains(t, got, "Probe pod cnf-suite/certsuite-probe-bf2nm live status: phase=Running node=worker-0 podIP=10.0.0.12")
	assert.Contains(t, got, "container container-00 ready=true restarts=0")
	assert.Contains(t, got, "state=running(startedAt=2026-08-18T20:00:00Z)")
	assert.Contains(t, got, "lastTermination=none")
}

func TestFormatProbePodStatus_OOMKilled(t *testing.T) {
	t.Parallel()

	finished := metav1.NewTime(time.Date(2026, 8, 18, 20, 2, 17, 0, time.UTC))
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "certsuite-probe-bf2nm", Namespace: "cnf-suite"},
		Spec:       corev1.PodSpec{NodeName: "worker-0"},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			PodIP: "10.0.0.12",
			ContainerStatuses: []corev1.ContainerStatus{{
				Name:         "container-00",
				Ready:        false,
				RestartCount: 3,
				State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{
					Reason:  "CrashLoopBackOff",
					Message: "back-off 5m0s restarting failed container",
				}},
				LastTerminationState: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{
					Reason:     "OOMKilled",
					ExitCode:   137,
					FinishedAt: finished,
				}},
			}},
		},
	}

	got := formatProbePodStatus(pod)
	assert.Contains(t, got, "phase=Running")
	assert.Contains(t, got, "container container-00 ready=false restarts=3")
	assert.Contains(t, got, "state=waiting(reason=CrashLoopBackOff")
	assert.Contains(t, got, "lastTermination=reason=OOMKilled exit=137 oom=true finishedAt=2026-08-18T20:02:17Z")
}

func TestFormatProbePodStatus_Terminated(t *testing.T) {
	t.Parallel()

	finished := metav1.NewTime(time.Date(2026, 8, 18, 20, 5, 0, 0, time.UTC))
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "probe-a", Namespace: "cnf-suite"},
		Spec:       corev1.PodSpec{NodeName: "worker-1"},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
			ContainerStatuses: []corev1.ContainerStatus{{
				Name: "container-00",
				State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{
					Reason:     "Error",
					ExitCode:   1,
					FinishedAt: finished,
				}},
			}},
		},
	}

	got := formatProbePodStatus(pod)
	assert.Contains(t, got, "phase=Failed")
	assert.Contains(t, got, "state=terminated(reason=Error exit=1 finishedAt=2026-08-18T20:05:00Z)")
}

func TestFormatProbePodStatus_NoContainers(t *testing.T) {
	t.Parallel()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "probe-a", Namespace: "cnf-suite"},
		Spec:       corev1.PodSpec{NodeName: "worker-1"},
		Status:     corev1.PodStatus{Phase: corev1.PodPending, PodIP: ""},
	}

	got := formatProbePodStatus(pod)
	assert.Equal(t, "Probe pod cnf-suite/probe-a live status: phase=Pending node=worker-1 podIP=", got)
	assert.NotContains(t, got, "container ")
}

func TestFormatProbePodStatus_NilPod(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Probe pod live status: <nil>", formatProbePodStatus(nil))
}

func TestFormatContainerState(t *testing.T) {
	t.Parallel()

	started := metav1.NewTime(time.Date(2026, 8, 18, 20, 0, 0, 0, time.UTC))
	finished := metav1.NewTime(time.Date(2026, 8, 18, 20, 5, 0, 0, time.UTC))

	tests := []struct {
		name  string
		state corev1.ContainerState
		want  string
	}{
		{name: "unknown empty", state: corev1.ContainerState{}, want: "unknown"},
		{
			name:  "waiting",
			state: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff", Message: "backoff"}},
			want:  "waiting(reason=CrashLoopBackOff message=backoff)",
		},
		{
			name:  "running",
			state: corev1.ContainerState{Running: &corev1.ContainerStateRunning{StartedAt: started}},
			want:  "running(startedAt=2026-08-18T20:00:00Z)",
		},
		{
			name:  "terminated",
			state: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: "Error", ExitCode: 1, FinishedAt: finished}},
			want:  "terminated(reason=Error exit=1 finishedAt=2026-08-18T20:05:00Z)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, formatContainerState(tt.state))
		})
	}
}

func TestFormatLastTermination(t *testing.T) {
	t.Parallel()

	finished := metav1.NewTime(time.Date(2026, 8, 18, 20, 2, 17, 0, time.UTC))

	assert.Equal(t, "none", formatLastTermination(corev1.ContainerState{}))
	assert.Equal(t, "reason=OOMKilled exit=137 oom=true finishedAt=2026-08-18T20:02:17Z",
		formatLastTermination(corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{
			Reason: "OOMKilled", ExitCode: 137, FinishedAt: finished,
		}}))
	assert.Equal(t, "reason=Error exit=1 oom=false finishedAt=2026-08-18T20:02:17Z",
		formatLastTermination(corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{
			Reason: "Error", ExitCode: 1, FinishedAt: finished,
		}}))
}

func TestHandleProbeExecError_Passthrough(t *testing.T) {
	ctx := clientsholder.NewContext("cnf-suite", "probe-a", "container-00")
	assert.Nil(t, handleProbeExecError(ctx, nil))

	origErr := errors.New("connection refused")
	got := handleProbeExecError(ctx, origErr)
	assert.Equal(t, origErr, got)
}

func TestHandleProbeExecError_WrapsAndDumps(t *testing.T) {
	var logArchive bytes.Buffer
	log.SetupLogger(&logArchive, "ERROR")

	pod := runningProbePod()
	stubProbePodGetter(t, func(namespace, name string) (*corev1.Pod, error) {
		assert.Equal(t, "cnf-suite", namespace)
		assert.Equal(t, "certsuite-probe-bf2nm", name)
		return pod, nil
	})

	ctx := clientsholder.NewContext("cnf-suite", "certsuite-probe-bf2nm", "container-00")
	inner := errors.New(`unable to upgrade connection: container not found ("container-00")`)
	got := handleProbeExecError(ctx, inner)

	require.Error(t, got)
	assert.ErrorIs(t, got, inner)
	assert.Contains(t, got.Error(), `probe pod exec failed in cnf-suite/certsuite-probe-bf2nm container "container-00" (not a CNF container failure)`)
	assert.Contains(t, logArchive.String(), "Probe pod cnf-suite/certsuite-probe-bf2nm live status")
	assert.Contains(t, logArchive.String(), "container container-00")
}

func TestDumpProbePodStatusOnce_DumpsOncePerProbe(t *testing.T) {
	var logArchive bytes.Buffer
	log.SetupLogger(&logArchive, "ERROR")

	var calls atomic.Int32
	stubProbePodGetter(t, func(_, _ string) (*corev1.Pod, error) {
		calls.Add(1)
		return runningProbePod(), nil
	})

	ctx := clientsholder.NewContext("cnf-suite", "certsuite-probe-bf2nm", "container-00")
	inner := errors.New(`container not found ("container-00")`)

	_ = handleProbeExecError(ctx, inner)
	_ = handleProbeExecError(ctx, inner)
	_ = handleProbeExecError(ctx, inner)

	assert.Equal(t, int32(1), calls.Load())
	assert.Equal(t, 1, strings.Count(logArchive.String(), "Probe pod cnf-suite/certsuite-probe-bf2nm live status"))
}

func TestDumpProbePodStatusOnce_SeparateProbes(t *testing.T) {
	var logArchive bytes.Buffer
	log.SetupLogger(&logArchive, "ERROR")

	var calls atomic.Int32
	stubProbePodGetter(t, func(_, _ string) (*corev1.Pod, error) {
		calls.Add(1)
		return runningProbePod(), nil
	})

	inner := errors.New(`container not found ("container-00")`)
	_ = handleProbeExecError(clientsholder.NewContext("cnf-suite", "probe-a", "container-00"), inner)
	_ = handleProbeExecError(clientsholder.NewContext("cnf-suite", "probe-b", "container-00"), inner)

	assert.Equal(t, int32(2), calls.Load())
}

func TestDumpProbePodStatus_NotFound(t *testing.T) {
	var logArchive bytes.Buffer
	log.SetupLogger(&logArchive, "ERROR")

	notFound := apierrors.NewNotFound(schema.GroupResource{Resource: "pods"}, "certsuite-probe-bf2nm")
	stubProbePodGetter(t, func(_, _ string) (*corev1.Pod, error) {
		return nil, notFound
	})

	ctx := clientsholder.NewContext("cnf-suite", "certsuite-probe-bf2nm", "container-00")
	got := handleProbeExecError(ctx, errors.New(`container not found ("container-00")`))

	require.Error(t, got)
	assert.Contains(t, logArchive.String(), "Could not get live status of probe pod cnf-suite/certsuite-probe-bf2nm after exec failure")
	assert.Contains(t, logArchive.String(), "not found")
}

func runningProbePod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "certsuite-probe-bf2nm", Namespace: "cnf-suite"},
		Spec: corev1.PodSpec{
			NodeName:   "worker-0",
			Containers: []corev1.Container{{Name: "container-00"}},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			PodIP: "10.0.0.12",
			ContainerStatuses: []corev1.ContainerStatus{{
				Name:         "container-00",
				Ready:        true,
				RestartCount: 1,
				State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{
					StartedAt: metav1.NewTime(time.Date(2026, 8, 18, 19, 30, 0, 0, time.UTC)),
				}},
			}},
		},
	}
}

func TestGetPidFromContainer_ProbeContainerNotFound(t *testing.T) {
	var logArchive bytes.Buffer
	log.SetupLogger(&logArchive, "ERROR")

	inner := errors.New(`unable to upgrade connection: container not found ("container-00")`)
	mock := &clientsholder.MockCommand{
		ExecFunc: func(_ clientsholder.Context, _ string) (string, string, error) {
			return "", "", inner
		},
	}
	stubProbeCommand(t, mock)

	var dumps atomic.Int32
	stubProbePodGetter(t, func(_, _ string) (*corev1.Pod, error) {
		dumps.Add(1)
		return runningProbePod(), nil
	})

	ctx := clientsholder.NewContext("cnf-suite", "certsuite-probe-bf2nm", "container-00")
	pid, err := GetPidFromContainer(testCNFContainer(), ctx)

	assert.Equal(t, 0, pid)
	require.Error(t, err)
	assert.ErrorIs(t, err, inner)
	assert.True(t, IsProbeExecFailure(err))
	assert.Contains(t, err.Error(), `probe pod exec failed in cnf-suite/certsuite-probe-bf2nm container "container-00" (not a CNF container failure)`)
	assert.Equal(t, 1, mock.CallCount())
	assert.Equal(t, int32(1), dumps.Load())
	assert.Contains(t, logArchive.String(), "Probe pod cnf-suite/certsuite-probe-bf2nm live status")
}

func TestExecCommandContainerNSEnter_ProbeContainerNotFound(t *testing.T) {
	var logArchive bytes.Buffer
	log.SetupLogger(&logArchive, "ERROR")

	inner := errors.New(`unable to upgrade connection: container not found ("container-00")`)
	mock := &clientsholder.MockCommand{
		ExecFunc: func(_ clientsholder.Context, command string) (string, string, error) {
			if strings.Contains(command, "crictl inspect") {
				return "1234\n", "", nil
			}
			return "", "", inner
		},
	}
	stubProbeCommand(t, mock)

	probe := runningProbePod()
	stubTestEnvironment(t, &provider.TestEnvironment{
		ProbePods: map[string]*corev1.Pod{probe.Spec.NodeName: probe},
	})

	var dumps atomic.Int32
	stubProbePodGetter(t, func(_, _ string) (*corev1.Pod, error) {
		dumps.Add(1)
		return probe, nil
	})

	_, _, err := ExecCommandContainerNSEnter("ss -tpln | grep sshd", testCNFContainer())
	require.Error(t, err)
	assert.ErrorIs(t, err, inner)
	assert.True(t, IsProbeExecFailure(err))
	assert.Contains(t, err.Error(), `probe pod exec failed in cnf-suite/certsuite-probe-bf2nm container "container-00" (not a CNF container failure)`)
	assert.Equal(t, 2, mock.CallCount())
	assert.Equal(t, int32(1), dumps.Load())
	assert.Contains(t, logArchive.String(), "Probe pod cnf-suite/certsuite-probe-bf2nm live status")
}

func TestDefaultGetProbePod(t *testing.T) {
	t.Cleanup(clientsholder.ClearTestClientsHolder)

	pod := runningProbePod()
	clientsholder.GetTestClientsHolder([]runtime.Object{pod})

	got, err := defaultGetProbePod("cnf-suite", "certsuite-probe-bf2nm")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, corev1.PodRunning, got.Status.Phase)
	assert.Equal(t, "worker-0", got.Spec.NodeName)
	assert.Equal(t, "10.0.0.12", got.Status.PodIP)

	_, err = defaultGetProbePod("cnf-suite", "missing-probe")
	require.Error(t, err)
	assert.True(t, apierrors.IsNotFound(err))
}
