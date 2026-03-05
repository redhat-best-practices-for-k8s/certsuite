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
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"

	goversion "github.com/hashicorp/go-version"
	configv1 "github.com/openshift/api/config/v1"
	clientconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	opensslFlagTLS12      = "-tls1_2"
	opensslProtoNameTLS12 = "TLSv1.2"
	truncateLen           = 200

	// openssl output patterns used to classify TLS probe results.
	opensslCipherNone       = "Cipher is (NONE)"
	opensslAlertProtoVer    = "alert protocol version"
	opensslAlertHandshake   = "alert handshake failure"
	opensslHandshakeFailure = "handshake failure"
	opensslNoCiphers        = "no ciphers available"
	opensslConnected        = "CONNECTED"
	opensslConnErrno        = "errno"

	// OCPTLSProfileEnforcementVersion is the minimum OCP version at which the
	// APIServer CR's TLS security profile is reliably enforced. On older versions,
	// we ignore the CR and fall back to the Intermediate profile (min TLS 1.2).
	OCPTLSProfileEnforcementVersion = "4.22"
)

// TLSPolicy holds the resolved effective TLS policy for the cluster.
type TLSPolicy struct {
	ProfileType      string
	MinTLSVersion    uint16   // Go crypto/tls constant (e.g. tls.VersionTLS12)
	AllowedCipherIDs []uint16 // Go cipher suite IDs allowed for TLS 1.2
}

// TLSProbeResult holds the outcome of a TLS probe against a single service port.
type TLSProbeResult struct {
	Compliant     bool
	IsTLS         bool
	Reachable     bool
	NegotiatedVer string
	Reason        string
}

// opensslToGoCipher maps OpenSSL cipher suite names to Go crypto/tls cipher suite IDs.
//
// The OpenSSL names come from the OpenShift TLS security profiles defined in:
//
//	https://github.com/openshift/api/blob/master/config/v1/types_tlssecurityprofile.go
//
// which follow Mozilla's Server Side TLS recommendations:
//
//	https://wiki.mozilla.org/Security/Server_Side_TLS
//
// The Go cipher suite constants are documented at:
//
//	https://pkg.go.dev/crypto/tls#pkg-constants
//
// DHE ciphers (e.g. DHE-RSA-AES128-GCM-SHA256) have no Go equivalent and are omitted;
// they are tested via the openssl exec probe path.
var opensslToGoCipher = map[string]uint16{
	"ECDHE-ECDSA-AES128-GCM-SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	"ECDHE-RSA-AES128-GCM-SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	"ECDHE-ECDSA-AES256-GCM-SHA384": tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	"ECDHE-RSA-AES256-GCM-SHA384":   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	"ECDHE-ECDSA-CHACHA20-POLY1305": tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	"ECDHE-RSA-CHACHA20-POLY1305":   tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
	"AES128-GCM-SHA256":             tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	"AES256-GCM-SHA384":             tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	"AES128-SHA256":                 tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
	"AES128-SHA":                    tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	"AES256-SHA":                    tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	"ECDHE-ECDSA-AES128-SHA256":     tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	"ECDHE-RSA-AES128-SHA256":       tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
	"ECDHE-ECDSA-AES128-SHA":        tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	"ECDHE-RSA-AES128-SHA":          tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	"ECDHE-ECDSA-AES256-SHA384":     tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA, // closest Go mapping
	"ECDHE-RSA-AES256-SHA384":       tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,   // closest Go mapping
	"ECDHE-ECDSA-AES256-SHA":        tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	"ECDHE-RSA-AES256-SHA":          tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	"DES-CBC3-SHA":                  tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
}

// tls13CipherNames lists TLS 1.3 cipher names from OpenShift profiles.
// TLS 1.3 ciphers are not configurable in Go (always enabled when TLS 1.3 is used)
// so we skip them when building allowed/disallowed sets.
var tls13CipherNames = map[string]bool{
	"TLS_AES_128_GCM_SHA256":       true,
	"TLS_AES_256_GCM_SHA384":       true,
	"TLS_CHACHA20_POLY1305_SHA256": true,
}

// ocpVersionToGoVersion converts OpenShift TLSProtocolVersion strings to Go tls constants.
func ocpVersionToGoVersion(v configv1.TLSProtocolVersion) uint16 {
	switch v {
	case configv1.VersionTLS10:
		return tls.VersionTLS10
	case configv1.VersionTLS11:
		return tls.VersionTLS11
	case configv1.VersionTLS12:
		return tls.VersionTLS12
	case configv1.VersionTLS13:
		return tls.VersionTLS13
	default:
		return tls.VersionTLS12 // safe default
	}
}

// IsOCPVersionAtLeast returns true if ocpVersion is >= minVersion.
// Returns false if either version string is empty or cannot be parsed.
func IsOCPVersionAtLeast(ocpVersion, minVersion string) bool {
	if ocpVersion == "" || minVersion == "" {
		return false
	}

	current, err := goversion.NewVersion(ocpVersion)
	if err != nil {
		return false
	}

	minimum, err := goversion.NewVersion(minVersion)
	if err != nil {
		return false
	}

	return current.GreaterThanOrEqual(minimum)
}

// DefaultTLSPolicy returns the Intermediate profile (OpenShift default).
func DefaultTLSPolicy() TLSPolicy {
	return ResolveTLSProfile(nil)
}

// GetClusterTLSPolicy reads the cluster's TLS security profile from the APIServer CR
// on OpenShift, or falls back to the Intermediate default on non-OCP clusters.
// On OCP versions below OCPTLSProfileEnforcementVersion, the APIServer CR is ignored
// because TLS profile enforcement is not reliable on those versions.
func GetClusterTLSPolicy(ocpClient clientconfigv1.ConfigV1Interface, isOCP bool, ocpVersion string) TLSPolicy {
	if !isOCP {
		return DefaultTLSPolicy()
	}

	if !IsOCPVersionAtLeast(ocpVersion, OCPTLSProfileEnforcementVersion) {
		return DefaultTLSPolicy()
	}

	apiServer, err := ocpClient.APIServers().Get(context.TODO(), "cluster", metav1.GetOptions{})
	if err != nil {
		return DefaultTLSPolicy()
	}

	return ResolveTLSProfile(apiServer.Spec.TLSSecurityProfile)
}

// ResolveTLSProfile converts an OpenShift TLSSecurityProfile into a TLSPolicy.
// A nil profile resolves to the Intermediate profile (the OpenShift default).
func ResolveTLSProfile(profile *configv1.TLSSecurityProfile) TLSPolicy {
	if profile == nil {
		return resolveTLSProfileSpec(string(configv1.TLSProfileIntermediateType),
			configv1.TLSProfiles[configv1.TLSProfileIntermediateType])
	}

	switch profile.Type {
	// Old: most permissive profile. Min TLS 1.0, allows 28 ciphers including
	// legacy suites (DES-CBC3-SHA, RC4, SHA-1 based). Intended for backward
	// compatibility with very old clients.
	case configv1.TLSProfileOldType:
		return resolveTLSProfileSpec(string(configv1.TLSProfileOldType),
			configv1.TLSProfiles[configv1.TLSProfileOldType])

	// Modern: most restrictive profile. Min TLS 1.3, no TLS 1.2 ciphers
	// (only the 3 mandatory TLS 1.3 suites). Provides the strongest security
	// but drops support for all TLS 1.2 clients.
	case configv1.TLSProfileModernType:
		return resolveTLSProfileSpec(string(configv1.TLSProfileModernType),
			configv1.TLSProfiles[configv1.TLSProfileModernType])

	// Custom: user-defined profile with explicit min TLS version and cipher
	// list. Falls back to Intermediate if the Custom spec is nil.
	case configv1.TLSProfileCustomType:
		if profile.Custom != nil {
			spec := &profile.Custom.TLSProfileSpec
			return resolveTLSProfileSpec(string(configv1.TLSProfileCustomType), spec)
		}
		return resolveTLSProfileSpec(string(configv1.TLSProfileIntermediateType),
			configv1.TLSProfiles[configv1.TLSProfileIntermediateType])

	// Intermediate (default): balanced profile. Min TLS 1.2, allows 8 strong
	// TLS 1.2 ciphers (ECDHE/DHE with GCM or ChaCha20-Poly1305) plus the
	// 3 TLS 1.3 suites. This is the OpenShift default and the fallback for
	// unrecognized profile types.
	default:
		return resolveTLSProfileSpec(string(configv1.TLSProfileIntermediateType),
			configv1.TLSProfiles[configv1.TLSProfileIntermediateType])
	}
}

// resolveTLSProfileSpec converts an OpenShift TLSProfileSpec (profile type name +
// min version + cipher list) into our internal TLSPolicy. It maps each OpenSSL cipher
// name to its Go constant, skipping TLS 1.3 ciphers (not configurable in Go) and DHE
// ciphers (no Go equivalent).
func resolveTLSProfileSpec(profileType string, spec *configv1.TLSProfileSpec) TLSPolicy {
	policy := TLSPolicy{
		ProfileType:   profileType,
		MinTLSVersion: ocpVersionToGoVersion(spec.MinTLSVersion),
	}

	for _, cipherName := range spec.Ciphers {
		if tls13CipherNames[cipherName] {
			continue // TLS 1.3 ciphers are handled by the protocol, not configurable
		}
		if id, ok := opensslToGoCipher[cipherName]; ok {
			policy.AllowedCipherIDs = append(policy.AllowedCipherIDs, id)
		}
	}

	return policy
}

// versionBelow returns the TLS version one step below the given version,
// or 0 if there is no lower version to test.
func versionBelow(ver uint16) uint16 {
	switch ver {
	case tls.VersionTLS13:
		return tls.VersionTLS12
	case tls.VersionTLS12:
		return tls.VersionTLS11
	case tls.VersionTLS11:
		return tls.VersionTLS10
	default:
		return 0
	}
}

// opensslHandshakeRejected returns true if the openssl output indicates
// a rejected TLS handshake (cipher none, protocol version alert, or failure).
func opensslHandshakeRejected(output string) bool {
	return strings.Contains(output, opensslCipherNone) ||
		strings.Contains(output, opensslAlertProtoVer) ||
		strings.Contains(output, opensslHandshakeFailure)
}

// ProbeServicePortViaExec runs openssl s_client inside a probe pod to test
// TLS compliance against the given policy.
func ProbeServicePortViaExec(ch clientsholder.Command, ctx clientsholder.Context, address string, port int32, policy TLSPolicy) TLSProbeResult {
	endpoint := net.JoinHostPort(address, strconv.Itoa(int(port)))

	// Try connecting at the minimum version
	opensslVerFlag := opensslVersionFlag(policy.MinTLSVersion)
	cmdMin := fmt.Sprintf("echo | timeout 5 openssl s_client -connect %s %s 2>&1", endpoint, opensslVerFlag)
	stdoutMin, _, errMin := ch.ExecCommandContainer(ctx, cmdMin)

	// openssl exits non-zero on handshake failure, which causes ExecCommandContainer
	// to return an error. The stdout still contains useful output (CONNECTED, Protocol,
	// Cipher, etc.) that we need to parse. Only treat as truly unreachable if both
	// the error is set AND stdout has no recognizable openssl output.
	if errMin != nil && !strings.Contains(stdoutMin, opensslConnected) && !strings.Contains(stdoutMin, opensslConnErrno) &&
		!strings.Contains(stdoutMin, "SSL") && !strings.Contains(stdoutMin, "Cipher") {
		return TLSProbeResult{
			Compliant: true,
			IsTLS:     false,
			Reachable: false,
			Reason:    fmt.Sprintf("exec probe failed: %v (stdout=%s)", errMin, truncate(stdoutMin, truncateLen)),
		}
	}

	// Check for rejection indicators BEFORE parsing the Protocol line.
	// openssl reports the *attempted* protocol version even when the handshake fails,
	// so "Protocol: TLSv1.2" appears in output even if the server rejected TLS 1.2.
	// "Cipher is (NONE)" or "alert protocol version" means the handshake did not complete.
	rejected := opensslHandshakeRejected(stdoutMin)

	minVerOK := false
	if !rejected {
		minVerOK = strings.Contains(stdoutMin, fmt.Sprintf("Protocol  : %s", opensslVersionName(policy.MinTLSVersion))) ||
			strings.Contains(stdoutMin, fmt.Sprintf("Protocol: %s", opensslVersionName(policy.MinTLSVersion)))
		// Also check for TLS 1.3 if min is lower (server may negotiate higher)
		if !minVerOK && policy.MinTLSVersion < tls.VersionTLS13 {
			minVerOK = strings.Contains(stdoutMin, "Protocol  : TLSv1.3") ||
				strings.Contains(stdoutMin, "Protocol: TLSv1.3")
		}
	}

	if !minVerOK {
		return classifyExecNoMinVersion(stdoutMin, policy)
	}

	// Check if server accepts versions below minimum
	belowVer := versionBelow(policy.MinTLSVersion)
	if belowVer > 0 {
		result := probeExecBelow(ch, ctx, endpoint, belowVer, policy)
		if !result.Compliant {
			return result
		}
	}

	// Check cipher compliance via exec
	cipherResult := probeExecCipherCompliance(ch, ctx, endpoint, policy)
	if cipherResult != nil {
		return *cipherResult
	}

	return TLSProbeResult{
		Compliant:     true,
		IsTLS:         true,
		Reachable:     true,
		NegotiatedVer: tlsVersionString(policy.MinTLSVersion),
		Reason:        fmt.Sprintf("server honors %s profile (via exec probe)", policy.ProfileType),
	}
}

// classifyExecNoMinVersion categorizes an openssl exec probe result when the server
// did not negotiate the policy's minimum version successfully.
//
// This function distinguishes between:
//   - Port unreachable (connection refused, no route)
//   - TLS server that rejected our version (has TLS alert in output)
//   - Non-TLS service (openssl connected but got no TLS response)
//   - Unrecognizable output (openssl not installed, timeout, etc.)
//
// IMPORTANT: Do NOT use Protocol lines or "SSL handshake" as TLS indicators.
// openssl always prints "Protocol  : TLSv1.2" (the *attempted* version) and
// "SSL handshake has read N bytes" in its output regardless of whether the
// server actually speaks TLS. The only reliable TLS indicator is the presence
// of a TLS alert message (e.g. "alert protocol version", "handshake failure").
func classifyExecNoMinVersion(stdout string, policy TLSPolicy) TLSProbeResult {
	// Connection refused or unreachable
	if strings.Contains(stdout, "connect:errno=") || strings.Contains(stdout, "Connection refused") {
		return TLSProbeResult{
			Compliant: true,
			IsTLS:     false,
			Reachable: false,
			Reason:    "port unreachable via exec probe",
		}
	}

	// TLS alert indicators are conclusive evidence of a genuine TLS server
	// that rejected our protocol version. Plain HTTP/gRPC services never
	// generate TLS alerts — they produce errors like "packet length too long"
	// or "record layer failure" instead.
	if strings.Contains(stdout, opensslAlertProtoVer) ||
		strings.Contains(stdout, opensslAlertHandshake) ||
		strings.Contains(stdout, opensslHandshakeFailure) {
		return TLSProbeResult{
			Compliant:     false,
			IsTLS:         true,
			Reachable:     true,
			NegotiatedVer: fmt.Sprintf("< %s", tlsVersionString(policy.MinTLSVersion)),
			Reason:        fmt.Sprintf("server does not support %s via exec probe", tlsVersionString(policy.MinTLSVersion)),
		}
	}

	// If stdout is empty or doesn't contain any recognizable openssl output,
	// this likely means openssl is not installed or the command failed to run.
	if !strings.Contains(stdout, opensslConnected) && !strings.Contains(stdout, opensslConnErrno) {
		return TLSProbeResult{
			Compliant: true,
			IsTLS:     false,
			Reachable: false,
			Reason:    fmt.Sprintf("exec probe produced no recognizable output (stdout=%s)", truncate(stdout, truncateLen)),
		}
	}

	// Connected but no TLS alerts — non-TLS service (plain HTTP, gRPC, etc.)
	// The output may contain "Cipher is (NONE)", "packet length too long",
	// "record layer failure", "write:errno=", or "SSL handshake has read N bytes",
	// but none of these indicate TLS — they're artifacts of openssl trying to
	// parse non-TLS responses.
	return TLSProbeResult{
		Compliant: true,
		IsTLS:     false,
		Reachable: true,
		Reason:    "non-TLS service (informational, via exec probe)",
	}
}

// probeExecBelow is the exec-based version check. It runs openssl s_client
// with a version flag below the policy's minimum to check if the server accepts a
// downgraded TLS version. Looks for rejection indicators like "Cipher is (NONE)" or
// "alert protocol version" to determine if the server correctly refused the connection.
func probeExecBelow(ch clientsholder.Command, ctx clientsholder.Context, endpoint string, belowVer uint16, policy TLSPolicy) TLSProbeResult {
	belowFlag := opensslVersionFlag(belowVer)
	cmdBelow := fmt.Sprintf("echo | timeout 5 openssl s_client -connect %s %s 2>&1", endpoint, belowFlag)
	stdoutBelow, _, _ := ch.ExecCommandContainer(ctx, cmdBelow)

	// Check for rejection indicators
	belowRejected := opensslHandshakeRejected(stdoutBelow)

	if belowRejected {
		return TLSProbeResult{
			Compliant:     true,
			IsTLS:         true,
			Reachable:     true,
			NegotiatedVer: tlsVersionString(policy.MinTLSVersion),
			Reason:        fmt.Sprintf("server enforces %s minimum (via exec probe)", tlsVersionString(policy.MinTLSVersion)),
		}
	}

	belowVerName := opensslVersionName(belowVer)
	belowOK := strings.Contains(stdoutBelow, fmt.Sprintf("Protocol  : %s", belowVerName)) ||
		strings.Contains(stdoutBelow, fmt.Sprintf("Protocol: %s", belowVerName))

	if belowOK {
		return TLSProbeResult{
			Compliant:     false,
			IsTLS:         true,
			Reachable:     true,
			NegotiatedVer: tlsVersionString(belowVer),
			Reason:        fmt.Sprintf("server accepts %s (%s minimum required, via exec probe)", tlsVersionString(belowVer), tlsVersionString(policy.MinTLSVersion)),
		}
	}

	return TLSProbeResult{
		Compliant:     true,
		IsTLS:         true,
		Reachable:     true,
		NegotiatedVer: tlsVersionString(policy.MinTLSVersion),
		Reason:        fmt.Sprintf("server enforces %s minimum (via exec probe)", tlsVersionString(policy.MinTLSVersion)),
	}
}

// probeExecCipherCompliance checks cipher compliance via openssl exec.
// Returns nil if compliant, or a non-compliant result.
func probeExecCipherCompliance(ch clientsholder.Command, ctx clientsholder.Context, endpoint string, policy TLSPolicy) *TLSProbeResult {
	if policy.MinTLSVersion > tls.VersionTLS12 {
		return nil // TLS 1.3 ciphers are not configurable, skip check
	}

	disallowedNames := computeDisallowedOpenSSLCiphers(policy)
	if len(disallowedNames) == 0 {
		return nil
	}

	cipherStr := strings.Join(disallowedNames, ":")
	cmd := fmt.Sprintf("echo | timeout 5 openssl s_client -connect %s -cipher %s -tls1_2 2>&1", endpoint, cipherStr)
	stdout, _, _ := ch.ExecCommandContainer(ctx, cmd)

	// Check for rejection
	rejected := opensslHandshakeRejected(stdout) ||
		strings.Contains(stdout, opensslNoCiphers) ||
		strings.Contains(stdout, opensslAlertHandshake)

	if rejected {
		return nil // compliant
	}

	// Check if a cipher was negotiated
	if strings.Contains(stdout, "Cipher    :") || strings.Contains(stdout, "Cipher:") {
		// Extract cipher name from output
		cipherName := extractOpenSSLCipher(stdout)
		result := TLSProbeResult{
			Compliant:     false,
			IsTLS:         true,
			Reachable:     true,
			NegotiatedVer: "TLS 1.2",
			Reason:        fmt.Sprintf("server accepted disallowed cipher %s (not in %s profile, via exec probe)", cipherName, policy.ProfileType),
		}
		return &result
	}

	return nil
}

// computeDisallowedOpenSSLCiphers returns OpenSSL cipher names that are NOT in the policy's allowed list.
func computeDisallowedOpenSSLCiphers(policy TLSPolicy) []string {
	// Build set of allowed OpenSSL names from the profile
	allowedSet := make(map[string]bool)
	profileSpec := getProfileSpec(policy)
	if profileSpec != nil {
		for _, name := range profileSpec.Ciphers {
			if !tls13CipherNames[name] {
				allowedSet[name] = true
			}
		}
	}

	var disallowed []string
	for name := range opensslToGoCipher {
		if !allowedSet[name] {
			disallowed = append(disallowed, name)
		}
	}
	return disallowed
}

// getProfileSpec returns the TLSProfileSpec for the policy's profile type.
func getProfileSpec(policy TLSPolicy) *configv1.TLSProfileSpec {
	switch policy.ProfileType {
	case string(configv1.TLSProfileOldType):
		return configv1.TLSProfiles[configv1.TLSProfileOldType]
	case string(configv1.TLSProfileIntermediateType):
		return configv1.TLSProfiles[configv1.TLSProfileIntermediateType]
	case string(configv1.TLSProfileModernType):
		return configv1.TLSProfiles[configv1.TLSProfileModernType]
	default:
		return configv1.TLSProfiles[configv1.TLSProfileIntermediateType]
	}
}

// extractOpenSSLCipher parses the "Cipher    :" or "Cipher:" line from openssl
// s_client output and returns the negotiated cipher name.
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

// opensslVersionFlag converts a Go TLS version constant to the corresponding
// openssl s_client command-line flag (e.g. tls.VersionTLS12 -> "-tls1_2").
func opensslVersionFlag(ver uint16) string {
	switch ver {
	case tls.VersionTLS10:
		return "-tls1"
	case tls.VersionTLS11:
		return "-tls1_1"
	case tls.VersionTLS12:
		return opensslFlagTLS12
	case tls.VersionTLS13:
		return "-tls1_3"
	default:
		return opensslFlagTLS12
	}
}

// opensslVersionName converts a Go TLS version constant to the protocol name
// as it appears in openssl s_client output (e.g. tls.VersionTLS12 -> "TLSv1.2").
func opensslVersionName(ver uint16) string {
	switch ver {
	case tls.VersionTLS10:
		return "TLSv1"
	case tls.VersionTLS11:
		return "TLSv1.1"
	case tls.VersionTLS12:
		return opensslProtoNameTLS12
	case tls.VersionTLS13:
		return "TLSv1.3"
	default:
		return opensslProtoNameTLS12
	}
}

// CheckServiceTLSCompliance iterates env.Services and all their TCP ports,
// probes each for TLS compliance against the given policy via a probe pod,
// and returns report objects.
func CheckServiceTLSCompliance(check *checksdb.Check, env *provider.TestEnvironment, policy TLSPolicy) (
	compliant, nonCompliant []*testhelper.ReportObject) {
	for _, svc := range env.Services {
		// Skip headless services
		if svc.Spec.ClusterIP == "" || svc.Spec.ClusterIP == "None" {
			check.LogInfo("Skipping headless service %q in namespace %q (no ClusterIP)", svc.Name, svc.Namespace)
			continue
		}

		for _, port := range svc.Spec.Ports {
			// Skip non-TCP ports
			if port.Protocol != corev1.ProtocolTCP {
				check.LogInfo("Skipping non-TCP port %d/%s on service %q", port.Port, port.Protocol, svc.Name)
				continue
			}

			check.LogInfo("Probing service %q port %d in namespace %q (profile: %s, min: %s)",
				svc.Name, port.Port, svc.Namespace, policy.ProfileType, tlsVersionString(policy.MinTLSVersion))

			result := probeServicePortViaExecPod(check, env, svc.Spec.ClusterIP, port.Port, policy)

			// If no probe pods are available, treat as compliant (informational).
			if result == nil {
				check.LogInfo("No probe pods available to test service %q port %d, treating as compliant (informational)",
					svc.Name, port.Port)
				ro := buildReportObject(svc, port.Port, TLSProbeResult{
					Compliant: true,
					Reason:    "no probe pods available to test TLS compliance",
				})
				compliant = append(compliant, ro)
				continue
			}

			check.LogInfo("Exec probe result for %s:%d: compliant=%v, isTLS=%v, reachable=%v, reason=%q",
				svc.Spec.ClusterIP, port.Port, result.Compliant, result.IsTLS, result.Reachable, result.Reason)

			// If the probe could not reach the service, treat it as compliant
			// (informational) — the port isn't listening or is unreachable,
			// so there's no TLS to validate.
			if !result.Reachable {
				check.LogInfo("Service %q port %d unreachable (%s), treating as compliant (informational)",
					svc.Name, port.Port, result.Reason)
				result.Compliant = true
			}

			ro := buildReportObject(svc, port.Port, *result)
			if result.Compliant {
				compliant = append(compliant, ro)
			} else {
				nonCompliant = append(nonCompliant, ro)
			}
		}
	}

	return compliant, nonCompliant
}

// probeServicePortViaExecPod probes a service port from inside a probe pod using
// openssl s_client. It finds the first available probe pod and runs
// ProbeServicePortViaExec through it. Returns nil if no probe pods are available.
func probeServicePortViaExecPod(check *checksdb.Check, env *provider.TestEnvironment, address string, port int32, policy TLSPolicy) *TLSProbeResult {
	for nodeName, probePod := range env.ProbePods {
		if probePod == nil || len(probePod.Spec.Containers) == 0 {
			continue
		}
		check.LogInfo("Using probe pod on node %q for exec probe to %s:%d", nodeName, address, port)
		ch := clientsholder.GetClientsHolder()
		ctx := clientsholder.NewContext(
			probePod.Namespace,
			probePod.Name,
			probePod.Spec.Containers[0].Name,
		)
		result := ProbeServicePortViaExec(ch, ctx, address, port, policy)
		return &result
	}
	return nil
}

// buildReportObject constructs a ReportObject from a TLS probe result, populating
// it with the service metadata (namespace, name, port) and negotiated TLS version.
func buildReportObject(svc *corev1.Service, port int32, result TLSProbeResult) *testhelper.ReportObject {
	ro := testhelper.NewReportObject(result.Reason, testhelper.ServiceType, result.Compliant)
	ro.AddField(testhelper.Namespace, svc.Namespace)
	ro.AddField(testhelper.ServiceName, svc.Name)
	ro.AddField(testhelper.PortNumber, strconv.Itoa(int(port)))
	ro.AddField(testhelper.PortProtocol, "TCP")
	if result.NegotiatedVer != "" {
		ro.AddField(testhelper.TLSVersion, result.NegotiatedVer)
	}
	return ro
}

// TLSVersionName returns a human-readable name for a TLS version constant.
func TLSVersionName(ver uint16) string {
	return tlsVersionString(ver)
}

// tlsVersionString converts a Go TLS version constant to a human-readable string
// (e.g. tls.VersionTLS12 -> "TLS 1.2").
func tlsVersionString(ver uint16) string {
	switch ver {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("unknown (0x%04x)", ver)
	}
}

// truncate returns the first n bytes of s, appending "..." if truncated.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
