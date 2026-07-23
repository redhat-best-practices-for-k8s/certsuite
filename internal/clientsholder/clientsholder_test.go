// Copyright (C) 2020-2022 Red Hat, Inc.
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

package clientsholder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
)

func TestNewContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		namespace string
		pod       string
		container string
	}{
		{
			name:      "valid inputs",
			namespace: "test-ns",
			pod:       "test-pod",
			container: "test-container",
		},
		{
			name:      "empty strings",
			namespace: "",
			pod:       "",
			container: "",
		},
		{
			name:      "namespace only",
			namespace: "my-namespace",
			pod:       "",
			container: "",
		},
		{
			name:      "special characters",
			namespace: "ns-with-dashes",
			pod:       "pod.with.dots",
			container: "container_with_underscores",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := NewContext(tt.namespace, tt.pod, tt.container)
			assert.Equal(t, tt.namespace, ctx.GetNamespace())
			assert.Equal(t, tt.pod, ctx.GetPodName())
			assert.Equal(t, tt.container, ctx.GetContainerName())
		})
	}
}

func TestGetClientTimeout(t *testing.T) {
	tests := []struct {
		name     string
		envVal   string
		setEnv   bool
		expected time.Duration
	}{
		{
			name:     "env var unset returns default",
			setEnv:   false,
			expected: DefaultTimeout,
		},
		{
			name:     "valid duration 30s",
			envVal:   "30s",
			setEnv:   true,
			expected: 30 * time.Second,
		},
		{
			name:     "valid duration 2m",
			envVal:   "2m",
			setEnv:   true,
			expected: 2 * time.Minute,
		},
		{
			name:     "invalid duration falls back to default",
			envVal:   "not-a-duration",
			setEnv:   true,
			expected: DefaultTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(clientTimeoutEnvVar, tt.envVal)
			}
			result := getClientTimeout()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetClientConfigFromRestConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		restConfig *rest.Config
		wantHost   string
		wantToken  string
		wantCAFile string
	}{
		{
			name: "config with host and token",
			restConfig: &rest.Config{
				Host:        "https://api.example.com:6443",
				BearerToken: "my-token-123",
			},
			wantHost:  "https://api.example.com:6443",
			wantToken: "my-token-123",
		},
		{
			name: "config with cert data",
			restConfig: &rest.Config{
				Host: "https://api.cluster.local:6443",
				TLSClientConfig: rest.TLSClientConfig{
					CAFile: "/var/run/secrets/ca.crt",
				},
			},
			wantHost:   "https://api.cluster.local:6443",
			wantCAFile: "/var/run/secrets/ca.crt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			config := GetClientConfigFromRestConfig(tt.restConfig)

			require.NotNil(t, config)
			assert.Equal(t, "Config", config.Kind)
			assert.Equal(t, "v1", config.APIVersion)
			assert.Equal(t, defaultContext, config.CurrentContext)

			cluster, ok := config.Clusters[defaultCluster]
			require.True(t, ok)
			assert.Equal(t, tt.wantHost, cluster.Server)

			if tt.wantCAFile != "" {
				assert.Equal(t, tt.wantCAFile, cluster.CertificateAuthority)
			}

			authInfo, ok := config.AuthInfos[defaultUser]
			require.True(t, ok)
			assert.Equal(t, tt.wantToken, authInfo.Token)

			ctx, ok := config.Contexts[defaultContext]
			require.True(t, ok)
			assert.Equal(t, defaultCluster, ctx.Cluster)
			assert.Equal(t, defaultUser, ctx.AuthInfo)
		})
	}
}
