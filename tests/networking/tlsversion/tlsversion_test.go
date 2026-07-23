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

type mockPattern struct {
	key    string
	stdout string
	err    error
}

func newMockCommand(patterns ...mockPattern) *clientsholder.MockCommand {
	return &clientsholder.MockCommand{
		ExecFunc: func(_ clientsholder.Context, command string) (string, string, error) {
			for _, p := range patterns {
				if strings.Contains(command, p.key) {
					return p.stdout, "", p.err
				}
			}
			return "", "", fmt.Errorf("unexpected command: %s", command)
		},
	}
}

func TestIsPortTLS_TLSServer(t *testing.T) {
	mock := newMockCommand(
		mockPattern{key: "s_client",
			stdout: "CONNECTED(00000003)\n---\nProtocol  : TLSv1.3\nCipher    : TLS_AES_256_GCM_SHA384\n---",
		},
	)
	ctx := clientsholder.NewContext("ns", "pod", "container")
	isTLS, reachable, reason := IsPortTLS(mock, ctx, "10.0.0.1", 443)
	assert.True(t, isTLS, "expected TLS, reason: %s", reason)
	assert.True(t, reachable)
	assert.Contains(t, reason, "TLS negotiated")
}

func TestIsPortTLS_PlaintextService(t *testing.T) {
	mock := newMockCommand(
		mockPattern{key: "s_client",
			stdout: "CONNECTED(00000003)\n---\nCipher is (NONE)\npacket length too long\n---",
			err:    fmt.Errorf("exit status 1"),
		},
	)
	ctx := clientsholder.NewContext("ns", "pod", "container")
	isTLS, reachable, reason := IsPortTLS(mock, ctx, "10.0.0.1", 8080)
	assert.False(t, isTLS, "expected plaintext, reason: %s", reason)
	assert.True(t, reachable)
	assert.Contains(t, reason, "plaintext")
}

func TestIsPortTLS_ConnectionRefused(t *testing.T) {
	mock := newMockCommand(
		mockPattern{key: "s_client",
			stdout: "connect:errno=111\nConnection refused",
			err:    fmt.Errorf("exit status 1"),
		},
	)
	ctx := clientsholder.NewContext("ns", "pod", "container")
	isTLS, reachable, reason := IsPortTLS(mock, ctx, "10.0.0.1", 9999)
	assert.False(t, isTLS)
	assert.False(t, reachable)
	assert.Contains(t, reason, "unreachable")
}

func TestIsPortTLS_TLSHandshakeAlert(t *testing.T) {
	mock := newMockCommand(
		mockPattern{key: "s_client",
			stdout: "CONNECTED(00000003)\n---\nalert handshake failure\nCipher is (NONE)\n---",
			err:    fmt.Errorf("exit status 1"),
		},
	)
	ctx := clientsholder.NewContext("ns", "pod", "container")
	isTLS, reachable, reason := IsPortTLS(mock, ctx, "10.0.0.1", 443)
	assert.True(t, isTLS, "TLS alert means genuine TLS server, reason: %s", reason)
	assert.True(t, reachable)
	assert.Contains(t, reason, "TLS server detected")
}

func TestIsPortTLS_ExecFailed(t *testing.T) {
	mock := newMockCommand(
		mockPattern{key: "s_client",
			stdout: "",
			err:    fmt.Errorf("command not found"),
		},
	)
	ctx := clientsholder.NewContext("ns", "pod", "container")
	isTLS, reachable, _ := IsPortTLS(mock, ctx, "10.0.0.1", 443)
	assert.False(t, isTLS)
	assert.False(t, reachable)
}

func TestIsPortTLS_AlertProtocolVersion(t *testing.T) {
	mock := newMockCommand(
		mockPattern{key: "s_client",
			stdout: "CONNECTED(00000003)\n---\nalert protocol version\nCipher is (NONE)\n---",
			err:    fmt.Errorf("exit status 1"),
		},
	)
	ctx := clientsholder.NewContext("ns", "pod", "container")
	isTLS, reachable, reason := IsPortTLS(mock, ctx, "10.0.0.1", 443)
	assert.True(t, isTLS, "protocol version alert means TLS server, reason: %s", reason)
	assert.True(t, reachable)
	assert.Contains(t, reason, "TLS server detected")
}

func TestIsPortTLS_HandshakeFailureWithoutAlertPrefix(t *testing.T) {
	mock := newMockCommand(
		mockPattern{key: "s_client",
			stdout: "CONNECTED(00000003)\n---\nhandshake failure\n---",
			err:    fmt.Errorf("exit status 1"),
		},
	)
	ctx := clientsholder.NewContext("ns", "pod", "container")
	isTLS, reachable, reason := IsPortTLS(mock, ctx, "10.0.0.1", 443)
	assert.True(t, isTLS, "handshake failure means TLS server, reason: %s", reason)
	assert.True(t, reachable)
	assert.Contains(t, reason, "TLS server detected")
}

func TestIsPortTLS_CipherIs0000(t *testing.T) {
	mock := newMockCommand(
		mockPattern{key: "s_client",
			stdout: "CONNECTED(00000003)\n---\nCipher    : 0000\nProtocol  : TLSv1.2\n---",
		},
	)
	ctx := clientsholder.NewContext("ns", "pod", "container")
	isTLS, reachable, reason := IsPortTLS(mock, ctx, "10.0.0.1", 443)
	assert.False(t, isTLS, "cipher 0000 means rejected, reason: %s", reason)
	assert.True(t, reachable)
	assert.Contains(t, reason, "plaintext")
}

func TestIsPortTLS_NoRecognizableOutput(t *testing.T) {
	mock := newMockCommand(
		mockPattern{key: "s_client",
			stdout: "some random garbage output",
		},
	)
	ctx := clientsholder.NewContext("ns", "pod", "container")
	isTLS, reachable, reason := IsPortTLS(mock, ctx, "10.0.0.1", 443)
	assert.False(t, isTLS)
	assert.False(t, reachable)
	assert.Contains(t, reason, "no recognizable openssl output")
}

func TestExtractOpenSSLCipher(t *testing.T) {
	tests := []struct {
		name     string
		stdout   string
		expected string
	}{
		{"spaced format", "Cipher    : ECDHE-RSA-AES128-GCM-SHA256\nProtocol  : TLSv1.2", "ECDHE-RSA-AES128-GCM-SHA256"},
		{"compact format", "Cipher:TLS_AES_256_GCM_SHA384\nProtocol: TLSv1.3", "TLS_AES_256_GCM_SHA384"},
		{"no cipher line", "Protocol  : TLSv1.2\nSome other output", cipherUnknown},
		{"empty output", "", cipherUnknown},
		{"cipher 0000", "Cipher    : 0000\nProtocol  : TLSv1.2", "0000"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, extractOpenSSLCipher(tc.stdout))
		})
	}
}
