package resources

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
)

// HasRequestsSet checks if a container has resource requests set.
//
// It takes a Kubernetes container definition and a logger, examines the
// CPU and memory request fields, logs any errors encountered during
// parsing, and returns true if either request is non‑zero,
// otherwise false.
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

// HasExclusiveCPUsAssigned reports whether a container has exclusive CPUs assigned.
//
// It examines the CPU resource limits of the given container and determines if
// those limits correspond to an exclusive CPU allocation according to Kubernetes'
// cpu management policies. The function returns true when the container's CPU
// configuration indicates exclusivity, otherwise it returns false. A logger can
// be supplied for debug output.
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
