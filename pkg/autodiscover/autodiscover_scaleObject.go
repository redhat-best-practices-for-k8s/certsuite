package autodiscover

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/stringhelper"
	scalingv1 "k8s.io/api/autoscaling/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Scaleobject struct {
	Scale               *scalingv1.Scale
	GroupResourceSchema schema.GroupResource
}

func GetScaleCrUnderTest(namespaces []string, crds []*apiextv1.CustomResourceDefinition, testData []configuration.CrdFilter) []Scaleobject {
	var scalableItems []Scaleobject
	clients := clientsholder.GetClientsHolder()
	for _, aCrd := range crds {
		for _, crdFilter := range testData {
			if strings.HasSuffix(aCrd.Name, crdFilter.NameSuffix) {
				for _, version := range aCrd.Spec.Versions {
					gvr := schema.GroupVersionResource{
						Group:    aCrd.Spec.Group,
						Version:  version.Name,
						Resource: aCrd.Spec.Names.Plural,
					}

					logrus.Debugf("Looking for CRs from CRD: %s api version:%s group:%s plural:%s", aCrd.Name, version.Name, aCrd.Spec.Group, aCrd.Spec.Names.Plural)
					crs, err := clients.DynamicClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})

					if err != nil {
						logrus.Fatalf("error getting Resource for %s: %v\n", aCrd.Name, err)
					}
					scalableItems = append(scalableItems, appendCrItems(crs, aCrd, namespaces)...)
				}
			}
		}
	}
	return scalableItems
}

func appendCrItems(crs *unstructured.UnstructuredList, aCrd *apiextv1.CustomResourceDefinition, namespaces []string) []Scaleobject {
	var scalableItems []Scaleobject
	clients := clientsholder.GetClientsHolder()
	for _, cr := range crs.Items {
		groupResourceSchema := schema.GroupResource{
			Group:    aCrd.Spec.Group,
			Resource: aCrd.Spec.Names.Plural,
		}
		namespace := cr.Object["metadata"].(map[string]interface{})["namespace"].(string)
		if !stringhelper.StringInSlice(namespaces, namespace, false) {
			continue
		}
		name := cr.Object["metadata"].(map[string]interface{})["name"].(string)

		crScale, err := clients.ScalingClient.Scales(namespace).Get(context.TODO(), groupResourceSchema, name, metav1.GetOptions{})
		if err != nil {
			logrus.Fatalf("error while getting the scaling fileds %e", err)
		}
		scalableItems = append(scalableItems, Scaleobject{Scale: crScale, GroupResourceSchema: groupResourceSchema})
	}
	return scalableItems
}
