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

package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	v1core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAutomountServiceAccountSetOnSA(t *testing.T) {
	testCases := []struct {
		automountServiceTokenSet bool
	}{
		{
			automountServiceTokenSet: true,
		},
		{
			automountServiceTokenSet: false,
		},
	}

	for _, tc := range testCases {
		testSA := v1core.ServiceAccount{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "podNS",
				Name:      "testSA",
			},
			AutomountServiceAccountToken: &tc.automountServiceTokenSet,
		}

		var testRuntimeObjects []runtime.Object
		testRuntimeObjects = append(testRuntimeObjects, &testSA)

		obj := NewAutomountTokenTester(clientsholder.GetTestClientsHolder(testRuntimeObjects))
		assert.NotNil(t, obj)
		isSet, err := obj.AutomountServiceAccountSetOnSA("testSA", "podNS")
		assert.Nil(t, err)
		assert.Equal(t, tc.automountServiceTokenSet, *isSet)
	}
}
