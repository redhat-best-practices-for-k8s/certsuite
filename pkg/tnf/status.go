package tnf

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// ClaimFilePrintf prints to stdout.
// ToDo: Remove?
func ClaimFilePrintf(format string, args ...interface{}) {
	Logf(logrus.TraceLevel, format, args...)
}

// Logf prints to stdout.
func Logf(level logrus.Level, format string, args ...interface{}) {
	message := fmt.Sprintf(format+"\n", args...)
	logrus.StandardLogger().Log(level, message)
}
