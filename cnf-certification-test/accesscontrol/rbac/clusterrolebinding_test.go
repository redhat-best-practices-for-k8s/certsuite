// Copyright (C) 2022 Red Hat, Inc.
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

package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
)

func TestGetClusterRoleBinding(t *testing.T) {
	rb := NewClusterRoleBindingTester("testCR", "podNS", clientsholder.GetTestClientsHolder(buildTestObjects()))
	assert.NotNil(t, rb)
	gatheredCRBs, err := rb.GetClusterRoleBindings()
	assert.Nil(t, err)
	assert.Equal(t, "testNS:testCR", gatheredCRBs[0])
}
