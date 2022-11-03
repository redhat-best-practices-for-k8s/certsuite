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
	"strings"

	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
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
			return nil, fmt.Errorf("failed to get installPlans for csv %s (ns %s), err: %s", csv.Name, csv.Namespace, err)
		}

		for _, installPlan := range csvInstallPlans {
			indexImage, catalogErr := getCatalogSourceImageIndexFromInstallPlan(installPlan)
			if catalogErr != nil {
				return nil, fmt.Errorf("failed to get installPlan image index for csv %s (ns %s) installPlan %s, err: %s",
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
