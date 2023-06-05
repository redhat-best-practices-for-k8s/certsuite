package pdb

import (
	"fmt"

	policyv1 "k8s.io/api/policy/v1"
)

func CheckPDBIsValid(pdb *policyv1.PodDisruptionBudget, replicas *int32) (bool, error) {
	var replicaCount int32
	if replicas != nil {
		replicaCount = *replicas
	} else {
		replicaCount = 1 // default value
	}

	if pdb.Spec.MinAvailable != nil && pdb.Spec.MinAvailable.IntValue() == 0 {
		return false, fmt.Errorf("field .spec.minAvailable cannot be zero")
	}

	if pdb.Spec.MaxUnavailable != nil && pdb.Spec.MaxUnavailable.IntValue() >= int(replicaCount) {
		return false, fmt.Errorf("field .spec.maxUnavailable cannot be greater than or equal to the number of pods in the replica")
	}

	return true, nil
}
