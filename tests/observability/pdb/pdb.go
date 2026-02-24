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

// ZoneAwareCheckResult contains the result of a zone-aware PDB validation
type ZoneAwareCheckResult struct {
	IsValid             bool
	BasicCheckError     error
	ZoneCheckError      error
	MaxReplicasPerZone  int
	NumZones            int
	Replicas            int
	MinAvailableValue   int
	MaxUnavailableValue int
}

// percentageToFloat converts a percentage string to a float
func percentageToFloat(percentage string) (float64, error) {
	var percentageFloat float64
	_, err := fmt.Sscanf(percentage, "%f%%", &percentageFloat)
	if err != nil {
		return 0, err
	}
	return percentageFloat / percentageDivisor, nil
}

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

// intOrStringToValue converts a PDB value (integer or percentage) to an absolute integer.
//
// For integers: returns the value directly (e.g., 3 → 3)
// For percentages: converts to absolute value using replicas (e.g., "60%" with 5 replicas → 3)
//
// This allows uniform comparison regardless of whether the PDB uses integers or percentages.
func intOrStringToValue(intOrStr *intstr.IntOrString, replicas int32) (int, error) {
	switch intOrStr.Type {
	case intstr.Int:
		return intOrStr.IntValue(), nil
	case intstr.String:
		// Convert percentage to absolute value: "60%" with 5 replicas → 0.60 * 5 = 3
		v, err := percentageToFloat(intOrStr.StrVal)
		if err != nil {
			return 0, fmt.Errorf("invalid value %q: %v", intOrStr.StrVal, err)
		}
		return int(math.RoundToEven(v * float64(replicas))), nil
	}
	return 0, fmt.Errorf("invalid type: neither int nor percentage")
}

// isZoneAwareForDraining checks if the PDB's constraints allow draining all pods in a single zone.
// For percentage values, intOrStringToValue() converts them to absolute values (e.g., "60%" with 5 replicas → 3).
// A PDB typically specifies EITHER maxUnavailable OR minAvailable (not both).
// The check passes if the specified constraint allows zone draining.
func isZoneAwareForDraining(pdb *policyv1.PodDisruptionBudget, maxReplicasPerZone int,
	minAvailableValue, maxUnavailableValue int, replicaCount int32) bool {
	if pdb.Spec.MaxUnavailable != nil {
		if maxUnavailableValue >= maxReplicasPerZone {
			return true
		}
	}

	if pdb.Spec.MinAvailable != nil {
		maxAllowedMinAvailable := int(replicaCount) - maxReplicasPerZone
		if minAvailableValue <= maxAllowedMinAvailable {
			return true
		}
	}

	return false
}

// CheckPDBIsZoneAware validates that a PDB can tolerate an entire zone going offline.
// This is important during platform upgrades where all workers in a zone may be unavailable.
//
// Formula:
//
//	max_replicas_per_zone = ceil(nr_replicas / nr_zones)
//
// The PDB must satisfy ONE of these conditions:
//
//	For maxUnavailable:
//	  - Integer: maxUnavailable >= max_replicas_per_zone
//	  - Percentage: maxUnavailable >= (max_replicas_per_zone / nr_replicas) * 100
//
//	For minAvailable:
//	  - Integer: minAvailable <= (nr_replicas - max_replicas_per_zone)
//	  - Percentage: minAvailable <= ((nr_replicas - max_replicas_per_zone) / nr_replicas) * 100
//
// Example 1: 2 zones, 5 replicas → max_replicas_per_zone = ceil(5/2) = 3
//   - maxUnavailable >= 3 (int) or >= 60% (3/5 * 100)
//   - minAvailable <= 2 (int) or <= 40% (2/5 * 100)
//
// Example 2: 3 zones, 9 replicas → max_replicas_per_zone = 9/3 = 3
//   - maxUnavailable >= 3 (int) or >= 33% (3/9 * 100)
//   - minAvailable <= 6 (int) or <= 66% (6/9 * 100)
func CheckPDBIsZoneAware(pdb *policyv1.PodDisruptionBudget, replicas *int32, numZones int) *ZoneAwareCheckResult {
	result := &ZoneAwareCheckResult{
		NumZones: numZones,
	}

	var replicaCount int32
	if replicas != nil {
		replicaCount = *replicas
	} else {
		replicaCount = 1 // default value
	}
	result.Replicas = int(replicaCount)

	// Skip zone-aware check for single-zone clusters or SNO
	if numZones <= 1 {
		result.IsValid = true
		return result
	}

	// Calculate max replicas that could be in a single zone
	// Using ceiling division: ceil(replicas / zones)
	maxReplicasPerZone := int(math.Ceil(float64(replicaCount) / float64(numZones)))
	result.MaxReplicasPerZone = maxReplicasPerZone

	// Get minAvailable and maxUnavailable values
	var minAvailableValue int
	var maxUnavailableValue int

	if pdb.Spec.MinAvailable != nil {
		var err error
		minAvailableValue, err = intOrStringToValue(pdb.Spec.MinAvailable, replicaCount)
		if err != nil {
			result.ZoneCheckError = fmt.Errorf("failed to parse minAvailable: %v", err)
			return result
		}
		result.MinAvailableValue = minAvailableValue
	}

	if pdb.Spec.MaxUnavailable != nil {
		var err error
		maxUnavailableValue, err = intOrStringToValue(pdb.Spec.MaxUnavailable, replicaCount)
		if err != nil {
			result.ZoneCheckError = fmt.Errorf("failed to parse maxUnavailable: %v", err)
			return result
		}
		result.MaxUnavailableValue = maxUnavailableValue
	}

	if !isZoneAwareForDraining(pdb, maxReplicasPerZone, minAvailableValue, maxUnavailableValue, replicaCount) {
		minAllowedMaxUnavailable := maxReplicasPerZone
		maxAllowedMinAvailable := int(replicaCount) - maxReplicasPerZone
		result.ZoneCheckError = fmt.Errorf(
			"PDB is not zone-aware: with %d replicas across %d zones, max %d pods could be in one zone. "+
				"Either set maxUnavailable >= %d or minAvailable <= %d to survive a zone failure",
			replicaCount, numZones, maxReplicasPerZone,
			minAllowedMaxUnavailable, maxAllowedMinAvailable)
		return result
	}

	result.IsValid = true
	return result
}
