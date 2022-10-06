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
package onlinecheck_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/internal/api/onlinecheck"
)

func TestIsContainerCertified(t *testing.T) {
	/*
		Notes for this unit test:
		See: https://catalog.redhat.com/api/containers/v1/repositories/registry/registry.connect.redhat.com/repository/rocketchat/rocketchat/tag/0.56.0-1
		We have found that there may be a mismatch between the docker_image_digest and the parsed.docker_image_digest.
		If this test fails, double check these values.
	*/

	client := onlinecheck.NewOnlineValidator()
	var v bool
	v = client.IsContainerCertified("registry.connect.redhat.com", "rocketchat/rocketchat", "", "")
	assert.Equal(t, true, v) // true
	v = client.IsContainerCertified("registry.connect.redhat.com", "rocketchat/rocketchat", "0.56.0-1", "")
	assert.Equal(t, true, v) // true

	v = client.IsContainerCertified("registry.connect.redhat.com", "rocketchat/rocketchat", "0.56.0-1", "sha256:b1d5b80d4c119c4316d9fa38a6a21383f30b07b67d8efc762530283a8d070070")
	assert.Equal(t, true, v) // true

	// wrong tag, valid digest, should be true
	v = client.IsContainerCertified("registry.connect.redhat.com", "rocketchat/rocketchat", "0.56.0-100", "sha256:b1d5b80d4c119c4316d9fa38a6a21383f30b07b67d8efc762530283a8d070070")
	assert.Equal(t, true, v) // true

	// wrong tag, everything else is valid
	v = client.IsContainerCertified("registry.connect.redhat.com", "rocketchat/rocketchat", "0.56.0-XX", "")
	assert.Equal(t, false, v) // false

	// wrong digest, everything else is valid
	v = client.IsContainerCertified("registry.connect.redhat.com", "rocketchat/rocketchat", "0.56.0-1", "sha256:c358eee360a1e7754c2d555ec5fba4e6a42f1ede2bc9dd9e59068dd287113b35")
	assert.Equal(t, false, v) // false
}
