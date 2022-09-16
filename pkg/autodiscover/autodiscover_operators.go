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

func findIstioNamespace(oc corev1client.CoreV1Interface) bool {
	nsList, err := oc.Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorln("Error when listing", "err: ", err)
		return false
	}
	for index := range nsList.Items {
		if nsList.Items[index].ObjectMeta.Name == istioNamespace {
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
func findSubscriptions(olmClient clientOlm.Interface, namespaces []string) []olmv1Alpha.Subscription {
	subscriptions := []olmv1Alpha.Subscription{}
	for _, ns := range namespaces {
		logrus.Debugf("Searching subscriptions in namespace %s", ns)
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
