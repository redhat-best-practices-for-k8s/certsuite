// Copyright (C) 2020-2021 Red Hat, Inc.
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

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestFindTestCrdNames(t *testing.T) {
	// Function to generate some runtime objects for the k8s mock client
	generateObjects := func() []runtime.Object {
		testCRD := apiextv1.CustomResourceDefinition{
			ObjectMeta: v1.ObjectMeta{
				Name: "testCRD_testsuffix",
			},
		}

		var testRuntimeObjects []runtime.Object
		testRuntimeObjects = append(testRuntimeObjects, &testCRD)
		return testRuntimeObjects
	}

	testCases := []struct {
		spoofCRDFilters    []configuration.CrdFilter
		expectedTargetCRDs []*apiextv1.CustomResourceDefinition
	}{
		{ // Test Case #1 - testsuffix happy path
			spoofCRDFilters: []configuration.CrdFilter{
				{
					NameSuffix: "testsuffix",
				},
			},
			expectedTargetCRDs: []*apiextv1.CustomResourceDefinition{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: "testCRD_testsuffix",
					},
				},
			},
		},
		{ // Test Case #2 - empty filters so cannot find CRD
			spoofCRDFilters:    []configuration.CrdFilter{},
			expectedTargetCRDs: nil,
		},
	}

	for _, tc := range testCases {
		_ = clientsholder.GetTestClientsHolder(generateObjects())
		// Run the function and assert the results
		assert.Equal(t, tc.expectedTargetCRDs, FindTestCrdNames(tc.spoofCRDFilters))
	}
}
