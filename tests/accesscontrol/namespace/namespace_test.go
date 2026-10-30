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

package namespace

import (
	"io"
	"os"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestGetInvalidCRsNum(t *testing.T) {
	testCases := []struct {
		invalidCrs         map[string]map[string][]string
		expectedInvalidCRs int
	}{
		{
			invalidCrs: map[string]map[string][]string{
				"cr1": {
					"ns1": {
						"testCRDs",
					},
				},
			},
			expectedInvalidCRs: 1,
		},
		{
			invalidCrs:         map[string]map[string][]string{},
			expectedInvalidCRs: 0,
		},
	}

	for _, tc := range testCases {
		log.SetupLogger(os.Stdout, "INFO")
		result := GetInvalidCRsNum(tc.invalidCrs, log.GetLogger())
		assert.Equal(t, tc.expectedInvalidCRs, result)
	}
}

// TestGetCrsPerNamespacesSkipsUnservedVersions verifies that a CRD which retires
// an API version is still enumerated. Listing a version with Served=false
// returns NotFound from a real API server, which used to abort the whole check.
func TestGetCrsPerNamespacesSkipsUnservedVersions(t *testing.T) {
	log.SetupLogger(io.Discard, "INFO")

	const (
		group   = "example.com"
		plural  = "widgets"
		version = "v1"
	)
	servedGVR := schema.GroupVersionResource{Group: group, Version: version, Resource: plural}
	unservedGVR := schema.GroupVersionResource{Group: group, Version: "v1beta1", Resource: plural}

	crd := &apiextv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: plural + "." + group},
		Spec: apiextv1.CustomResourceDefinitionSpec{
			Group: group,
			Names: apiextv1.CustomResourceDefinitionNames{Plural: plural, Kind: "Widget", ListKind: "WidgetList"},
			Versions: []apiextv1.CustomResourceDefinitionVersion{
				{Name: version, Served: true, Storage: true},
				{Name: unservedGVR.Version, Served: false},
			},
		},
	}

	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			servedGVR:   "WidgetList",
			unservedGVR: "WidgetList",
		})
	client.PrependReactor("list", plural, func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetResource().Version != version {
			return true, nil, apierrors.NewNotFound(action.GetResource().GroupResource(), plural)
		}
		return true, &unstructured.UnstructuredList{
			Object: map[string]interface{}{"apiVersion": group + "/" + version, "kind": "WidgetList"},
			Items: []unstructured.Unstructured{{
				Object: map[string]interface{}{
					"apiVersion": group + "/" + version,
					"kind":       "Widget",
					"metadata": map[string]interface{}{
						"name":      "widget-1",
						"namespace": "ns1",
					},
				},
			}},
		}, nil
	})
	clientsholder.SetTestK8sDynamicClientsHolder(client)
	t.Cleanup(func() {
		clientsholder.SetTestK8sDynamicClientsHolder(nil)
		clientsholder.ClearTestClientsHolder()
	})

	got, err := getCrsPerNamespaces(crd)
	require.NoError(t, err)
	assert.Equal(t, map[string][]string{"ns1": {"widget-1"}}, got)
}
