// Copyright (C) 2020-2024 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

/*
Package operator provides CNFCERT tests used to validate operator CNF facets.
*/

package operator

import (
	"testing"

	"github.com/blang/semver/v4"
	opFrameworkVersion "github.com/operator-framework/api/pkg/lib/version"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSplitCsv(t *testing.T) {
	tests := []struct {
		input       string
		expectedCsv string
		expectedNs  string
	}{
		{
			input:       "hazelcast-platform-operator.v5.12.0, ns=tnf",
			expectedCsv: "hazelcast-platform-operator.v5.12.0",
			expectedNs:  "tnf",
		},
		{
			input:       "example-operator.v1.0.0, ns=example-ns",
			expectedCsv: "example-operator.v1.0.0",
			expectedNs:  "example-ns",
		},
		{
			input:       "another-operator.v2.3.1, ns=another-ns",
			expectedCsv: "another-operator.v2.3.1",
			expectedNs:  "another-ns",
		},
		{
			input:       "no-namespace",
			expectedCsv: "no-namespace",
			expectedNs:  "",
		},
		{
			input:       "ns=onlynamespace",
			expectedCsv: "",
			expectedNs:  "onlynamespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SplitCsv(tt.input)
			if result.NameCsv != tt.expectedCsv {
				t.Errorf("splitCsv(%q) got namecsv %q, want %q", tt.input, result.NameCsv, tt.expectedCsv)
			}
			if result.Namespace != tt.expectedNs {
				t.Errorf("splitCsv(%q) got namespace %q, want %q", tt.input, result.Namespace, tt.expectedNs)
			}
		})
	}
}

func TestOperatorInstalledMoreThanOnce(t *testing.T) {
	generateOperator := func(name, csvName string, major, minor, patch int) *provider.Operator {
		return &provider.Operator{
			Name: name,
			Csv: &v1alpha1.ClusterServiceVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name: csvName,
				},
				Spec: v1alpha1.ClusterServiceVersionSpec{
					Version: opFrameworkVersion.OperatorVersion{
						Version: semver.Version{
							Major: uint64(major),
							Minor: uint64(minor),
							Patch: uint64(patch),
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		testOp1        *provider.Operator
		testOp2        *provider.Operator
		expectedOutput bool
	}{
		{ // Test Case #1 - Both operators are nil
			testOp1:        nil,
			testOp2:        nil,
			expectedOutput: false,
		},
		{ // Test Case #2 - One operator is nil
			testOp1:        &provider.Operator{},
			testOp2:        nil,
			expectedOutput: false,
		},
		{ // Test Case #3 - Both operators are different
			testOp1:        generateOperator("test-operator-1", "test-operator-1.v1.0.0", 1, 0, 0),
			testOp2:        generateOperator("test-operator-2", "test-operator-2.v1.0.0", 1, 0, 0),
			expectedOutput: false,
		},
		{ // Test Case #4 - Both operators are the same but different versions.
			// OLM does not allow the same operator to be installed more than once with the same version.
			testOp1:        generateOperator("test-operator", "test-operator.v1.0.0", 1, 0, 0),
			testOp2:        generateOperator("test-operator", "test-operator.v1.0.1", 1, 0, 1),
			expectedOutput: true,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, OperatorInstalledMoreThanOnce(tc.testOp1, tc.testOp2))
	}
}

func TestIsCsvInNamespaceClusterWide(t *testing.T) {
	allCsvs := []*v1alpha1.ClusterServiceVersion{
		&v1alpha1.ClusterServiceVersion{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ClusterServiceVersionKind,
				APIVersion: v1alpha1.ClusterServiceVersionAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "a.1.0",
				Namespace: "a",
				Annotations: map[string]string{
					"olm.targetNamespaces":  "",
					"olm.operatorNamespace": "a",
				},
			},
		},
		&v1alpha1.ClusterServiceVersion{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ClusterServiceVersionKind,
				APIVersion: v1alpha1.ClusterServiceVersionAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "a.1.0",
				Namespace: "b",
				Annotations: map[string]string{
					"olm.operatorNamespace": "a",
				},
			},
		},
		&v1alpha1.ClusterServiceVersion{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ClusterServiceVersionKind,
				APIVersion: v1alpha1.ClusterServiceVersionAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "a.1.0",
				Namespace: "c",
				Annotations: map[string]string{
					"olm.operatorNamespace": "a",
				},
			},
		},
		&v1alpha1.ClusterServiceVersion{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ClusterServiceVersionKind,
				APIVersion: v1alpha1.ClusterServiceVersionAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "b.1.0",
				Namespace: "b",
				Annotations: map[string]string{
					"olm.targetNamespaces":  "c,d",
					"olm.operatorNamespace": "b",
				},
			},
		},
		&v1alpha1.ClusterServiceVersion{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ClusterServiceVersionKind,
				APIVersion: v1alpha1.ClusterServiceVersionAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "b.1.0",
				Namespace: "d",
				Annotations: map[string]string{
					"olm.operatorNamespace": "b",
				},
			},
		},
	}

	testCases := []struct {
		csvName        string
		allCsvs        []*v1alpha1.ClusterServiceVersion
		expectedOutput bool
	}{
		{
			csvName:        "a.1.0",
			allCsvs:        allCsvs,
			expectedOutput: true,
		},
		{
			csvName:        "b.1.0",
			allCsvs:        allCsvs,
			expectedOutput: false,
		},
		{
			csvName:        "c.1.0",
			allCsvs:        allCsvs,
			expectedOutput: true,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, isCsvInNamespaceClusterWide(tc.csvName, allCsvs))
	}
}
