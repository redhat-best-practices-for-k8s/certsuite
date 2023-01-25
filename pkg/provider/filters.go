// Copyright (C) 2022 Red Hat, Inc.
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

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
)

func (env *TestEnvironment) GetGuaranteedPodsWithExlusiveCPUs() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.IsPodGuaranteedWithExclusiveCPUs() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

func (env *TestEnvironment) GetGuaranteedPods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.IsPodGuaranteed() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

func (env *TestEnvironment) GetNonGuaranteedPods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if !p.IsPodGuaranteed() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

func (env *TestEnvironment) GetPodsWithoutAffinityRequiredLabel() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if !p.AffinityRequired() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

func (env *TestEnvironment) GetAffinityRequiredPods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.AffinityRequired() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

func (env *TestEnvironment) GetHugepagesPods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.HasHugepages() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}

func (env *TestEnvironment) GetCPUPinningPodsWithDpdk() []*Pod {
	return filterDPDKRunningPods(env.GetGuaranteedPodsWithExlusiveCPUs())
}

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
			logrus.Errorf("Failed to execute command %s in debug %s, errStr: %s, err: %v", findCommand, pod.String(), errStr, err)
			continue
		}
		if strings.Contains(outStr, dpdkDriver) {
			filteredPods = append(filteredPods, pod)
		}
	}
	return filteredPods
}

func (env *TestEnvironment) GetShareProcessNamespacePods() []*Pod {
	var filteredPods []*Pod
	for _, p := range env.Pods {
		if p.IsShareProcessNamespace() {
			filteredPods = append(filteredPods, p)
		}
	}
	return filteredPods
}
