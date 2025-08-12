// Copyright (C) 2020-2024 Red Hat, Inc.
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

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	appsv1 "k8s.io/api/apps/v1"
	scalingv1 "k8s.io/api/autoscaling/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	appv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/scale"
)

// FindDeploymentByNameByNamespace retrieves a Deployment by its name and namespace.
//
// It takes an AppsV1Interface client, the deployment name, and the namespace,
// and returns a pointer to the Deployment object or an error if not found or
// if any API call fails. The function uses the client's Deployments() method
// to perform a Get request for the specified resource.
func FindDeploymentByNameByNamespace(appClient appv1client.AppsV1Interface, namespace, name string) (*appsv1.Deployment, error) {
	dp, err := appClient.Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Error("Cannot retrieve deployment in ns=%s name=%s", namespace, name)
		return nil, err
	}
	return dp, nil
}

// FindStatefulsetByNameByNamespace retrieves a StatefulSet from the given namespace by name.
//
// It accepts a Kubernetes AppsV1Interface client, the namespace string,
// and the name of the desired StatefulSet. The function uses the client's
// StatefulSets() method to perform a Get operation for the specified
// resource. If found, it returns a pointer to the appsv1.StatefulSet object
// and a nil error; otherwise it returns nil and an error describing the failure.
func FindStatefulsetByNameByNamespace(appClient appv1client.AppsV1Interface, namespace, name string) (*appsv1.StatefulSet, error) {
	ss, err := appClient.StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Error("Cannot retrieve deployment in ns=%s name=%s", namespace, name)
		return nil, err
	}
	return ss, nil
}

// FindCrObjectByNameByNamespace retrieves a scaling object by name and namespace.
//
// It accepts a ScalesGetter client, the target namespace, the CR name,
// and the GroupResource describing the custom resource type.
// The function returns a pointer to a Scale object representing the
// current scale of the specified custom resource and an error if the
// operation fails. If the requested scaling object cannot be found,
// the returned error will describe the failure reason.
func FindCrObjectByNameByNamespace(scalesGetter scale.ScalesGetter, ns, name string, groupResourceSchema schema.GroupResource) (*scalingv1.Scale, error) {
	crScale, err := scalesGetter.Scales(ns).Get(context.TODO(), groupResourceSchema, name, metav1.GetOptions{})
	if err != nil {
		log.Error("Cannot retrieve deployment in ns=%s name=%s", ns, name)
		return nil, err
	}
	return crScale, nil
}

// isDeploymentsPodsMatchingAtLeastOneLabel checks if any pod in a deployment matches at least one label from a set.
//
// It receives a slice of label objects, the namespace of the deployment, and a pointer to the deployment object.
// The function iterates over the pods belonging to the specified deployment, comparing each pod's labels against
// the provided list. If it finds a pod that contains any of the given labels, it returns true; otherwise,
// it returns false. Logging is performed at debug and info levels to trace the evaluation process.
func isDeploymentsPodsMatchingAtLeastOneLabel(labels []labelObject, namespace string, deployment *appsv1.Deployment) bool {
	for _, aLabelObject := range labels {
		log.Debug("Searching pods in deployment %q found in ns %q using label %s=%s", deployment.Name, namespace, aLabelObject.LabelKey, aLabelObject.LabelValue)
		if deployment.Spec.Template.Labels[aLabelObject.LabelKey] == aLabelObject.LabelValue {
			log.Info("Deployment %s found in ns=%s", deployment.Name, namespace)
			return true
		}
	}
	return false
}

// Find deployments that match given label criteria in a Kubernetes cluster.
//
// It accepts an AppsV1 client, a slice of label objects to match against,
// and a list of namespace names. The function lists all deployments
// across the specified namespaces, filters them by checking if at least
// one pod within each deployment contains any of the provided labels,
// and returns the matching deployments. If no matches are found or an
// error occurs during listing, it logs appropriate warnings and
// returns an empty slice. The returned slice is a list of appsv1.Deployment objects that satisfy the label conditions.
func findDeploymentsByLabels(
	appClient appv1client.AppsV1Interface,
	labels []labelObject,
	namespaces []string,
) []appsv1.Deployment {
	allDeployments := []appsv1.Deployment{}
	for _, ns := range namespaces {
		dps, err := appClient.Deployments(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Error("Failed to list deployments in ns=%s, err: %v . Trying to proceed.", ns, err)
			continue
		}
		if len(dps.Items) == 0 {
			log.Warn("Did not find any deployments in ns=%s", ns)
		}
		for i := 0; i < len(dps.Items); i++ {
			if len(labels) > 0 {
				// The deployment is added only once if at least one pod matches one label in the Deployment
				if isDeploymentsPodsMatchingAtLeastOneLabel(labels, ns, &dps.Items[i]) {
					allDeployments = append(allDeployments, dps.Items[i])
					continue
				}
			} else {
				// If labels are not provided, all deployments in the namespaces under test, are tested by the CNF suite
				log.Debug("Searching pods in deployment %q found in ns %q without label", dps.Items[i].Name, ns)
				allDeployments = append(allDeployments, dps.Items[i])
				log.Info("Deployment %s found in ns=%s", dps.Items[i].Name, ns)
			}
		}
	}
	if len(allDeployments) == 0 {
		log.Warn("Did not find any deployment in the configured namespaces %v", namespaces)
	}
	return allDeployments
}

// isStatefulSetsMatchingAtLeastOneLabel checks whether a StatefulSet contains at least one label that matches any of the provided label objects.
//
// It accepts a slice of labelObject, a key to look for on the StatefulSet's labels, and a pointer to the StatefulSet.
// The function returns true if the StatefulSet has the specified key and its value satisfies at least one of the label objects' criteria; otherwise it returns false.
func isStatefulSetsMatchingAtLeastOneLabel(labels []labelObject, namespace string, statefulSet *appsv1.StatefulSet) bool {
	for _, aLabelObject := range labels {
		log.Debug("Searching pods in statefulset %q found in ns %q using label %s=%s", statefulSet.Name, namespace, aLabelObject.LabelKey, aLabelObject.LabelValue)
		if statefulSet.Spec.Template.Labels[aLabelObject.LabelKey] == aLabelObject.LabelValue {
			log.Info("StatefulSet %s found in ns=%s", statefulSet.Name, namespace)
			return true
		}
	}
	return false
}

// findStatefulSetsByLabels retrieves StatefulSet objects that match any of the supplied label selectors.
//
// It queries the Kubernetes API for all StatefulSets in the cluster, then filters
// them to include only those whose labels satisfy at least one of the provided
// labelObject criteria. The function accepts an AppsV1 client interface,
// a slice of labelObject structures representing desired label matches, and a
// list of namespace strings to limit the search. It returns a slice of StatefulSet
// objects that meet the criteria or an empty slice if none are found. Errors
// encountered during listing are logged but do not halt execution; only a log
// entry is produced for each failure.
func findStatefulSetsByLabels(
	appClient appv1client.AppsV1Interface,
	labels []labelObject,
	namespaces []string,
) []appsv1.StatefulSet {
	allStatefulSets := []appsv1.StatefulSet{}
	for _, ns := range namespaces {
		statefulSet, err := appClient.StatefulSets(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Error("Failed to list statefulsets in ns=%s, err: %v . Trying to proceed.", ns, err)
			continue
		}
		if len(statefulSet.Items) == 0 {
			log.Warn("Did not find any statefulSet in ns=%s", ns)
		}
		for i := 0; i < len(statefulSet.Items); i++ {
			if len(labels) > 0 {
				// The StatefulSet is added only once if at least one pod matches one label in the Statefulset
				if isStatefulSetsMatchingAtLeastOneLabel(labels, ns, &statefulSet.Items[i]) {
					allStatefulSets = append(allStatefulSets, statefulSet.Items[i])
					continue
				}
			} else {
				// If labels are not provided, all statefulsets in the namespaces under test, are tested by the CNF suite
				log.Debug("Searching pods in statefulset %q found in ns %q without label", statefulSet.Items[i].Name, ns)
				allStatefulSets = append(allStatefulSets, statefulSet.Items[i])
				log.Info("StatefulSet %s found in ns=%s", statefulSet.Items[i].Name, ns)
			}
		}
	}
	if len(allStatefulSets) == 0 {
		log.Warn("Did not find any statefulset in the configured namespaces %v", namespaces)
	}
	return allStatefulSets
}

// findHpaControllers retrieves all HorizontalPodAutoscaler objects that manage the given pod sets.
//
// It accepts a Kubernetes client interface and a slice of pod set names. The function queries
// the autoscaling/v1 API for HorizontalPodAutoscalers in all namespaces, filters those whose
// scale target references match any of the provided pod set names, and returns a slice of
// pointers to the matching HPA objects. If no HPAs are found or an error occurs during listing,
// it logs relevant information and returns an empty slice.
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
