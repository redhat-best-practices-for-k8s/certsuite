package chaostesting

import (
	"fmt"
	"time"

	"github.com/test-network-function/cnf-certification-test/pkg/provider"

	"github.com/sirupsen/logrus"

	poddelete "github.com/test-network-function/cnf-certification-test/cnf-certification-test/chaostesting/pod_delete"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"

	"github.com/onsi/ginkgo/v2"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
)

const (
	// timeout for eventually call
	testCaseTimeout = 180 * time.Second
	deployment      = "deployment"
)

var _ = ginkgo.Describe(common.ChaosTesting, func() {
	var env provider.TestEnvironment

	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})
	ginkgo.ReportAfterEach(results.RecordResult)

	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodDeleteIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testPodDelete(&env)
	})

})

func testPodDelete(env *provider.TestEnvironment) {
	for _, dep := range env.Deployments {
		namespace := dep.Namespace
		var label string
		var err error
		if label, err = poddelete.GetLabelDeploymetValue(env, dep.Spec.Template.Labels); err != nil {
			logrus.Errorf("didnt find a match label for the deployment %s ", dep.Name)
			ginkgo.Fail(fmt.Sprintf("There is no label for the deployment%s ", dep.Name))
		}
		if err := poddelete.ApplyAndCreateFiles(label, deployment, namespace); err != nil {
			ginkgo.Fail(fmt.Sprintf("test failed while creating the files %s", err))
		}
		if completed := poddelete.WaitForTestFinish(testCaseTimeout); !completed {
			poddelete.DeleteAllResources(namespace)
			logrus.Debug("test failed to be completed")
			ginkgo.Fail("test failed to be completed")
		}
		if result := poddelete.IsChaosResultVerdictPass(); !result {
			// delete the chaos engin crd
			poddelete.DeleteAllResources(namespace)
			ginkgo.Fail("test completed but it failed with reason ")
		}

		poddelete.DeleteAllResources(namespace)
	}
}
