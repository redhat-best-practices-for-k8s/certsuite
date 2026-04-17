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
	"net"
	"strconv"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
)

const (
	opensslCipherNone       = "Cipher is (NONE)"
	opensslAlertProtoVer    = "alert protocol version"
	opensslAlertHandshake   = "alert handshake failure"
	opensslHandshakeFailure = "handshake failure"
	opensslConnected        = "CONNECTED"
	opensslConnErrno        = "errno"
)

func extractOpenSSLCipher(stdout string) string {
	for line := range strings.SplitSeq(stdout, "\n") {
		trimmed := strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(trimmed, "Cipher    :"); ok {
			return strings.TrimSpace(after)
		}
		if after, ok := strings.CutPrefix(trimmed, "Cipher:"); ok {
			return strings.TrimSpace(after)
		}
	}
	return "unknown"
}

// IsPortTLS probes a single TCP port via openssl to determine whether it speaks TLS.
// It does not validate TLS version or cipher compliance — only whether TLS is present.
func IsPortTLS(ch clientsholder.Command, ctx clientsholder.Context, address string, port int32) (isTLS, reachable bool, reason string) {
	endpoint := net.JoinHostPort(address, strconv.Itoa(int(port)))
	cmd := fmt.Sprintf("echo | timeout 5 openssl s_client -connect %s 2>&1", endpoint)
	stdout, _, err := ch.ExecCommandContainer(ctx, cmd)

	if err != nil && !strings.Contains(stdout, opensslConnected) && !strings.Contains(stdout, opensslConnErrno) &&
		!strings.Contains(stdout, "SSL") && !strings.Contains(stdout, "Cipher") {
		return false, false, fmt.Sprintf("exec probe failed: %v", err)
	}

	if strings.Contains(stdout, "connect:errno=") || strings.Contains(stdout, "Connection refused") {
		return false, false, "port unreachable"
	}

	if !strings.Contains(stdout, opensslConnected) && !strings.Contains(stdout, opensslConnErrno) {
		return false, false, "no recognizable openssl output"
	}

	// TLS alerts indicate a genuine TLS server (even if it rejected our params)
	if strings.Contains(stdout, opensslAlertProtoVer) ||
		strings.Contains(stdout, opensslAlertHandshake) ||
		strings.Contains(stdout, opensslHandshakeFailure) {
		return true, true, "TLS server detected (handshake alert)"
	}

	// Successful TLS handshake: cipher is not NONE
	if !strings.Contains(stdout, opensslCipherNone) {
		if strings.Contains(stdout, "Cipher    :") || strings.Contains(stdout, "Cipher:") {
			cipher := extractOpenSSLCipher(stdout)
			if cipher != "" && cipher != "0000" && cipher != "(NONE)" {
				return true, true, fmt.Sprintf("TLS negotiated (cipher: %s)", cipher)
			}
		}
	}

	// Connected but no TLS indicators — plaintext service
	return false, true, "plaintext service (no TLS)"
}
