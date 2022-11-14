// Copyright (C) 2022 Red Hat, Inc.
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

package provider

import (
	"context"
	"fmt"
	"strings"

	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Operator struct {
	Name             string                            `yaml:"name" json:"name"`
	Namespace        string                            `yaml:"namespace" json:"namespace"`
	Csv              *olmv1Alpha.ClusterServiceVersion `yaml:"csv" json:"csv"`
	SubscriptionName string                            `yaml:"subscriptionName" json:"subscriptionName"`
	InstallPlans     []CsvInstallPlan                  `yaml:"installPlans,omitempty" json:"installPlans,omitempty"`
	Package          string                            `yaml:"package" json:"package"`
	Org              string                            `yaml:"org" json:"org"`
	Version          string                            `yaml:"version" json:"version"`
	Channel          string                            `yaml:"channel" json:"channel"`
}

type CsvInstallPlan struct {
	// Operator's installPlan name
	Name string `yaml:"name" json:"name"`
	// BundleImage is the URL referencing the bundle image
	BundleImage string `yaml:"bundleImage" json:"bundleImage"`
	// IndexImage is the URL referencing the index image
	IndexImage string `yaml:"indexImage" json:"indexImage"`
}

func (op *Operator) String() string {
	return fmt.Sprintf("csv: %s ns:%s subscription:%s", op.Name, op.Namespace, op.SubscriptionName)
}

//nolint:funlen
func createOperators(csvs []olmv1Alpha.ClusterServiceVersion, subscriptions []olmv1Alpha.Subscription) ([]*Operator, error) {
	installPlans := map[string][]olmv1Alpha.InstallPlan{} // Helper: maps a namespace name to all its installplans.
	operators := []*Operator{}
	for i := range csvs {
		csv := &csvs[i]
		op := &Operator{Name: csv.Name, Namespace: csv.Namespace, Csv: csv}

		packageAndVersion := strings.SplitN(csv.Name, ".", 2) //nolint:gomnd // ok
		op.Version = packageAndVersion[1]

		for s := range subscriptions {
			subscription := &subscriptions[s]
			if subscription.Status.InstalledCSV != csv.Name || subscription.Namespace != csv.Namespace {
				continue
			}

			op.SubscriptionName = subscription.Name
			op.Package = subscription.Spec.Package
			op.Org = subscription.Spec.CatalogSource
			op.Channel = subscription.Spec.Channel
			break
		}

		csvInstallPlans, err := getCsvInstallPlans(csv.Namespace, csv.Name, installPlans)
		if err != nil {
			return nil, fmt.Errorf("failed to get installPlans for csv %s (ns %s), err: %v", csv.Name, csv.Namespace, err)
		}

		for _, installPlan := range csvInstallPlans {
			indexImage, catalogErr := getCatalogSourceImageIndexFromInstallPlan(installPlan)
			if catalogErr != nil {
				return nil, fmt.Errorf("failed to get installPlan image index for csv %s (ns %s) installPlan %s, err: %v",
					csv.Name, csv.Namespace, installPlan.Name, catalogErr)
			}

			op.InstallPlans = append(op.InstallPlans, CsvInstallPlan{
				Name:        installPlan.Name,
				BundleImage: installPlan.Status.BundleLookups[0].Path,
				IndexImage:  indexImage,
			})
		}

		operators = append(operators, op)
	}

	return operators, nil
}

func CsvToString(csv *olmv1Alpha.ClusterServiceVersion) string {
	return fmt.Sprintf("operator csv: %s ns: %s",
		csv.Name,
		csv.Namespace,
	)
}

// getInstallPlansInNamespace is a helper function to get the installPlans in a namespace. The
// map installPlans is used to store them in order to avoid repeating http requests for a namespace
// whose installPlans were already obtained.
func getInstallPlansInNamespace(namespace string, clusterInstallPlans map[string][]olmv1Alpha.InstallPlan) ([]olmv1Alpha.InstallPlan, error) {
	// Check if installplans were stored before.
	nsInstallPlans, exist := clusterInstallPlans[namespace]
	if exist {
		return nsInstallPlans, nil
	}

	clients := clientsholder.GetClientsHolder()
	installPlanList, err := clients.OlmClient.OperatorsV1alpha1().InstallPlans(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable get installplans in namespace %s, err: %v", namespace, err)
	}

	nsInstallPlans = installPlanList.Items
	clusterInstallPlans[namespace] = nsInstallPlans

	return nsInstallPlans, nil
}

// getCsvInstallPlans is a helper function that returns the installPlans for a CSV in a namespace.
// The map clusterInstallPlans is used to store previously retrieved installPlans, in order to save
// http requests.
func getCsvInstallPlans(namespace, csv string, clusterInstallPlans map[string][]olmv1Alpha.InstallPlan) ([]*olmv1Alpha.InstallPlan, error) {
	nsInstallPlans, err := getInstallPlansInNamespace(namespace, clusterInstallPlans)
	if err != nil {
		return nil, err
	}

	installPlans := []*olmv1Alpha.InstallPlan{}
	for i := range nsInstallPlans {
		nsInstallPlan := &nsInstallPlans[i]
		for _, csvName := range nsInstallPlan.Spec.ClusterServiceVersionNames {
			if csv != csvName {
				continue
			}

			if nsInstallPlan.Status.BundleLookups == nil {
				logrus.Warnf("InstallPlan %s for csv %s (ns %s) does not have bundle lookups. It will be skipped.", nsInstallPlan.Name, csv, namespace)
				continue
			}

			installPlans = append(installPlans, nsInstallPlan)
		}
	}

	if len(installPlans) == 0 {
		return nil, fmt.Errorf("no installplans found for csv %s (ns %s)", csv, namespace)
	}

	return installPlans, nil
}

func getCatalogSourceImageIndexFromInstallPlan(installPlan *olmv1Alpha.InstallPlan) (string, error) {
	// ToDo/Technical debt: what to do if installPlan has more than one BundleLookups entries.
	catalogSourceName := installPlan.Status.BundleLookups[0].CatalogSourceRef.Name
	catalogSourceNamespace := installPlan.Status.BundleLookups[0].CatalogSourceRef.Namespace

	clients := clientsholder.GetClientsHolder()
	catalogSource, err := clients.OlmClient.OperatorsV1alpha1().CatalogSources(catalogSourceNamespace).Get(context.TODO(), catalogSourceName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get catalogsource: %s", err)
	}

	return catalogSource.Spec.Image, nil
}
