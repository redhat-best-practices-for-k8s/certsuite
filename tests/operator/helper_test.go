// Copyright (C) 2020-2026 Red Hat, Inc.
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
	corev1 "k8s.io/api/core/v1"
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

func TestGetAllPodsBy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		namespace string
		pods      []*provider.Pod
		wantCount int
	}{
		{
			name:      "matching namespace",
			namespace: "ns-a",
			pods: []*provider.Pod{
				{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns-a"}}},
				{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: "ns-a"}}},
			},
			wantCount: 2,
		},
		{
			name:      "non-matching namespace",
			namespace: "ns-b",
			pods: []*provider.Pod{
				{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns-a"}}},
			},
			wantCount: 0,
		},
		{
			name:      "empty list",
			namespace: "ns-a",
			pods:      []*provider.Pod{},
			wantCount: 0,
		},
		{
			name:      "mixed namespaces",
			namespace: "ns-a",
			pods: []*provider.Pod{
				{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns-a"}}},
				{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: "ns-b"}}},
				{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod3", Namespace: "ns-a"}}},
			},
			wantCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getAllPodsBy(tt.namespace, tt.pods)
			assert.Len(t, got, tt.wantCount)
			for _, p := range got {
				assert.Equal(t, tt.namespace, p.Namespace)
			}
		})
	}
}

func TestGetCsvsBy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		namespace string
		csvs      []*v1alpha1.ClusterServiceVersion
		wantCount int
	}{
		{
			name:      "matching namespace",
			namespace: "ns-a",
			csvs: []*v1alpha1.ClusterServiceVersion{
				{ObjectMeta: metav1.ObjectMeta{Name: "csv1", Namespace: "ns-a"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "csv2", Namespace: "ns-a"}},
			},
			wantCount: 2,
		},
		{
			name:      "non-matching namespace",
			namespace: "ns-b",
			csvs: []*v1alpha1.ClusterServiceVersion{
				{ObjectMeta: metav1.ObjectMeta{Name: "csv1", Namespace: "ns-a"}},
			},
			wantCount: 0,
		},
		{
			name:      "empty list",
			namespace: "ns-a",
			csvs:      []*v1alpha1.ClusterServiceVersion{},
			wantCount: 0,
		},
		{
			name:      "mixed namespaces",
			namespace: "ns-a",
			csvs: []*v1alpha1.ClusterServiceVersion{
				{ObjectMeta: metav1.ObjectMeta{Name: "csv1", Namespace: "ns-a"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "csv2", Namespace: "ns-b"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "csv3", Namespace: "ns-a"}},
			},
			wantCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCsvsBy(tt.namespace, tt.csvs)
			assert.Len(t, got, tt.wantCount)
			for _, csv := range got {
				assert.Equal(t, tt.namespace, csv.Namespace)
			}
		})
	}
}

func TestIsSingleNamespacedOperator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		operatorNamespace string
		targetNamespaces  []string
		want              bool
	}{
		{
			name:              "single target different from operator namespace",
			operatorNamespace: "operator-ns",
			targetNamespaces:  []string{"target-ns"},
			want:              true,
		},
		{
			name:              "single target same as operator namespace",
			operatorNamespace: "operator-ns",
			targetNamespaces:  []string{"operator-ns"},
			want:              false,
		},
		{
			name:              "multiple targets",
			operatorNamespace: "operator-ns",
			targetNamespaces:  []string{"ns-a", "ns-b"},
			want:              false,
		},
		{
			name:              "empty targets",
			operatorNamespace: "operator-ns",
			targetNamespaces:  []string{},
			want:              false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSingleNamespacedOperator(tt.operatorNamespace, tt.targetNamespaces)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsMultiNamespacedOperator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		operatorNamespace string
		targetNamespaces  []string
		want              bool
	}{
		{
			name:              "multiple targets not including operator namespace",
			operatorNamespace: "operator-ns",
			targetNamespaces:  []string{"ns-a", "ns-b"},
			want:              true,
		},
		{
			name:              "multiple targets including operator namespace",
			operatorNamespace: "operator-ns",
			targetNamespaces:  []string{"ns-a", "operator-ns"},
			want:              false,
		},
		{
			name:              "single target",
			operatorNamespace: "operator-ns",
			targetNamespaces:  []string{"ns-a"},
			want:              false,
		},
		{
			name:              "empty targets",
			operatorNamespace: "operator-ns",
			targetNamespaces:  []string{},
			want:              false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isMultiNamespacedOperator(tt.operatorNamespace, tt.targetNamespaces)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsCsvInNamespaceClusterWide(t *testing.T) {
	allCsvs := []*v1alpha1.ClusterServiceVersion{
		{
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
		{
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
		{
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
		{
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
		{
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
