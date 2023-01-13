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
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	HugePages2Mi          = "hugepages-2Mi"
	HugePages1Gi          = "hugepages-1Gi"
	hugePages             = "hugepages"
	replicationController = "ReplicationController"
	deploymentConfig      = "DeploymentConfig"
)

type Pod struct {
	*corev1.Pod
	Containers         []*Container
	MultusIPs          map[string][]string
	MultusPCIs         []string
	SkipNetTests       bool
	SkipMultusNetTests bool
}

func NewPod(aPod *corev1.Pod) (out Pod) {
	var err error
	out.Pod = aPod
	out.MultusIPs = make(map[string][]string)
	out.MultusIPs, err = GetPodIPsPerNet(aPod.GetAnnotations()[CniNetworksStatusKey])
	if err != nil {
		logrus.Errorf("Could not decode networks-status annotation, error: %v", err)
	}

	out.MultusPCIs, err = GetPciPerPod(aPod.GetAnnotations()[CniNetworksStatusKey])
	if err != nil {
		logrus.Errorf("Could not decode networks-status annotation, error: %v", err)
	}

	if _, ok := aPod.GetLabels()[skipConnectivityTestsLabel]; ok {
		out.SkipNetTests = true
	}
	if _, ok := aPod.GetLabels()[skipMultusConnectivityTestsLabel]; ok {
		out.SkipMultusNetTests = true
	}
	out.Containers = append(out.Containers, getPodContainers(aPod, false)...)
	return out
}

func ConvertArrayPods(pods []*corev1.Pod) (out []*Pod) {
	for i := range pods {
		aPodWrapper := NewPod(pods[i])
		out = append(out, &aPodWrapper)
	}
	return out
}

func (p *Pod) IsPodGuaranteedWithExclusiveCPUs() bool {
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

func (p *Pod) AffinityRequired() bool {
	if val, ok := p.Labels[AffinityRequiredKey]; ok {
		result, err := strconv.ParseBool(val)
		if err != nil {
			logrus.Warnf("failure to parse bool %v", val)
			return false
		}
		return result
	}
	return false
}

func (p *Pod) HasHugepages() bool {
	// Pods may contain more than one container.  All containers must conform to the CPU isolation requirements.
	for _, cut := range p.Containers {
		for name := range cut.Resources.Requests {
			if strings.Contains(name.String(), hugePages) {
				return true
			}
		}
		for _, name := range cut.Resources.Limits {
			if strings.Contains(name.String(), hugePages) {
				return true
			}
		}
	}
	return false
}

func (p *Pod) CheckResourceHugePagesSize(size string) bool {
	// check if hugepages configuration other than 2Mi is present
	for _, cut := range p.Containers {
		// Resources must be specified
		if len(cut.Resources.Requests) == 0 || len(cut.Resources.Limits) == 0 {
			continue
		}
		for name := range cut.Resources.Requests {
			if strings.Contains(name.String(), hugePages) && name.String() != size {
				return false
			}
		}
		for name := range cut.Resources.Limits {
			if strings.Contains(name.String(), hugePages) && name.String() != size {
				return false
			}
		}
	}
	return true
}

func (p *Pod) IsAffinityCompliant() (bool, error) {
	if p.Spec.Affinity == nil {
		return false, fmt.Errorf("%s has been found with an AffinityRequired flag but is missing corresponding affinity rules", p.String())
	}
	if p.Spec.Affinity.PodAntiAffinity != nil {
		return false, fmt.Errorf("%s has been found with an AffinityRequired flag but has anti-affinity rules", p.String())
	}
	if p.Spec.Affinity.PodAffinity == nil && p.Spec.Affinity.NodeAffinity == nil {
		return false, fmt.Errorf("%s has been found with an AffinityRequired flag but is missing corresponding pod/node affinity rules", p.String())
	}
	return true, nil
}

func (p *Pod) IsShareProcessNamespace() bool {
	return p.Spec.ShareProcessNamespace != nil && *p.Spec.ShareProcessNamespace
}

func (p *Pod) ContainsIstioProxy() bool {
	for _, container := range p.Containers {
		if container.Name == "istio-proxy" {
			return true
		}
	}
	return false
}

func (p *Pod) CreatedByDeploymentConfig() (bool, error) {
	oc := clientsholder.GetClientsHolder()
	for _, podOwner := range p.ObjectMeta.GetOwnerReferences() {
		if podOwner.Kind == replicationController {
			replicationControllers, err := oc.K8sClient.CoreV1().ReplicationControllers(p.Namespace).Get(context.TODO(), podOwner.Name, v1.GetOptions{})
			if err != nil {
				return false, err
			}
			for _, rcOwner := range replicationControllers.GetOwnerReferences() {
				if rcOwner.Name == podOwner.Name && rcOwner.Kind == deploymentConfig {
					return true, err
				}
			}
		}
	}
	return false, nil
}

func (p *Pod) HasNodeAssignment() bool {
	// Checks whether or not the pod has a nodeSelector or a NodeName supplied
	return p.Spec.NodeSelector != nil || p.Spec.NodeName != ""
}
