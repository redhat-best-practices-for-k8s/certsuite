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

package provider

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUID(t *testing.T) {
	testCases := []struct {
		testCID     string
		expectedErr error
		expectedUID string
	}{
		{
			testCID:     "cid://testing",
			expectedErr: nil,
			expectedUID: "testing",
		},
		{
			testCID:     "cid://",
			expectedErr: errors.New("cannot determine container UID"),
			expectedUID: "",
		},
	}

	for _, tc := range testCases {
		c := GetContainer()
		c.Status.ContainerID = tc.testCID
		uid, err := c.GetUID()
		assert.Equal(t, tc.expectedErr, err)
		assert.Equal(t, tc.expectedUID, uid)
	}
}
