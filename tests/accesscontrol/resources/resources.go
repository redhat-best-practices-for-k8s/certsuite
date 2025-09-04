package resources

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
)

// HasRequestsSet Determines if a container has resource requests defined
//
// This function examines the request fields of a container's resource
// specification. It checks that there is at least one request entry, and that
// both CPU and memory requests are non‑zero values. If any requirement is
// missing it logs an error and returns false; otherwise it returns true.
func HasRequestsSet(cut *provider.Container, logger *log.Logger) bool {
	passed := true

	// Parse the requests.
	if len(cut.Resources.Requests) == 0 {
		logger.Error("Container %q has been found missing resource requests", cut)
		passed = false
	} else {
		if cut.Resources.Requests.Cpu().IsZero() {
			logger.Error("Container %q has been found missing CPU requests", cut)
			passed = false
		}

		if cut.Resources.Requests.Memory().IsZero() {
			logger.Error("Container %q has been found missing memory requests", cut)
			passed = false
		}
	}
	return passed
}

// HasExclusiveCPUsAssigned Determines if a container runs with exclusive CPU allocation
//
// The function examines the CPU and memory limits and requests of a container
// to decide whether it belongs to an exclusive CPU pool. If either limit is
// missing, non‑integer, or mismatched with its request, the container is
// considered shared; otherwise it is marked exclusive. The result is returned
// as a boolean.
func HasExclusiveCPUsAssigned(cut *provider.Container, logger *log.Logger) bool {
	cpuLimits := cut.Resources.Limits.Cpu()
	memLimits := cut.Resources.Limits.Memory()

	// if no cpu or memory limits are specified the container will run in the shared cpu pool
	if cpuLimits.IsZero() || memLimits.IsZero() {
		logger.Debug("Container %q has been found missing cpu/memory resource limits", cut)
		return false
	}

	// if the cpu limits quantity is not an integer the container will run in the shared cpu pool
	cpuLimitsVal, isInteger := cpuLimits.AsInt64()
	if !isInteger {
		logger.Debug("Container %q cpu resource limit is not an integer", cut)
		return false
	}

	// if the cpu and memory limits and requests are equal to each other the container will run in the exclusive cpu pool
	cpuRequestsVal, _ := cut.Resources.Requests.Cpu().AsInt64()
	memRequestsVal, _ := cut.Resources.Requests.Memory().AsInt64()
	memLimitsVal, _ := memLimits.AsInt64()
	if cpuLimitsVal == cpuRequestsVal && memLimitsVal == memRequestsVal {
		return true
	}

	// if the cpu limits and request are different, the container will run in the shared cpu pool
	logger.Debug("Container %q cpu/memory resources and limits are not equal to each other", cut)
	return false
}
