package chaostesting

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
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
	RequestTimeout = 40 * time.Second
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
	oc := clientsholder.GetClientsHolder()
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	// create the chaos experment resources
	_, err := createResource("chaostesting/deleteExperment.yaml")
	if err != nil {
		ginkgo.Fail(fmt.Sprintf("%e error create the chaos experment resources.", err))
	}
	// create the service account
	_, err = createResource("chaostesting/servicAccount.yaml")
	if err != nil {
		ginkgo.Fail(fmt.Sprintf("error create the service account: %e .", err))
	}
	dep := env.Deployments
	kind := "deployment"
	for i := range dep {
		namespace := dep[i].ObjectMeta.Namespace
		label := "app=" + dep[i].Spec.Template.ObjectMeta.Labels["app"]
		err := applyTemplate(label, kind, namespace)
		if err != nil {
			logrus.Debugf("cant create the file of the test: %e", err)
			ginkgo.Fail(fmt.Sprintf("cant create the file of the test: %e.", err))

		}
		// create the chaos engine for every deployment in the cluster
		_, err = createResource("chaostesting/chaosEngine-temp.yaml")
		if err != nil {
			ginkgo.Fail(fmt.Sprintf("%e error create the chaos engine.", err))
		}
		r := waitForTestFinish(RequestTimeout)
		if r {
			finalResult := returnResult()
			log.Print(finalResult)
			// delete the chaos engin crd
			gvr := schema.GroupVersionResource{Group: "litmuschaos.io", Version: "v1alpha1", Resource: "chaosengines"}
			if err := oc.DynamicClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), "engine-test", deleteOptions); err != nil {
				logrus.Debugf("error while removing the chaos engine resources %e", err)

			}
			e := os.Remove("chaostesting/chaosEngine-temp.yaml")
			if e != nil {
				logrus.Debugf("error while removing the temp file of the chaos engine %e", e)
			}
		}
	}
}

func applyTemplate(appLabel, appKind, Namespace string) error {
	input, err := ioutil.ReadFile("chaostesting/chaosEngine.yaml")
	if err != nil {
		fmt.Println(err)
		return err
	}
	output := bytes.Replace(input, []byte("{{ APP_NAMESPACE }}"), []byte(Namespace), -1)
	output = bytes.Replace(output, []byte("{{ APP_LABEL }}"), []byte(appLabel), -1)
	output = bytes.Replace(output, []byte("{{ APP_KIND }}"), []byte(appKind), -1)
	if err = ioutil.WriteFile("chaostesting/chaosEngine-temp.yaml", output, 0700); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
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

func returnResult() bool {
	oc := clientsholder.GetClientsHolder()
	gvr := schema.GroupVersionResource{Group: "litmuschaos.io", Version: "v1alpha1", Resource: "chaosresults"}
	crs, err := oc.DynamicClient.Resource(gvr).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("error getting : %v\n", err)
	}
	for _, cr := range crs.Items {
		result := cr.Object["status"].(map[string]interface{})["experimentStatus"].(map[string]interface{})["failStep"]
		expKind := cr.Object["spec"].(map[string]interface{})["experiment"]
		if expKind == "pod-delete" {
			if result == "N/A" {
				return true
			} else {
				return false
			}
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
		typ := cr.Object["status"].(map[string]interface{})["experiments"].([]interface{})
		status := cr.Object["status"].(map[string]interface{})["engineStatus"]
		if typ[0].(map[string]interface{})["name"] == "pod-delete" {
			if status == "completed" {
				return true
			} else {
				return false
			}
		}
	}
	return false
}

func createResource(filepath string) (bool, error) {
	oc := clientsholder.GetClientsHolder()

	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return false, err
	}
	log.Printf("%q \n", string(b))

	c := oc.K8sClient

	dd := oc.DynamicClient

	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(b), 100)
	for {
		var rawObj runtime.RawExtension
		if err = decoder.Decode(&rawObj); err != nil {
			break
		}

		obj, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			return false, err
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
