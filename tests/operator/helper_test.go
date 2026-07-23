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

package operator

import (
	"testing"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
