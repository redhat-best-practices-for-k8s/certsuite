// Copyright (C) 2022-2023 Red Hat, Inc.
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
	"github.com/sirupsen/logrus"
)

func AreResourcesIdentical(p *Pod) bool {
	// Pods may contain more than one container.  All containers must conform to the CPU isolation requirements.
	for _, cut := range p.Containers {
		// At least limits must be specified (requests default to limits if not specified)
		if len(cut.Resources.Limits) == 0 {
			logrus.Debugf("%s has been found with undefined limits.", cut.String())
			return false
		}

		// Gather the values
		cpuRequests := cut.Resources.Requests.Cpu()
		cpuLimits := cut.Resources.Limits.Cpu()
		memoryRequests := cut.Resources.Requests.Memory()
		memoryLimits := cut.Resources.Limits.Memory()

		// Check for mismatches
		if !cpuRequests.Equal(*cpuLimits) {
			logrus.Debugf("%s has CPU requests %f and limits %f that do not match.", cut.String(), cpuRequests.AsApproximateFloat64(), cpuLimits.AsApproximateFloat64())
			return false
		}

		if !memoryRequests.Equal(*memoryLimits) {
			logrus.Debugf("%s has memory requests %f and limits %f that do not match.", cut.String(), memoryRequests.AsApproximateFloat64(), memoryLimits.AsApproximateFloat64())
			return false
		}
	}

	return true
}

func AreCPUResourcesWholeUnits(p *Pod) bool {
	isInteger := func(val int64) bool {
		return val%1000 == 0
	}

	// Pods may contain more than one container.  All containers must conform to the CPU isolation requirements.
	for _, cut := range p.Containers {
		// Resources must be specified
		if len(cut.Resources.Requests) == 0 || len(cut.Resources.Limits) == 0 {
			logrus.Debugf("%s has been found with undefined requests or limits.", cut.String())
			return false
		}

		// Gather the values
		cpuRequests := cut.Resources.Requests.Cpu().MilliValue()
		cpuLimits := cut.Resources.Limits.Cpu().MilliValue()

		if !isInteger(cpuRequests) {
			logrus.Debugf("%s has CPU requests %d (milli) that has to be a whole unit.", cut.String(), cpuRequests)
			return false
		}
		if !isInteger(cpuLimits) {
			logrus.Debugf("%s has CPU limits %d (milli) that has to be a whole unit.", cut.String(), cpuLimits)
			return false
		}
	}

	return true
}

func LoadBalancingDisabled(p *Pod) bool {
	const (
		disableVar = "disable"
	)

	cpuLoadBalancingDisabled := false
	irqLoadBalancingDisabled := false

	if v, ok := p.ObjectMeta.Annotations["cpu-load-balancing.crio.io"]; ok {
		if v == disableVar {
			cpuLoadBalancingDisabled = true
		} else {
			logrus.Debugf("Annotation cpu-load-balancing.crio.io has an invalid value for CPU isolation.  Must be 'disable'.")
		}
	} else {
		logrus.Debugf("Annotation cpu-load-balancing.crio.io is missing.")
	}

	if v, ok := p.ObjectMeta.Annotations["irq-load-balancing.crio.io"]; ok {
		if v == disableVar {
			irqLoadBalancingDisabled = true
		} else {
			logrus.Debugf("Annotation irq-load-balancing.crio.io has an invalid value for CPU isolation.  Must be 'disable'.")
		}
	} else {
		logrus.Debugf("Annotation irq-load-balancing.crio.io is missing.")
	}

	// Both conditions have to be set to 'disable'
	if cpuLoadBalancingDisabled && irqLoadBalancingDisabled {
		return true
	}

	return false
}
