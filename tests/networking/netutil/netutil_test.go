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

package netutil

import (
	"context"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseListeningPorts(t *testing.T) {
	testCases := []struct {
		inputStr               string
		expectedListeningPorts map[PortInfo]bool
	}{
		{
			inputStr:               "tcp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\n",
			expectedListeningPorts: map[PortInfo]bool{{PortNumber: 8080, Protocol: "TCP"}: true},
		},
		{
			inputStr:               "",
			expectedListeningPorts: map[PortInfo]bool{},
		},
		{
			inputStr:               "\n",
			expectedListeningPorts: map[PortInfo]bool{},
		},
		{
			inputStr:               "tcp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\ntcp LISTEN 0      128    0.0.0.0:7878 0.0.0.0:*\n",
			expectedListeningPorts: map[PortInfo]bool{{PortNumber: 8080, Protocol: "TCP"}: true, {PortNumber: 7878, Protocol: "TCP"}: true},
		},
		{
			inputStr:               "udp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\nudp LISTEN 0      128    0.0.0.0:7878 0.0.0.0:*\n",
			expectedListeningPorts: map[PortInfo]bool{{PortNumber: 8080, Protocol: "UDP"}: true, {PortNumber: 7878, Protocol: "UDP"}: true},
		},
		{
			inputStr:               "tcp LISTEN 0      128    [::]:22\n",
			expectedListeningPorts: map[PortInfo]bool{{PortNumber: 22, Protocol: "TCP"}: true},
		},
	}
	for _, tc := range testCases {
		listeningPorts, err := parseListeningPorts(tc.inputStr)
		if assert.NoError(t, err) {
			assert.Equal(t, tc.expectedListeningPorts, listeningPorts)
		}
	}
}

func TestWrapNSEnterError(t *testing.T) {
	t.Parallel()

	cut := &provider.Container{Namespace: "smf1", Podname: "nf-alert"}
	deadline := wrapNSEnterError("ss -tpln", cut, "", context.DeadlineExceeded)
	require.Error(t, deadline)
	assert.ErrorIs(t, deadline, context.DeadlineExceeded)
	assert.Contains(t, deadline.Error(), "failed to execute command ss -tpln")

	stderrOnly := wrapNSEnterError("ss -tulwnH", cut, "command not found", nil)
	require.Error(t, stderrOnly)
	assert.Contains(t, stderrOnly.Error(), "command not found")
}
