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

package networking

import (
	"github.com/sirupsen/logrus"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"

	"github.com/onsi/ginkgo/v2"
)

//
// All actual test code belongs below here.  Utilities belong above.
//
var _ = ginkgo.Describe(common.NetworkingTestKey, func() {
	logrus.Debug(common.PlatformAlterationTestKey, " not moved yet to new framework")
})
