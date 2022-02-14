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
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/ownerreference"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	v1 "k8s.io/api/core/v1"
)

//
// All actual test code belongs below here.  Utilities belong above.
//
var _ = ginkgo.Describe(common.LifecycleTestKey, func() {
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		provider.BuildTestEnvironment()
		env = provider.GetTestEnvironment()
	})
	testContainersPreStop(&env)
	testContainersImagePolicy(&env)
	testContainersReadinessProbe(&env)
	testContainersLivenessProbe(&env)
	testPodsOwnerReference(&env)
})

func testContainersPreStop(env *provider.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestShudtownIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		badcontainers := []string{}
		for _, cut := range env.Containers {
			logrus.Debugln("check container ", cut.Namespace, " ", cut.Podname, " ", cut.Data.Name, " pre stop lifecycle ")
			if cut.Data.Lifecycle.PreStop == nil {
				badcontainers = append(badcontainers, cut.Data.Name)
				logrus.Errorln("container ", cut.Data.Name, " does not have preStop defined")
			}
		}
		if len(badcontainers) > 0 {
			tnf.ClaimFilePrintf("bad containers %v", badcontainers)
		}
		gomega.Expect(0).To(gomega.Equal(len(badcontainers)))
	})
}

func testContainersImagePolicy(env *provider.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestImagePullPolicyIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		badcontainers := []string{}
		for _, cut := range env.Containers {
			logrus.Debugln("check container ", cut.Namespace, " ", cut.Podname, " ", cut.Data.Name, " pull policy, should be ", v1.PullIfNotPresent)
			if cut.Data.ImagePullPolicy != v1.PullIfNotPresent {
				s := cut.Namespace + ":" + cut.Podname + ":" + cut.Data.Name
				badcontainers = append(badcontainers, s)
				logrus.Errorln("container ", cut.Data.Name, " is using ", cut.Data.ImagePullPolicy, " as image policy")
			}
		}
		if len(badcontainers) > 0 {
			tnf.ClaimFilePrintf("bad containers %v", badcontainers)
		}
		gomega.Expect(0).To(gomega.Equal(len(badcontainers)))
	})
}

//nolint:dupl
func testContainersReadinessProbe(env *provider.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestReadinessProbeIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		badcontainers := []string{}
		for _, cut := range env.Containers {
			logrus.Debugln("check container ", cut.Namespace, " ", cut.Podname, " ", cut.Data.Name, " readiness probe ")
			if cut.Data.ReadinessProbe == nil {
				s := cut.Namespace + ":" + cut.Podname + ":" + cut.Data.Name
				badcontainers = append(badcontainers, s)
				logrus.Errorln("container ", cut.Data.Name, " does not have ReadinessProbe defined")
			}
		}
		if len(badcontainers) > 0 {
			tnf.ClaimFilePrintf("bad containers %v", badcontainers)
		}
		gomega.Expect(0).To(gomega.Equal(len(badcontainers)))
	})
}

//nolint:dupl
func testContainersLivenessProbe(env *provider.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestLivenessProbeIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		badcontainers := []string{}
		for _, cut := range env.Containers {
			logrus.Debugln("check container ", cut.Namespace, " ", cut.Podname, " ", cut.Data.Name, " liveness probe ")
			if cut.Data.LivenessProbe == nil {
				s := cut.Namespace + ":" + cut.Podname + ":" + cut.Data.Name
				badcontainers = append(badcontainers, s)
				logrus.Errorln("container ", cut.Data.Name, " does not have livenessProbe defined")
			}
		}
		if len(badcontainers) > 0 {
			tnf.ClaimFilePrintf("bad containers %v", badcontainers)
		}
		gomega.Expect(0).To(gomega.Equal(len(badcontainers)))
	})
}

func testPodsOwnerReference(env *provider.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodDeploymentBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		ginkgo.By("Testing owners of CNF pod, should be replicas Set")
		badPods := []string{}
		for _, put := range env.Pods {
			logrus.Debugln("check pod ", put.Namespace, " ", put.Name, " owner reference")
			o := ownerreference.NewOwnerReference(put)
			o.RunTest()
			if o.GetResults() == tnf.SUCCESS {
				s := put.Namespace + ":" + put.Name
				badPods = append(badPods, s)
			}
		}
		if len(badPods) > 0 {
			tnf.ClaimFilePrintf("bad containers %v", badPods)
		}
		gomega.Expect(0).To(gomega.Equal(len(badPods)))
	})
}
