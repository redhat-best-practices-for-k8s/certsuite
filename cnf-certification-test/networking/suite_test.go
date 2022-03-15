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
)

func TestParseVariables(t *testing.T) {
	// expected inputs
	testCases := []struct {
		// inputRes is string that include the result after we run the command ""oc get pod %s -n %s -o json  | jq -r '.spec.containers[%d].ports'""
		inputRes string
		// now is empty but maybe in the future has be not empty.
		listeningPorts map[key]string
		// expected outputs here
		expectedlisteningPorts map[key]string
		expectedRes            string
	}{
		{
			inputRes:               "tcp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\n",
			listeningPorts:         map[key]string{},
			expectedlisteningPorts: map[key]string{{port: 8080, protocol: "TCP"}: ""},
			expectedRes:            "tcp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\n",
		},
		{
			inputRes:               "",
			listeningPorts:         map[key]string{},
			expectedlisteningPorts: map[key]string{},
			expectedRes:            "",
		},
		{
			inputRes:               "\n",
			listeningPorts:         map[key]string{},
			expectedlisteningPorts: map[key]string{},
			expectedRes:            "\n",
		},
		{
			inputRes:               "tcp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\ntcp LISTEN 0      128    0.0.0.0:7878 0.0.0.0:*\n",
			listeningPorts:         map[key]string{},
			expectedlisteningPorts: map[key]string{{port: 8080, protocol: "TCP"}: "", {port: 7878, protocol: "TCP"}: ""},
			expectedRes:            "tcp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\ntcp LISTEN 0      128    0.0.0.0:7878 0.0.0.0:*\n",
		},
		{
			inputRes:               "udp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\nudp LISTEN 0      128    0.0.0.0:7878 0.0.0.0:*\n",
			listeningPorts:         map[key]string{},
			expectedlisteningPorts: map[key]string{{port: 8080, protocol: "UDP"}: "", {port: 7878, protocol: "UDP"}: ""},
			expectedRes:            "udp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\nudp LISTEN 0      128    0.0.0.0:7878 0.0.0.0:*\n",
		},
	}
	for _, tc := range testCases {
		parseListening(tc.inputRes, tc.listeningPorts)
		assert.Equal(t, tc.expectedlisteningPorts, tc.listeningPorts)
	}
}

func TestCheckIfListenIsDeclared(t *testing.T) {
	// expected inputs
	testCases := []struct {
		// inputs
		listeningPorts map[key]string
		declaredPorts  map[key]string

		// expected outputs here
		expectedres map[key]string
	}{
		{
			listeningPorts: map[key]string{},
			declaredPorts:  map[key]string{},
			expectedres:    map[key]string{},
		},
		{
			listeningPorts: map[key]string{{port: 8080, protocol: "TCP"}: ""},
			declaredPorts:  map[key]string{{port: 8080, protocol: "TCP"}: "http-probe"},
			expectedres:    map[key]string{},
		},

		{
			listeningPorts: map[key]string{{port: 8080, protocol: "TCP"}: ""},
			declaredPorts:  map[key]string{},
			expectedres:    map[key]string{{port: 8080, protocol: "TCP"}: ""},
		},
		{
			listeningPorts: map[key]string{{port: 8080, protocol: "TCP"}: "", {port: 8443, protocol: "TCP"}: ""},
			declaredPorts:  map[key]string{{port: 8080, protocol: "TCP"}: "http-probe"},
			expectedres:    map[key]string{{port: 8443, protocol: "TCP"}: ""},
		},
		{
			listeningPorts: map[key]string{},
			declaredPorts:  map[key]string{{port: 8080, protocol: "TCP"}: "http-probe"},
			expectedres:    map[key]string{},
		},
		{
			listeningPorts: map[key]string{{port: 8080, protocol: "TCP"}: "", {port: 8443, protocol: "TCP"}: ""},
			declaredPorts:  map[key]string{{port: 8080, protocol: "TCP"}: "http-probe", {port: 8443, protocol: "TCP"}: "https"},
			expectedres:    map[key]string{},
		},
	}
	for _, tc := range testCases {
		res := checkIfListenIsDeclared(tc.listeningPorts, tc.declaredPorts)
		assert.Equal(t, res, tc.expectedres)
	}
}
