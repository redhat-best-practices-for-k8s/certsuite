package chaostesting

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/restmapper"

	"github.com/onsi/ginkgo/v2"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	// timeout for eventually call
	testCaseTimeout = 180 * time.Second
	deployment      = "deployment"
	// costum rescources file to create them
	serviceAccountFile = "chaostesting/service-account.yaml"
	experimentFile     = "chaostesting/experiment-delete.yaml"
	chaosEngineFile    = "chaostesting/chaos-engine.yaml"
	chaosTestName      = "pod-delete" // test name
	completedResult    = "completed"
	pass               = "Pass"
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
		if label = getLabelDeploymetValue(env, dep.Spec.Template.Labels); label == "" {
			logrus.Errorf("didnt find a match label for the deployment %s ", dep.Name)
			ginkgo.Fail(fmt.Sprintf("There is no label for the deployment%s ", dep.Name))
		}
		if err := applyAndCreateFiles(label, deployment, namespace); err != nil {
			ginkgo.Fail(fmt.Sprintf("test failed while creating the files %s", err))
		}
		if completed := waitForTestFinish(testCaseTimeout); !completed {
			deleteAllResources(namespace)
			logrus.Debug("test failed to be completed")
			ginkgo.Fail("test failed to be completed")
		}
		if result := IsChaosResultVerdictPass(); !result {
			// delete the chaos engin crd
			deleteAllResources(namespace)
			ginkgo.Fail("test completed but it failed with reason ")
		}

		deleteAllResources(namespace)
	}
}

// a function to search the right label for the deployment that we wanr to test pod delete on it
func getLabelDeploymetValue(env *provider.TestEnvironment, labelsMap map[string]string) string {
	var key string
	for _, label := range env.Config.TargetPodLabels {
		if label.Prefix != "" {
			key = fmt.Sprintf("%s/%s", label.Prefix, label.Name)
		} else {
			key = label.Name
		}
		if l, ok := labelsMap[key]; ok && l == label.Value {
			return fmt.Sprintf("%s=%s", key, label.Value)
		}
	}
	return ""
}

func applyAndCreateFiles(appLabel, appKind, namespace string) error {
	fileName, err := applyTemplate(appLabel, appKind, namespace, experimentFile)
	if err != nil {
		logrus.Errorf("cant create the file of the test: %s", err)
		return err
	}
	if err = createResource(fileName); err != nil {
		logrus.Errorf("%s error create the chaos experment resources.", err)
		return err
	}
	fileName, err = applyTemplate(appLabel, appKind, namespace, serviceAccountFile)
	if err != nil {
		logrus.Errorf("cant create the file of the test: %s", err)
		return err
	}
	if err = createResource(fileName); err != nil {
		logrus.Errorf("error create the service account: %s .", err)
		return err
	}
	fileName, err = applyTemplate(appLabel, appKind, namespace, chaosEngineFile)
	if err != nil {
		logrus.Errorf("cant create the file of the test: %s", err)
		return err
	}
	// create the chaos engine for every deployment in the cluster
	if err = createResource(fileName); err != nil {
		logrus.Errorf("%s error create the chaos engine.", err)
		return err
	}
	return nil
}

func deleteAllResources(namespace string) {
	oc := clientsholder.GetClientsHolder()
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	gvr := schema.GroupVersionResource{Group: "litmuschaos.io", Version: "v1alpha1", Resource: "chaosengines"}
	if err := oc.DynamicClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), "engine-test", deleteOptions); err != nil {
		logrus.Errorf("error while removing the chaos engine resources %e", err)
	}
	err := oc.K8sClient.CoreV1().ServiceAccounts(namespace).Delete(context.TODO(), "test-sa", deleteOptions)
	if err != nil {
		logrus.Errorf("error while removing the ServiceAccountsresources %e", err)
	}
	if err = oc.K8sClient.RbacV1().Roles(namespace).Delete(context.TODO(), "test-sa", deleteOptions); err != nil {
		logrus.Errorf("error while removing the chaos engine resources %e", err)
	}
	if err = oc.K8sClient.RbacV1().RoleBindings(namespace).Delete(context.TODO(), "test-sa", deleteOptions); err != nil {
		logrus.Errorf("error while removing the chaos engine resources %e", err)
	}
	gvr = schema.GroupVersionResource{Group: "litmuschaos.io", Version: "v1alpha1", Resource: "chaosexperiments"}
	if err := oc.DynamicClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), chaosTestName, deleteOptions); err != nil {
		logrus.Errorf("error while removing the chaos engine resources %e", err)
	}
	e := os.Remove(chaosEngineFile + ".tmp")
	if e != nil {
		logrus.Errorf("error while removing the temp file of the chaos engine %e", e)
	}
	e = os.Remove(serviceAccountFile + ".tmp")
	if e != nil {
		logrus.Errorf("error while removing the temp file of the servicAccount %e", e)
	}
	e = os.Remove(experimentFile + ".tmp")
	if e != nil {
		logrus.Errorf("error while removing the temp file of the deleteExperment %e", e)
	}
}

func applyTemplate(appLabel, appKind, namespace, filename string) (string, error) {
	input, err := os.ReadFile(filename)
	if err != nil {
		logrus.Errorf("error while reading the yaml file : %s ,%s", filename, err)
		tnf.ClaimFilePrintf("error while reading the yaml file : %s ,%s", filename, err)
		return "", err
	}
	output := bytes.ReplaceAll(input, []byte("{{APP_NAMESPACE}}"), []byte(namespace))
	output = bytes.ReplaceAll(output, []byte("{{APP_LABEL}}"), []byte(appLabel))
	output = bytes.ReplaceAll(output, []byte("{{APP_KIND}}"), []byte(appKind))
	fileName := filename + ".tmp"
	const filePermission = 0o600
	if err = os.WriteFile(fileName, output, filePermission); err != nil {
		logrus.Errorf("error: %s while writing the new template file %s ", err, fileName)
		tnf.ClaimFilePrintf("error: %s while writing the new template file %s ", err, fileName)
		return "", err
	}
	return fileName, nil
}

func waitForTestFinish(timeout time.Duration) bool {
	const pollingPeriod = 1 * time.Second
	var elapsed time.Duration
	var result bool
	for elapsed < timeout {
		result = waitForResult()

		if result {
			break
		}
		time.Sleep(pollingPeriod)
		elapsed += pollingPeriod
	}
	return result
}

func IsChaosResultVerdictPass() bool {
	oc := clientsholder.GetClientsHolder()
	gvr := schema.GroupVersionResource{Group: "litmuschaos.io", Version: "v1alpha1", Resource: "chaosresults"}
	crs, err := oc.DynamicClient.Resource(gvr).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("error getting : %v\n", err)
	}
	for _, cr := range crs.Items {
		failResult := cr.Object["status"].(map[string]interface{})["experimentStatus"].(map[string]interface{})["failStep"]
		verdictValue := cr.Object["status"].(map[string]interface{})["experimentStatus"].(map[string]interface{})["verdict"]
		expKind := cr.Object["spec"].(map[string]interface{})["experiment"]
		if expKind == chaosTestName {
			if verdictValue == pass {
				return true
			}
			logrus.Debugf("test completed but it failed with reason %s", failResult.(string))
			return false
		}
	}
	return false
}

func waitForResult() bool {
	oc := clientsholder.GetClientsHolder()
	gvr := schema.GroupVersionResource{Group: "litmuschaos.io", Version: "v1alpha1", Resource: "chaosengines"}
	crs, err := oc.DynamicClient.Resource(gvr).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("error getting : %v\n", err)
	}
	for _, cr := range crs.Items {
		if status := cr.Object["status"]; status != nil {
			if exp := status.(map[string]interface{})["experiments"]; exp != nil {
				typ := exp.([]interface{})
				status := cr.Object["status"].(map[string]interface{})["engineStatus"]
				if typ[0].(map[string]interface{})["name"] == chaosTestName {
					return status == completedResult
				}
			}
		}
	}
	return false
}

//nolint:funlen //
func createResource(filepath string) error {
	oc := clientsholder.GetClientsHolder()
	b, ferr := os.ReadFile(filepath)
	if ferr != nil {
		return ferr
	}
	k8sClient := oc.K8sClient
	dynamicClient := oc.DynamicClient
	const oneh = 100
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(b), oneh)
	for {
		var rawObj runtime.RawExtension
		if err := decoder.Decode(&rawObj); err != nil {
			break
		}
		obj, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			return err
		}
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return err
		}

		unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

		gr, err := restmapper.GetAPIGroupResources(k8sClient.Discovery())
		if err != nil {
			return err
		}

		mapper := restmapper.NewDiscoveryRESTMapper(gr)
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}

		var dri dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			if unstructuredObj.GetNamespace() == "" {
				unstructuredObj.SetNamespace("default")
			}
			dri = dynamicClient.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())
		} else {
			dri = dynamicClient.Resource(mapping.Resource)
		}

		if _, err := dri.Create(context.Background(), unstructuredObj, metav1.CreateOptions{}); err != nil {
			return err
		}
	}
	if ferr != io.EOF {
		return ferr
	}
	return nil
}
