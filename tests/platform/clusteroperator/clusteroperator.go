package clusteroperator

import (
	configv1 "github.com/openshift/api/config/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
)

// IsClusterOperatorAvailable checks whether a ClusterOperator is available and reports its status.
//
// It accepts a pointer to a ClusterOperator configuration object, logs information about the operator,
// and returns true if the operator is considered available based on its internal state.
// If the input is nil or the operator cannot be determined as available, it returns false.
func IsClusterOperatorAvailable(co *configv1.ClusterOperator) bool {
	// Loop through the conditions, looking for the 'Available' state.
	for _, condition := range co.Status.Conditions {
		if condition.Type == configv1.OperatorAvailable {
			log.Info("ClusterOperator %q is in an 'Available' state", co.Name)
			return true
		}
	}

	log.Info("ClusterOperator %q is not in an 'Available' state", co.Name)
	return false
}
