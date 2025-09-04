package clusteroperator

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/ocplite"
)

func IsClusterOperatorAvailable(co *ocplite.ClusterOperator) bool {
	// Loop through the conditions, looking for the 'Available' state.
	for _, condition := range co.Status.Conditions {
		if condition.Type == ocplite.OperatorAvailable {
			log.Info("ClusterOperator %q is in an 'Available' state", co.Name)
			return true
		}
	}

	log.Info("ClusterOperator %q is not in an 'Available' state", co.Name)
	return false
}
