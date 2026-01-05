// Copyright (C) 2022-2026 Red Hat, Inc.
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
)

// GetGuaranteedPodsWithExclusiveCPUs returns a slice of Pod objects that are guaranteed to have exclusive CPUs.
// It iterates over the Pods in the TestEnvironment and filters out the Pods that do not have exclusive CPUs.
// The filtered Pods are then returned as a slice.
func (env *TestEnvironment) GetGuaranteedPodsWithExclusiveCPUs() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.IsPodGuaranteedWithExclusiveCPUs() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// GetGuaranteedPodsWithIsolatedCPUs returns a list of pods from the TestEnvironment
// that are guaranteed to have isolated CPUs and are CPU isolation compliant.
func (env *TestEnvironment) GetGuaranteedPodsWithIsolatedCPUs() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.IsPodGuaranteedWithExclusiveCPUs() && p.IsCPUIsolationCompliant() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// GetGuaranteedPods returns a slice of guaranteed pods in the test environment.
// A guaranteed pod is a pod that meets certain criteria specified by the IsPodGuaranteed method.
// The method iterates over all pods in the environment and filters out the guaranteed ones.
// It returns the filtered pods as a slice.
func (env *TestEnvironment) GetGuaranteedPods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.IsPodGuaranteed() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// GetNonGuaranteedPods returns a slice of non-guaranteed pods in the test environment.
func (env *TestEnvironment) GetNonGuaranteedPods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if !p.IsPodGuaranteed() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// GetPodsWithoutAffinityRequiredLabel returns a slice of Pod objects that do not have the affinity required label.
// It iterates over the Pods in the TestEnvironment and filters out the ones that do not have the affinity required label.
// The filtered Pods are returned as a slice.
func (env *TestEnvironment) GetPodsWithoutAffinityRequiredLabel() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if !p.AffinityRequired() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// GetAffinityRequiredPods returns a slice of Pod objects that have affinity required.
// It iterates over the Pods in the TestEnvironment and filters out the Pods that have affinity required.
// The filtered Pods are returned as a slice.
func (env *TestEnvironment) GetAffinityRequiredPods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.AffinityRequired() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// GetHugepagesPods returns a slice of Pod objects that have hugepages enabled.
// It iterates over the Pods in the TestEnvironment and filters out the ones that do not have hugepages.
// The filtered Pods are returned as a []*Pod.
func (env *TestEnvironment) GetHugepagesPods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.HasHugepages() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

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

// GetShareProcessNamespacePods returns a slice of Pod objects that have the ShareProcessNamespace flag set to true.
// It iterates over the Pods in the TestEnvironment and filters out the ones that do not have the ShareProcessNamespace flag set.
// The filtered Pods are then returned as a slice.
func (env *TestEnvironment) GetShareProcessNamespacePods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.IsShareProcessNamespace() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

// GetPodsUsingSRIOV returns a list of pods that are using SR-IOV.
// It iterates through the pods in the TestEnvironment and checks if each pod is using SR-IOV.
// If an error occurs while checking the SR-IOV usage for a pod, it returns an error.
// The filtered pods that are using SR-IOV are returned along with a nil error.
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

func getContainers(pods []*Pod) []*Container {
	var containers []*Container

	for _, pod := range pods {
		containers = append(containers, pod.Containers...)
	}
	return containers
}

// GetGuaranteedPodContainersWithExclusiveCPUs returns a slice of Container objects representing the containers
// that have exclusive CPUs in the TestEnvironment.
func (env *TestEnvironment) GetGuaranteedPodContainersWithExclusiveCPUs() []*Container {
	return getContainers(env.GetGuaranteedPodsWithExclusiveCPUs())
}

// GetNonGuaranteedPodContainersWithoutHostPID returns a slice of containers from the test environment
// that belong to non-guaranteed pods without the HostPID setting enabled.
func (env *TestEnvironment) GetNonGuaranteedPodContainersWithoutHostPID() []*Container {
	return getContainers(filterPodsWithoutHostPID(env.GetNonGuaranteedPods()))
}

// GetGuaranteedPodContainersWithExclusiveCPUsWithoutHostPID returns a slice of containers from the test environment
// that belong to pods with exclusive CPUs and do not have the host PID enabled.
func (env *TestEnvironment) GetGuaranteedPodContainersWithExclusiveCPUsWithoutHostPID() []*Container {
	return getContainers(filterPodsWithoutHostPID(env.GetGuaranteedPodsWithExclusiveCPUs()))
}

// GetGuaranteedPodContainersWithIsolatedCPUsWithoutHostPID returns a slice of containers from the TestEnvironment
// that have guaranteed pods with isolated CPUs and without the HostPID flag set.
func (env *TestEnvironment) GetGuaranteedPodContainersWithIsolatedCPUsWithoutHostPID() []*Container {
	return getContainers(filterPodsWithoutHostPID(env.GetGuaranteedPodsWithIsolatedCPUs()))
}
