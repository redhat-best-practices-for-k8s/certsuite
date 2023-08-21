package tnf

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
)

// ClaimFilePrintf prints to claim and junit report files.
func ClaimFilePrintf(format string, args ...interface{}) {
	Log(logrus.TraceLevel, format, args)
}

// Log prints to stdout and claim and junit report files.
func Log(level logrus.Level, format string, args ...interface{}) {
	message := fmt.Sprintf(format+"\n", args...)
	_, err := ginkgo.GinkgoWriter.Write([]byte(message))
	if err != nil {
		logrus.Errorf("Ginkgo writer could not write msg '%s' because: %s", message, err)
	}

	logrus.StandardLogger().Log(level, message)
}
