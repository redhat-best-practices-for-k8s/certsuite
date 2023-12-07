package chaostesting

/*

import (
	"fmt"
	"time"

	"github.com/test-network-function/cnf-certification-test/pkg/provider"


	poddelete "github.com/test-network-function/cnf-certification-test/cnf-certification-test/chaostesting/pod_delete"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/internal/log"

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
	log.Debug("Entering %s suite", common.ChaosTesting)
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

	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodDeleteIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testPodDelete(&env)
	})

})

func testPodDelete(env *provider.TestEnvironment) {
	ginkgo.Skip("This TC is under construction.")

	for _, dep := range env.Deployments {
		var label string
		var err error
		if label, err = poddelete.GetLabelDeploymentValue(env, dep.Spec.Template.Labels); err != nil {
			log.Error("did not find a match label for the deployment %s ", dep.ToString())
			ginkgo.Fail(fmt.Sprintf("There is no label for the deployment %s ", dep.ToString()))
		}
		if err := poddelete.ApplyAndCreatePodDeleteResources(label, deployment, dep.Namespace); err != nil {
			ginkgo.Fail(fmt.Sprintf("test failed while creating the resources err:%s", err))
		}
		if completed := poddelete.WaitForTestFinish(testCaseTimeout); !completed {
			poddelete.DeleteAllResources(dep.Namespace)
			log.Error("deployment %s timed-out the litmus test", dep.ToString())
			ginkgo.Fail(fmt.Sprintf("deployment %s timed-out the litmus test", dep.ToString()))
		}
		if result := poddelete.IsChaosResultVerdictPass(); !result {
			// delete the chaos engin crd
			poddelete.DeleteAllResources(dep.Namespace)
			log.Error("deployment %s failed the litmus test", dep.ToString())
			ginkgo.Fail(fmt.Sprintf("deployment %s failed the litmus test", dep.ToString()))
		}
		poddelete.DeleteAllResources(dep.Namespace)
	}
}

*/
