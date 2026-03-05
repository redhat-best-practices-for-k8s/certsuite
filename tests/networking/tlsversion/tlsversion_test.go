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
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		mockPattern{"-tls1_3", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nProtocol  : TLSv1.3\nCipher    : TLS_AES_256_GCM_SHA384\n---",
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

func TestVersionsAbove(t *testing.T) {
	tests := []struct {
		name     string
		minVer   uint16
		expected []uint16
	}{
		{"TLS 1.0 returns 1.1, 1.2, 1.3", tls.VersionTLS10, []uint16{tls.VersionTLS11, tls.VersionTLS12, tls.VersionTLS13}},
		{"TLS 1.1 returns 1.2, 1.3", tls.VersionTLS11, []uint16{tls.VersionTLS12, tls.VersionTLS13}},
		{"TLS 1.2 returns 1.3", tls.VersionTLS12, []uint16{tls.VersionTLS13}},
		{"TLS 1.3 returns empty", tls.VersionTLS13, nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := versionsAbove(tc.minVer)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestOpensslVersionNegotiated(t *testing.T) {
	tests := []struct {
		name     string
		stdout   string
		ver      uint16
		expected bool
	}{
		{"TLS 1.2 with spaces", "Protocol  : TLSv1.2\nCipher    : AES", tls.VersionTLS12, true},
		{"TLS 1.2 without spaces", "Protocol: TLSv1.2\nCipher: AES", tls.VersionTLS12, true},
		{"TLS 1.3 negotiated", "Protocol  : TLSv1.3\nCipher    : TLS_AES_256_GCM_SHA384", tls.VersionTLS13, true},
		{"wrong version", "Protocol  : TLSv1.2\nCipher    : AES", tls.VersionTLS13, false},
		{"empty output", "", tls.VersionTLS12, false},
		{"cipher none still matches protocol", "Cipher is (NONE)\nProtocol  : TLSv1.2", tls.VersionTLS12, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := opensslVersionNegotiated(tc.stdout, tc.ver)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestProbeServicePortViaExec_IntermediateRejectsTLS13Only(t *testing.T) {
	// Server only supports TLS 1.2, rejects TLS 1.3. Intermediate profile requires
	// both TLS 1.2 and TLS 1.3, so this should be non-compliant.
	mock := newMockCommand(
		mockPattern{"-cipher", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nssl handshake failure\n---",
		}},
		mockPattern{"-tls1_3", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nalert protocol version\nCipher is (NONE)\n---",
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
	assert.False(t, result.Compliant, "expected non-compliant: TLS 1.2-only server doesn't support TLS 1.3 required by Intermediate")
	assert.True(t, result.IsTLS)
	assert.Contains(t, result.Reason, "rejected TLS 1.3")
}

func TestProbeServicePortViaExec_OldProfileAllVersions(t *testing.T) {
	// Old profile (min TLS 1.0) requires the server to accept all versions 1.0-1.3.
	oldPolicy := ResolveTLSProfile(&configv1.TLSSecurityProfile{Type: configv1.TLSProfileOldType})
	mock := newMockCommand(
		mockPattern{"-cipher", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nssl handshake failure\n---",
		}},
		mockPattern{"-tls1_3", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nProtocol  : TLSv1.3\nCipher    : TLS_AES_256_GCM_SHA384\n---",
		}},
		mockPattern{"-tls1_2", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nProtocol  : TLSv1.2\nCipher    : ECDHE-RSA-AES128-GCM-SHA256\n---",
		}},
		mockPattern{"-tls1_1", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nProtocol  : TLSv1.1\nCipher    : AES128-SHA\n---",
		}},
		mockPattern{"-tls1 ", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nProtocol  : TLSv1\nCipher    : AES128-SHA\n---",
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := ProbeServicePortViaExec(mock, ctx, "10.0.0.1", 443, oldPolicy)
	assert.True(t, result.Compliant, "expected compliant: server accepts all TLS versions, got: %s", result.Reason)
}

func TestProbeServicePortViaExec_OldProfileRejectsTLS11(t *testing.T) {
	// Old profile server that rejects TLS 1.1 should be non-compliant.
	oldPolicy := ResolveTLSProfile(&configv1.TLSSecurityProfile{Type: configv1.TLSProfileOldType})
	mock := newMockCommand(
		mockPattern{"-tls1_1", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nalert protocol version\nCipher is (NONE)\n---",
		}},
		mockPattern{"-tls1 ", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nProtocol  : TLSv1\nCipher    : AES128-SHA\n---",
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := ProbeServicePortViaExec(mock, ctx, "10.0.0.1", 443, oldPolicy)
	assert.False(t, result.Compliant, "expected non-compliant: server rejects TLS 1.1 under Old profile")
	assert.Contains(t, result.Reason, "rejected TLS 1.1")
}

func TestProbeExecVersion_BelowMinimum_ServerRejects(t *testing.T) {
	mock := newMockCommand(
		mockPattern{"-tls1_1", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nalert protocol version\nCipher is (NONE)\n---",
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := probeExecVersion(mock, ctx, "10.0.0.1:443", tls.VersionTLS11, false, intermediatePolicy())
	assert.True(t, result.Compliant, "expected compliant: server correctly rejected below-minimum version")
}

func TestProbeExecVersion_BelowMinimum_ServerAccepts(t *testing.T) {
	mock := newMockCommand(
		mockPattern{"-tls1_1", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nProtocol  : TLSv1.1\nCipher    : AES128-SHA\n---",
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := probeExecVersion(mock, ctx, "10.0.0.1:443", tls.VersionTLS11, false, intermediatePolicy())
	assert.False(t, result.Compliant, "expected non-compliant: server accepted below-minimum TLS 1.1")
	assert.Contains(t, result.Reason, "accepts TLS 1.1")
}

func TestProbeExecVersion_AboveMinimum_ServerAccepts(t *testing.T) {
	mock := newMockCommand(
		mockPattern{"-tls1_3", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nProtocol  : TLSv1.3\nCipher    : TLS_AES_256_GCM_SHA384\n---",
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := probeExecVersion(mock, ctx, "10.0.0.1:443", tls.VersionTLS13, true, intermediatePolicy())
	assert.True(t, result.Compliant, "expected compliant: server accepts TLS 1.3")
}

func TestProbeExecVersion_AboveMinimum_ServerRejects(t *testing.T) {
	mock := newMockCommand(
		mockPattern{"-tls1_3", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nalert protocol version\nCipher is (NONE)\n---",
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := probeExecVersion(mock, ctx, "10.0.0.1:443", tls.VersionTLS13, true, intermediatePolicy())
	assert.False(t, result.Compliant, "expected non-compliant: server rejected TLS 1.3 required by profile")
	assert.Contains(t, result.Reason, "rejected TLS 1.3")
}

func TestClassifyExecNoMinVersion(t *testing.T) {
	policy := intermediatePolicy()

	tests := []struct {
		name            string
		stdout          string
		expectCompliant bool
		expectIsTLS     bool
		expectReachable bool
		expectReason    string
	}{
		{
			name:            "connection refused",
			stdout:          "connect:errno=111\nConnection refused",
			expectCompliant: true,
			expectIsTLS:     false,
			expectReachable: false,
			expectReason:    "port unreachable",
		},
		{
			name:            "errno only",
			stdout:          "connect:errno=113",
			expectCompliant: true,
			expectIsTLS:     false,
			expectReachable: false,
			expectReason:    "port unreachable",
		},
		{
			name:            "TLS alert protocol version",
			stdout:          "CONNECTED(00000003)\nalert protocol version\nCipher is (NONE)",
			expectCompliant: false,
			expectIsTLS:     true,
			expectReachable: true,
			expectReason:    "does not support TLS 1.2",
		},
		{
			name:            "TLS alert handshake failure",
			stdout:          "CONNECTED(00000003)\nalert handshake failure\nCipher is (NONE)",
			expectCompliant: false,
			expectIsTLS:     true,
			expectReachable: true,
			expectReason:    "does not support TLS 1.2",
		},
		{
			name:            "TLS handshake failure",
			stdout:          "CONNECTED(00000003)\nhandshake failure",
			expectCompliant: false,
			expectIsTLS:     true,
			expectReachable: true,
			expectReason:    "does not support TLS 1.2",
		},
		{
			name:            "no recognizable output",
			stdout:          "",
			expectCompliant: true,
			expectIsTLS:     false,
			expectReachable: false,
			expectReason:    "no recognizable output",
		},
		{
			name:            "garbage output",
			stdout:          "some random text with no openssl markers",
			expectCompliant: true,
			expectIsTLS:     false,
			expectReachable: false,
			expectReason:    "no recognizable output",
		},
		{
			name:            "non-TLS service with CONNECTED",
			stdout:          "CONNECTED(00000003)\nwrite:errno=104\nCipher is (NONE)",
			expectCompliant: true,
			expectIsTLS:     false,
			expectReachable: true,
			expectReason:    "non-TLS service",
		},
		{
			name:            "non-TLS service with packet length error",
			stdout:          "CONNECTED(00000003)\npacket length too long\nCipher is (NONE)\nProtocol  : TLSv1.2",
			expectCompliant: true,
			expectIsTLS:     false,
			expectReachable: true,
			expectReason:    "non-TLS service",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := classifyExecNoMinVersion(tc.stdout, policy)
			assert.Equal(t, tc.expectCompliant, result.Compliant, "Compliant mismatch")
			assert.Equal(t, tc.expectIsTLS, result.IsTLS, "IsTLS mismatch")
			assert.Equal(t, tc.expectReachable, result.Reachable, "Reachable mismatch")
			assert.Contains(t, result.Reason, tc.expectReason, "Reason mismatch")
		})
	}
}

func TestComputeDisallowedOpenSSLCiphers(t *testing.T) {
	// Helper: build a set from a slice for easier membership checks.
	toSet := func(s []string) map[string]bool {
		m := make(map[string]bool, len(s))
		for _, v := range s {
			m[v] = true
		}
		return m
	}

	t.Run("Old profile has no disallowed ciphers", func(t *testing.T) {
		policy := ResolveTLSProfile(&configv1.TLSSecurityProfile{Type: configv1.TLSProfileOldType})
		disallowed := computeDisallowedOpenSSLCiphers(policy)
		assert.Empty(t, disallowed, "Old profile allows all ciphers in opensslToGoCipher")
	})

	t.Run("Modern profile disallows all TLS 1.2 ciphers", func(t *testing.T) {
		policy := ResolveTLSProfile(&configv1.TLSSecurityProfile{Type: configv1.TLSProfileModernType})
		disallowed := computeDisallowedOpenSSLCiphers(policy)
		assert.Equal(t, len(opensslToGoCipher), len(disallowed),
			"Modern profile should disallow every cipher in opensslToGoCipher")
	})

	t.Run("Intermediate profile disallows some ciphers", func(t *testing.T) {
		policy := intermediatePolicy()
		disallowed := computeDisallowedOpenSSLCiphers(policy)
		disallowedSet := toSet(disallowed)

		assert.NotEmpty(t, disallowed)
		assert.Less(t, len(disallowed), len(opensslToGoCipher),
			"Intermediate should disallow fewer ciphers than the full set")

		// Intermediate allows ECDHE-RSA-AES128-GCM-SHA256 — it must NOT be disallowed.
		assert.False(t, disallowedSet["ECDHE-RSA-AES128-GCM-SHA256"],
			"ECDHE-RSA-AES128-GCM-SHA256 is allowed by Intermediate and should not be disallowed")

		// Intermediate does not allow DES-CBC3-SHA — it must be disallowed.
		assert.True(t, disallowedSet["DES-CBC3-SHA"],
			"DES-CBC3-SHA is not in the Intermediate profile and should be disallowed")
	})

	t.Run("disallowed and allowed are disjoint and cover all ciphers", func(t *testing.T) {
		policy := intermediatePolicy()
		disallowed := computeDisallowedOpenSSLCiphers(policy)
		disallowedSet := toSet(disallowed)

		profileSpec := getProfileSpec(policy)
		allowedSet := make(map[string]bool)
		for _, name := range profileSpec.Ciphers {
			if !tls13CipherNames[name] {
				if _, hasGoMapping := opensslToGoCipher[name]; hasGoMapping {
					allowedSet[name] = true
				}
			}
		}

		for name := range opensslToGoCipher {
			inAllowed := allowedSet[name]
			inDisallowed := disallowedSet[name]
			assert.NotEqual(t, inAllowed, inDisallowed,
				"cipher %s must be in exactly one of allowed or disallowed (allowed=%v, disallowed=%v)",
				name, inAllowed, inDisallowed)
		}
	})
}

func TestOpensslHandshakeRejected(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected bool
	}{
		{"cipher none", "CONNECTED(00000003)\nCipher is (NONE)\nProtocol  : TLSv1.2", true},
		{"alert protocol version", "CONNECTED(00000003)\nalert protocol version", true},
		{"handshake failure", "CONNECTED(00000003)\nhandshake failure", true},
		{"successful handshake", "CONNECTED(00000003)\nProtocol  : TLSv1.3\nCipher    : TLS_AES_256_GCM_SHA384", false},
		{"empty output", "", false},
		{"unrelated error", "CONNECTED(00000003)\npacket length too long", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, opensslHandshakeRejected(tc.output))
		})
	}
}

func TestOcpVersionToGoVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    configv1.TLSProtocolVersion
		expected uint16
	}{
		{"TLS 1.0", configv1.VersionTLS10, tls.VersionTLS10},
		{"TLS 1.1", configv1.VersionTLS11, tls.VersionTLS11},
		{"TLS 1.2", configv1.VersionTLS12, tls.VersionTLS12},
		{"TLS 1.3", configv1.VersionTLS13, tls.VersionTLS13},
		{"empty string defaults to TLS 1.2", "", tls.VersionTLS12},
		{"unknown defaults to TLS 1.2", "VersionTLS99", tls.VersionTLS12},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ocpVersionToGoVersion(tc.input))
		})
	}
}

func TestTLSVersionName(t *testing.T) {
	assert.Equal(t, "TLS 1.2", TLSVersionName(tls.VersionTLS12))
	assert.Equal(t, "TLS 1.3", TLSVersionName(tls.VersionTLS13))
}

func TestOpensslVersionFlag_Default(t *testing.T) {
	assert.Equal(t, opensslFlagTLS12, opensslVersionFlag(0x0000))
}

func TestOpensslVersionName_Default(t *testing.T) {
	assert.Equal(t, opensslProtoNameTLS12, opensslVersionName(0x0000))
}

func TestExtractOpenSSLCipher(t *testing.T) {
	tests := []struct {
		name     string
		stdout   string
		expected string
	}{
		{"with spaces", "Cipher    : ECDHE-RSA-AES128-GCM-SHA256\nProtocol  : TLSv1.2", "ECDHE-RSA-AES128-GCM-SHA256"},
		{"without spaces", "Cipher: TLS_AES_256_GCM_SHA384\nProtocol: TLSv1.3", "TLS_AES_256_GCM_SHA384"},
		{"cipher 0000", "Cipher    : 0000\nProtocol  : TLSv1.2", "0000"},
		{"no cipher line", "Protocol  : TLSv1.2\nSome other output", "unknown"},
		{"empty output", "", "unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, extractOpenSSLCipher(tc.stdout))
		})
	}
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "short", truncate("short", 200))
	assert.Equal(t, "abc...", truncate("abcdef", 3))
	assert.Equal(t, "", truncate("", 10))
	assert.Equal(t, "exact", truncate("exact", 5))
}

func TestGetProfileSpec(t *testing.T) {
	tests := []struct {
		name        string
		profileType string
		expectedMin configv1.TLSProtocolVersion
	}{
		{"Old", string(configv1.TLSProfileOldType), configv1.VersionTLS10},
		{"Intermediate", string(configv1.TLSProfileIntermediateType), configv1.VersionTLS12},
		{"Modern", string(configv1.TLSProfileModernType), configv1.VersionTLS13},
		{"unknown defaults to Intermediate", "Unknown", configv1.VersionTLS12},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec := getProfileSpec(TLSPolicy{ProfileType: tc.profileType})
			assert.NotNil(t, spec)
			assert.Equal(t, tc.expectedMin, spec.MinTLSVersion)
		})
	}
}

func TestBuildReportObject(t *testing.T) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-service",
			Namespace: "test-ns",
		},
	}

	t.Run("compliant with TLS version", func(t *testing.T) {
		result := TLSProbeResult{
			Compliant:     true,
			NegotiatedVer: "TLS 1.3",
			Reason:        "server honors profile",
		}
		ro := buildReportObject(svc, 443, result)
		assert.Equal(t, testhelper.ServiceType, ro.ObjectType)
		assertReportField(t, ro, testhelper.Namespace, "test-ns")
		assertReportField(t, ro, testhelper.ServiceName, "my-service")
		assertReportField(t, ro, testhelper.PortNumber, "443")
		assertReportField(t, ro, testhelper.PortProtocol, "TCP")
		assertReportField(t, ro, testhelper.TLSVersion, "TLS 1.3")
	})

	t.Run("non-compliant without TLS version", func(t *testing.T) {
		result := TLSProbeResult{
			Compliant: false,
			Reason:    "server rejected TLS 1.2",
		}
		ro := buildReportObject(svc, 8080, result)
		assert.Equal(t, testhelper.ServiceType, ro.ObjectType)
		assertReportField(t, ro, testhelper.PortNumber, "8080")
		// NegotiatedVer is empty, so TLSVersion field should not be present.
		for _, key := range ro.ObjectFieldsKeys {
			assert.NotEqual(t, testhelper.TLSVersion, key, "TLSVersion should not be set when NegotiatedVer is empty")
		}
	})
}

func assertReportField(t *testing.T, ro *testhelper.ReportObject, key, expectedValue string) {
	t.Helper()
	for i, k := range ro.ObjectFieldsKeys {
		if k == key {
			assert.Equal(t, expectedValue, ro.ObjectFieldsValues[i], "field %s", key)
			return
		}
	}
	t.Errorf("field %q not found in report object", key)
}

func TestProbeExecCipherCompliance_ServerAcceptsDisallowed(t *testing.T) {
	mock := newMockCommand(
		mockPattern{"-cipher", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nProtocol  : TLSv1.2\nCipher    : DES-CBC3-SHA\n---",
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := probeExecCipherCompliance(mock, ctx, "10.0.0.1:443", intermediatePolicy())
	assert.NotNil(t, result, "expected non-nil result for accepted disallowed cipher")
	assert.False(t, result.Compliant)
	assert.Contains(t, result.Reason, "disallowed cipher")
}

func TestProbeExecCipherCompliance_TLS13SkipsCipherCheck(t *testing.T) {
	mock := newMockCommand()
	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := probeExecCipherCompliance(mock, ctx, "10.0.0.1:443", modernPolicy())
	assert.Nil(t, result, "Modern profile (TLS 1.3) should skip cipher check")
}

func TestProbeExecCipherCompliance_ServerRejectsDisallowed(t *testing.T) {
	mock := newMockCommand(
		mockPattern{"-cipher", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nhandshake failure\nCipher is (NONE)\n---",
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := probeExecCipherCompliance(mock, ctx, "10.0.0.1:443", intermediatePolicy())
	assert.Nil(t, result, "expected nil (compliant) when server rejects disallowed ciphers")
}

func TestProbeExecMinVersion_ServerNegotiatesTLS13WhenMinIs12(t *testing.T) {
	// Server negotiates TLS 1.3 when we ask for TLS 1.2 minimum — this is fine.
	mock := newMockCommand(
		mockPattern{"-tls1_2", mockExecResult{
			stdout: "CONNECTED(00000003)\n---\nProtocol  : TLSv1.3\nCipher    : TLS_AES_256_GCM_SHA384\n---",
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := probeExecMinVersion(mock, ctx, "10.0.0.1:443", intermediatePolicy())
	assert.Nil(t, result, "expected nil (success) when server negotiates TLS 1.3 for Intermediate profile")
}

func TestResolveTLSProfile_UnknownType(t *testing.T) {
	policy := ResolveTLSProfile(&configv1.TLSSecurityProfile{Type: "SomethingElse"})
	assert.Equal(t, "Intermediate", policy.ProfileType)
	assert.Equal(t, uint16(tls.VersionTLS12), policy.MinTLSVersion)
}

func TestResolveTLSProfile_CustomNilSpec(t *testing.T) {
	policy := ResolveTLSProfile(&configv1.TLSSecurityProfile{
		Type:   configv1.TLSProfileCustomType,
		Custom: nil,
	})
	assert.Equal(t, "Intermediate", policy.ProfileType)
}
