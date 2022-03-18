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

package networking

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/declaredandlistening"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	v1 "k8s.io/api/core/v1"
)

//nolint:dupl,funlen
func TestParseVariables(t *testing.T) {
	// expected inputs
	testCases := []struct {
		// inputRes is string that include the result after we run the command ""oc get pod %s -n %s -o json  | jq -r '.spec.containers[%d].ports'""
		inputRes string
		// now is empty but maybe in the future has be not empty.
		listeningPorts map[declaredandlistening.Key]*provider.Container
		container      *provider.Container
		// expected outputs here
		expectedlisteningPorts map[declaredandlistening.Key]*provider.Container
		expectedRes            string
		expectContainer        *provider.Container
	}{
		{
			inputRes:       "tcp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\n",
			listeningPorts: map[declaredandlistening.Key]*provider.Container{},
			container: &provider.Container{
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
			expectedlisteningPorts: map[declaredandlistening.Key]*provider.Container{{Port: 8080, Protocol: "TCP"}: {
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
			},
			expectedRes: "tcp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\n",
			expectContainer: &provider.Container{
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
		},

		{
			inputRes:       "",
			listeningPorts: map[declaredandlistening.Key]*provider.Container{},
			container: &provider.Container{
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
			expectedlisteningPorts: map[declaredandlistening.Key]*provider.Container{},
			expectedRes:            "",
			expectContainer: &provider.Container{
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
		},
		{
			inputRes:       "\n",
			listeningPorts: map[declaredandlistening.Key]*provider.Container{},
			container: &provider.Container{
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
			expectedlisteningPorts: map[declaredandlistening.Key]*provider.Container{},
			expectedRes:            "\n",
			expectContainer: &provider.Container{
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
		},

		{
			inputRes:       "tcp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\ntcp LISTEN 0      128    0.0.0.0:7878 0.0.0.0:*\n",
			listeningPorts: map[declaredandlistening.Key]*provider.Container{},
			container: &provider.Container{
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
			expectedlisteningPorts: map[declaredandlistening.Key]*provider.Container{{Port: 8080, Protocol: "TCP"}: {
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
				{Port: 7878, Protocol: "TCP"}: {
					Data: &v1.Container{
						Name: "",
					},
					Namespace: "",
					Podname:   "",
				},
			},
			expectedRes: "tcp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\ntcp LISTEN 0      128    0.0.0.0:7878 0.0.0.0:*\n",
			expectContainer: &provider.Container{
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
		},

		{
			inputRes:       "udp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\nudp LISTEN 0      128    0.0.0.0:7878 0.0.0.0:*\n",
			listeningPorts: map[declaredandlistening.Key]*provider.Container{},
			container: &provider.Container{
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
			expectedlisteningPorts: map[declaredandlistening.Key]*provider.Container{{Port: 8080, Protocol: "UDP"}: {
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
				{Port: 7878, Protocol: "UDP"}: {
					Data: &v1.Container{
						Name: "",
					},
					Namespace: "",
					Podname:   "",
				},
			},
			expectedRes: "udp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\nudp LISTEN 0      128    0.0.0.0:7878 0.0.0.0:*\n",
			expectContainer: &provider.Container{
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
		},
	}
	for _, tc := range testCases {
		declaredandlistening.ParseListening(tc.inputRes, tc.listeningPorts, tc.container)
		assert.Equal(t, tc.expectedlisteningPorts, tc.listeningPorts)
	}
}

//nolint:funlen
func TestCheckIfListenIsDeclared(t *testing.T) {
	// expected inputs
	testCases := []struct {
		// input
		listeningPorts map[declaredandlistening.Key]*provider.Container
		declaredPorts  map[declaredandlistening.Key]*provider.Container

		// expected outputs here
		expectedres map[declaredandlistening.Key]*provider.Container
	}{

		{
			listeningPorts: map[declaredandlistening.Key]*provider.Container{},
			declaredPorts:  map[declaredandlistening.Key]*provider.Container{},
			expectedres:    map[declaredandlistening.Key]*provider.Container{},
		},
		{
			listeningPorts: map[declaredandlistening.Key]*provider.Container{{Port: 8080, Protocol: "TCP"}: {
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
			},
			declaredPorts: map[declaredandlistening.Key]*provider.Container{{Port: 8080, Protocol: "TCP"}: {
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
			},
			expectedres: map[declaredandlistening.Key]*provider.Container{},
		},

		{
			listeningPorts: map[declaredandlistening.Key]*provider.Container{{Port: 8080, Protocol: "TCP"}: {
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
			},
			declaredPorts: map[declaredandlistening.Key]*provider.Container{},
			expectedres: map[declaredandlistening.Key]*provider.Container{{Port: 8080, Protocol: "TCP"}: {
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
			},
		},

		{
			listeningPorts: map[declaredandlistening.Key]*provider.Container{{Port: 8080, Protocol: "TCP"}: {
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
				{Port: 8443, Protocol: "TCP"}: {
					Data: &v1.Container{
						Name: "",
					},
					Namespace: "",
					Podname:   "",
				},
			},
			declaredPorts: map[declaredandlistening.Key]*provider.Container{{Port: 8080, Protocol: "TCP"}: {
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
			},
			expectedres: map[declaredandlistening.Key]*provider.Container{{Port: 8443, Protocol: "TCP"}: {
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
			},
		},
		{
			listeningPorts: map[declaredandlistening.Key]*provider.Container{},
			declaredPorts: map[declaredandlistening.Key]*provider.Container{{Port: 8080, Protocol: "TCP"}: {
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
			},
			expectedres: map[declaredandlistening.Key]*provider.Container{},
		},

		{
			listeningPorts: map[declaredandlistening.Key]*provider.Container{{Port: 8080, Protocol: "TCP"}: {
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
				{Port: 8443, Protocol: "TCP"}: {
					Data: &v1.Container{
						Name: "",
					},
					Namespace: "",
					Podname:   "",
				},
			},
			declaredPorts: map[declaredandlistening.Key]*provider.Container{{Port: 8080, Protocol: "TCP"}: {
				Data: &v1.Container{
					Name: "",
				},
				Namespace: "",
				Podname:   "",
			},
				{Port: 8443, Protocol: "TCP"}: {
					Data: &v1.Container{
						Name: "",
					},
					Namespace: "",
					Podname:   "",
				},
			},
			expectedres: map[declaredandlistening.Key]*provider.Container{},
		},
	}
	for _, tc := range testCases {
		res := declaredandlistening.CheckIfListenIsDeclared(tc.listeningPorts, tc.declaredPorts)
		assert.Equal(t, res, tc.expectedres)
	}
}
