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
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
)

// modernPolicy is the Modern profile (TLS 1.3 only).
func modernPolicy() TLSPolicy {
	return ResolveTLSProfile(&configv1.TLSSecurityProfile{Type: configv1.TLSProfileModernType})
}

// intermediatePolicy is the Intermediate profile (TLS 1.2 minimum).
func intermediatePolicy() TLSPolicy {
	return ResolveTLSProfile(nil) // nil = Intermediate (default)
}

func TestProbeServicePortTLS_TLS13Only_ModernProfile(t *testing.T) {
	// Server enforces TLS 1.3 only — compliant with Modern profile
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	server.TLS = &tls.Config{
		MinVersion: tls.VersionTLS13,
		MaxVersion: tls.VersionTLS13,
	}
	server.StartTLS()
	defer server.Close()

	host, port := parseHostPort(t, server.Listener.Addr().String())

	result := ProbeServicePortTLS(host, port, modernPolicy())
	if !result.Compliant {
		t.Errorf("expected compliant, got non-compliant: %s", result.Reason)
	}
	if !result.IsTLS {
		t.Error("expected IsTLS=true")
	}
	if result.NegotiatedVer != "TLS 1.3" {
		t.Errorf("expected negotiated version TLS 1.3, got %q", result.NegotiatedVer)
	}
}

func TestProbeServicePortTLS_AcceptsTLS12_ModernProfile(t *testing.T) {
	// Server accepts TLS 1.2 and 1.3 — non-compliant with Modern profile
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	server.TLS = &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
	}
	server.StartTLS()
	defer server.Close()

	host, port := parseHostPort(t, server.Listener.Addr().String())

	result := ProbeServicePortTLS(host, port, modernPolicy())
	if result.Compliant {
		t.Errorf("expected non-compliant, got compliant: %s", result.Reason)
	}
	if !result.IsTLS {
		t.Error("expected IsTLS=true")
	}
	if result.NegotiatedVer != "TLS 1.2" {
		t.Errorf("expected negotiated version TLS 1.2, got %q", result.NegotiatedVer)
	}
}

func TestProbeServicePortTLS_AcceptsTLS12_IntermediateProfile(t *testing.T) {
	// Server accepts TLS 1.2 and 1.3 with only Intermediate-profile ciphers — compliant
	policy := intermediatePolicy()
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	server.TLS = &tls.Config{
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS13,
		CipherSuites: policy.AllowedCipherIDs,
	}
	server.StartTLS()
	defer server.Close()

	host, port := parseHostPort(t, server.Listener.Addr().String())

	result := ProbeServicePortTLS(host, port, policy)
	if !result.Compliant {
		t.Errorf("expected compliant with Intermediate profile, got non-compliant: %s", result.Reason)
	}
	if !result.IsTLS {
		t.Error("expected IsTLS=true")
	}
}

func TestProbeServicePortTLS_TLS13Only_IntermediateProfile(t *testing.T) {
	// Server enforces TLS 1.3 only — also compliant with Intermediate profile
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	server.TLS = &tls.Config{
		MinVersion: tls.VersionTLS13,
		MaxVersion: tls.VersionTLS13,
	}
	server.StartTLS()
	defer server.Close()

	host, port := parseHostPort(t, server.Listener.Addr().String())

	result := ProbeServicePortTLS(host, port, intermediatePolicy())
	if !result.Compliant {
		t.Errorf("expected compliant with Intermediate profile, got non-compliant: %s", result.Reason)
	}
}

func TestProbeServicePortTLS_PlainTCP(t *testing.T) {
	// Plain TCP server (no TLS)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer listener.Close()

	// Accept connections and immediately close them
	go func() {
		for {
			conn, acceptErr := listener.Accept()
			if acceptErr != nil {
				return
			}
			// Write some plain text to make it clear this is not TLS
			conn.Write([]byte("hello\n")) //nolint:errcheck
			conn.Close()
		}
	}()

	host, port := parseHostPort(t, listener.Addr().String())

	result := ProbeServicePortTLS(host, port, intermediatePolicy())
	if !result.Compliant {
		t.Errorf("expected compliant (non-TLS informational), got non-compliant: %s", result.Reason)
	}
	if result.IsTLS {
		t.Error("expected IsTLS=false for plain TCP")
	}
}

func TestProbeServicePortTLS_PortNotListening(t *testing.T) {
	// Use a port that nothing is listening on
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	_, port := parseHostPort(t, listener.Addr().String())
	listener.Close() // Close immediately so port is not listening

	result := ProbeServicePortTLS("127.0.0.1", port, intermediatePolicy())
	if !result.Compliant {
		t.Errorf("expected compliant (unreachable), got non-compliant: %s", result.Reason)
	}
	if result.Reachable {
		t.Error("expected Reachable=false")
	}
}

// mockCommand implements clientsholder.Command for testing the exec fallback.
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
	if !result.Compliant {
		t.Errorf("expected compliant, got non-compliant: %s", result.Reason)
	}
	if result.NegotiatedVer != "TLS 1.3" {
		t.Errorf("expected TLS 1.3, got %q", result.NegotiatedVer)
	}
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
	if result.Compliant {
		t.Errorf("expected non-compliant, got compliant: %s", result.Reason)
	}
	if result.NegotiatedVer != "TLS 1.2" {
		t.Errorf("expected TLS 1.2, got %q", result.NegotiatedVer)
	}
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
	if !result.Compliant {
		t.Errorf("expected compliant with Intermediate profile, got non-compliant: %s", result.Reason)
	}
}

func TestProbeServicePortViaExec_TLS13OnlyRejectsTLS12_ModernProfile(t *testing.T) {
	// Reproduces real openssl output when a TLS 1.3-only server rejects TLS 1.2.
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
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := ProbeServicePortViaExec(mock, ctx, "10.217.4.83", 443, modernPolicy())
	if !result.Compliant {
		t.Errorf("expected compliant (TLS 1.2 rejected), got non-compliant: %s", result.Reason)
	}
	if result.NegotiatedVer != "TLS 1.3" {
		t.Errorf("expected TLS 1.3, got %q", result.NegotiatedVer)
	}
}

func TestProbeServicePortViaExec_ConnectionRefused(t *testing.T) {
	mock := newMockCommand(
		mockPattern{"-tls1_2", mockExecResult{
			stdout: "connect:errno=111\nConnection refused",
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := ProbeServicePortViaExec(mock, ctx, "10.0.0.1", 443, intermediatePolicy())
	if !result.Compliant {
		t.Errorf("expected compliant (unreachable), got non-compliant: %s", result.Reason)
	}
	if result.Reachable {
		t.Error("expected Reachable=false")
	}
}

func TestProbeServicePortViaExec_NonTLS(t *testing.T) {
	mock := newMockCommand(
		mockPattern{"-tls1_2", mockExecResult{
			stdout: "CONNECTED(00000003)\nwrite:errno=104\n",
		}},
	)

	ctx := clientsholder.NewContext("ns", "pod", "container")
	result := ProbeServicePortViaExec(mock, ctx, "10.0.0.1", 8080, intermediatePolicy())
	if !result.Compliant {
		t.Errorf("expected compliant (non-TLS informational), got non-compliant: %s", result.Reason)
	}
	if result.IsTLS {
		t.Error("expected IsTLS=false for non-TLS service")
	}
}

func TestResolveTLSProfile_Nil(t *testing.T) {
	policy := ResolveTLSProfile(nil)
	if policy.ProfileType != "Intermediate" {
		t.Errorf("expected Intermediate profile, got %q", policy.ProfileType)
	}
	if policy.MinTLSVersion != tls.VersionTLS12 {
		t.Errorf("expected min TLS 1.2, got 0x%04x", policy.MinTLSVersion)
	}
	if len(policy.AllowedCipherIDs) == 0 {
		t.Error("expected non-empty allowed cipher list for Intermediate")
	}
}

func TestResolveTLSProfile_Old(t *testing.T) {
	policy := ResolveTLSProfile(&configv1.TLSSecurityProfile{Type: configv1.TLSProfileOldType})
	if policy.ProfileType != "Old" {
		t.Errorf("expected Old profile, got %q", policy.ProfileType)
	}
	if policy.MinTLSVersion != tls.VersionTLS10 {
		t.Errorf("expected min TLS 1.0, got 0x%04x", policy.MinTLSVersion)
	}
	// Old profile should have more allowed ciphers than Intermediate
	intermediate := ResolveTLSProfile(nil)
	if len(policy.AllowedCipherIDs) <= len(intermediate.AllowedCipherIDs) {
		t.Errorf("expected Old to have more ciphers (%d) than Intermediate (%d)",
			len(policy.AllowedCipherIDs), len(intermediate.AllowedCipherIDs))
	}
}

func TestResolveTLSProfile_Modern(t *testing.T) {
	policy := ResolveTLSProfile(&configv1.TLSSecurityProfile{Type: configv1.TLSProfileModernType})
	if policy.ProfileType != "Modern" {
		t.Errorf("expected Modern profile, got %q", policy.ProfileType)
	}
	if policy.MinTLSVersion != tls.VersionTLS13 {
		t.Errorf("expected min TLS 1.3, got 0x%04x", policy.MinTLSVersion)
	}
	if len(policy.AllowedCipherIDs) != 0 {
		t.Errorf("expected no TLS 1.2 allowed ciphers for Modern, got %d", len(policy.AllowedCipherIDs))
	}
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
	if policy.ProfileType != "Custom" {
		t.Errorf("expected Custom profile, got %q", policy.ProfileType)
	}
	if policy.MinTLSVersion != tls.VersionTLS12 {
		t.Errorf("expected min TLS 1.2, got 0x%04x", policy.MinTLSVersion)
	}
	if len(policy.AllowedCipherIDs) != 2 {
		t.Errorf("expected 2 allowed ciphers, got %d", len(policy.AllowedCipherIDs))
	}
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
		got := versionBelow(tc.ver)
		if got != tc.expected {
			t.Errorf("versionBelow(0x%04x) = 0x%04x, want 0x%04x", tc.ver, got, tc.expected)
		}
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
			got := tlsVersionString(tc.ver)
			if got != tc.expected {
				t.Errorf("tlsVersionString(0x%04x) = %q, want %q", tc.ver, got, tc.expected)
			}
		})
	}
}

func TestComputeDisallowedCiphers(t *testing.T) {
	// Allow only 2 ciphers
	allowed := []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	}
	disallowed := computeDisallowedCiphers(allowed)

	// Should have all Go-supported ciphers minus the 2 allowed
	allCiphers := allGoTLS12CipherIDs()
	expectedLen := len(allCiphers) - 2
	if len(disallowed) != expectedLen {
		t.Errorf("expected %d disallowed ciphers, got %d", expectedLen, len(disallowed))
	}

	// Verify allowed ciphers are not in disallowed
	disallowedSet := make(map[uint16]bool)
	for _, id := range disallowed {
		disallowedSet[id] = true
	}
	for _, id := range allowed {
		if disallowedSet[id] {
			t.Errorf("allowed cipher 0x%04x found in disallowed set", id)
		}
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
		got := opensslVersionFlag(tc.ver)
		if got != tc.expected {
			t.Errorf("opensslVersionFlag(0x%04x) = %q, want %q", tc.ver, got, tc.expected)
		}
	}
}

func TestDefaultTLSPolicy(t *testing.T) {
	policy := DefaultTLSPolicy()
	if policy.ProfileType != "Intermediate" {
		t.Errorf("expected Intermediate, got %q", policy.ProfileType)
	}
	if policy.MinTLSVersion != tls.VersionTLS12 {
		t.Errorf("expected TLS 1.2, got 0x%04x", policy.MinTLSVersion)
	}
}

func parseHostPort(t *testing.T, addr string) (host string, port int32) {
	t.Helper()
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("failed to parse address %q: %v", addr, err)
	}
	// Port numbers from test listeners are always valid, no overflow risk
	portInt := 0
	for _, c := range p {
		portInt = portInt*10 + int(c-'0')
	}
	return h, int32(portInt)
}
