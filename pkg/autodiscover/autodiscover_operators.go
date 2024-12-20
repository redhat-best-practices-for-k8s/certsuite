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
	"fmt"

	helmclient "github.com/mittwald/go-helm-client"
	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	clientOlm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"

	olmpkgv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	olmpkgclient "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/client/clientset/versioned/typed/operators/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/stringhelper"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

const (
	istioNamespace      = "istio-system"
	istioDeploymentName = "istiod"
)

func isIstioServiceMeshInstalled(appClient appv1client.AppsV1Interface, allNs []string) bool {
	// The Istio namespace must be present
	if !stringhelper.StringInSlice(allNs, istioNamespace, false) {
		log.Info("Istio Service Mesh not present (the namespace %q does not exists)", istioNamespace)
		return false
	}

	// The Deployment "istiod" must be present in an active service mesh
	_, err := appClient.Deployments(istioNamespace).Get(context.TODO(), istioDeploymentName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		log.Warn("The Istio Deployment %q is missing (but the Istio namespace exists)", istioDeploymentName)
		return false
	} else if err != nil {
		log.Error("Failed getting Deployment %q", istioDeploymentName)
		return false
	}

	log.Info("Istio Service Mesh detected")

	return true
}

func findOperatorsMatchingAtLeastOneLabel(olmClient clientOlm.Interface, labels []labelObject, namespace configuration.Namespace) *olmv1Alpha.ClusterServiceVersionList {
	csvList := &olmv1Alpha.ClusterServiceVersionList{}
	for _, l := range labels {
		log.Debug("Searching CSVs in namespace %q with label %q", namespace, l)
		csv, err := olmClient.OperatorsV1alpha1().ClusterServiceVersions(namespace.Name).List(context.TODO(), metav1.ListOptions{
			LabelSelector: l.LabelKey + "=" + l.LabelValue,
		})
		if err != nil {
			log.Error("Error when listing csvs in namespace %q with label %q, err: %v", namespace, l.LabelKey+"="+l.LabelValue, err)
			continue
		}
		csvList.Items = append(csvList.Items, csv.Items...)
	}
	return csvList
}

func findOperatorsByLabels(olmClient clientOlm.Interface, labels []labelObject, namespaces []configuration.Namespace) (csvs []*olmv1Alpha.ClusterServiceVersion) {
	const nsAnnotation = "olm.operatorNamespace"

	// Helper namespaces map to do quick search of the operator's controller namespace.
	namespacesMap := map[string]bool{}
	for _, ns := range namespaces {
		namespacesMap[ns.Name] = true
	}

	csvs = []*olmv1Alpha.ClusterServiceVersion{}
	var csvList *olmv1Alpha.ClusterServiceVersionList
	for _, ns := range namespaces {
		if len(labels) > 0 {
			csvList = findOperatorsMatchingAtLeastOneLabel(olmClient, labels, ns)
		} else {
			// If labels are not provided in the namespace under test, they are tested by the CNF suite
			log.Debug("Searching CSVs in namespace %s without label", ns)
			var err error
			csvList, err = olmClient.OperatorsV1alpha1().ClusterServiceVersions(ns.Name).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				log.Error("Error when listing csvs in namespace %q , err: %v", ns, err)
				continue
			}
		}
		for i := range csvList.Items {
			csv := &csvList.Items[i]

			// Filter out CSV if operator's controller pod/s is/are not running in any configured/test namespace.
			controllerNamespace, found := csv.Annotations[nsAnnotation]
			if !found {
				log.Error("Failed to get ns annotation %q from csv %v/%v", nsAnnotation, csv.Namespace, csv.Name)
				continue
			}

			if namespacesMap[controllerNamespace] {
				csvs = append(csvs, csv)
			}
		}
	}
	for i := range csvs {
		log.Info("Found CSV %q (namespace %q)", csvs[i].Name, csvs[i].Namespace)
	}
	return csvs
}

func getAllNamespaces(oc corev1client.CoreV1Interface) (allNs []string, err error) {
	nsList, err := oc.Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return allNs, fmt.Errorf("error getting all namespaces, err: %v", err)
	}
	for index := range nsList.Items {
		allNs = append(allNs, nsList.Items[index].ObjectMeta.Name)
	}
	return allNs, nil
}
func getAllOperators(olmClient clientOlm.Interface) ([]*olmv1Alpha.ClusterServiceVersion, error) {
	csvs := []*olmv1Alpha.ClusterServiceVersion{}

	csvList, err := olmClient.OperatorsV1alpha1().ClusterServiceVersions("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error when listing CSVs in all namespaces, err: %v", err)
	}
	for i := range csvList.Items {
		csvs = append(csvs, &csvList.Items[i])
	}

	for i := range csvs {
		log.Info("Found CSV %q (ns %q)", csvs[i].Name, csvs[i].Namespace)
	}
	return csvs, nil
}

func findSubscriptions(olmClient clientOlm.Interface, namespaces []string) []olmv1Alpha.Subscription {
	subscriptions := []olmv1Alpha.Subscription{}
	for _, ns := range namespaces {
		displayNs := ns
		if ns == "" {
			displayNs = "All Namespaces"
		}
		log.Debug("Searching subscriptions in namespace %q", displayNs)
		subscription, err := olmClient.OperatorsV1alpha1().Subscriptions(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Error("Error when listing subscriptions in namespace %q", ns)
			continue
		}
		subscriptions = append(subscriptions, subscription.Items...)
	}

	for i := range subscriptions {
		log.Info("Found subscription %q (ns %q)", subscriptions[i].Name, subscriptions[i].Namespace)
	}
	return subscriptions
}

func getHelmList(restConfig *rest.Config, namespaces []string) map[string][]*release.Release {
	helmChartReleases := map[string][]*release.Release{}
	for _, ns := range namespaces {
		opt := &helmclient.RestConfClientOptions{
			Options: &helmclient.Options{
				Namespace:        ns,
				RepositoryCache:  "/tmp/.helmcache",
				RepositoryConfig: "/tmp/.helmrepo",
				Debug:            true,
				Linting:          true,
				DebugLog:         log.Info,
			},
			RestConfig: restConfig,
		}

		helmClient, err := helmclient.NewClientFromRestConf(opt)
		if err != nil {
			panic(err)
		}
		nsHelmchartreleases, _ := helmClient.ListDeployedReleases()
		helmChartReleases[ns] = nsHelmchartreleases
	}
	return helmChartReleases
}

// getAllInstallPlans is a helper function to get the all the installPlans in a cluster.
func getAllInstallPlans(olmClient clientOlm.Interface) (out []*olmv1Alpha.InstallPlan) {
	installPlanList, err := olmClient.OperatorsV1alpha1().InstallPlans("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error("Unable get installplans in cluster, err: %v", err)
		return out
	}
	for index := range installPlanList.Items {
		out = append(out, &installPlanList.Items[index])
	}
	return out
}

// getAllCatalogSources is a helper function to get the all the CatalogSources in a cluster.
func getAllCatalogSources(olmClient clientOlm.Interface) (out []*olmv1Alpha.CatalogSource) {
	catalogSourcesList, err := olmClient.OperatorsV1alpha1().CatalogSources("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error("Unable get CatalogSources in cluster, err: %v", err)
		return out
	}
	for index := range catalogSourcesList.Items {
		out = append(out, &catalogSourcesList.Items[index])
	}
	return out
}

// getAllPackageManifests is a helper function to get the all the PackageManifests in a cluster.
func getAllPackageManifests(olmPkgClient olmpkgclient.OperatorsV1Interface) (out []*olmpkgv1.PackageManifest) {
	packageManifestsList, err := olmPkgClient.PackageManifests("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error("Unable get Package Manifests in cluster, err: %v", err)
		return out
	}
	for index := range packageManifestsList.Items {
		out = append(out, &packageManifestsList.Items[index])
	}
	return out
}
