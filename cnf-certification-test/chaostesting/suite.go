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
	ginkgo.AfterEach(func() {
		// Note: Reloading test environment here because it is possible
		// that some of the resources that are gathered prior to the chaos test
		// have been deleted and need another record created.
		env.SetNeedsRefresh()
	})

	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodDeleteIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testPodDelete(&env)
	})

})

func testPodDelete(env *provider.TestEnvironment) {
	ginkgo.Skip("This TC is under construction.")

	for _, dep := range env.Deployments {
		namespace := dep.Namespace
		var label string
		var err error
		if label, err = poddelete.GetLabelDeploymetValue(env, dep.Spec.Template.Labels); err != nil {
			logrus.Errorf("didn't find a match label for the deployment %s ", provider.DeploymentToString(dep))
			ginkgo.Fail(fmt.Sprintf("There is no label for the deployment %s ", provider.DeploymentToString(dep)))
		}
		if err := poddelete.ApplyAndCreatePodDeleteResources(label, deployment, namespace); err != nil {
			ginkgo.Fail(fmt.Sprintf("test failed while creating the resources err:%s", err))
		}
		if completed := poddelete.WaitForTestFinish(testCaseTimeout); !completed {
			poddelete.DeleteAllResources(namespace)
			logrus.Errorf("deployment %s timed-out the litmus test", provider.DeploymentToString(dep))
			ginkgo.Fail(fmt.Sprintf("deployment %s timed-out the litmus test", provider.DeploymentToString(dep)))
		}
		if result := poddelete.IsChaosResultVerdictPass(); !result {
			// delete the chaos engin crd
			poddelete.DeleteAllResources(namespace)
			logrus.Errorf("deployment %s failed the litmus test", provider.DeploymentToString(dep))
			ginkgo.Fail(fmt.Sprintf("deployment %s failed the litmus test", provider.DeploymentToString(dep)))
		}
		poddelete.DeleteAllResources(namespace)
	}
}
