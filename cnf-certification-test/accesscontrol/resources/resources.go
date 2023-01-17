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
	memLimits := cut.Resources.Limits.Memory()
	memRequests := cut.Resources.Requests.Memory()

	// if no cpu or memory limits are specified the container will run in the shared cpu pool
	if cpuLimits.IsZero() || memLimits.IsZero() {
		tnf.ClaimFilePrintf("Container has been found missing cpu/memory resource limits: %s", cut.String())
		return false
	}

	// if the cpu limits quantity is not an integer the container will run in the shared cpu pool
	cpuLimitsVal, isInteger := cpuLimits.AsInt64()
	if !isInteger {
		tnf.ClaimFilePrintf("Container's cpu resource limit is not an integer: %s", cut.String())
		return false
	}

	// if the cpu and memory limits and requests are equal to each other the container will run in the exclusive cpu pool
	cpuRequestsVal, _ := cpuRequests.AsInt64()
	memRequestsVal, _ := memRequests.AsInt64()
	memLimitsVal, _ := memLimits.AsInt64()
	if cpuLimitsVal == cpuRequestsVal && memLimitsVal == memRequestsVal {
		return true
	}

	// if the cpu limits and request are different, the container will run in the shared cpu pool
	tnf.ClaimFilePrintf("Container's cpu/memory resources and limits are not equal to each other: %s", cut.String())
	return false
}
