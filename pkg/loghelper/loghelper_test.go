// Copyright (C) 2020-2021 Red Hat, Inc.
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

package loghelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogLines(t *testing.T) {
	SetLogFormat()
	ll := CuratedLogLines{}
	ll = ll.Init("one", "two", "three")
	assert.Equal(t, []string{"one", "two", "three"}, ll.GetLogLines())
	ll = ll.AddLogLine("four") // adds a newline
	assert.Equal(t, []string{"one", "two", "three", "four\n"}, ll.GetLogLines())
}
