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
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	probeExecErrContainerNotFound = "container not found"
	probeExecErrUnableToUpgrade   = "unable to upgrade connection"
)

var getProbePod = defaultGetProbePod

// probeDumpOnce dumps live probe status at most once per namespace/name.
var probeDumpOnce sync.Map // map[string]*sync.Once

func defaultGetProbePod(namespace, name string) (*corev1.Pod, error) {
	ch := clientsholder.GetClientsHolder()
	return ch.K8sClient.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func resetProbeStatusDumpState() {
	probeDumpOnce = sync.Map{}
}

// IsProbeExecFailure reports whether err is a probe-pod exec outage
// (container not found, SPDY upgrade failure, or deadline exceeded)
// rather than a CNF finding.
func IsProbeExecFailure(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, probeExecErrContainerNotFound) ||
		strings.Contains(msg, probeExecErrUnableToUpgrade)
}

func handleProbeExecError(ctx clientsholder.Context, err error) error {
	if err == nil || !IsProbeExecFailure(err) {
		return err
	}
	dumpProbePodStatusOnce(ctx)
	return fmt.Errorf("probe pod exec failed in %s/%s container %q (not a CNF container failure): %w",
		ctx.GetNamespace(), ctx.GetPodName(), ctx.GetContainerName(), err)
}

func dumpProbePodStatusOnce(ctx clientsholder.Context) {
	key := ctx.GetNamespace() + "/" + ctx.GetPodName()
	onceVal, _ := probeDumpOnce.LoadOrStore(key, &sync.Once{})
	onceVal.(*sync.Once).Do(func() {
		dumpProbePodStatus(ctx)
	})
}

func dumpProbePodStatus(ctx clientsholder.Context) {
	ns, name := ctx.GetNamespace(), ctx.GetPodName()
	pod, err := getProbePod(ns, name)
	if err != nil {
		log.Error("Could not get live status of probe pod %s/%s after exec failure: %v", ns, name, err)
		return
	}
	log.Error("%s", formatProbePodStatus(pod))
}

func formatProbePodStatus(pod *corev1.Pod) string {
	if pod == nil {
		return "Probe pod live status: <nil>"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Probe pod %s/%s live status: phase=%s node=%s podIP=%s",
		pod.Namespace, pod.Name, pod.Status.Phase, pod.Spec.NodeName, pod.Status.PodIP)
	for i := range pod.Status.ContainerStatuses {
		cs := &pod.Status.ContainerStatuses[i]
		fmt.Fprintf(&b, "\n  container %s ready=%v restarts=%d state=%s lastTermination=%s",
			cs.Name, cs.Ready, cs.RestartCount, formatContainerState(cs.State), formatLastTermination(cs.LastTerminationState))
	}
	return b.String()
}

func formatContainerState(state corev1.ContainerState) string {
	switch {
	case state.Waiting != nil:
		return fmt.Sprintf("waiting(reason=%s message=%s)", state.Waiting.Reason, state.Waiting.Message)
	case state.Running != nil:
		return fmt.Sprintf("running(startedAt=%s)", state.Running.StartedAt.UTC().Format(time.RFC3339))
	case state.Terminated != nil:
		t := state.Terminated
		return fmt.Sprintf("terminated(reason=%s exit=%d finishedAt=%s)", t.Reason, t.ExitCode, t.FinishedAt.UTC().Format(time.RFC3339))
	default:
		return "unknown"
	}
}

func formatLastTermination(state corev1.ContainerState) string {
	if state.Terminated == nil {
		return "none"
	}
	t := state.Terminated
	return fmt.Sprintf("reason=%s exit=%d oom=%v finishedAt=%s",
		t.Reason, t.ExitCode, t.Reason == "OOMKilled", t.FinishedAt.UTC().Format(time.RFC3339))
}
