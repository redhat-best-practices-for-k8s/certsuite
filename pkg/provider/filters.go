// Copyright (C) 2022-2024 Red Hat, Inc.
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
	"fmt"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
)

// TestEnvironment.GetGuaranteedPodsWithExclusiveCPUs Retrieves pods that have guaranteed exclusive CPU allocation
//
// The method examines each pod in the test environment, applying a check to
// determine if the pod is guaranteed with exclusive CPUs. Pods passing this
// check are collected into a slice and returned. This list can be used by other
// functions to identify containers or pods suitable for CPU‑pinning
// scenarios.
func (env *TestEnvironment) GetGuaranteedPodsWithExclusiveCPUs() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.IsPodGuaranteedWithExclusiveCPUs() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// TestEnvironment.GetGuaranteedPodsWithIsolatedCPUs Retrieves pods that are guaranteed to have isolated CPUs
//
// This method scans all pods in the test environment, selecting only those
// whose CPU requests match whole units and whose resources are identical across
// containers. It further checks that each pod meets CPU isolation compliance
// criteria, such as having appropriate annotations and a specified runtime
// class name. The resulting slice of pods is returned for use by other
// filtering functions.
func (env *TestEnvironment) GetGuaranteedPodsWithIsolatedCPUs() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.IsPodGuaranteedWithExclusiveCPUs() && p.IsCPUIsolationCompliant() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// TestEnvironment.GetGuaranteedPods Retrieves all pods that satisfy the guaranteed condition
//
// This method scans every pod in the test environment, checks each one with its
// own guarantee logic, and collects those that pass into a slice. The resulting
// slice contains only the pods deemed guaranteed, which are then returned to
// the caller.
func (env *TestEnvironment) GetGuaranteedPods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.IsPodGuaranteed() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// TestEnvironment.GetNonGuaranteedPods retrieves all pods that are not guaranteed in the test environment
//
// The function iterates over every pod in the TestEnvironment, checks if each
// pod is not guaranteed by calling IsPodGuaranteed, and collects those pods
// into a slice. It returns this slice of non‑guaranteed pods for further
// processing or analysis.
func (env *TestEnvironment) GetNonGuaranteedPods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if !p.IsPodGuaranteed() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// TestEnvironment.GetPodsWithoutAffinityRequiredLabel Retrieves pods missing the required affinity label
//
// The method scans all pods in the test environment, checks each pod for the
// presence of an affinity-required label using the Pod.AffinityRequired helper,
// and collects those that lack it. It returns a slice containing only these
// pods, allowing callers to identify which resources need proper labeling.
func (env *TestEnvironment) GetPodsWithoutAffinityRequiredLabel() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if !p.AffinityRequired() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// TestEnvironment.GetAffinityRequiredPods Retrieves pods that require affinity
//
// This method scans the test environment's collection of pod objects and
// selects those that have an affinity requirement flag set in their labels. It
// returns a slice containing only the matching pods, enabling callers to focus
// on affinity-dependent resources.
func (env *TestEnvironment) GetAffinityRequiredPods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.AffinityRequired() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// TestEnvironment.GetHugepagesPods returns all pods that request or limit hugepages
//
// The method scans the environment’s pod collection, checks each pod for any
// container using a hugepage resource via HasHugepages, and collects those that
// do. The resulting slice of pointers to Pod objects is returned; if none have
// hugepages, an empty slice is produced.
func (env *TestEnvironment) GetHugepagesPods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.HasHugepages() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// TestEnvironment.GetCPUPinningPodsWithDpdk Lists guaranteed pods that pin CPUs with DPDK
//
// This method retrieves all pods in the test environment that are guaranteed to
// have exclusive CPU resources and then filters them to include only those
// running DPDK drivers. It calls a helper function that checks each pod’s
// container for DPDK device presence via a system command. The resulting slice
// contains pointers to pods meeting both criteria, suitable for further
// validation or manipulation.
func (env *TestEnvironment) GetCPUPinningPodsWithDpdk() []*Pod {
	return filterDPDKRunningPods(env.GetGuaranteedPodsWithExclusiveCPUs())
}

// filterPodsWithoutHostPID filters out pods that enable HostPID
//
// The function receives a slice of pod objects and iterates through each one,
// checking whether the HostPID flag is set in the pod specification. Pods with
// this flag enabled are skipped; all others are collected into a new slice. The
// resulting slice contains only those pods that do not use the host's PID
// namespace.
func filterPodsWithoutHostPID(pods []*Pod) []*Pod {
	var withoutHostPIDPods []*Pod

	for _, pod := range pods {
		if pod.Spec.HostPID {
			continue
		}
		withoutHostPIDPods = append(withoutHostPIDPods, pod)
	}
	return withoutHostPIDPods
}

// filterDPDKRunningPods Filters pods that are running DPDK-enabled devices
//
// This function examines a slice of pod objects, executing a command inside
// each container to locate the device file path specified by the pod’s Multus
// PCI annotation. If the output contains the string "vfio-pci", indicating the
// presence of a DPDK driver, the pod is added to a new list. The resulting
// slice contains only pods that have confirmed DPDK support.
func filterDPDKRunningPods(pods []*Pod) []*Pod {
	var filteredPods []*Pod
	const (
		dpdkDriver           = "vfio-pci"
		findDeviceSubCommand = "find /sys -name"
	)
	o := clientsholder.GetClientsHolder()
	for _, pod := range pods {
		if len(pod.MultusPCIs) == 0 {
			continue
		}
		ctx := clientsholder.NewContext(pod.Namespace, pod.Name, pod.Spec.Containers[0].Name)
		findCommand := fmt.Sprintf("%s '%s'", findDeviceSubCommand, pod.MultusPCIs[0])
		outStr, errStr, err := o.ExecCommandContainer(ctx, findCommand)
		if err != nil || errStr != "" {
			log.Error("Failed to execute command %s in probe %s, errStr: %s, err: %v", findCommand, pod.String(), errStr, err)
			continue
		}
		if strings.Contains(outStr, dpdkDriver) {
			filteredPods = append(filteredPods, pod)
		}
	}
	return filteredPods
}

// TestEnvironment.GetShareProcessNamespacePods Retrieves pods that enable shared process namespaces
//
// The function scans the TestEnvironment's collection of Pod objects, selecting
// those whose ShareProcessNamespace flag is true. It accumulates these matching
// pods into a new slice and returns it. The returned slice contains only pods
// configured for shared process namespace operation.
func (env *TestEnvironment) GetShareProcessNamespacePods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.IsShareProcessNamespace() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// TestEnvironment.GetPodsUsingSRIOV Collects all pods that are using SR-IOV
//
// The method scans every pod in the test environment, checking each one for
// SR‑IOV usage by calling its helper function. If a pod reports SR‑IOV
// support, it is added to a slice of matching pods. The function returns this
// list and an error if any pod check fails.
func (env *TestEnvironment) GetPodsUsingSRIOV() ([]*Pod, error) {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		usesSRIOV, err := p.IsUsingSRIOV()
		if err != nil {
			return nil, fmt.Errorf("failed to check sriov usage for pod %s: %v", p, err)
		}

		if usesSRIOV {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods, nil
}

// getContainers collects all containers from a list of pods
//
// The function iterates over each pod in the provided slice, appending every
// container within those pods to a new slice. It returns this aggregated slice,
// allowing callers to work with a flat list of containers regardless of their
// originating pod.
func getContainers(pods []*Pod) []*Container {
	var containers []*Container

	for _, pod := range pods {
		containers = append(containers, pod.Containers...)
	}
	return containers
}

// TestEnvironment.GetGuaranteedPodContainersWithExclusiveCPUs Retrieves containers with guaranteed exclusive CPUs
//
// This method returns a slice of container objects that belong to pods which
// have been marked as guaranteed to use exclusive CPUs. It gathers the relevant
// pods via GetGuaranteedPodsWithExclusiveCPUs and then collects their
// containers into a single list for further processing or inspection.
func (env *TestEnvironment) GetGuaranteedPodContainersWithExclusiveCPUs() []*Container {
	return getContainers(env.GetGuaranteedPodsWithExclusiveCPUs())
}

// TestEnvironment.GetNonGuaranteedPodContainersWithoutHostPID Lists containers in non-guaranteed pods that do not use HostPID
//
// This method retrieves all non-guaranteed pods from the test environment,
// filters out any pods with the HostPID setting enabled, then collects every
// container within those remaining pods. The result is a slice of container
// objects representing workloads that are both non‑guaranteed and run without
// shared PID namespaces.
func (env *TestEnvironment) GetNonGuaranteedPodContainersWithoutHostPID() []*Container {
	return getContainers(filterPodsWithoutHostPID(env.GetNonGuaranteedPods()))
}

// TestEnvironment.GetGuaranteedPodContainersWithExclusiveCPUsWithoutHostPID Retrieves containers from guaranteed pods that use exclusive CPUs but do not enable host PID
//
// It first selects all pods in the test environment marked as guaranteed with
// exclusive CPUs, then filters out any pod where HostPID is enabled. Finally it
// collects and returns every container belonging to the remaining pods.
func (env *TestEnvironment) GetGuaranteedPodContainersWithExclusiveCPUsWithoutHostPID() []*Container {
	return getContainers(filterPodsWithoutHostPID(env.GetGuaranteedPodsWithExclusiveCPUs()))
}

// TestEnvironment.GetGuaranteedPodContainersWithIsolatedCPUsWithoutHostPID returns containers from guaranteed pods with isolated CPUs that do not use HostPID
//
// It first collects all pods in the environment that are guaranteed to have
// exclusive CPU allocation and comply with CPU isolation rules. Then it filters
// out any pod where the HostPID flag is enabled, ensuring only non-HostPID pods
// remain. Finally, it aggregates and returns a slice of containers from those
// remaining pods.
func (env *TestEnvironment) GetGuaranteedPodContainersWithIsolatedCPUsWithoutHostPID() []*Container {
	return getContainers(filterPodsWithoutHostPID(env.GetGuaranteedPodsWithIsolatedCPUs()))
}
