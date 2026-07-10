// Copyright (C) 2020-2026 Red Hat, Inc.
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
package autodiscover

import (
	"context"
	"fmt"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	appsv1 "k8s.io/api/apps/v1"
	scalingv1 "k8s.io/api/autoscaling/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	appv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/scale"
)

// isMatchingAtLeastOneLabel is a generic helper that checks if a controller's pod template
// matches at least one of the provided labels.
func isMatchingAtLeastOneLabel[T any](
	labels []labelObject,
	namespace string,
	item *T,
	resourceType string,
	getLabelsFn func(*T) map[string]string,
	getNameFn func(*T) string,
) bool {
	name := getNameFn(item)
	templateLabels := getLabelsFn(item)

	for _, labelObj := range labels {
		log.Debug("Searching pods in %s %q found in ns %q using label %s=%s",
			resourceType, name, namespace, labelObj.LabelKey, labelObj.LabelValue)

		if templateLabels[labelObj.LabelKey] == labelObj.LabelValue {
			log.Info("%s %s found in ns=%s", resourceType, name, namespace)
			return true
		}
	}
	return false
}

// findControllersByLabels is a generic implementation for finding pod controllers by labels.
// It works with any Kubernetes resource type (Deployment, StatefulSet, DaemonSet, etc.) by using
// accessor functions to extract the necessary fields.
func findControllersByLabels[T any](
	appClient appv1client.AppsV1Interface,
	labels []labelObject,
	namespaces []string,
	resourceType string,
	lister func(appv1client.AppsV1Interface, string) ([]T, error),
	getLabelsFn func(*T) map[string]string,
	getNameFn func(*T) string,
) []T {
	allResults := []T{}

	for _, ns := range namespaces {
		items, err := lister(appClient, ns)
		if err != nil {
			log.Error("Failed to list %s resources in ns=%s, err: %v. Trying to proceed.", resourceType, ns, err)
			continue
		}

		if len(items) == 0 {
			log.Warn("Did not find any %s in ns=%s", resourceType, ns)
		}

		for i := range items {
			if len(labels) > 0 {
				// The resource is added only once if at least one pod matches one label
				if isMatchingAtLeastOneLabel(labels, ns, &items[i], resourceType, getLabelsFn, getNameFn) {
					allResults = append(allResults, items[i])
				}
			} else {
				// If labels are not provided, all resources in the namespaces under test are included
				name := getNameFn(&items[i])
				log.Debug("Searching pods in %s %q found in ns %q without label", resourceType, name, ns)
				allResults = append(allResults, items[i])
				log.Info("%s %s found in ns=%s", resourceType, name, ns)
			}
		}
	}

	if len(allResults) == 0 {
		log.Warn("Did not find any %s in the configured namespaces %v", resourceType, namespaces)
	}

	return allResults
}

func listDeployments(appClient appv1client.AppsV1Interface, ns string) ([]appsv1.Deployment, error) {
	dps, err := appClient.Deployments(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return dps.Items, nil
}

func listStatefulSets(appClient appv1client.AppsV1Interface, ns string) ([]appsv1.StatefulSet, error) {
	ss, err := appClient.StatefulSets(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return ss.Items, nil
}

func getDeploymentTemplateLabels(d *appsv1.Deployment) map[string]string {
	return d.Spec.Template.Labels
}

func getDeploymentName(d *appsv1.Deployment) string {
	return d.Name
}

func getStatefulSetTemplateLabels(ss *appsv1.StatefulSet) map[string]string {
	return ss.Spec.Template.Labels
}

func getStatefulSetName(ss *appsv1.StatefulSet) string {
	return ss.Name
}

func FindDeploymentByNameByNamespace(appClient appv1client.AppsV1Interface, namespace, name string) (*appsv1.Deployment, error) {
	dp, err := appClient.Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment %s/%s: %w", namespace, name, err)
	}
	return dp, nil
}
func FindStatefulsetByNameByNamespace(appClient appv1client.AppsV1Interface, namespace, name string) (*appsv1.StatefulSet, error) {
	ss, err := appClient.StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get statefulset %s/%s: %w", namespace, name, err)
	}
	return ss, nil
}

func FindCrObjectByNameByNamespace(scalesGetter scale.ScalesGetter, ns, name string, groupResourceSchema schema.GroupResource) (*scalingv1.Scale, error) {
	crScale, err := scalesGetter.Scales(ns).Get(context.TODO(), groupResourceSchema, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get scale for %s/%s: %w", ns, name, err)
	}
	return crScale, nil
}

func findDeploymentsByLabels(
	appClient appv1client.AppsV1Interface,
	labels []labelObject,
	namespaces []string,
) []appsv1.Deployment {
	return findControllersByLabels(
		appClient,
		labels,
		namespaces,
		"Deployment",
		listDeployments,
		getDeploymentTemplateLabels,
		getDeploymentName,
	)
}

func findStatefulSetsByLabels(
	appClient appv1client.AppsV1Interface,
	labels []labelObject,
	namespaces []string,
) []appsv1.StatefulSet {
	return findControllersByLabels(
		appClient,
		labels,
		namespaces,
		"StatefulSet",
		listStatefulSets,
		getStatefulSetTemplateLabels,
		getStatefulSetName,
	)
}

func findHpaControllers(cs kubernetes.Interface, namespaces []string) []*scalingv1.HorizontalPodAutoscaler {
	var m []*scalingv1.HorizontalPodAutoscaler
	for _, ns := range namespaces {
		hpas, err := cs.AutoscalingV1().HorizontalPodAutoscalers(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Error("Cannot list HorizontalPodAutoscalers on namespace %q, err: %v", ns, err)
			return m
		}
		for i := 0; i < len(hpas.Items); i++ {
			m = append(m, &hpas.Items[i])
		}
	}
	if len(m) == 0 {
		log.Info("Cannot find any deployed HorizontalPodAutoscaler")
	}
	return m
}
