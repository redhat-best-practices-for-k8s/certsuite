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
	"fmt"
	"sort"
	"strings"

	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/sirupsen/logrus"
)

type Operator struct {
	Name               string                            `yaml:"name" json:"name"`
	Namespace          string                            `yaml:"namespace" json:"namespace"`
	Csv                *olmv1Alpha.ClusterServiceVersion `yaml:"csv" json:"csv"`
	SubscriptionName   string                            `yaml:"subscriptionName" json:"subscriptionName"`
	InstallPlans       []CsvInstallPlan                  `yaml:"installPlans,omitempty" json:"installPlans,omitempty"`
	Package            string                            `yaml:"package" json:"package"`
	Org                string                            `yaml:"org" json:"org"`
	Version            string                            `yaml:"version" json:"version"`
	Channel            string                            `yaml:"channel" json:"channel"`
	PackageFromCsvName string                            `yaml:"packagefromcsvname" json:"packagefromcsvname"`
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

// TODO: Fix lint properly
//
//nolint:funlen,gocyclo,lll
func createOperators(csvs []olmv1Alpha.ClusterServiceVersion, subscriptions []olmv1Alpha.Subscription, allInstallPlans []*olmv1Alpha.InstallPlan, allCatalogSources []*olmv1Alpha.CatalogSource, catalogSourceNotRequired, succeededRequired bool) []*Operator {
	const (
		maxSize = 2
	)

	operators := []*Operator{}
	for i := range csvs {
		csv := &csvs[i]
		op := &Operator{Name: csv.Name, Namespace: csv.Namespace, Csv: csv}

		packageAndVersion := strings.SplitN(csv.Name, ".", maxSize)
		if len(packageAndVersion) == 0 {
			continue
		}
		op.PackageFromCsvName = packageAndVersion[0]
		op.Version = csv.Spec.Version.String()

		atLeastOneSubscription := false
		for s := range subscriptions {
			subscription := &subscriptions[s]
			if subscription.Status.InstalledCSV != csv.Name || subscription.Namespace != csv.Namespace {
				continue
			}

			op.SubscriptionName = subscription.Name
			op.Package = subscription.Spec.Package
			op.Org = subscription.Spec.CatalogSource
			op.Channel = subscription.Spec.Channel
			atLeastOneSubscription = true
			break
		}
		if !atLeastOneSubscription {
			logrus.Warnf("Subscription not found for csv %s (ns %s) This Operator will not receive updates.", csv.Name, csv.Namespace)
		}

		atLeastOneInstallPlan := false
		for _, installPlan := range allInstallPlans {
			if installPlan.Namespace != csv.Namespace {
				continue
			}
			alLeastOneCsv := false
			for _, csvName := range installPlan.Spec.ClusterServiceVersionNames {
				if csv.Name != csvName {
					continue
				}

				if installPlan.Status.BundleLookups == nil {
					logrus.Warnf("InstallPlan %s for csv %s (ns %s) does not have bundle lookups. It will be skipped.", installPlan.Name, csv.Name, csv.Namespace)
					continue
				}
				alLeastOneCsv = true
				break
			}
			if !alLeastOneCsv {
				continue
			}
			indexImage, catalogErr := getCatalogSourceImageIndexFromInstallPlan(installPlan, allCatalogSources)
			if catalogErr != nil {
				logrus.Tracef("failed to get installPlan image index for csv %s (ns %s) installPlan %s, err: %v",
					csv.Name, csv.Namespace, installPlan.Name, catalogErr)
				continue
			}

			op.InstallPlans = append(op.InstallPlans, CsvInstallPlan{
				Name:        installPlan.Name,
				BundleImage: installPlan.Status.BundleLookups[0].Path,
				IndexImage:  indexImage,
			})
			atLeastOneInstallPlan = true
			break
		}
		if !atLeastOneInstallPlan {
			logrus.Warnf("InstallPlan with BundleLookups not found for csv %s (ns %s) not present. Catalog source not available", csv.Name, csv.Namespace)
		}
		if !(atLeastOneInstallPlan || catalogSourceNotRequired) {
			continue
		}
		if !(csv.Status.Phase == olmv1Alpha.CSVPhaseSucceeded || !succeededRequired) {
			continue
		}
		operators = append(operators, op)
	}
	return operators
}

func CsvToString(csv *olmv1Alpha.ClusterServiceVersion) string {
	return fmt.Sprintf("operator csv: %s ns: %s",
		csv.Name,
		csv.Namespace,
	)
}

func getSummaryAllOperators(operators []*Operator) (summary []string) {
	for _, o := range operators {
		targetNamespaces := ""
		value, ok := o.Csv.ObjectMeta.Annotations["olm.targetNamespaces"]

		if !ok || value == "" {
			targetNamespaces = "All namespaces"
		} else {
			targetNamespaces = value + " Single namespace"
		}

		summary = append(summary, string(o.Csv.Status.Phase)+" operator: "+o.PackageFromCsvName+" ver: "+o.Version+" in ns: "+o.Namespace+" ( "+targetNamespaces+" managed )")
	}
	return summary
}

func getShortSummaryAllOperators(operators []*Operator) (summary []string) {
	operatorMap := map[string]bool{}
	for _, o := range operators {
		targetNamespaces := ""
		value, ok := o.Csv.ObjectMeta.Annotations["olm.targetNamespaces"]

		if !ok || value == "" {
			targetNamespaces = " ( All namespaces managed )"
		} else {
			targetNamespaces = " in ns: " + o.Namespace + " ( Single namespace )"
		}

		operatorMap[string(o.Csv.Status.Phase)+" operator: "+o.PackageFromCsvName+" ver: "+o.Version+targetNamespaces] = true
	}
	for s := range operatorMap {
		summary = append(summary, s)
	}
	sort.Strings(summary)
	return summary
}

func getCatalogSourceImageIndexFromInstallPlan(installPlan *olmv1Alpha.InstallPlan, allCatalogSources []*olmv1Alpha.CatalogSource) (string, error) {
	// ToDo/Technical debt: what to do if installPlan has more than one BundleLookups entries.
	catalogSourceName := installPlan.Status.BundleLookups[0].CatalogSourceRef.Name
	catalogSourceNamespace := installPlan.Status.BundleLookups[0].CatalogSourceRef.Namespace

	for _, s := range allCatalogSources {
		if s.Namespace == catalogSourceNamespace && s.Name == catalogSourceName {
			return s.Spec.Image, nil
		}
	}

	return "", fmt.Errorf("failed to get catalogsource: not found")
}
