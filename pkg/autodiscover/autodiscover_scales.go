package autodiscover

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	scalingv1 "k8s.io/api/autoscaling/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ScaleObject struct {
	Scale               *scalingv1.Scale
	GroupResourceSchema schema.GroupResource
}

func GetScaleCrUnderTest(namespaces []string, crds []*apiextv1.CustomResourceDefinition) []ScaleObject {
	dynamicClient := clientsholder.GetClientsHolder().DynamicClient

	var scaleObjects []ScaleObject
	for _, crd := range crds {
		if crd.Spec.Scope != apiextv1.NamespaceScoped {
			logrus.Warnf("Target CRD %s is cluster-wide scoped. Skipping search of scale objects.", crd.Name)
			continue
		}

		for i := range crd.Spec.Versions {
			crdVersion := crd.Spec.Versions[i]
			gvr := schema.GroupVersionResource{
				Group:    crd.Spec.Group,
				Version:  crdVersion.Name,
				Resource: crd.Spec.Names.Plural,
			}

			// Filter out non-scalable CRDs.
			if crdVersion.Subresources == nil || crdVersion.Subresources.Scale == nil {
				logrus.Infof("Target CRD %s is not scalable. Skipping search of scalable CRs.", crd.Name)
				continue
			}

			logrus.Debugf("Looking for Scalable CRs of CRD %s (api version %s, group %s, plural %s) in target namespaces.",
				crd.Name, crdVersion.Name, crd.Spec.Group, crd.Spec.Names.Plural)

			for _, ns := range namespaces {
				crs, err := dynamicClient.Resource(gvr).Namespace(ns).List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					logrus.Fatalf("Error getting CRs of CRD %s in namespace %s: %v", crd.Name, ns, err)
				}

				if len(crs.Items) > 0 {
					scaleObjects = append(scaleObjects, getCrScaleObjects(crs.Items, crd)...)
				} else {
					logrus.Warnf("No CRs of CRD %s found in the target namespaces.", crd.Name)
				}
			}
		}
	}

	return scaleObjects
}

func getCrScaleObjects(crs []unstructured.Unstructured, crd *apiextv1.CustomResourceDefinition) []ScaleObject {
	var scaleObjects []ScaleObject
	clients := clientsholder.GetClientsHolder()
	for _, cr := range crs {
		groupResourceSchema := schema.GroupResource{
			Group:    crd.Spec.Group,
			Resource: crd.Spec.Names.Plural,
		}

		name := cr.GetName()
		namespace := cr.GetNamespace()
		crScale, err := clients.ScalingClient.Scales(namespace).Get(context.TODO(), groupResourceSchema, name, metav1.GetOptions{})
		if err != nil {
			logrus.Fatalf("Error while getting the scale of CR=%s (CRD=%s) in namespace %s: %v", name, crd.Name, namespace, err)
		}

		scaleObjects = append(scaleObjects, ScaleObject{Scale: crScale, GroupResourceSchema: groupResourceSchema})
	}
	return scaleObjects
}
