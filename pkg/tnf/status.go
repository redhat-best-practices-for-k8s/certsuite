package tnf

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// ClaimFilePrintf prints to claim and junit report files.
func ClaimFilePrintf(format string, args ...interface{}) {
	Logf(logrus.TraceLevel, format, args)
}

// Logf prints to stdout and claim and junit report files.
func Logf(level logrus.Level, format string, args ...interface{}) {
	message := fmt.Sprintf(format+"\n", args...)
	logrus.StandardLogger().Log(level, message)
	// _, err := ginkgo.GinkgoWriter.Write([]byte(message))
	// if err != nil {
	// 	logrus.Errorf("Ginkgo writer could not write msg '%s' because: %s", message, err)
	// }
}
