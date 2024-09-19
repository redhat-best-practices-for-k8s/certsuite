// Copyright (C) 2022-2023 Red Hat, Inc.
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

package autodiscover

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/stretchr/testify/assert"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetPodDisruptionBudgets(t *testing.T) {
	generatePodDisruptionBudget := func(name, namespace string) *policyv1.PodDisruptionBudget {
		return &policyv1.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: policyv1.PodDisruptionBudgetSpec{},
		}
	}

	testCases := []struct {
		pdbName      string
		pdbNamespace string
		expectedPDBs []policyv1.PodDisruptionBudget
	}{
		{
			pdbName:      "testPdb",
			pdbNamespace: "tnf",
			expectedPDBs: []policyv1.PodDisruptionBudget{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testPdb",
						Namespace: "tnf",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		var testRuntimeObjects []runtime.Object
		testRuntimeObjects = append(testRuntimeObjects, generatePodDisruptionBudget(tc.pdbName, tc.pdbNamespace))
		oc := clientsholder.GetTestClientsHolder(testRuntimeObjects)
		pdbs, err := getPodDisruptionBudgets(oc.K8sClient.PolicyV1(), []string{tc.pdbNamespace})
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedPDBs, pdbs)
	}
}
