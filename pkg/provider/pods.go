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
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	corev1 "k8s.io/api/core/v1"
)

type Pod struct {
	*corev1.Pod
	Containers         []*Container
	MultusIPs          map[string][]string
	SkipNetTests       bool
	SkipMultusNetTests bool
}

func NewPod(aPod *corev1.Pod) (out Pod) {
	var err error
	out.Pod = aPod
	out.MultusIPs = make(map[string][]string)
	out.MultusIPs, err = GetPodIPsPerNet(aPod.GetAnnotations()[CniNetworksStatusKey])
	if err != nil {
		logrus.Errorf("Could not decode networks-status annotation, error: %s", err)
	}

	if _, ok := aPod.GetLabels()[skipConnectivityTestsLabel]; ok {
		out.SkipNetTests = true
	}
	if _, ok := aPod.GetLabels()[skipMultusConnectivityTestsLabel]; ok {
		out.SkipMultusNetTests = true
	}
	out.Containers = append(out.Containers, getPodContainers(aPod)...)
	return out
}

func ConvertArrayPods(pods []*corev1.Pod) (out []*Pod) {
	for i := range pods {
		aPodWrapper := NewPod(pods[i])
		out = append(out, &aPodWrapper)
	}
	return out
}

func (p *Pod) IsPodGuaranteed() bool {
	return AreCPUResourcesWholeUnits(p) && AreResourcesIdentical(p)
}

func (p *Pod) IsCPUIsolationCompliant() bool {
	isCPUIsolated := true

	if !LoadBalancingDisabled(p) {
		errMsg := fmt.Sprintf("%s has been found to not have annotations set correctly for CPU isolation.", p.String())
		logrus.Debugf(errMsg)
		tnf.ClaimFilePrintf(errMsg)
		isCPUIsolated = false
	}

	if !IsRuntimeClassNameSpecified(p) {
		errMsg := fmt.Sprintf("%s has been found to not have runtimeClassName specified.", p.String())
		logrus.Debugf(errMsg)
		tnf.ClaimFilePrintf(errMsg)
		isCPUIsolated = false
	}

	return isCPUIsolated
}

func (p *Pod) String() string {
	return fmt.Sprintf("pod: %s ns: %s",
		p.Name,
		p.Namespace,
	)
}

func (p *Pod) ColocationEnabled() bool {
	if val, ok := p.Labels[AffinityAllowedKey]; ok {
		result, err := strconv.ParseBool(val)
		if err != nil {
			return false
		}
		return result
	}
	return false
}
