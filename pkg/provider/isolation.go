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
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
)

// AreResourcesIdentical reports whether the requested and allocated resources for a pod match.
//
// It examines each container in the pod, comparing CPU and memory requests to limits.
// If all containers have equal request and limit values for both CPU and memory,
// the function returns true; otherwise it returns false. The comparison uses
// approximate equality to account for unit conversions. No other fields of the
// Pod are considered.
func AreResourcesIdentical(p *Pod) bool {
	// Pods may contain more than one container.  All containers must conform to the CPU isolation requirements.
	for _, cut := range p.Containers {
		// At least limits must be specified (requests default to limits if not specified)
		if len(cut.Resources.Limits) == 0 {
			log.Debug("%s has been found with undefined limits.", cut.String())
			return false
		}

		// Gather the values
		cpuRequests := cut.Resources.Requests.Cpu()
		cpuLimits := cut.Resources.Limits.Cpu()
		memoryRequests := cut.Resources.Requests.Memory()
		memoryLimits := cut.Resources.Limits.Memory()

		// Check for mismatches
		if !cpuRequests.Equal(*cpuLimits) {
			log.Debug("%s has CPU requests %f and limits %f that do not match.", cut.String(), cpuRequests.AsApproximateFloat64(), cpuLimits.AsApproximateFloat64())
			return false
		}

		if !memoryRequests.Equal(*memoryLimits) {
			log.Debug("%s has memory requests %f and limits %f that do not match.", cut.String(), memoryRequests.AsApproximateFloat64(), memoryLimits.AsApproximateFloat64())
			return false
		}
	}

	return true
}

// AreCPUResourcesWholeUnits determines whether all CPU resource specifications in a pod use whole unit values.
//
// It inspects the pod's containers and initContainers, evaluating both
// CPU request and limit fields. If any value contains a fractional part,
// the function logs a debug message and returns false; otherwise it
// returns true indicating that every CPU resource is an integer number
// of cores. The input parameter is a pointer to a Pod object from
// k8s.io/api/core/v1, and the return type is a boolean.
func AreCPUResourcesWholeUnits(p *Pod) bool {
	isInteger := func(val int64) bool {
		return val%1000 == 0
	}

	// Pods may contain more than one container.  All containers must conform to the CPU isolation requirements.
	for _, cut := range p.Containers {
		// Resources must be specified
		cpuRequestsMillis := cut.Resources.Requests.Cpu().MilliValue()
		cpuLimitsMillis := cut.Resources.Limits.Cpu().MilliValue()

		if cpuRequestsMillis == 0 || cpuLimitsMillis == 0 {
			log.Debug("%s has been found with undefined requests or limits.", cut.String())
			return false
		}

		if !isInteger(cpuRequestsMillis) {
			log.Debug("%s has CPU requests %d (milli) that has to be a whole unit.", cut.String(), cpuRequestsMillis)
			return false
		}
		if !isInteger(cpuLimitsMillis) {
			log.Debug("%s has CPU limits %d (milli) that has to be a whole unit.", cut.String(), cpuLimitsMillis)
			return false
		}
	}

	return true
}

// LoadBalancingDisabled determines whether load balancing is disabled for a pod.
//
// It examines the given Pod's annotations or labels to decide if the
// pod should skip network load‑balancing checks. The function returns true
// when load balancing is explicitly turned off, and false otherwise.
func LoadBalancingDisabled(p *Pod) bool {
	const (
		disableVar = "disable"
	)

	cpuLoadBalancingDisabled := false
	irqLoadBalancingDisabled := false

	if v, ok := p.Annotations["cpu-load-balancing.crio.io"]; ok {
		if v == disableVar {
			cpuLoadBalancingDisabled = true
		} else {
			log.Debug("Annotation cpu-load-balancing.crio.io has an invalid value for CPU isolation.  Must be 'disable'.")
		}
	} else {
		log.Debug("Annotation cpu-load-balancing.crio.io is missing.")
	}

	if v, ok := p.Annotations["irq-load-balancing.crio.io"]; ok {
		if v == disableVar {
			irqLoadBalancingDisabled = true
		} else {
			log.Debug("Annotation irq-load-balancing.crio.io has an invalid value for CPU isolation.  Must be 'disable'.")
		}
	} else {
		log.Debug("Annotation irq-load-balancing.crio.io is missing.")
	}

	// Both conditions have to be set to 'disable'
	if cpuLoadBalancingDisabled && irqLoadBalancingDisabled {
		return true
	}

	return false
}
