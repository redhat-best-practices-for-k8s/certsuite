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
