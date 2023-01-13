package resources

import (
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

func HasRequestsAndLimitsSet(cut *provider.Container) bool {
	passed := true
	// Parse the limits.
	if len(cut.Resources.Limits) == 0 {
		tnf.ClaimFilePrintf("Container has been found missing resource limits: %s", cut.String())
		passed = false
	} else {
		if cut.Resources.Limits.Cpu().Value() == 0 {
			tnf.ClaimFilePrintf("Container has been found missing CPU limits: %s", cut.String())
			passed = false
		}

		if cut.Resources.Limits.Memory().Value() == 0 {
			tnf.ClaimFilePrintf("Container has been found missing memory limits: %s", cut.String())
			passed = false
		}
	}

	// Parse the requests.
	if len(cut.Resources.Requests) == 0 {
		tnf.ClaimFilePrintf("Container has been found missing resource requests: %s", cut.String())
		passed = false
	} else {
		if cut.Resources.Requests.Cpu().Value() == 0 {
			tnf.ClaimFilePrintf("Container has been found missing CPU requests: %s", cut.String())
			passed = false
		}

		if cut.Resources.Requests.Memory().Value() == 0 {
			tnf.ClaimFilePrintf("Container has been found missing memory requests: %s", cut.String())
			passed = false
		}
	}
	return passed
}

// For more info on cpu mgmt polcies see https://kubernetes.io/docs/tasks/administer-cluster/cpu-management-policies/.
func HasExclusiveCPUsAssigned(cut *provider.Container) bool {
	cpuLimits := cut.Resources.Limits.Cpu()
	cpuRequests := cut.Resources.Requests.Cpu()
	// if no cpu limits are specified the container will run in the shared cpu pool
	if cpuLimits.IsZero() {
		return false
	}

	// if the cpu limits quantity is not an integer the container will run in the shared cpu pool
	cpuLimitsVal, isInteger := cpuLimits.AsInt64()
	if !isInteger {
		return false
	}

	// if the cpu requests are not specified they are set equal to limits, so the container will run in the exclusive cpu pool
	if cpuRequests.IsZero() {
		return true
	}

	// if the cpu requests quantity is not an integer the container will run the shared cpu pool
	cpuRequestsVal, isInteger := cpuRequests.AsInt64()
	if !isInteger {
		return false
	}

	// if the cpu limits and requests are integers and equal the container will run in the exclusive cpu pool
	if cpuLimitsVal == cpuRequestsVal {
		return true
	}

	// if the cpu limits and request are different, the container will run in the shared cpu pool
	return false
}
