package clusteroperator

import (
	configv1 "github.com/openshift/api/config/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
)

// IsClusterOperatorAvailable Determines if a ClusterOperator reports an 'Available' status
//
// The function inspects the conditions of a given cluster operator, checking
// for one whose type indicates availability. If such a condition is found, it
// logs that the operator is available and returns true; otherwise it logs that
// the operator is not available and returns false.
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
