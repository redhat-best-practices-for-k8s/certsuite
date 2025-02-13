// Copyright (C) 2022-2024 Red Hat, Inc.
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
	"errors"
	"fmt"
	defaultLog "log"
	"os"
	"sort"
	"strings"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	olmv1 "github.com/operator-framework/api/pkg/operators/v1"
	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmpkgv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-openshift-ecosystem/openshift-preflight/artifacts"
	plibRuntime "github.com/redhat-openshift-ecosystem/openshift-preflight/certification"
	plibOperator "github.com/redhat-openshift-ecosystem/openshift-preflight/operator"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type Operator struct {
	Name                  string                                `yaml:"name" json:"name"`
	Namespace             string                                `yaml:"namespace" json:"namespace"`
	TargetNamespaces      []string                              `yaml:"targetNamespaces" json:"targetNamespaces,omitempty"`
	IsClusterWide         bool                                  `yaml:"isClusterWide" json:"isClusterWide"`
	Csv                   *olmv1Alpha.ClusterServiceVersion     `yaml:"csv,omitempty" json:"csv,omitempty"`
	Phase                 olmv1Alpha.ClusterServiceVersionPhase `yaml:"csvphase" json:"csvphase"`
	SubscriptionName      string                                `yaml:"subscriptionName" json:"subscriptionName"`
	SubscriptionNamespace string                                `yaml:"subscriptionNamespace" json:"subscriptionNamespace"`
	InstallPlans          []CsvInstallPlan                      `yaml:"installPlans,omitempty" json:"installPlans,omitempty"`
	Package               string                                `yaml:"package" json:"package"`
	Org                   string                                `yaml:"org" json:"org"`
	Version               string                                `yaml:"version" json:"version"`
	Channel               string                                `yaml:"channel" json:"channel"`
	PackageFromCsvName    string                                `yaml:"packagefromcsvname" json:"packagefromcsvname"`
	PreflightResults      PreflightResultsDB                    `yaml:"operandPods" json:"operandPods"`
	OperandPods           map[string]*Pod
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
	return fmt.Sprintf("csv: %s ns:%s subscription:%s targetNamespaces=%v", op.Name, op.Namespace, op.SubscriptionName, op.TargetNamespaces)
}

func (op *Operator) SetPreflightResults(env *TestEnvironment) error {
	if len(op.InstallPlans) == 0 {
		log.Warn("Operator %q has no InstallPlans. Skipping setting Preflight results", op)
		return nil
	}

	bundleImage := op.InstallPlans[0].BundleImage
	indexImage := op.InstallPlans[0].IndexImage
	oc := clientsholder.GetClientsHolder()

	// Create artifacts handler
	artifactsWriter, err := artifacts.NewMapWriter()
	if err != nil {
		return err
	}
	ctx := artifacts.ContextWithWriter(context.TODO(), artifactsWriter)
	opts := []plibOperator.Option{}
	opts = append(opts, plibOperator.WithDockerConfigJSONFromFile(env.GetDockerConfigFile()))
	if env.IsPreflightInsecureAllowed() {
		log.Info("Insecure connections are being allowed to Preflight")
		opts = append(opts, plibOperator.WithInsecureConnection())
	}

	// Add logger output to the context
	logbytes := bytes.NewBuffer([]byte{})
	checklogger := defaultLog.Default()
	checklogger.SetOutput(logbytes)
	logger := stdr.New(checklogger)
	ctx = logr.NewContext(ctx, logger)

	check := plibOperator.NewCheck(bundleImage, indexImage, oc.KubeConfig, opts...)

	results, runtimeErr := check.Run(ctx)
	if runtimeErr != nil {
		_, checks, err := check.List(ctx)
		if err != nil {
			return fmt.Errorf("could not get preflight container test list")
		}
		for _, c := range checks {
			results.PassedOverall = false
			result := plibRuntime.Result{Check: c, ElapsedTime: 0}
			results.Errors = append(results.Errors, *result.WithError(runtimeErr))
		}
	}

	// Take all of the preflight logs and stick them into our log.
	log.Info("%s", logbytes.String())

	e := os.RemoveAll("artifacts/")
	if e != nil {
		log.Fatal("%v", e)
	}

	log.Info("Storing operator Preflight results into object for %q", bundleImage)
	op.PreflightResults = GetPreflightResultsDB(&results)
	return nil
}

// getUniqueCsvListByName returns a CSV list with unique names from a list which may contain
// more than one CSV with the same name. The output CSV list is sorted by CSV name.
func getUniqueCsvListByName(csvs []*olmv1Alpha.ClusterServiceVersion) []*olmv1Alpha.ClusterServiceVersion {
	uniqueCsvsMap := map[string]*olmv1Alpha.ClusterServiceVersion{}
	for _, csv := range csvs {
		uniqueCsvsMap[csv.Name] = csv
	}

	uniqueCsvsList := []*olmv1Alpha.ClusterServiceVersion{}
	log.Info("Found %d unique CSVs", len(uniqueCsvsMap))
	for name, csv := range uniqueCsvsMap {
		log.Info("  CSV: %s", name)
		uniqueCsvsList = append(uniqueCsvsList, csv)
	}

	// Sort by name: (1) creates a deterministic output, (2) makes UT easier.
	sort.Slice(uniqueCsvsList, func(i, j int) bool { return uniqueCsvsList[i].Name < uniqueCsvsList[j].Name })
	return uniqueCsvsList
}

func createOperators(csvs []*olmv1Alpha.ClusterServiceVersion,
	allSubscriptions []olmv1Alpha.Subscription,
	allPackageManifests []*olmpkgv1.PackageManifest,
	allInstallPlans []*olmv1Alpha.InstallPlan,
	allCatalogSources []*olmv1Alpha.CatalogSource,
	succeededRequired,
	keepCsvDetails bool) []*Operator {
	const (
		maxSize = 2
	)

	operators := []*Operator{}

	// Make map with unique csv names to original index in the env.Csvs slice.
	// Otherwise, cluster-wide operators info will be repeated unnecessarily.
	uniqueCsvs := getUniqueCsvListByName(csvs)

	for _, csv := range uniqueCsvs {
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
			log.Debug("Empty CSV Name (package.version), cannot extract a package or a version, skipping. Csv: %+v", csv)
			continue
		}
		op.PackageFromCsvName = packageAndVersion[0]
		op.Version = csv.Spec.Version.String()
		// Get at least one subscription and update the Operator object with it.
		if getAtLeastOneSubscription(op, csv, allSubscriptions, allPackageManifests) {
			targetNamespaces, err := getOperatorTargetNamespaces(op.SubscriptionNamespace)
			if err != nil {
				log.Error("Failed to get target namespaces for operator %s: %v", csv.Name, err)
			} else {
				op.TargetNamespaces = targetNamespaces
				op.IsClusterWide = len(targetNamespaces) == 0
			}
		} else {
			log.Warn("Subscription not found for CSV: %s (ns %s)", csv.Name, csv.Namespace)
		}
		log.Info("Getting installplans for op %s (subs %s ns %s)", op.Name, op.SubscriptionName, op.SubscriptionNamespace)
		// Get at least one Install Plan and update the Operator object with it.
		getAtLeastOneInstallPlan(op, csv, allInstallPlans, allCatalogSources)
		operators = append(operators, op)
	}
	return operators
}

func getAtLeastOneSubscription(op *Operator, csv *olmv1Alpha.ClusterServiceVersion, subscriptions []olmv1Alpha.Subscription, packageManifests []*olmpkgv1.PackageManifest) (atLeastOneSubscription bool) {
	atLeastOneSubscription = false
	for s := range subscriptions {
		subscription := &subscriptions[s]
		if subscription.Status.InstalledCSV != csv.Name {
			continue
		}

		op.SubscriptionName = subscription.Name
		op.SubscriptionNamespace = subscription.Namespace
		op.Package = subscription.Spec.Package
		op.Org = subscription.Spec.CatalogSource
		op.Channel = subscription.Spec.Channel
		atLeastOneSubscription = true

		// If the channel is not present in the subscription, get the default channel from the package manifest
		if op.Channel == "" {
			aPackageManifest := getPackageManifestWithSubscription(subscription, packageManifests)
			if aPackageManifest != nil {
				op.Channel = aPackageManifest.Status.DefaultChannel
			} else {
				log.Error("Could not determine the default channel, this operator will always fail certification")
			}
		}
		break
	}
	return atLeastOneSubscription
}

func getPackageManifestWithSubscription(subscription *olmv1Alpha.Subscription, packageManifests []*olmpkgv1.PackageManifest) *olmpkgv1.PackageManifest {
	for index := range packageManifests {
		if packageManifests[index].Status.PackageName == subscription.Spec.Package &&
			packageManifests[index].Namespace == subscription.Spec.CatalogSourceNamespace &&
			packageManifests[index].Status.CatalogSource == subscription.Spec.CatalogSource {
			return packageManifests[index]
		}
	}
	return nil
}

func getAtLeastOneCsv(csv *olmv1Alpha.ClusterServiceVersion, installPlan *olmv1Alpha.InstallPlan) (atLeastOneCsv bool) {
	atLeastOneCsv = false
	for _, csvName := range installPlan.Spec.ClusterServiceVersionNames {
		if csv.Name != csvName {
			continue
		}

		if installPlan.Status.BundleLookups == nil {
			log.Warn("InstallPlan %s for csv %s (ns %s) does not have bundle lookups. It will be skipped.", installPlan.Name, csv.Name, csv.Namespace)
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
		if installPlan.Namespace != op.SubscriptionNamespace {
			continue
		}

		// If the install plan does not deploys this CSV, check the next one
		if !getAtLeastOneCsv(csv, installPlan) {
			continue
		}

		indexImage, catalogErr := getCatalogSourceImageIndexFromInstallPlan(installPlan, allCatalogSources)
		if catalogErr != nil {
			log.Debug("failed to get installPlan image index for csv %s (ns %s) installPlan %s, err: %v",
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
		key := fmt.Sprintf("%s operator: %s ver: %s", o.Phase, o.PackageFromCsvName, o.Version)
		if o.IsClusterWide {
			key += " (all namespaces)"
		} else {
			key += fmt.Sprintf(" in ns: %v", o.TargetNamespaces)
		}
		operatorMap[key] = true
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

func getOperatorTargetNamespaces(namespace string) ([]string, error) {
	client := clientsholder.GetClientsHolder()

	list, err := client.OlmClient.OperatorsV1().OperatorGroups(namespace).List(
		context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, errors.New("no OperatorGroup found")
	}

	return list.Items[0].Spec.TargetNamespaces, nil
}

func GetAllOperatorGroups() ([]*olmv1.OperatorGroup, error) {
	client := clientsholder.GetClientsHolder()

	list, err := client.OlmClient.OperatorsV1().OperatorGroups("").List(context.TODO(), metav1.ListOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, err
	}

	if k8serrors.IsNotFound(err) {
		log.Warn("No OperatorGroup(s) found in the cluster")
		return nil, nil
	}

	if len(list.Items) == 0 {
		log.Warn("OperatorGroup API resource found but no OperatorGroup(s) found in the cluster")
		return nil, nil
	}

	// Collect all OperatorGroup pointers
	var operatorGroups []*olmv1.OperatorGroup
	for i := range list.Items {
		operatorGroups = append(operatorGroups, &list.Items[i])
	}

	return operatorGroups, nil
}

func searchPodInSlice(name, namespace string, pods []*Pod) *Pod {
	// Helper map to filter pods that have been already added
	podsMap := map[types.NamespacedName]*Pod{}
	for _, testPod := range pods {
		podsMap[types.NamespacedName{Namespace: testPod.Namespace, Name: testPod.Name}] = testPod
	}

	// Search by namespace+name key
	podKey := types.NamespacedName{Namespace: namespace, Name: name}
	if pod, found := podsMap[podKey]; found {
		return pod
	}

	return nil
}

func addOperatorPodsToTestPods(operatorPods []*Pod, env *TestEnvironment) {
	for _, operatorPod := range operatorPods {
		// Check whether the pod was already discovered
		testPod := searchPodInSlice(operatorPod.Name, operatorPod.Namespace, env.Pods)
		if testPod != nil {
			log.Info("Operator pod %v/%v already discovered.", testPod.Namespace, testPod.Name)
			// Make sure it's flagged as operator pod.
			testPod.IsOperator = true
		} else {
			log.Info("Operator pod %v/%v added to test pod list", operatorPod.Namespace, operatorPod.Name)
			// Append pod to the test pod list.
			env.Pods = append(env.Pods, operatorPod)
		}
	}
}

func addOperandPodsToTestPods(operandPods []*Pod, env *TestEnvironment) {
	for _, operandPod := range operandPods {
		// Check whether the pod was already discovered
		testPod := searchPodInSlice(operandPod.Name, operandPod.Namespace, env.Pods)
		if testPod != nil {
			log.Info("Operand pod %v/%v already discovered.", testPod.Namespace, testPod.Name)
			// Make sure it's flagged as operand pod.
			testPod.IsOperand = true
		} else {
			log.Info("Operand pod %v/%v added to test pod list", operandPod.Namespace, operandPod.Name)
			// Append pod to the test pod list.
			env.Pods = append(env.Pods, operandPod)
		}
	}
}
