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

package loghelper

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
)

// CuratedLogLines
type CuratedLogLines struct {
	lines []string
}

// AddLogLine checks a slice for a given string.
//
//nolint:goprintffuncname
func (list *CuratedLogLines) AddLogLine(format string, args ...interface{}) {
	message := fmt.Sprintf(format+"\n", args...)
	list.lines = append(list.lines, message)
	logrus.Debug(message)
}

// Init checks a slice for a given string.
func (list *CuratedLogLines) Init(lines ...string) {
	list.lines = append(list.lines, lines...)
}

func (list *CuratedLogLines) GetLogLines() []string {
	return list.lines
}

// SetLogFormat sets the log format for logrus
func SetLogFormat() {
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
}

func SetHooks(f *os.File) {
	logrus.AddHook(&writer.Hook{ // Send all logs to file
		Writer: f,
		LogLevels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
			logrus.InfoLevel,
			logrus.DebugLevel,
			logrus.TraceLevel,
		},
	})
}

// setLogLevel sets the log level for logrus based on the "TNF_LOG_LEVEL" environment variable
func SetLogLevel() {
	params := configuration.GetTestParameters()

	var logLevel, err = logrus.ParseLevel(params.LogLevel)
	if err != nil {
		logrus.Error("TNF_LOG_LEVEL environment set with an invalid value, defaulting to DEBUG \n Valid values are:  trace, debug, info, warn, error, fatal, panic")
		logLevel = logrus.DebugLevel
	}

	logrus.Info("Log level set to: ", logLevel)
	logrus.SetLevel(logLevel)
}
