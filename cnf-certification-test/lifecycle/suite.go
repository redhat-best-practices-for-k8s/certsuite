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
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
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
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestShudtownIdentifier)
	It(testID, Label(testID), func() {
		Expect(true).To(Equal(true))
		badcontainers := []string{}
		for _, cut := range env.Containers {
			fmt.Println("container ", cut.Data.Name)
			logrus.Debugln("check container ", cut.Data.Name) //maybe use different platform ?
			if cut.Data.Lifecycle.PreStop == nil {
				badcontainers = append(badcontainers, cut.Data.Name)
				logrus.Errorln("container ", cut.Data.Name, " does not have preStop defined")
			}

		}
		Expect(0).To(Equal(len(badcontainers)))
	})

})
