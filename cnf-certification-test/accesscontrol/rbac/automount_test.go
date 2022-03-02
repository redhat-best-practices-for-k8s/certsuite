package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetermineStatus(t *testing.T) {
	testCases := []struct {
		testField     string
		expectedSet   bool
		expectedValue bool
	}{
		{ // Test Case #1 - Nil field
			testField:     "nil",
			expectedSet:   false,
			expectedValue: false,
		},
		{ // Test Case #2 - True field
			testField:     "true",
			expectedSet:   true,
			expectedValue: true,
		},
		{ // Test Case #2 - False field
			testField:     "false",
			expectedSet:   true,
			expectedValue: false,
		},
	}

	for _, tc := range testCases {
		var tf *bool
		trueVar := true
		falseVar := false
		if tc.testField == "nil" {
			tf = nil
		} else if tc.testField == "true" {
			tf = &trueVar
		} else if tc.testField == "false" {
			tf = &falseVar
		}

		saStatus := DetermineStatus(tf)
		assert.Equal(t, tc.expectedSet, saStatus.TokenSet)
		assert.Equal(t, tc.expectedValue, saStatus.TokenValue)
	}
}
