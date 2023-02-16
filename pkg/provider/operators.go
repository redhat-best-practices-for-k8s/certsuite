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
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/redhat-openshift-ecosystem/openshift-preflight/artifacts"
	plibRuntime "github.com/redhat-openshift-ecosystem/openshift-preflight/certification"
	plibOperator "github.com/redhat-openshift-ecosystem/openshift-preflight/operator"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
)

const (
	allns = "All namespaces"
)

type Operator struct {
	Name               string                                `yaml:"name" json:"name"`
	Namespace          string                                `yaml:"namespace" json:"namespace"`
	TargetNamespace    string                                `yaml:"targetnamespace" json:"targetnamespace"`
	Csv                *olmv1Alpha.ClusterServiceVersion     `yaml:"csv,omitempty" json:"csv,omitempty"`
	Phase              olmv1Alpha.ClusterServiceVersionPhase `yaml:"csvphase" json:"csvphase"`
	SubscriptionName   string                                `yaml:"subscriptionName" json:"subscriptionName"`
	InstallPlans       []CsvInstallPlan                      `yaml:"installPlans,omitempty" json:"installPlans,omitempty"`
	Package            string                                `yaml:"package" json:"package"`
	Org                string                                `yaml:"org" json:"org"`
	Version            string                                `yaml:"version" json:"version"`
	Channel            string                                `yaml:"channel" json:"channel"`
	PackageFromCsvName string                                `yaml:"packagefromcsvname" json:"packagefromcsvname"`
	PreflightResults   plibRuntime.Results
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
	return fmt.Sprintf("csv: %s ns:%s subscription:%s targetNamespace=%s", op.Name, op.Namespace, op.SubscriptionName, op.TargetNamespace)
}

//nolint:funlen
func (op *Operator) SetPreflightResults(env *TestEnvironment) error {
	bundleImage := op.InstallPlans[0].BundleImage
	indexImage := op.InstallPlans[0].IndexImage
	oc := clientsholder.GetClientsHolder()

	// Create artifacts handler
	artifactsWriter, err := artifacts.NewMapWriter()
	if err != nil {
		return err
	}
	ctx := artifacts.ContextWithWriter(context.Background(), artifactsWriter)
	opts := []plibOperator.Option{}
	opts = append(opts, plibOperator.WithDockerConfigJSONFromFile(env.GetDockerConfigFile()))
	if env.IsPreflightInsecureAllowed() {
		logrus.Info("Insecure connections are being allowed to preflight")
		opts = append(opts, plibOperator.WithInsecureConnection())
	}

	// Add logger output to the context
	logbytes := bytes.NewBuffer([]byte{})
	checklogger := log.Default()
	checklogger.SetOutput(logbytes)
	logger := stdr.New(checklogger)
	ctx = logr.NewContext(ctx, logger)

	check := plibOperator.NewCheck(bundleImage, indexImage, oc.KubeConfig, opts...)
	results, err := check.Run(ctx)
	if err != nil {
		return err
	}

	// Take all of the preflight logs and stick them into logrus.
	logrus.Info(logbytes.String())

	e := os.RemoveAll("artifacts/")
	if e != nil {
		logrus.Fatal(e)
	}

	logrus.Infof("Storing operator preflight results into object for %s", bundleImage)
	op.PreflightResults = results
	return nil
}

//nolint:funlen // adding 1 log 26 > 25
func createOperators(csvs []olmv1Alpha.ClusterServiceVersion,
	subscriptions []olmv1Alpha.Subscription,
	allInstallPlans []*olmv1Alpha.InstallPlan,
	allCatalogSources []*olmv1Alpha.CatalogSource,
	catalogSourceNotRequired,
	succeededRequired,
	keepCsvDetails bool) []*Operator {
	const (
		maxSize = 2
	)

	operators := []*Operator{}
	for i := range csvs {
		csv := &csvs[i]
		if !(csv.Status.Phase == olmv1Alpha.CSVPhaseSucceeded || !succeededRequired) {
			continue
		}
		op := &Operator{Name: csv.Name, Namespace: csv.Namespace}
		if keepCsvDetails {
			op.Csv = csv
		}
		op.Phase = csv.Status.Phase
		packageAndVersion := strings.SplitN(csv.Name, ".", maxSize)
		if len(packageAndVersion) == 0 {
			logrus.Tracef("Empty CSV Name (package.version), cannot extract a package or a version, skipping. Csv: %+v", *csv)
			continue
		}
		op.PackageFromCsvName = packageAndVersion[0]
		op.Version = csv.Spec.Version.String()
		// Get at least one subscription and update the Operator object with them. Not needed for operator tests to pass or to be part of properly installed cluster operators.
		if !getAtLeastOneSubscription(op, csv, subscriptions) {
			logrus.Tracef("Subscription not found for csv %s (ns %s) This Operator will not receive updates.", csv.Name, csv.Namespace)
		}
		// Get at least one Install Plan and update the Operator object with them. Needed to pass tests (including catalog source) but not needed to be part of properly installed cluster operators.
		atLeastOneInstallPlan := getAtLeastOneInstallPlan(op, csv, allInstallPlans, allCatalogSources)
		if !atLeastOneInstallPlan {
			logrus.Tracef("InstallPlan with BundleLookups not found for csv %s (ns %s) not present. Catalog source not available", csv.Name, csv.Namespace)
		}
		if !(atLeastOneInstallPlan || catalogSourceNotRequired) {
			continue
		}
		op.TargetNamespace = getTargetNamespace(csv)
		operators = append(operators, op)
	}
	return operators
}

func getAtLeastOneSubscription(op *Operator, csv *olmv1Alpha.ClusterServiceVersion, subscriptions []olmv1Alpha.Subscription) (atLeastOneSubscription bool) {
	atLeastOneSubscription = false
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
	return atLeastOneSubscription
}

func getAtLeastOneCsv(csv *olmv1Alpha.ClusterServiceVersion, installPlan *olmv1Alpha.InstallPlan) (atLeastOneCsv bool) {
	atLeastOneCsv = false
	for _, csvName := range installPlan.Spec.ClusterServiceVersionNames {
		if csv.Name != csvName {
			continue
		}

		if installPlan.Status.BundleLookups == nil {
			logrus.Warnf("InstallPlan %s for csv %s (ns %s) does not have bundle lookups. It will be skipped.", installPlan.Name, csv.Name, csv.Namespace)
			continue
		}
		atLeastOneCsv = true
		break
	}
	return atLeastOneCsv
}

func getAtLeastOneInstallPlan(op *Operator, csv *olmv1Alpha.ClusterServiceVersion, allInstallPlans []*olmv1Alpha.InstallPlan, allCatalogSources []*olmv1Alpha.CatalogSource) (atLeastOneInstallPlan bool) {
	atLeastOneInstallPlan = false
	for _, installPlan := range allInstallPlans {
		if installPlan.Namespace != csv.Namespace {
			continue
		}

		// If the install plan does not deploys this CSV, check the next one
		if !getAtLeastOneCsv(csv, installPlan) {
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
	}
	return atLeastOneInstallPlan
}

func CsvToString(csv *olmv1Alpha.ClusterServiceVersion) string {
	return fmt.Sprintf("operator csv: %s ns: %s",
		csv.Name,
		csv.Namespace,
	)
}

func getSummaryAllOperators(operators []*Operator) (summary []string) {
	operatorMap := map[string]bool{}
	for _, o := range operators {
		if o.TargetNamespace == allns {
			operatorMap[string(o.Phase)+" operator: "+o.PackageFromCsvName+" ver: "+o.Version+" ( "+o.TargetNamespace+" )"] = true
		} else {
			operatorMap[string(o.Phase)+" operator: "+o.PackageFromCsvName+" ver: "+o.Version+" in ns: "+o.Namespace+" ( "+o.TargetNamespace+" )"] = true
		}
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

func getTargetNamespace(csv *olmv1Alpha.ClusterServiceVersion) (targetNamespaces string) {
	value, ok := csv.ObjectMeta.Annotations["olm.targetNamespaces"]

	if !ok || value == "" {
		targetNamespaces = allns
	} else {
		targetNamespaces = value + " Single namespace"
	}
	return targetNamespaces
}
