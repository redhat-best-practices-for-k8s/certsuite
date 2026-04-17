// Copyright (C) 2024-2026 Red Hat, Inc.
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

package tlsversion

import (
	"fmt"
	"strings"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/stretchr/testify/assert"
)

type mockCommand struct {
	patterns []mockPattern
}

type mockPattern struct {
	key    string
	result mockExecResult
}

type mockExecResult struct {
	stdout string
	err    error
}

func newMockCommand(patterns ...mockPattern) *mockCommand {
	return &mockCommand{patterns: patterns}
}

func (m *mockCommand) ExecCommandContainer(_ clientsholder.Context, command string) (stdout, stderr string, err error) {
	for _, p := range m.patterns {
		if strings.Contains(command, p.key) {
			return p.result.stdout, "", p.result.err
		}
	}
	return "", "", fmt.Errorf("unexpected command: %s", command)
}

func TestIsPortTLS_TLSServer(t *testing.T) {
	mock := newMockCommand(
		mockPattern{"s_client", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nProtocol  : TLSv1.3\nCipher    : TLS_AES_256_GCM_SHA384\n---",
		}},
	)
	ctx := clientsholder.NewContext("ns", "pod", "container")
	isTLS, reachable, reason := IsPortTLS(mock, ctx, "10.0.0.1", 443)
	assert.True(t, isTLS, "expected TLS, reason: %s", reason)
	assert.True(t, reachable)
	assert.Contains(t, reason, "TLS negotiated")
}

func TestIsPortTLS_PlaintextService(t *testing.T) {
	mock := newMockCommand(
		mockPattern{"s_client", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nCipher is (NONE)\npacket length too long\n---",
			err:    fmt.Errorf("exit status 1"),
		}},
	)
	ctx := clientsholder.NewContext("ns", "pod", "container")
	isTLS, reachable, reason := IsPortTLS(mock, ctx, "10.0.0.1", 8080)
	assert.False(t, isTLS, "expected plaintext, reason: %s", reason)
	assert.True(t, reachable)
	assert.Contains(t, reason, "plaintext")
}

func TestIsPortTLS_ConnectionRefused(t *testing.T) {
	mock := newMockCommand(
		mockPattern{"s_client", mockExecResult{
			stdout: "connect:errno=111\nConnection refused",
			err:    fmt.Errorf("exit status 1"),
		}},
	)
	ctx := clientsholder.NewContext("ns", "pod", "container")
	isTLS, reachable, reason := IsPortTLS(mock, ctx, "10.0.0.1", 9999)
	assert.False(t, isTLS)
	assert.False(t, reachable)
	assert.Contains(t, reason, "unreachable")
}

func TestIsPortTLS_TLSHandshakeAlert(t *testing.T) {
	mock := newMockCommand(
		mockPattern{"s_client", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nalert handshake failure\nCipher is (NONE)\n---",
			err:    fmt.Errorf("exit status 1"),
		}},
	)
	ctx := clientsholder.NewContext("ns", "pod", "container")
	isTLS, reachable, reason := IsPortTLS(mock, ctx, "10.0.0.1", 443)
	assert.True(t, isTLS, "TLS alert means genuine TLS server, reason: %s", reason)
	assert.True(t, reachable)
	assert.Contains(t, reason, "TLS server detected")
}

func TestIsPortTLS_ExecFailed(t *testing.T) {
	mock := newMockCommand(
		mockPattern{"s_client", mockExecResult{
			stdout: "",
			err:    fmt.Errorf("command not found"),
		}},
	)
	ctx := clientsholder.NewContext("ns", "pod", "container")
	isTLS, reachable, _ := IsPortTLS(mock, ctx, "10.0.0.1", 443)
	assert.False(t, isTLS)
	assert.False(t, reachable)
}
