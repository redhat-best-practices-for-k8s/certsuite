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
