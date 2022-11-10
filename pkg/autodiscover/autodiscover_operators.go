// Copyright (C) 2020-2022 Red Hat, Inc.
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
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

const (
	istioNamespace = "istio-system"
)

func findIstioNamespace(allNs *[]string) bool {
	for index := range *allNs {
		if (*allNs)[index] == istioNamespace {
			return true
		}
	}
	return false
}
func findOperatorsByLabel(olmClient clientOlm.Interface, labels []configuration.Label, namespaces []configuration.Namespace) []olmv1Alpha.ClusterServiceVersion {
	csvs := []olmv1Alpha.ClusterServiceVersion{}
	for _, ns := range namespaces {
		logrus.Debugf("Searching CSVs in namespace %s", ns)
		for _, label := range labels {
			logrus.Debugf("Searching CSVs with label %+v", label)
			label := buildLabelQuery(label)
			csvList, err := olmClient.OperatorsV1alpha1().ClusterServiceVersions(ns.Name).List(context.TODO(), metav1.ListOptions{
				LabelSelector: label,
			})
			if err != nil {
				logrus.Errorln("error when listing csvs in ns=", ns, " label=", label)
				continue
			}
			csvs = append(csvs, csvList.Items...)
		}
	}

	logrus.Infof("Found %d CSVs:", len(csvs))
	for i := range csvs {
		logrus.Infof(" CSV name: %s (ns: %s)", csvs[i].Name, csvs[i].Namespace)
	}

	return csvs
}
func getAllNamespaces(oc corev1client.CoreV1Interface) (allNs []string, err error) {
	nsList, err := oc.Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorln("Error when listing", "err: ", err)
		return allNs, fmt.Errorf("error getting all namespaces, err: %s", err)
	}
	for index := range nsList.Items {
		allNs = append(allNs, nsList.Items[index].ObjectMeta.Name)
	}
	return allNs, nil
}
func getAllOperators(olmClient clientOlm.Interface) []olmv1Alpha.ClusterServiceVersion {
	csvs := []olmv1Alpha.ClusterServiceVersion{}

	logrus.Debugf("Searching CSVs in namespace All")
	csvList, err := olmClient.OperatorsV1alpha1().ClusterServiceVersions("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorln("error when listing csvs in all namespaces")
	}
	csvs = append(csvs, csvList.Items...)

	logrus.Infof("Found %d CSVs:", len(csvs))
	for i := range csvs {
		logrus.Infof(" CSV name: %s (ns: %s)", csvs[i].Name, csvs[i].Namespace)
	}
	return csvs
}

func findSubscriptions(olmClient clientOlm.Interface, namespaces []string) []olmv1Alpha.Subscription {
	subscriptions := []olmv1Alpha.Subscription{}
	for _, ns := range namespaces {
		displayNs := ns
		if ns == "" {
			displayNs = "All Namespaces"
		}
		logrus.Debugf("Searching subscriptions in namespace %s", displayNs)
		subscription, err := olmClient.OperatorsV1alpha1().Subscriptions(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logrus.Errorln("error when listing subscriptions in ns=", ns)
			continue
		}
		subscriptions = append(subscriptions, subscription.Items...)
	}

	logrus.Infof("Found %d subscriptions in the target namespaces:", len(subscriptions))
	for i := range subscriptions {
		logrus.Infof(" Subscriptions name: %s (ns: %s)", subscriptions[i].Name, subscriptions[i].Namespace)
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
				DebugLog:         logrus.Printf,
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
func getAllInstallPlans(olmClient clientOlm.Interface) (out []*olmv1Alpha.InstallPlan, err error) {
	installPlanList, err := olmClient.OperatorsV1alpha1().InstallPlans("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable get installplans in cluster, err: %s", err)
	}
	for index := range installPlanList.Items {
		out = append(out, &installPlanList.Items[index])
	}
	return out, nil
}

// getAllCatalogSources is a helper function to get the all the CatalogSources in a cluster.
func getAllCatalogSources(olmClient clientOlm.Interface) (out []*olmv1Alpha.CatalogSource, err error) {
	catalogSourcesList, err := olmClient.OperatorsV1alpha1().CatalogSources("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable get CatalogSources in cluster, err: %s", err)
	}
	for index := range catalogSourcesList.Items {
		out = append(out, &catalogSourcesList.Items[index])
	}
	return out, nil
}
