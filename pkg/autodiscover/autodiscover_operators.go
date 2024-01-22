// Copyright (C) 2020-2023 Red Hat, Inc.
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
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/stringhelper"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

const (
	istioNamespace = "istio-system"
	istioCR        = "installed-state"
)

func isIstioServiceMeshInstalled(allNs []string) bool {
	// the Istio namespace must be present
	if !stringhelper.StringInSlice(allNs, istioNamespace, false) {
		return false
	}

	// the Istio CR used for installation must be present
	oc := clientsholder.GetClientsHolder()
	gvr := schema.GroupVersionResource{Group: "install.istio.io", Version: "v1alpha1", Resource: "istiooperators"}
	cr, err := oc.DynamicClient.Resource(gvr).Namespace(istioNamespace).Get(context.TODO(), istioCR, metav1.GetOptions{})
	if err != nil {
		log.Error("Failed when checking the Istio CR, err: %v", err)
		return false
	}
	if cr == nil {
		log.Warn("The Istio installation CR is missing (but the Istio namespace exists)")
		return false
	}

	log.Info("Istio Service Mesh detected")

	return true
}

func findOperatorsByLabel(olmClient clientOlm.Interface, labels []labelObject, namespaces []configuration.Namespace) []*olmv1Alpha.ClusterServiceVersion {
	csvs := []*olmv1Alpha.ClusterServiceVersion{}
	for _, ns := range namespaces {
		for _, aLabelObject := range labels {
			label := aLabelObject.LabelKey + "=" + aLabelObject.LabelValue
			log.Debug("Searching CSVs in namespace %q with label %q", ns, label)
			csvList, err := olmClient.OperatorsV1alpha1().ClusterServiceVersions(ns.Name).List(context.TODO(), metav1.ListOptions{
				LabelSelector: label,
			})
			if err != nil {
				log.Error("Error when listing csvs in namespace %q with label %q", ns, label)
				continue
			}

			for i := range csvList.Items {
				csvs = append(csvs, &csvList.Items[i])
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
