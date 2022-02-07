package diagnostic

import (
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe(common.DiagnosticTestKey, func() {
	logrus.Debugf("%s not moved yet to new framework", common.DiagnosticTestKey)
})
