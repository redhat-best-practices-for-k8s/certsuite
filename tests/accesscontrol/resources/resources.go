package resources

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
)

// HasRequestsSet checks if a container has resource requests set.
// Returns :
//   - bool : true if resource requests are set for the container, otherwise return false.
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

// For more info on cpu management policies see https://kubernetes.io/docs/tasks/administer-cluster/cpu-management-policies/.
// HasExclusiveCPUsAssigned checks if a container has exclusive CPU's assigned.
// Returns:
//   - bool : true if a container has exclusive CPU's assigned, otherwise return false.
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
