package pdb

import (
	"fmt"
	"math"

	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	percentageDivisor = 100
)

// percentageToFloat Parses a percentage string into a decimal value
//
// The function reads a string that represents a and extracts the numeric part
// using formatted scanning. It then converts this number to a float64 and
// divides by a divisor to express it as a proportion, such as 0.25 for
// twenty‑five percent. If the input is not in the expected format, an error
// is returned.
func percentageToFloat(percentage string) (float64, error) {
	var percentageFloat float64
	_, err := fmt.Sscanf(percentage, "%f%%", &percentageFloat)
	if err != nil {
		return 0, err
	}
	return percentageFloat / percentageDivisor, nil
}

// CheckPDBIsValid Validates a PodDisruptionBudget against replica count
//
// The function checks the .spec.minAvailable and .spec.maxUnavailable fields of
// a PodDisruptionBudget, converting them to integer values based on the
// provided replica count or a default of one. It ensures minAvailable is
// non‑zero and does not exceed replicas, and that maxUnavailable is less than
// the number of pods. If any rule fails, it returns false with an explanatory
// error; otherwise it returns true.
func CheckPDBIsValid(pdb *policyv1.PodDisruptionBudget, replicas *int32) (bool, error) {
	var replicaCount int32
	if replicas != nil {
		replicaCount = *replicas
	} else {
		replicaCount = 1 // default value
	}

	var minAvailableValue int
	var maxUnavailableValue int

	if pdb.Spec.MinAvailable != nil {
		var err error
		minAvailableValue, err = intOrStringToValue(pdb.Spec.MinAvailable, replicaCount)
		if err != nil {
			return false, err
		}

		// Tests for the minAvailable spec
		if minAvailableValue == 0 {
			return false, fmt.Errorf("field .spec.minAvailable cannot be zero. Currently set to: %d. Replicas set to: %d", minAvailableValue, replicaCount)
		}

		if minAvailableValue > int(replicaCount) {
			return false, fmt.Errorf("minAvailable cannot be greater than replicas. Currently set to: %d. Replicas set to: %d", minAvailableValue, replicaCount)
		}
	}

	if pdb.Spec.MaxUnavailable != nil {
		var err error
		maxUnavailableValue, err = intOrStringToValue(pdb.Spec.MaxUnavailable, replicaCount)
		if err != nil {
			return false, err
		}

		// Tests for the maxUnavailable spec
		if maxUnavailableValue >= int(replicaCount) {
			return false, fmt.Errorf("field .spec.maxUnavailable cannot be greater than or equal to the number of pods in the replica. Currently set to: %d. Replicas set to: %d", maxUnavailableValue, replicaCount)
		}
	}

	return true, nil
}

// intOrStringToValue Converts an IntOrString to a concrete integer based on replica count
//
// The function examines the type of the input value; if it is an integer, that
// value is returned directly. If it is a string representing a percentage, the
// percentage is parsed and multiplied by the number of replicas, rounding to
// the nearest even integer. Errors are produced for unsupported types or
// invalid percentage strings.
func intOrStringToValue(intOrStr *intstr.IntOrString, replicas int32) (int, error) {
	switch intOrStr.Type {
	case intstr.Int:
		return intOrStr.IntValue(), nil
	case intstr.String:
		v, err := percentageToFloat(intOrStr.StrVal)
		if err != nil {
			return 0, fmt.Errorf("invalid value %q: %v", intOrStr.StrVal, err)
		}
		return int(math.RoundToEven(v * float64(replicas))), nil
	}
	return 0, fmt.Errorf("invalid type: neither int nor percentage")
}
