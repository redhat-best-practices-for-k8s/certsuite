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
	"path"
	"runtime"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// CuratedLogLines
type CuratedLogLines struct {
	Lines []string
}

// AddLogLine checks a slice for a given string.
func (list CuratedLogLines) AddLogLine(format string, args ...interface{}) CuratedLogLines {
	message := fmt.Sprintf(format+"\n", args...)
	list.Lines = append(list.Lines, message)
	logrus.Debug(message)
	return list
}

func (list CuratedLogLines) GetLogLines() []string {
	return list.Lines
}

// SetLogFormat sets the log format for logrus
func SetLogFormat() {
	logrus.Info("debug format initialization: start")
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = time.StampMilli
	customFormatter.PadLevelText = true
	customFormatter.FullTimestamp = true
	customFormatter.ForceColors = true
	logrus.SetReportCaller(true)
	customFormatter.CallerPrettyfier = func(f *runtime.Frame) (string, string) {
		_, filename := path.Split(f.File)
		return strconv.Itoa(f.Line) + "]", fmt.Sprintf("[%s:", filename)
	}
	logrus.SetFormatter(customFormatter)
	logrus.Info("debug format initialization: done")
	logrus.SetLevel(logrus.TraceLevel)
}
