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

// AreResourcesIdentical Verifies that CPU and memory requests match limits for every container in a pod
//
// The function iterates over all containers in the supplied pod, ensuring each
// has defined resource limits. It compares the request values to their
// corresponding limits for both CPU and memory; if any mismatch is found, it
// logs a debug message and returns false. When all containers satisfy these
// conditions, the function returns true.
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

// AreCPUResourcesWholeUnits Verifies that all CPU requests and limits are whole units
//
// The function iterates over each container in a pod, ensuring both CPU
// requests and limits are defined and expressed as multiples of one . If any
// container lacks these specifications or has non‑whole‑unit values, it
// logs the issue and returns false. When all containers meet the criteria, it
// returns true.
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

// LoadBalancingDisabled Determines if both CPU and IRQ load balancing are disabled via annotations
//
// The function checks a pod’s annotations for "cpu-load-balancing.crio.io"
// and "irq-load-balancing.crio.io", verifying each is set to the value
// "disable". If either annotation is missing or has an invalid value, it logs a
// debug message. It returns true only when both annotations are present with
// the correct value; otherwise it returns false.
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
