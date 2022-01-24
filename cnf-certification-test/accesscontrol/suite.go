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

package accesscontrol

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
)

const (
	// ocGetCrPluralNameFormat is the CR name to use with "oc get <resource_name>".
	ocGetCrPluralNameFormat = "oc get crd %s -o jsonpath='{.spec.names.plural}'"

	// ocGetCrNamespaceFormat is the "oc get" format string to get the namespaced-only resources created for a given CRD.
	ocGetCrNamespaceFormat = "oc get %s -A -o go-template='{{range .items}}{{if .metadata.namespace}}{{.metadata.name}},{{.metadata.namespace}}{{\"\n\"}}{{end}}{{end}}'"
)

var (
	invalidNamespacePrefixes = []string{
		"default",
		"openshift-",
		"istio-",
		"aspenmesh-",
	}
)

var _ = ginkgo.Describe(common.AccessControlTestKey, func() {
	logrus.Debug(common.AccessControlTestKey, " not moved yet to new framework")
})
