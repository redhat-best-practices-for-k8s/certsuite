// Copyright (C) 2023-2026 Red Hat, Inc.
package results

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateFormField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		fieldName  string
		fieldValue string
	}{
		{
			name:       "valid field creation",
			fieldName:  "type",
			fieldValue: "RhocpBestPracticeTestResult",
		},
		{
			name:       "field with numeric value",
			fieldName:  "certId",
			fieldValue: "12345",
		},
		{
			name:       "empty value",
			fieldName:  "description",
			fieldValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			err := createFormField(writer, tt.fieldName, tt.fieldValue)
			require.NoError(t, err)

			err = writer.Close()
			require.NoError(t, err)

			reader := multipart.NewReader(&buf, writer.Boundary())
			form, err := reader.ReadForm(1 << 20)
			require.NoError(t, err)

			values, ok := form.Value[tt.fieldName]
			require.True(t, ok)
			require.Len(t, values, 1)
			assert.Equal(t, tt.fieldValue, values[0])
		})
	}
}

func TestSetProxy(t *testing.T) {
	tests := []struct {
		name        string
		proxyURL    string
		proxyPort   string
		expectProxy bool
	}{
		{
			name:        "valid proxy URL and port",
			proxyURL:    "http://proxy.example.com",
			proxyPort:   "8080",
			expectProxy: true,
		},
		{
			name:        "empty URL skips proxy",
			proxyURL:    "",
			proxyPort:   "8080",
			expectProxy: false,
		},
		{
			name:        "empty port skips proxy",
			proxyURL:    "http://proxy.example.com",
			proxyPort:   "",
			expectProxy: false,
		},
		{
			name:        "both empty skips proxy",
			proxyURL:    "",
			proxyPort:   "",
			expectProxy: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{}
			setProxy(client, tt.proxyURL, tt.proxyPort)

			if tt.expectProxy {
				require.NotNil(t, client.Transport)
				transport, ok := client.Transport.(*http.Transport)
				require.True(t, ok)
				assert.NotNil(t, transport.Proxy)
			} else {
				assert.Nil(t, client.Transport)
			}
		})
	}
}
