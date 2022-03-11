package tnf

import (
	"fmt"
	"os"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
)

// ClaimFilePrintf prints to claim and junit report files.
func ClaimFilePrintf(format string, args ...interface{}) {
	message := fmt.Sprintf(format+"\n", args...)
	_, err := ginkgo.GinkgoWriter.Write([]byte(message))
	if err != nil {
		logrus.Errorf("Ginkgo writer could not write msg '%s' because: %s", message, err)
	} else {
		logrus.Trace(message)
	}
}

func IsUnitTest() bool {
	//nolint:goconst
	return strings.Contains(os.Args[1], "-test.") || strings.Contains(os.Args[0], ".test") || os.Getenv("UNIT_TEST") == "true"
}
