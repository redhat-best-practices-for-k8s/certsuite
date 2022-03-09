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
	"fmt"
	"strconv"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

const (
	indexprotocolname = 0
	indexport         = 4
)

type key struct {
	port     int
	protocol string
}

type Port []struct {
	ContainerPort int
	Name          string
	Protocol      string
}

//
// All actual test code belongs below here.  Utilities belong above.
//
var _ = ginkgo.Describe(common.NetworkingTestKey, func() {
	logrus.Debugf("%s not moved yet to new framework", common.NetworkingTestKey)
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		provider.BuildTestEnvironment()
		env = provider.GetTestEnvironment()
	})
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestUndeclaredContainerPortsUsage)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testListenAndDeclared(&env)
	})
})

func parseListening(res string, listeningPorts map[key]string) {
	var k key
	lines := strings.Split(res, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if !strings.Contains(line, "LISTEN") {
			continue
		}
		if indexprotocolname > len(fields) || indexport > len(fields) {
			return
		}
		s := strings.Split(fields[indexport], ":")
		p, _ := strconv.Atoi(s[1])
		k.port = p
		k.protocol = strings.ToUpper(fields[indexprotocolname])
		k.protocol = strings.ReplaceAll(k.protocol, "\"", "")
		listeningPorts[k] = ""
	}
}

func checkIfListenIsDeclared(listeningPorts, declaredPorts map[key]string) map[key]string {
	res := make(map[key]string)
	if len(listeningPorts) == 0 {
		return res
	}
	for k := range listeningPorts {
		_, ok := declaredPorts[k]
		if !ok {
			tnf.ClaimFilePrintf(fmt.Sprintf("The port %d on protocol %s in pod %s is not declared.", k.port, k.protocol, listeningPorts[k]))
			res[k] = listeningPorts[k]
		}
	}
	return res
}

//nolint:funlen
func testListenAndDeclared(env *provider.TestEnvironment) {
	var k key
	declaredPorts := make(map[key]string)
	listeningPorts := make(map[key]string)
	var failedPods string
	var skippedPods string

	for i := 0; i < len(env.Containers); i++ {
		ports := env.Containers[i].Data.Ports
		if ports == nil {
			tnf.ClaimFilePrintf("Failed to get declared port for container %d due to %v, skipping pod %s", i, env.Containers[i].Namespace+"."+env.Containers[i].NodeName)
			skippedPods += env.Containers[i].Namespace + "." + env.Containers[i].NodeName + "\n"
		}

		for i := 0; i < len(ports); i++ {
			k.port = int(ports[i].ContainerPort)
			k.protocol = string(ports[i].Protocol)
			declaredPorts[k] = ports[i].Name
		}
	}

	for _, dp := range env.DebugPods {
		oc := clientsholder.NewClientsHolder()
		output, outerr, err := oc.ExecCommandContainer(clientsholder.Context{Namespace: dp.Namespace,
			Podname: dp.Name, Containername: dp.Spec.Containers[0].Name}, `ss -tulwnH`)
		if err != nil {
			logrus.Errorln("can't execute command on container ", err)
			continue
		}
		if outerr != "" {
			logrus.Errorln("error when running listening command ", outerr)
			continue
		}
		parseListening(output, listeningPorts)
		if len(listeningPorts) == 0 {
			tnf.ClaimFilePrintf("Failed to get listening port for pod name %s in pod namespace %s, skipping this pod", dp.Name, dp.Namespace, err)
			continue
		}
		// compare between declaredPort,listeningPort
		undeclaredPorts := checkIfListenIsDeclared(listeningPorts, declaredPorts)
		for k := range undeclaredPorts {
			tnf.ClaimFilePrintf("The port %d on protocol %s in pod name %s and pod namespace is %s not declared.", k.port, k.protocol, dp.Name, dp.Namespace)
		}
		if len(undeclaredPorts) != 0 {
			for x := range undeclaredPorts {
				p := strconv.Itoa(x.port)
				failedPods += p + " " + x.protocol
			}
		}
	}

	if nf, ns := len(failedPods), len(skippedPods); nf > 0 || ns > 0 {
		ginkgo.Fail("Found %d pods with listening ports not declared and Skipped %d pods due to unexpected error", nf, ns)
	}
}
