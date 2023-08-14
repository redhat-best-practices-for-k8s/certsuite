// Copyright (C) 2020-2023 Red Hat, Inc.
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

package performance

import (
	"testing"

	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"gotest.tools/v3/assert"
)

func Test_appendToCompliantObject(t *testing.T) {
	testCases := []struct {
		compliantObjects       []*testhelper.ReportObject
		cut                    *provider.Container
		put                    *provider.Pod
		resultCompliantObjects []*testhelper.ReportObject
	}{
		{},
		{},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.resultCompliantObjects, appendToCompliantObject(tc.compliantObjects, tc.cut, tc.put))
	}
}

func Test_appendTononCompliantObject(t *testing.T) {
	testCases := []struct {
		noncompliantObjects       []*testhelper.ReportObject
		cut                       *provider.Container
		put                       *provider.Pod
		resultnonCompliantObjects []*testhelper.ReportObject
	}{
		{},
		{},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.resultnonCompliantObjects, appendToCompliantObject(tc.noncompliantObjects, tc.cut, tc.put))
	}
}
