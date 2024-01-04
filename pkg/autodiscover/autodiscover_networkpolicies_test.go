// Copyright (C) 2022 Red Hat, Inc.
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
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestGetNetworkPolicies(t *testing.T) {
	testCases := []struct {
		expectedNetworkPolicies []*networkingv1.NetworkPolicy
	}{
		{
			expectedNetworkPolicies: []*networkingv1.NetworkPolicy{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-network-policy",
						Namespace: "test-namespace",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object
		for _, networkPolicy := range testCase.expectedNetworkPolicies {
			runtimeObjects = append(runtimeObjects, networkPolicy)
		}

		// Create fake client
		client := k8sfake.NewSimpleClientset(runtimeObjects...)
		networkPolicies, err := getNetworkPolicies(client.NetworkingV1())
		assert.Nil(t, err)
		assert.Len(t, networkPolicies, len(testCase.expectedNetworkPolicies))
	}
}
