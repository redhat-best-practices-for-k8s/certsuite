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

package lifecycle

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf/testcases"
)

//
// All actual test code belongs below here.  Utilities belong above.
//
var _ = Describe(common.LifecycleTestKey, func() {
	var env provider.TestEnvironment
	BeforeEach(func() {
		provider.BuildTestEnvironment()
		env = provider.GetTestEnvironment()
	})
	conf, _ := GinkgoConfiguration()
	if testcases.IsInFocus(conf.FocusStrings, common.LifecycleTestKey) {
		testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestShudtownIdentifier)
		It(testID, Label(testID), func() {
			Expect(true).To(Equal(true))
			badcontainers := []string{}
			for _, cut := range env.Containers {
				logrus.Debugln("check container ", cut.Name)
				if cut.Lifecycle.PreStop == nil {
					badcontainers = append(badcontainers, cut.Name)
					logrus.Errorln("container ", cut.Name, " does not have preStop defined")
				}
			}
			Expect(0).To(Equal(len(badcontainers)))
		})
	}
})
