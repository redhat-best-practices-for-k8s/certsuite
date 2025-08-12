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

// GetGuaranteedPodsWithExclusiveCPUs returns all Pods that are guaranteed to have exclusive CPUs.
//
// It iterates over the Pods stored in the TestEnvironment, filters out those that do not
// satisfy the exclusive CPU guarantee condition, and returns a slice containing only the
// qualifying Pod objects. The returned slice may be empty if no such Pods exist.
func (env *TestEnvironment) GetGuaranteedPodsWithExclusiveCPUs() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.IsPodGuaranteedWithExclusiveCPUs() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// GetGuaranteedPodsWithIsolatedCPUs returns a slice of Pod pointers that are both guaranteed to use exclusive CPUs and comply with CPU isolation requirements.
//
// GetGuaranteedPodsWithIsolatedCPUs retrieves all pods from the TestEnvironment
// that meet two conditions: they must be scheduled as guaranteed (i.e., have
// CPU requests equal to limits) and their container configuration must satisfy
// the CPU isolation compliance rules. The function iterates over the internal
// pod list, filters each pod using IsPodGuaranteedWithExclusiveCPUs and
// IsCPUIsolationCompliant, and appends matching pods to the result slice,
// which it then returns.
func (env *TestEnvironment) GetGuaranteedPodsWithIsolatedCPUs() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.IsPodGuaranteedWithExclusiveCPUs() && p.IsCPUIsolationCompliant() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// GetGuaranteedPods returns all pods that are considered guaranteed in the test environment.
//
// It iterates over every pod managed by the TestEnvironment, checks each with IsPodGuaranteed,
// and collects those that satisfy the guarantee criteria into a slice.
// The resulting slice is returned to the caller.
func (env *TestEnvironment) GetGuaranteedPods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.IsPodGuaranteed() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// GetNonGuaranteedPods returns a slice of non‑guaranteed pods in the test environment.
//
// It iterates over all pods tracked by the TestEnvironment and filters out those
// whose resource requests do not equal their limits (i.e., not guaranteed).
// The returned slice contains pointers to Pod objects that are eligible for
// non‑guaranteed scheduling tests.
func (env *TestEnvironment) GetNonGuaranteedPods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if !p.IsPodGuaranteed() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// GetPodsWithoutAffinityRequiredLabel returns all Pods in the TestEnvironment that lack the affinity required label.
//
// It scans each Pod stored in the TestEnvironment, checks whether it has the
// affinity required label using the AffinityRequired function, and collects
// those that do not. The resulting slice of pointers to Pod is returned.
func (env *TestEnvironment) GetPodsWithoutAffinityRequiredLabel() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if !p.AffinityRequired() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// GetAffinityRequiredPods returns all Pods in the environment that require affinity.
//
// It iterates over the TestEnvironment's Pods slice, checks each Pod with
// the AffinityRequired helper, and collects those that have an affinity
// requirement into a new slice which is returned to the caller.
func (env *TestEnvironment) GetAffinityRequiredPods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.AffinityRequired() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// GetHugepagesPods returns all pods with hugepage memory enabled.
//
// It iterates over the TestEnvironment's pod list, checks each pod using HasHugepages,
// and collects those that expose a hugepage volume into a slice of *Pod.
// The resulting slice is returned to the caller.
func (env *TestEnvironment) GetHugepagesPods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.HasHugepages() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// GetCPUPinningPodsWithDpdk returns Pods with CPU pinning enabled and DPDK usage.
//
// It filters the current environment for pods that are running DPDK
// and have exclusive CPUs guaranteed, then returns a slice of those
// Pod objects. The returned slice may be empty if no such pods exist.
func (env *TestEnvironment) GetCPUPinningPodsWithDpdk() []*Pod {
	return filterDPDKRunningPods(env.GetGuaranteedPodsWithExclusiveCPUs())
}

// filterPodsWithoutHostPID removes pods that run with the HostPID option set.
//
// It iterates over a slice of Pod pointers and returns a new slice
// containing only those pods whose specification does not enable
// host process namespace sharing (i.e., HostPID is false or unset).
// The function preserves the order of the remaining pods.
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

// filterDPDKRunningPods filters a slice of Pod pointers, returning only those
// pods that are confirmed to be running the Data Plane Development Kit (DPDK).
//
// For each pod in the input slice it checks whether any container inside the
// pod is executing a DPDK process by invoking an exec command. Pods for which
// the command succeeds and returns output containing a DPDK indicator string
// are retained; all others are discarded. The function preserves the order of
// matching pods in the returned slice. It does not modify the original pods.
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

// GetShareProcessNamespacePods returns all pods in the environment that have ShareProcessNamespace enabled.
//
// It scans the TestEnvironment's pod list, checks each Pod with IsShareProcessNamespace,
// and collects those that return true into a new slice which is then returned.
func (env *TestEnvironment) GetShareProcessNamespacePods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.IsShareProcessNamespace() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// GetPodsUsingSRIOV returns a slice of Pod pointers that are using SR-IOV.
//
// It iterates over the pods stored in the TestEnvironment and checks each pod
// with IsUsingSRIOV. If an error occurs during this check, it is returned immediately.
// Pods confirmed to use SR‑IOV are appended to the result slice, which is returned
// along with a nil error when all pods have been processed.
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

// getContainers extracts all containers from a slice of Pods and returns them as a single slice.
//
// It iterates over each Pod in the input slice, appends every container found
// within that Pod to a new slice, and then returns the aggregated list.
// The function does not modify the original Pods or their containers.
func getContainers(pods []*Pod) []*Container {
	var containers []*Container

	for _, pod := range pods {
		containers = append(containers, pod.Containers...)
	}
	return containers
}

// GetGuaranteedPodContainersWithExclusiveCPUs returns a slice of Container pointers representing the containers
//
// that are part of guaranteed pods which have exclusive CPU allocations in the TestEnvironment.
// It gathers all containers from the environment and filters them to include only those belonging
// to pods identified by GetGuaranteedPodsWithExclusiveCPUs. The result is useful for
// further analysis or validation of CPU affinity settings.
func (env *TestEnvironment) GetGuaranteedPodContainersWithExclusiveCPUs() []*Container {
	return getContainers(env.GetGuaranteedPodsWithExclusiveCPUs())
}

// GetNonGuaranteedPodContainersWithoutHostPID returns all containers from non‑guaranteed pods that do not have HostPID enabled.
//
// It first retrieves the list of non‑guaranteed pods, then filters out any pods with the HostPID setting,
// and finally collects the containers belonging to the remaining pods. The result is a slice of pointers
// to Container objects representing those containers.
func (env *TestEnvironment) GetNonGuaranteedPodContainersWithoutHostPID() []*Container {
	return getContainers(filterPodsWithoutHostPID(env.GetNonGuaranteedPods()))
}

// GetGuaranteedPodContainersWithExclusiveCPUsWithoutHostPID returns a slice of containers from the test environment
//
// It filters the pods in the environment to those that are guaranteed and have exclusive CPUs,
// then excludes any pod with HostPID enabled, finally returning all containers belonging to the remaining pods.
func (env *TestEnvironment) GetGuaranteedPodContainersWithExclusiveCPUsWithoutHostPID() []*Container {
	return getContainers(filterPodsWithoutHostPID(env.GetGuaranteedPodsWithExclusiveCPUs()))
}

// GetGuaranteedPodContainersWithIsolatedCPUsWithoutHostPID retrieves all containers from the TestEnvironment that belong to guaranteed pods with isolated CPUs and do not use HostPID.
//
// It first obtains the list of containers in the environment, filters those whose
// parent pod has isolated CPU annotations, and then removes any container whose
// pod sets the HostPID flag. The returned slice contains only the qualifying
// containers for further processing.
func (env *TestEnvironment) GetGuaranteedPodContainersWithIsolatedCPUsWithoutHostPID() []*Container {
	return getContainers(filterPodsWithoutHostPID(env.GetGuaranteedPodsWithIsolatedCPUs()))
}
