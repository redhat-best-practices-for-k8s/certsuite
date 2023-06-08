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
			expectedError: fmt.Errorf("field .spec.minAvailable cannot be zero. Currently set to: 0"),
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
			expectedError: fmt.Errorf("field .spec.maxUnavailable cannot be greater than or equal to the number of pods in the replica. Currently set to: 1"),
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
	}

	for _, tc := range testCases {
		result, err := CheckPDBIsValid(tc.testPDB, tc.replicas)
		assert.Equal(t, tc.expectedBool, result)
		assert.Equal(t, tc.expectedError, err)
	}
}
