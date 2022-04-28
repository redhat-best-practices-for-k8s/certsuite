package chaostesting

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/test-network-function/cnf-certification-test/pkg/provider"

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
	RequestTimeout     = 120 * time.Second
	Deployment         = "deployment"
	ServiceAccountFile = "chaostesting/servicAccount.yaml"
	ExperimentFile     = "chaostesting/deleteExperment.yaml"
	chaosEngineFile    = "chaostesting/chaosEngine.yaml"
	chaosname          = "pod-delete"
	completedResult    = "completed"
)

var _ = ginkgo.Describe(common.ChaosTesting, func() {
	var env provider.TestEnvironment

	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})
	ginkgo.ReportAfterEach(results.RecordResult)

	// container security context: privileged escalation
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodDeleteIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testPodDelete(&env)
	})

})

func testPodDelete(env *provider.TestEnvironment) {
	for _, dep := range env.Deployments {
		namespace := dep.ObjectMeta.Namespace
		label := "app=" + dep.Spec.Template.ObjectMeta.Labels["app"]
		fileName, err := applyTemplate(label, Deployment, namespace, ExperimentFile)
		if err != nil {
			logrus.Debugf("cant create the file of the test: %e", err)
			ginkgo.Fail(fmt.Sprintf("cant create the file of the test: %e.", err))
		}
		if _, err = createResource(fileName); err != nil {
			ginkgo.Fail(fmt.Sprintf("%e error create the chaos experment resources.", err))
		}
		fileName, err = applyTemplate(label, Deployment, namespace, ServiceAccountFile)
		if err != nil {
			logrus.Debugf("cant create the file of the test: %e", err)
			ginkgo.Fail(fmt.Sprintf("cant create the file of the test: %e.", err))
		}
		if _, err = createResource(fileName); err != nil {
			ginkgo.Fail(fmt.Sprintf("error create the service account: %e .", err))
		}
		fileName, err = applyTemplate(label, Deployment, namespace, chaosEngineFile)
		if err != nil {
			logrus.Debugf("cant create the file of the test: %e", err)
			ginkgo.Fail(fmt.Sprintf("cant create the file of the test: %e.", err))
		}
		// create the chaos engine for every deployment in the cluster
		if _, err = createResource(fileName); err != nil {
			ginkgo.Fail(fmt.Sprintf("%e error create the chaos engine.", err))
		}
		time.Sleep(1 * time.Second)
		completed := waitForTestFinish(RequestTimeout)
		if completed {
			if finalResult, result := returnResult(); finalResult {
				// delete the chaos engin crd
				deleteAllResources(namespace)
			} else {
				logrus.Debugf("test completed but it failed with reason %s", result)
				ginkgo.Fail(fmt.Sprintf("test completed but it failed with reason %s", result))
			}
		} else {
			logrus.Debug("test failed to be completed")
			ginkgo.Fail("test failed to be completed")
		}
	}
}

func deleteAllResources(namespace string) {
	oc := clientsholder.GetClientsHolder()
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	gvr := schema.GroupVersionResource{Group: "litmuschaos.io", Version: "v1alpha1", Resource: "chaosengines"}
	if err := oc.DynamicClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), "engine-test", deleteOptions); err != nil {
		logrus.Debugf("error while removing the chaos engine resources %e", err)
	}
	err := oc.K8sClient.CoreV1().ServiceAccounts(namespace).Delete(context.TODO(), "test-sa", deleteOptions)
	if err != nil {
		logrus.Debugf("error while removing the ServiceAccountsresources %e", err)
	}
	if err = oc.RbacClient.Roles(namespace).Delete(context.TODO(), "test-sa", deleteOptions); err != nil {
		logrus.Debugf("error while removing the chaos engine resources %e", err)
	}
	if err = oc.RbacClient.RoleBindings(namespace).Delete(context.TODO(), "test-sa", deleteOptions); err != nil {
		logrus.Debugf("error while removing the chaos engine resources %e", err)
	}
	gvr = schema.GroupVersionResource{Group: "litmuschaos.io", Version: "v1alpha1", Resource: "chaosexperiments"}
	if err := oc.DynamicClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), chaosname, deleteOptions); err != nil {
		logrus.Debugf("error while removing the chaos engine resources %e", err)
	}
	e := os.Remove("chaostesting/chaosEngine.yaml.tmp")
	if e != nil {
		logrus.Debugf("error while removing the temp file of the chaos engine %e", e)
	}
	e = os.Remove("chaostesting/servicAccount.yaml.tmp")
	if e != nil {
		logrus.Debugf("error while removing the temp file of the servicAccount %e", e)
	}
	e = os.Remove("chaostesting/deleteExperment.yaml.tmp")
	if e != nil {
		logrus.Debugf("error while removing the temp file of the deleteExperment %e", e)
	}
}

func applyTemplate(appLabel, appKind, namespace, fileename string) (string, error) {
	input, err := os.ReadFile(fileename)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	output := bytes.ReplaceAll(input, []byte("{{ APP_NAMESPACE }}"), []byte(namespace))
	output = bytes.ReplaceAll(output, []byte("{{ APP_LABEL }}"), []byte(appLabel))
	output = bytes.ReplaceAll(output, []byte("{{ APP_KIND }}"), []byte(appKind))
	fileName := fileename + ".tmp"
	if err = os.WriteFile(fileName, output, 0700); err != nil {
		fmt.Println(err)
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

func returnResult() (bool, string) {
	oc := clientsholder.GetClientsHolder()
	gvr := schema.GroupVersionResource{Group: "litmuschaos.io", Version: "v1alpha1", Resource: "chaosresults"}
	crs, err := oc.DynamicClient.Resource(gvr).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("error getting : %v\n", err)
	}
	for _, cr := range crs.Items {
		result := cr.Object["status"].(map[string]interface{})["experimentStatus"].(map[string]interface{})["failStep"]
		expKind := cr.Object["spec"].(map[string]interface{})["experiment"]
		if expKind == chaosname {
			if result == "N/A" {
				return true, ""
			}
			return false, result.(string)

		}
	}
	return false, ""
}

func waitForResult() bool {
	oc := clientsholder.GetClientsHolder()
	gvr := schema.GroupVersionResource{Group: "litmuschaos.io", Version: "v1alpha1", Resource: "chaosengines"}
	crs, err := oc.DynamicClient.Resource(gvr).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("error getting : %v\n", err)
	}
	for _, cr := range crs.Items {
		typ := cr.Object["status"].(map[string]interface{})["experiments"].([]interface{})
		status := cr.Object["status"].(map[string]interface{})["engineStatus"]
		if typ[0].(map[string]interface{})["name"] == chaosname {
			if status == completedResult {
				return true
			} else {
				return false
			}
		}
	}
	return false
}

//nolint:funlen //
func createResource(filepath string) (bool, error) {
	oc := clientsholder.GetClientsHolder()

	b, err := os.ReadFile(filepath)
	if err != nil {
		return false, err
	}
	log.Printf("%q \n", string(b))

	c := oc.K8sClient

	dd := oc.DynamicClient
	const oneh = 100
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(b), oneh)
	for {
		var rawObj runtime.RawExtension
		if err = decoder.Decode(&rawObj); err != nil {
			break
		}

		obj, gvk, error := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if error != nil {
			return false, error
		}
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return false, err
		}

		unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

		gr, err := restmapper.GetAPIGroupResources(c.Discovery())
		if err != nil {
			return false, err
		}

		mapper := restmapper.NewDiscoveryRESTMapper(gr)
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return false, err
		}

		var dri dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			if unstructuredObj.GetNamespace() == "" {
				unstructuredObj.SetNamespace("default")
			}
			dri = dd.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())
		} else {
			dri = dd.Resource(mapping.Resource)
		}

		if _, err := dri.Create(context.Background(), unstructuredObj, metav1.CreateOptions{}); err != nil {
			return false, err
		}
	}
	if err != io.EOF {
		return false, err
	}
	return true, nil
}
