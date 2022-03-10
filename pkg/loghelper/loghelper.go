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
	"fmt"

	"github.com/sirupsen/logrus"
)

// CuratedLogLines
type CuratedLogLines struct {
	lines []string
}

// AddLogLine checks a slice for a given string.
func (list CuratedLogLines) AddLogLine(format string, args ...interface{}) CuratedLogLines {
	message := fmt.Sprintf(format+"\n", args...)
	list.lines = append(list.lines, message)
	logrus.Debug(message)
	return list
}
func (list CuratedLogLines) Append(newLines CuratedLogLines) CuratedLogLines {
	list.lines = append(list.lines, newLines.lines...)
	return list
}
