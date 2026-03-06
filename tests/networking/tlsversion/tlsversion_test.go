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
	"crypto/tls"
	"fmt"
	"strings"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/stretchr/testify/assert"
)

// modernPolicy is the Modern profile (TLS 1.3 only).
func modernPolicy() TLSPolicy {
	return ResolveTLSProfile(&configv1.TLSSecurityProfile{Type: configv1.TLSProfileModernType})
}

// intermediatePolicy is the Intermediate profile (TLS 1.2 minimum).
func intermediatePolicy() TLSPolicy {
	return ResolveTLSProfile(nil) // nil = Intermediate (default)
}

// mockCommand implements clientsholder.Command for testing the exec probe.
// Patterns are checked in order; use the most specific pattern first.
type mockCommand struct {
	patterns []mockPattern
}

type mockPattern struct {
	key    string
	result mockExecResult
}

type mockExecResult struct {
	stdout string
	stderr string
	err    error
}

func newMockCommand(patterns ...mockPattern) *mockCommand {
	return &mockCommand{patterns: patterns}
}

func (m *mockCommand) ExecCommandContainer(_ clientsholder.Context, command string) (stdout, stderr string, err error) {
	for _, p := range m.patterns {
		if strings.Contains(command, p.key) {
			return p.result.stdout, p.result.stderr, p.result.err
		}
	}
	return "", "", fmt.Errorf("unexpected command: %s", command)
}

func TestProbeServicePortViaExec_TLS13Enforced_ModernProfile(t *testing.T) {
	mock := newMockCommand(
		mockPattern{"-tls1_3", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nProtocol  : TLSv1.3\nCipher    : TLS_AES_256_GCM_SHA384\n---",
		}},
		mockPattern{"-tls1_2", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nssl handshake failure\n---",
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := ProbeServicePortViaExec(mock, ctx, "10.0.0.1", 443, modernPolicy())
	assert.True(t, result.Compliant, "expected compliant, got: %s", result.Reason)
	assert.Equal(t, "TLS 1.3", result.NegotiatedVer)
}

func TestProbeServicePortViaExec_AcceptsTLS12_ModernProfile(t *testing.T) {
	mock := newMockCommand(
		mockPattern{"-tls1_3", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nProtocol  : TLSv1.3\nCipher    : TLS_AES_256_GCM_SHA384\n---",
		}},
		mockPattern{"-tls1_2", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nProtocol  : TLSv1.2\nCipher    : ECDHE-RSA-AES128-GCM-SHA256\n---",
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := ProbeServicePortViaExec(mock, ctx, "10.0.0.1", 443, modernPolicy())
	assert.False(t, result.Compliant, "expected non-compliant with Modern profile")
	assert.Equal(t, "TLS 1.2", result.NegotiatedVer)
}

func TestProbeServicePortViaExec_AcceptsTLS12_IntermediateProfile(t *testing.T) {
	// "-cipher" must be checked before "-tls1_2" since the cipher check command also contains "-tls1_2"
	mock := newMockCommand(
		mockPattern{"-cipher", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nssl handshake failure\n---",
		}},
		mockPattern{"-tls1_2", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nProtocol  : TLSv1.2\nCipher    : ECDHE-RSA-AES128-GCM-SHA256\n---",
		}},
		mockPattern{"-tls1_1", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nssl handshake failure\n---",
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := ProbeServicePortViaExec(mock, ctx, "10.0.0.1", 443, intermediatePolicy())
	assert.True(t, result.Compliant, "expected compliant with Intermediate profile, got: %s", result.Reason)
}

func TestProbeServicePortViaExec_TLS13OnlyRejectsTLS12_ModernProfile(t *testing.T) {
	// Reproduces real openssl output when a TLS 1.3-only server rejects TLS 1.2.
	// openssl exits non-zero on handshake failure, so ExecCommandContainer returns an error.
	mock := newMockCommand(
		mockPattern{"-tls1_3", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nProtocol  : TLSv1.3\nCipher    : TLS_AES_256_GCM_SHA384\n---",
		}},
		mockPattern{"-tls1_2", mockExecResult{
			stdout: `Connecting to 10.217.4.83
error:0A00042E:SSL routines:ssl3_read_bytes:tlsv1 alert protocol version:ssl/record/rec_layer_s3.c:916:SSL alert number 70
CONNECTED(00000003)
---
no peer certificate available
---
No client certificate CA names sent
---
SSL handshake has read 7 bytes and written 191 bytes
Verification: OK
---
New, (NONE), Cipher is (NONE)
Protocol: TLSv1.2
Secure Renegotiation IS NOT supported
Compression: NONE
Expansion: NONE
No ALPN negotiated
SSL-Session:
    Protocol  : TLSv1.2
    Cipher    : 0000
    Session-ID:
    Session-ID-ctx:
    Master-Key:
    PSK identity: None
    PSK identity hint: None
    SRP username: None
    Start Time: 1770929565
    Timeout   : 7200 (sec)
    Verify return code: 0 (ok)
    Extended master secret: no
---`,
			err: fmt.Errorf("command terminated with exit code 1"),
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := ProbeServicePortViaExec(mock, ctx, "10.217.4.83", 443, modernPolicy())
	assert.True(t, result.Compliant, "expected compliant (TLS 1.2 rejected), got: %s", result.Reason)
	assert.Equal(t, "TLS 1.3", result.NegotiatedVer)
}

func TestProbeServicePortViaExec_TLS13OnlyRejectsTLS12_IntermediateProfile(t *testing.T) {
	// Reproduces the CI smoke test failure: a TLS 1.3-only server is probed via exec
	// with Intermediate profile (min TLS 1.2). The openssl output for TLS 1.2 contains
	// "Protocol  : TLSv1.2" (the *attempted* version) but "Cipher is (NONE)" indicates
	// the handshake was rejected. The test must detect this as non-compliant.
	// openssl exits non-zero on handshake failure, so ExecCommandContainer returns an error.
	mock := newMockCommand(
		mockPattern{"-cipher", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nssl handshake failure\n---",
			err:    fmt.Errorf("command terminated with exit code 1"),
		}},
		mockPattern{"-tls1_2", mockExecResult{
			err: fmt.Errorf("command terminated with exit code 1"),
			stdout: `Connecting to 10.217.4.83
error:0A00042E:SSL routines:ssl3_read_bytes:tlsv1 alert protocol version:ssl/record/rec_layer_s3.c:916:SSL alert number 70
CONNECTED(00000003)
---
no peer certificate available
---
No client certificate CA names sent
---
SSL handshake has read 7 bytes and written 191 bytes
Verification: OK
---
New, (NONE), Cipher is (NONE)
Protocol: TLSv1.2
Secure Renegotiation IS NOT supported
Compression: NONE
Expansion: NONE
No ALPN negotiated
SSL-Session:
    Protocol  : TLSv1.2
    Cipher    : 0000
    Session-ID:
    Session-ID-ctx:
    Master-Key:
    PSK identity: None
    PSK identity hint: None
    SRP username: None
    Start Time: 1770929565
    Timeout   : 7200 (sec)
    Verify return code: 0 (ok)
    Extended master secret: no
---`,
		}},
		mockPattern{"-tls1_1", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nssl handshake failure\n---",
			err:    fmt.Errorf("command terminated with exit code 1"),
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := ProbeServicePortViaExec(mock, ctx, "10.217.4.83", 443, intermediatePolicy())
	assert.False(t, result.Compliant, "expected non-compliant: TLS 1.3-only server does not support TLS 1.2 required by Intermediate")
	assert.True(t, result.IsTLS, "expected IsTLS=true")
}

func TestProbeServicePortViaExec_ExecErrorWithOutput_IntermediateProfile(t *testing.T) {
	// Reproduces the actual CI behavior: when openssl fails a handshake, it exits
	// non-zero, and ExecCommandContainer returns a non-nil error. But stdout still
	// contains the openssl output that we need to parse. The code must NOT bail out
	// early on error; it must examine stdout to classify the result.
	rejectedOutput := `Connecting to 10.217.4.83
error:0A00042E:SSL routines:ssl3_read_bytes:tlsv1 alert protocol version:ssl/record/rec_layer_s3.c:916:SSL alert number 70
CONNECTED(00000003)
---
no peer certificate available
---
No client certificate CA names sent
---
SSL handshake has read 7 bytes and written 191 bytes
Verification: OK
---
New, (NONE), Cipher is (NONE)
Protocol: TLSv1.2
Secure Renegotiation IS NOT supported
Compression: NONE
Expansion: NONE
No ALPN negotiated
SSL-Session:
    Protocol  : TLSv1.2
    Cipher    : 0000
    Session-ID:
    Session-ID-ctx:
    Master-Key:
    PSK identity: None
    PSK identity hint: None
    SRP username: None
    Start Time: 1770929565
    Timeout   : 7200 (sec)
    Verify return code: 0 (ok)
    Extended master secret: no
---`
	mock := newMockCommand(
		mockPattern{"-cipher", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nssl handshake failure\n---",
			err:    fmt.Errorf("command terminated with exit code 1"),
		}},
		mockPattern{"-tls1_2", mockExecResult{
			stdout: rejectedOutput,
			err:    fmt.Errorf("command terminated with exit code 1"),
		}},
		mockPattern{"-tls1_1", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nssl handshake failure\n---",
			err:    fmt.Errorf("command terminated with exit code 1"),
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := ProbeServicePortViaExec(mock, ctx, "10.217.4.83", 443, intermediatePolicy())
	assert.False(t, result.Compliant, "expected non-compliant: TLS 1.3-only server does not support TLS 1.2 required by Intermediate")
	assert.True(t, result.IsTLS, "expected IsTLS=true")
}

func TestProbeServicePortViaExec_ExecErrorNoOutput(t *testing.T) {
	// When exec returns an error with no useful output (e.g., openssl not installed),
	// the result should be unreachable.
	mock := newMockCommand(
		mockPattern{"-tls1_2", mockExecResult{
			stdout: "",
			err:    fmt.Errorf("command not found: openssl"),
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := ProbeServicePortViaExec(mock, ctx, "10.0.0.1", 443, intermediatePolicy())
	assert.True(t, result.Compliant, "expected compliant (unreachable), got: %s", result.Reason)
	assert.False(t, result.Reachable, "expected Reachable=false")
}

func TestProbeServicePortViaExec_ConnectionRefused(t *testing.T) {
	mock := newMockCommand(
		mockPattern{"-tls1_2", mockExecResult{
			stdout: "connect:errno=111\nConnection refused",
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := ProbeServicePortViaExec(mock, ctx, "10.0.0.1", 443, intermediatePolicy())
	assert.True(t, result.Compliant, "expected compliant (unreachable), got: %s", result.Reason)
	assert.False(t, result.Reachable, "expected Reachable=false")
}

func TestProbeServicePortViaExec_NonTLS(t *testing.T) {
	mock := newMockCommand(
		mockPattern{"-tls1_2", mockExecResult{
			stdout: "CONNECTED(00000003)\nwrite:errno=104\n",
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := ProbeServicePortViaExec(mock, ctx, "10.0.0.1", 8080, intermediatePolicy())
	assert.True(t, result.Compliant, "expected compliant (non-TLS informational), got: %s", result.Reason)
	assert.False(t, result.IsTLS, "expected IsTLS=false for non-TLS service")
}

func TestProbeServicePortViaExec_NonTLS_RealisticOpenSSLOutput(t *testing.T) {
	// Reproduces the CI failure where openssl connecting to a plain HTTP service
	// (e.g. test-service-dualstack:8080) produces output containing
	// "Cipher is (NONE)" and "Protocol  : TLSv1.2" even though the service is
	// NOT TLS. The key insight is that openssl always reports the *attempted*
	// protocol version and "SSL handshake has read N bytes" even for non-TLS
	// connections. The only reliable TLS indicator is a TLS alert message
	// (e.g. "alert protocol version"). Non-TLS services produce errors like
	// "packet length too long" or "record layer failure" instead.
	mock := newMockCommand(
		mockPattern{"-tls1_2", mockExecResult{
			stdout: `Connecting to 10.96.200.200
A0EB58ABFFFF0000:error:0A0000C6:SSL routines:tls_get_more_records:packet length too long:ssl/record/methods/tls_common.c:662:
A0EB58ABFFFF0000:error:0A000139:SSL routines::record layer failure:ssl/record/rec_layer_s3.c:696:
CONNECTED(00000003)
---
no peer certificate available
---
No client certificate CA names sent
---
SSL handshake has read 5 bytes and written 198 bytes
Verification: OK
---
New, (NONE), Cipher is (NONE)
Protocol: TLSv1.2
Secure Renegotiation IS NOT supported
Compression: NONE
Expansion: NONE
No ALPN negotiated
SSL-Session:
    Protocol  : TLSv1.2
    Cipher    : 0000
    Session-ID:
    Session-ID-ctx:
    Master-Key:
    PSK identity: None
    PSK identity hint: None
    SRP username: None
    Start Time: 1770930000
    Timeout   : 7200 (sec)
    Verify return code: 0 (ok)
    Extended master secret: no
---`,
			err: fmt.Errorf("command terminated with exit code 1"),
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := ProbeServicePortViaExec(mock, ctx, "10.96.200.200", 8080, intermediatePolicy())
	assert.True(t, result.Compliant, "expected compliant (non-TLS service), got: %s", result.Reason)
	assert.False(t, result.IsTLS, "expected IsTLS=false for plain HTTP service")
	assert.True(t, result.Reachable, "expected Reachable=true (connection established)")
}

func TestResolveTLSProfile_Nil(t *testing.T) {
	policy := ResolveTLSProfile(nil)
	assert.Equal(t, "Intermediate", policy.ProfileType)
	assert.Equal(t, uint16(tls.VersionTLS12), policy.MinTLSVersion)
	assert.NotEmpty(t, policy.AllowedCipherIDs, "expected non-empty allowed cipher list for Intermediate")
}

func TestResolveTLSProfile_Old(t *testing.T) {
	policy := ResolveTLSProfile(&configv1.TLSSecurityProfile{Type: configv1.TLSProfileOldType})
	assert.Equal(t, "Old", policy.ProfileType)
	assert.Equal(t, uint16(tls.VersionTLS10), policy.MinTLSVersion)
	// Old profile should have more allowed ciphers than Intermediate
	intermediate := ResolveTLSProfile(nil)
	assert.Greater(t, len(policy.AllowedCipherIDs), len(intermediate.AllowedCipherIDs),
		"expected Old to have more ciphers than Intermediate")
}

func TestResolveTLSProfile_Modern(t *testing.T) {
	policy := ResolveTLSProfile(&configv1.TLSSecurityProfile{Type: configv1.TLSProfileModernType})
	assert.Equal(t, "Modern", policy.ProfileType)
	assert.Equal(t, uint16(tls.VersionTLS13), policy.MinTLSVersion)
	assert.Empty(t, policy.AllowedCipherIDs, "expected no TLS 1.2 allowed ciphers for Modern")
}

func TestResolveTLSProfile_Custom(t *testing.T) {
	policy := ResolveTLSProfile(&configv1.TLSSecurityProfile{
		Type: configv1.TLSProfileCustomType,
		Custom: &configv1.CustomTLSProfile{
			TLSProfileSpec: configv1.TLSProfileSpec{
				MinTLSVersion: configv1.VersionTLS12,
				Ciphers: []string{
					"ECDHE-ECDSA-AES128-GCM-SHA256",
					"ECDHE-RSA-AES128-GCM-SHA256",
				},
			},
		},
	})
	assert.Equal(t, "Custom", policy.ProfileType)
	assert.Equal(t, uint16(tls.VersionTLS12), policy.MinTLSVersion)
	assert.Len(t, policy.AllowedCipherIDs, 2)
}

func TestVersionBelow(t *testing.T) {
	tests := []struct {
		ver      uint16
		expected uint16
	}{
		{tls.VersionTLS13, tls.VersionTLS12},
		{tls.VersionTLS12, tls.VersionTLS11},
		{tls.VersionTLS11, tls.VersionTLS10},
		{tls.VersionTLS10, 0},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, versionBelow(tc.ver), "versionBelow(0x%04x)", tc.ver)
	}
}

func TestTLSVersionString(t *testing.T) {
	tests := []struct {
		ver      uint16
		expected string
	}{
		{tls.VersionTLS10, "TLS 1.0"},
		{tls.VersionTLS11, "TLS 1.1"},
		{tls.VersionTLS12, "TLS 1.2"},
		{tls.VersionTLS13, "TLS 1.3"},
		{0x0000, "unknown (0x0000)"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tlsVersionString(tc.ver))
		})
	}
}

func TestOpensslVersionFlag(t *testing.T) {
	tests := []struct {
		ver      uint16
		expected string
	}{
		{tls.VersionTLS10, "-tls1"},
		{tls.VersionTLS11, "-tls1_1"},
		{tls.VersionTLS12, "-tls1_2"},
		{tls.VersionTLS13, "-tls1_3"},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, opensslVersionFlag(tc.ver), "opensslVersionFlag(0x%04x)", tc.ver)
	}
}

func TestDefaultTLSPolicy(t *testing.T) {
	policy := DefaultTLSPolicy()
	assert.Equal(t, "Intermediate", policy.ProfileType)
	assert.Equal(t, uint16(tls.VersionTLS12), policy.MinTLSVersion)
}

func TestIsOCPVersionAtLeast(t *testing.T) {
	tests := []struct {
		name       string
		ocpVersion string
		minVersion string
		expected   bool
	}{
		{"4.21 below 4.22", "4.21", "4.22", false},
		{"4.22.0 equals 4.22", "4.22.0", "4.22", true},
		{"4.22 equals 4.22", "4.22", "4.22", true},
		{"5.0 above 4.22", "5.0", "4.22", true},
		{"4.23.1 above 4.22", "4.23.1", "4.22", true},
		{"empty ocpVersion", "", "4.22", false},
		{"empty minVersion", "4.22", "", false},
		{"invalid ocpVersion", "not-a-version", "4.22", false},
		{"0.0.0 below 4.22", "0.0.0", "4.22", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsOCPVersionAtLeast(tc.ocpVersion, tc.minVersion)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetClusterTLSPolicy_OCPBelowThreshold(t *testing.T) {
	// On OCP < 4.22, GetClusterTLSPolicy should return Intermediate default
	// without calling the API (so nil client is safe).
	policy := GetClusterTLSPolicy(nil, true, "4.21.3")
	assert.Equal(t, "Intermediate", policy.ProfileType)
	assert.Equal(t, uint16(tls.VersionTLS12), policy.MinTLSVersion)
}

func TestGetClusterTLSPolicy_NonOCP(t *testing.T) {
	// Non-OCP cluster should return Intermediate default without calling the API.
	policy := GetClusterTLSPolicy(nil, false, "")
	assert.Equal(t, "Intermediate", policy.ProfileType)
	assert.Equal(t, uint16(tls.VersionTLS12), policy.MinTLSVersion)
}
