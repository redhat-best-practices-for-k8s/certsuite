package pdb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestCheckPDBIsValid(t *testing.T) {
	int32Ptr := func(i int32) *int32 { return &i }

	testCases := []struct {
		testPDB       *policyv1.PodDisruptionBudget
		replicas      *int32
		expectedBool  bool
		expectedError error
	}{
		{ // Test Case #1 - MinAvailable is zero
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 0,
					},
				},
			},
			replicas:      nil,
			expectedBool:  false,
			expectedError: fmt.Errorf("field .spec.minAvailable cannot be zero. Currently set to: 0. Replicas set to: 1"),
		},
		{ // Test Case #2 - MaxUnavailable is greater than or equal to the number of pods in the replica
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1,
					},
				},
			},
			replicas:      nil,
			expectedBool:  false,
			expectedError: fmt.Errorf("field .spec.maxUnavailable cannot be greater than or equal to the number of pods in the replica. Currently set to: 1. Replicas set to: 1"),
		},
		{ // Test Case #3 - MinAvailable is not zero and MaxUnavailable is less than the number of pods in the replica
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1,
					},
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 0,
					},
				},
			},
			replicas:      nil,
			expectedBool:  true,
			expectedError: nil,
		},
		{ // Test Case #4 - MinAvailable is not zero and MaxUnavailable is nil
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1,
					},
					MaxUnavailable: nil,
				},
			},
			replicas:      nil,
			expectedBool:  true,
			expectedError: nil,
		},
		{ // Test Case #5 - MinAvailable is nil and MaxUnavailable is zero, replicas 1
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: nil,
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 0,
					},
				},
			},
			replicas:      int32Ptr(1),
			expectedBool:  true,
			expectedError: nil,
		},
		{ // Test Case #6 - Replicas is set to 1, MinAvailable is 1 and MaxUnavailable is nil
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1,
					},
					MaxUnavailable: nil,
				},
			},
			replicas:      int32Ptr(1),
			expectedBool:  true,
			expectedError: nil,
		},
		{ // Test Case #7 - MinAvailable is set to a percentage, replicas set to 1, fail because 0.5 replicas is not valid
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "50%",
					},
					MaxUnavailable: nil,
				},
			},
			replicas:      int32Ptr(1),
			expectedBool:  false,
			expectedError: fmt.Errorf("field .spec.minAvailable cannot be zero. Currently set to: 0. Replicas set to: 1"),
		},
		{ // Test Case #8 - MinAvailable is set to a percentage, replicas set to 2, passes because 1 replica is valid
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "50%",
					},
					MaxUnavailable: nil,
				},
			},
			replicas:      int32Ptr(2),
			expectedBool:  true,
			expectedError: nil,
		},
		{ // Test Case #9 - MaxAvailable and MinAvailable are set to a percentage, replicas set to 2, passes because 1 replica is valid
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "50%",
					},
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "50%",
					},
				},
			},
			replicas:      int32Ptr(2),
			expectedBool:  true,
			expectedError: nil,
		},
		{ // Test Case #10 - MaxAvailable set to 90%, the replicas is set to 10, passes because 1 replica is valid
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: nil,
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "90%",
					},
				},
			},
			replicas:      int32Ptr(10),
			expectedBool:  true,
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		result, err := CheckPDBIsValid(tc.testPDB, tc.replicas)
		assert.Equal(t, tc.expectedBool, result)
		assert.Equal(t, tc.expectedError, err)
	}
}

func TestIntOrStringToValue(t *testing.T) {
	testCases := []struct {
		intOrString  intstr.IntOrString
		testReplicas int32
		expectedInt  int
		expectedErr  error
	}{
		{
			intOrString: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 1,
			},
			testReplicas: 1,
			expectedInt:  1,
		},
		{
			intOrString: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "50%",
			},
			testReplicas: 1,
			expectedInt:  0,
		},
		{
			intOrString: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "50",
			},
			testReplicas: 1,
			expectedInt:  0,
			expectedErr:  fmt.Errorf("invalid value \"50\": unexpected EOF"),
		},
	}

	for index, tc := range testCases {
		result, err := intOrStringToValue(&testCases[index].intOrString, tc.testReplicas)
		assert.Equal(t, tc.expectedInt, result)
		assert.Equal(t, tc.expectedErr, err)
	}
}

func TestPercentageToFloat(t *testing.T) {
	testCases := []struct {
		percentage    string
		expectedFloat float64
		expectedErr   error
	}{
		{
			percentage:    "50%",
			expectedFloat: 0.5,
			expectedErr:   nil,
		},
		{
			percentage:    "50",
			expectedFloat: 0,
			expectedErr:   fmt.Errorf("unexpected EOF"),
		},
		{
			percentage:    "0%",
			expectedFloat: 0,
			expectedErr:   nil,
		},
		{
			percentage:    "100%",
			expectedFloat: 1,
			expectedErr:   nil,
		},
	}

	for _, tc := range testCases {
		result, err := percentageToFloat(tc.percentage)
		assert.Equal(t, tc.expectedFloat, result)
		assert.Equal(t, tc.expectedErr, err)
	}
}

func TestCheckPDBIsZoneAware(t *testing.T) {
	int32Ptr := func(i int32) *int32 { return &i }

	testCases := []struct {
		name         string
		testPDB      *policyv1.PodDisruptionBudget
		replicas     *int32
		numZones     int
		expectedPass bool
	}{
		{
			// Test Case #1 - Single zone cluster, should always pass
			name: "single zone cluster - always pass",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1,
					},
				},
			},
			replicas:     int32Ptr(1),
			numZones:     1,
			expectedPass: true,
		},
		{
			// Test Case #2 - 3 zones, 3 replicas, maxUnavailable=1 - FAIL
			// ceil(3/3)=1 pod per zone, need maxUnavailable>=1 to survive zone failure
			// but maxUnavailable=1 means all replicas could be unavailable
			name: "3 zones 3 replicas maxUnavailable=1 - pass (can lose 1 zone)",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1,
					},
				},
			},
			replicas:     int32Ptr(3),
			numZones:     3,
			expectedPass: true, // maxUnavailable(1) >= maxReplicasPerZone(1)
		},
		{
			// Test Case #3 - 3 zones, 6 replicas, maxUnavailable=1 - FAIL
			// ceil(6/3)=2 pods per zone, maxUnavailable=1 is not enough
			name: "3 zones 6 replicas maxUnavailable=1 - fail",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1,
					},
				},
			},
			replicas:     int32Ptr(6),
			numZones:     3,
			expectedPass: false, // maxUnavailable(1) < maxReplicasPerZone(2)
		},
		{
			// Test Case #4 - 3 zones, 6 replicas, maxUnavailable=2 - PASS
			// ceil(6/3)=2 pods per zone, maxUnavailable=2 is enough
			name: "3 zones 6 replicas maxUnavailable=2 - pass",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 2,
					},
				},
			},
			replicas:     int32Ptr(6),
			numZones:     3,
			expectedPass: true, // maxUnavailable(2) >= maxReplicasPerZone(2)
		},
		{
			// Test Case #5 - 3 zones, 6 replicas, minAvailable=4 - PASS
			// ceil(6/3)=2 pods per zone, minAvailable<=6-2=4 is enough
			name: "3 zones 6 replicas minAvailable=4 - pass",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 4,
					},
				},
			},
			replicas:     int32Ptr(6),
			numZones:     3,
			expectedPass: true, // minAvailable(4) <= replicas(6) - maxReplicasPerZone(2)
		},
		{
			// Test Case #6 - 3 zones, 6 replicas, minAvailable=5 - FAIL
			// ceil(6/3)=2 pods per zone, need minAvailable<=4 to survive zone failure
			name: "3 zones 6 replicas minAvailable=5 - fail",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 5,
					},
				},
			},
			replicas:     int32Ptr(6),
			numZones:     3,
			expectedPass: false, // minAvailable(5) exceeds replicas(6) minus maxReplicasPerZone(2)
		},
		{
			// Test Case #7 - 3 zones, 5 replicas, maxUnavailable=2 - PASS
			// ceil(5/3)=2 pods per zone, maxUnavailable=2 is enough
			name: "3 zones 5 replicas maxUnavailable=2 - pass (uneven distribution)",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 2,
					},
				},
			},
			replicas:     int32Ptr(5),
			numZones:     3,
			expectedPass: true, // maxUnavailable(2) meets ceil(5/3) threshold
		},
		{
			// Test Case #8 - percentage based: 3 zones, 9 replicas, maxUnavailable=34% (rounds to 3) - PASS
			// ceil(9/3)=3 pods per zone, 34% of 9 = 3.06 rounds to 3
			name: "3 zones 9 replicas maxUnavailable=34% - pass",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "34%",
					},
				},
			},
			replicas:     int32Ptr(9),
			numZones:     3,
			expectedPass: true, // 34% of 9 = 3.06 rounds to 3 >= maxReplicasPerZone(3)
		},
		{
			// Test Case #9 - percentage based: 3 zones, 9 replicas, maxUnavailable=30% (rounds to 3) - PASS
			// ceil(9/3)=3 pods per zone, 30% of 9 = 2.7 rounds to 3
			name: "3 zones 9 replicas maxUnavailable=30% - pass",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "30%",
					},
				},
			},
			replicas:     int32Ptr(9),
			numZones:     3,
			expectedPass: true, // 30% of 9 = 2.7 rounds to 3 >= maxReplicasPerZone(3)
		},
		{
			// Test Case #10 - percentage based: 3 zones, 9 replicas, maxUnavailable=20% (rounds to 2) - FAIL
			// ceil(9/3)=3 pods per zone, 20% of 9 = 1.8 rounds to 2
			name: "3 zones 9 replicas maxUnavailable=20% - fail",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "20%",
					},
				},
			},
			replicas:     int32Ptr(9),
			numZones:     3,
			expectedPass: false, // 20% of 9 = 1.8 rounds to 2 < maxReplicasPerZone(3)
		},
		{
			// Test Case #11 - 2 zones, 4 replicas, minAvailable=2 - PASS
			// ceil(4/2)=2 pods per zone, minAvailable<=4-2=2 is enough
			name: "2 zones 4 replicas minAvailable=2 - pass",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 2,
					},
				},
			},
			replicas:     int32Ptr(4),
			numZones:     2,
			expectedPass: true, // minAvailable(2) within replicas(4) minus maxReplicasPerZone(2) limit
		},
		{
			// Test Case #12 - 2 zones, 4 replicas, minAvailable=3 - FAIL
			// ceil(4/2)=2 pods per zone, need minAvailable<=2 to survive zone failure
			name: "2 zones 4 replicas minAvailable=3 - fail",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 3,
					},
				},
			},
			replicas:     int32Ptr(4),
			numZones:     2,
			expectedPass: false, // minAvailable(3) exceeds replicas(4) minus maxReplicasPerZone(2)
		},
		{
			// Test Case #13 - numZones=0 should be treated as single zone
			name: "0 zones (no zone labels) - treated as single zone, pass",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1,
					},
				},
			},
			replicas:     int32Ptr(1),
			numZones:     0,
			expectedPass: true, // No zones = single zone = skip check
		},
		// ======================================================================
		// User-specified test cases (Examples from requirements)
		// ======================================================================
		{
			// Example 1: 2 zones, 5 replicas, max_replicas_per_zone=ceil(5/2)=3
			// maxUnavailable >= 60% (3/5) should PASS
			name: "Example1: 2 zones 5 replicas maxUnavailable=60% - pass",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "60%",
					},
				},
			},
			replicas:     int32Ptr(5),
			numZones:     2,
			expectedPass: true, // 60% of 5 = 3 >= max_replicas_per_zone(3)
		},
		{
			// Example 1: 2 zones, 5 replicas, max_replicas_per_zone=ceil(5/2)=3
			// maxUnavailable < 60% (e.g., 50%) should FAIL
			name: "Example1: 2 zones 5 replicas maxUnavailable=50% - fail",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "50%",
					},
				},
			},
			replicas:     int32Ptr(5),
			numZones:     2,
			expectedPass: false, // 50% of 5 = 2.5 rounds to 2 < max_replicas_per_zone(3)
		},
		{
			// Example 1: 2 zones, 5 replicas, max_replicas_per_zone=ceil(5/2)=3
			// minAvailable <= 2 should PASS
			name: "Example1: 2 zones 5 replicas minAvailable=2 - pass",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 2,
					},
				},
			},
			replicas:     int32Ptr(5),
			numZones:     2,
			expectedPass: true, // minAvailable(2) within replicas(5) minus max_replicas_per_zone(3) limit
		},
		{
			// Example 1: 2 zones, 5 replicas, max_replicas_per_zone=ceil(5/2)=3
			// minAvailable > 2 (e.g., 3) should FAIL
			name: "Example1: 2 zones 5 replicas minAvailable=3 - fail",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 3,
					},
				},
			},
			replicas:     int32Ptr(5),
			numZones:     2,
			expectedPass: false, // minAvailable(3) exceeds replicas(5) minus max_replicas_per_zone(3)
		},
		{
			// Example 2: 3 zones, 9 replicas, max_replicas_per_zone=9/3=3
			// maxUnavailable >= 33% should PASS (33% of 9 = 2.97 rounds to 3)
			name: "Example2: 3 zones 9 replicas maxUnavailable=33% - pass",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "33%",
					},
				},
			},
			replicas:     int32Ptr(9),
			numZones:     3,
			expectedPass: true, // 33% of 9 = 2.97 rounds to 3 >= max_replicas_per_zone(3)
		},
		{
			// Example 2: 3 zones, 9 replicas, max_replicas_per_zone=9/3=3
			// maxUnavailable >= 34% should definitely PASS
			name: "Example2: 3 zones 9 replicas maxUnavailable=34% - pass",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "34%",
					},
				},
			},
			replicas:     int32Ptr(9),
			numZones:     3,
			expectedPass: true, // 34% of 9 = 3.06 rounds to 3 >= max_replicas_per_zone(3)
		},
		{
			// Example 2: 3 zones, 9 replicas, max_replicas_per_zone=9/3=3
			// maxUnavailable < 33% (e.g., 25%) should FAIL
			name: "Example2: 3 zones 9 replicas maxUnavailable=25% - fail",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "25%",
					},
				},
			},
			replicas:     int32Ptr(9),
			numZones:     3,
			expectedPass: false, // 25% of 9 = 2.25 rounds to 2 < max_replicas_per_zone(3)
		},
		{
			// Example 2: 3 zones, 9 replicas, max_replicas_per_zone=9/3=3
			// minAvailable <= 6 should PASS
			name: "Example2: 3 zones 9 replicas minAvailable=6 - pass",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 6,
					},
				},
			},
			replicas:     int32Ptr(9),
			numZones:     3,
			expectedPass: true, // minAvailable(6) within replicas(9) minus max_replicas_per_zone(3) limit
		},
		{
			// Example 2: 3 zones, 9 replicas, max_replicas_per_zone=9/3=3
			// minAvailable > 6 (e.g., 7) should FAIL
			name: "Example2: 3 zones 9 replicas minAvailable=7 - fail",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 7,
					},
				},
			},
			replicas:     int32Ptr(9),
			numZones:     3,
			expectedPass: false, // minAvailable(7) exceeds replicas(9) minus max_replicas_per_zone(3)
		},
		// ======================================================================
		// OR logic test cases: when BOTH minAvailable AND maxUnavailable are specified
		// A PDB typically specifies only ONE, but if both are set, either passing is OK
		// ======================================================================
		{
			// Both specified, both pass - should PASS
			name: "OR logic: both constraints pass - pass",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 4, // <= 6-2=4 ✓
					},
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 2, // >= 2 ✓
					},
				},
			},
			replicas:     int32Ptr(6),
			numZones:     3,
			expectedPass: true, // Both constraints allow zone draining
		},
		{
			// Both specified, maxUnavailable passes but minAvailable fails - should PASS (OR logic)
			name: "OR logic: maxUnavailable ok but minAvailable fails - pass",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 5, // > 6-2=4 ✗
					},
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 2, // >= 2 ✓
					},
				},
			},
			replicas:     int32Ptr(6),
			numZones:     3,
			expectedPass: true, // maxUnavailable passes, so PDB is zone-aware (OR logic)
		},
		{
			// Both specified, minAvailable passes but maxUnavailable fails - should PASS (OR logic)
			name: "OR logic: minAvailable ok but maxUnavailable fails - pass",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 4, // <= 6-2=4 ✓
					},
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1, // < 2 ✗
					},
				},
			},
			replicas:     int32Ptr(6),
			numZones:     3,
			expectedPass: true, // minAvailable passes, so PDB is zone-aware (OR logic)
		},
		{
			// Both specified, both fail - should FAIL
			name: "OR logic: both constraints fail - fail",
			testPDB: &policyv1.PodDisruptionBudget{
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 5, // > 6-2=4 ✗
					},
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1, // < 2 ✗
					},
				},
			},
			replicas:     int32Ptr(6),
			numZones:     3,
			expectedPass: false, // Both constraints fail, so PDB is NOT zone-aware
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CheckPDBIsZoneAware(tc.testPDB, tc.replicas, tc.numZones)
			assert.Equal(t, tc.expectedPass, result.IsValid, "Test case: %s", tc.name)
		})
	}
}
