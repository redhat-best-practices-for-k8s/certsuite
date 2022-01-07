package diagnostic

import (
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"

	"github.com/onsi/ginkgo/v2"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf/testcases"
)

var _ = ginkgo.Describe(common.DiagnosticTestKey, func() {
	conf, _ := ginkgo.GinkgoConfiguration()
	if testcases.IsInFocus(conf.FocusStrings, common.DiagnosticTestKey) {
		logrus.Debug(common.DiagnosticTestKey, " not moved yet to new framework")
	}
})
