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

package declaredandlistening

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/declaredandlistening"
)

func TestParseVariables(t *testing.T) {
	// expected inputs
	testCases := []struct {
		// inputRes is string that include the result after we run the command ""oc get pod %s -n %s -o json  | jq -r '.spec.containers[%d].ports'""
		inputRes string
		// expected outputs here
		expectedlisteningPorts map[declaredandlistening.Key]bool
		expectedRes            string
	}{
		{
			inputRes:               "tcp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\n",
			expectedlisteningPorts: map[declaredandlistening.Key]bool{{Port: 8080, Protocol: "TCP"}: true},
			expectedRes:            "tcp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\n",
		},
		{
			inputRes:               "",
			expectedlisteningPorts: map[declaredandlistening.Key]bool{},
			expectedRes:            "",
		},
		{
			inputRes:               "\n",
			expectedlisteningPorts: map[declaredandlistening.Key]bool{},
			expectedRes:            "\n",
		},
		{
			inputRes:               "tcp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\ntcp LISTEN 0      128    0.0.0.0:7878 0.0.0.0:*\n",
			expectedlisteningPorts: map[declaredandlistening.Key]bool{{Port: 8080, Protocol: "TCP"}: true, {Port: 7878, Protocol: "TCP"}: true},
			expectedRes:            "tcp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\ntcp LISTEN 0      128    0.0.0.0:7878 0.0.0.0:*\n",
		},
		{
			inputRes:               "udp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\nudp LISTEN 0      128    0.0.0.0:7878 0.0.0.0:*\n",
			expectedlisteningPorts: map[declaredandlistening.Key]bool{{Port: 8080, Protocol: "UDP"}: true, {Port: 7878, Protocol: "UDP"}: true},
			expectedRes:            "udp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\nudp LISTEN 0      128    0.0.0.0:7878 0.0.0.0:*\n",
		},
		{
			inputRes:               "tcp LISTEN 0      128    [::]:22\n",
			expectedlisteningPorts: map[declaredandlistening.Key]bool{{Port: 22, Protocol: "TCP"}: true},
			expectedRes:            "tcp LISTEN 0      128    [::]:22\n",
		},
		{
			inputRes:               ":tcp LISTEN 0      128   [::]:22\n",
			expectedlisteningPorts: map[declaredandlistening.Key]bool{{Port: 22, Protocol: "TCP"}: true},
			expectedRes:            ":tcp LISTEN 0      128   [::]:22\n",
		},
	}
	for _, tc := range testCases {
		listeningPorts := map[declaredandlistening.Key]bool{}
		declaredandlistening.ParseListening(tc.inputRes, listeningPorts)
		assert.Equal(t, tc.expectedlisteningPorts, listeningPorts)
	}
}

func TestCheckIfListenIsDeclared(t *testing.T) {
	// expected inputs
	testCases := []struct {
		// inputs
		listeningPorts map[declaredandlistening.Key]bool
		declaredPorts  map[declaredandlistening.Key]bool

		// expected outputs here
		expectedres map[declaredandlistening.Key]bool
	}{
		{
			listeningPorts: map[declaredandlistening.Key]bool{},
			declaredPorts:  map[declaredandlistening.Key]bool{},
			expectedres:    map[declaredandlistening.Key]bool{},
		},
		{
			listeningPorts: map[declaredandlistening.Key]bool{{Port: 8080, Protocol: "TCP"}: true},
			declaredPorts:  map[declaredandlistening.Key]bool{{Port: 8080, Protocol: "TCP"}: true},
			expectedres:    map[declaredandlistening.Key]bool{},
		},

		{
			listeningPorts: map[declaredandlistening.Key]bool{{Port: 8080, Protocol: "TCP"}: true},
			declaredPorts:  map[declaredandlistening.Key]bool{},
			expectedres:    map[declaredandlistening.Key]bool{{Port: 8080, Protocol: "TCP"}: true},
		},
		{
			listeningPorts: map[declaredandlistening.Key]bool{{Port: 8080, Protocol: "TCP"}: true, {Port: 8443, Protocol: "TCP"}: true},
			declaredPorts:  map[declaredandlistening.Key]bool{{Port: 8080, Protocol: "TCP"}: true},
			expectedres:    map[declaredandlistening.Key]bool{{Port: 8443, Protocol: "TCP"}: true},
		},
		{
			listeningPorts: map[declaredandlistening.Key]bool{},
			declaredPorts:  map[declaredandlistening.Key]bool{{Port: 8080, Protocol: "TCP"}: true},
			expectedres:    map[declaredandlistening.Key]bool{},
		},
		{
			listeningPorts: map[declaredandlistening.Key]bool{{Port: 8080, Protocol: "TCP"}: true, {Port: 8443, Protocol: "TCP"}: true},
			declaredPorts:  map[declaredandlistening.Key]bool{{Port: 8080, Protocol: "TCP"}: true, {Port: 8443, Protocol: "TCP"}: true},
			expectedres:    map[declaredandlistening.Key]bool{},
		},
	}
	for _, tc := range testCases {
		res := declaredandlistening.CheckIfListenIsDeclared(tc.listeningPorts, tc.declaredPorts)
		assert.Equal(t, res, tc.expectedres)
	}
}
