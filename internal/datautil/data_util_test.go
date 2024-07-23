package datautil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMapSubset(t *testing.T) {
	testCasesStr := []struct {
		m        map[string]string
		s        map[string]string
		expected bool
	}{
		{
			m:        map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"},
			s:        map[string]string{"k1": "v1", "k2": "v2"},
			expected: true,
		},
		{
			m:        map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"},
			s:        map[string]string{"k1": "v1", "k0": "v0"},
			expected: false,
		},
		{
			m:        map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"},
			s:        map[string]string{"k1": "v1", "k2": "v2", "k3": "v3", "k0": "v0"},
			expected: false,
		},
		{
			m:        map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"},
			s:        map[string]string{},
			expected: true,
		},
	}

	for _, tc := range testCasesStr {
		assert.Equal(t, tc.expected, IsMapSubset(tc.m, tc.s))
	}

	testCasesInt := []struct {
		m        map[string]int
		s        map[string]int
		expected bool
	}{
		{
			m:        map[string]int{"k1": 1, "k2": 2, "k3": 3},
			s:        map[string]int{"k1": 1, "k2": 2},
			expected: true,
		},
		{
			m:        map[string]int{"k1": 1, "k2": 2, "k3": 3},
			s:        map[string]int{"k1": 1, "k0": 0},
			expected: false,
		},
		{
			m:        map[string]int{"k1": 1, "k2": 2, "k3": 3},
			s:        map[string]int{"k1": 1, "k2": 2, "k3": 3, "k0": 0},
			expected: false,
		},
	}

	for _, tc := range testCasesInt {
		assert.Equal(t, tc.expected, IsMapSubset(tc.m, tc.s))
	}
}
