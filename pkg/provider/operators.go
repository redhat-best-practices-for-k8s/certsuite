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

// Operator represents an installed operator in a Kubernetes cluster.
// It contains metadata about the operator such as name, channel, version,
// and its installation status. The struct also tracks related resources
// like subscriptions, install plans, operand pods, and any preflight
// results collected during validation. This information is used by
// CertSuite to assess operator compliance and operational health.
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

// CsvInstallPlan represents the configuration needed to install a CSV from an operator bundle.
//
// It holds the image references for the bundle, the index that contains it,
// and the name of the ClusterServiceVersion to deploy. These fields are
// used by the provider package when creating or updating an installation plan.
type CsvInstallPlan struct {
	// Operator's installPlan name
	Name string `yaml:"name" json:"name"`
	// BundleImage is the URL referencing the bundle image
	BundleImage string `yaml:"bundleImage" json:"bundleImage"`
	// IndexImage is the URL referencing the index image
	IndexImage string `yaml:"indexImage" json:"indexImage"`
}

// String returns a human-readable representation of the Operator.
//
// It formats the operator's fields into a concise string, typically used for
// logging or debugging purposes. The output includes key attributes such as
// the operator type and relevant identifiers. This method does not modify
// the receiver.
func (op *Operator) String() string {
	return fmt.Sprintf("csv: %s ns:%s subscription:%s targetNamespaces=%v", op.Name, op.Namespace, op.SubscriptionName, op.TargetNamespaces)
}

// SetPreflightResults runs the preflight checks on a test environment and stores their results in the database.
//
// SetPreflightResults executes all configured preflight tests against the given TestEnvironment,
// collects any errors or warnings, writes the results to the provider's preflight results
// database, and returns an error if the operation fails. It uses the operator's client
// configuration, supports insecure connections based on environment settings, and logs
// progress through the provider's logging facilities. The function does not return the
// individual test outcomes directly; they are persisted for later retrieval via
// GetPreflightResultsDB.
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

// getUniqueCsvListByName returns a list of ClusterServiceVersions with unique names.
//
// It accepts a slice that may contain multiple CSV objects sharing the same name and
// produces a new slice containing only one instance per distinct name.
// The resulting slice is sorted by the CSV name field for deterministic ordering.
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

// createOperators constructs a slice of Operator objects from various operator resources.
//
// It takes lists of CSVs, Subscriptions, PackageManifests, InstallPlans, and CatalogSources,
// along with flags indicating whether to include catalog sources and subscriptions.
// The function deduplicates CSVs by name, associates subscriptions and install plans
// with their target namespaces, and creates Operator structs containing the relevant
// metadata. It returns a slice of pointers to these Operator objects for further processing.
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
		// Skip CSVs that are not in the Succeeded phase if the flag is set.
		if csv.Status.Phase != olmv1Alpha.CSVPhaseSucceeded && succeededRequired {
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

// getAtLeastOneSubscription checks that at least one subscription matches a package manifest.
//
// It takes an Operator, a ClusterServiceVersion, a slice of Subscriptions,
// and a list of PackageManifests. The function returns true if any
// subscription is found for which a corresponding package manifest can be
// retrieved via getPackageManifestWithSubscription. If no matching
// subscription exists or an error occurs while retrieving the manifest,
// it logs the error and returns false. This ensures that operators have
// at least one valid subscription before proceeding with further checks.
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

// getPackageManifestWithSubscription retrieves the package manifest that corresponds to a given Operator Lifecycle Manager subscription.
//
// It accepts a pointer to an olmv1Alpha.Subscription and a slice of
// olm.PackageManifest objects. The function examines each manifest in the
// slice, comparing relevant fields (such as name, catalog source, or namespace)
// against the subscription's specification to find a match. If a matching
// PackageManifest is found it is returned; otherwise the function returns nil.
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

// getAtLeastOneCsv reports whether a valid ClusterServiceVersion is available.
//
// It examines the supplied ClusterServiceVersion and its associated InstallPlan.
// If either is nil or indicates an error state, it logs a warning and returns false.
// Otherwise, it returns true to signal that at least one CSV has been successfully
// retrieved and can be used for further operator validation.
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

// getAtLeastOneInstallPlan checks whether an operator has at least one InstallPlan that can be used to install the specified CSV.
//
// It receives a pointer to the Operator, the desired ClusterServiceVersion,
// a slice of available InstallPlans and a slice of CatalogSources.
// The function first verifies that the CSV exists in the catalog sources,
// then iterates over the InstallPlans looking for one whose image index
// matches the CSV's bundle image. If such an InstallPlan is found it returns true;
// otherwise false.
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

// CsvToString converts a ClusterServiceVersion to its string representation.
//
// It takes a pointer to an olmv1Alpha.ClusterServiceVersion and returns a
// formatted string that includes the key fields of the object for display or
// logging purposes.
func CsvToString(csv *olmv1Alpha.ClusterServiceVersion) string {
	return fmt.Sprintf("operator csv: %s ns: %s",
		csv.Name,
		csv.Namespace,
	)
}

// getSummaryAllOperators extracts a concise summary from each Operator and returns them as a slice of strings.
//
// It takes a slice of pointers to Operator structs, iterates over them,
// formats each operator's name and status into a human‑readable string,
// and collects these strings in a new slice.
// The returned slice contains one entry per operator, preserving the
// original order. This summary is used for reporting or logging purposes.
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

// getCatalogSourceImageIndexFromInstallPlan extracts the image index URL from an Operator Lifecycle Manager install plan and a list of catalog sources.
//
// It searches the provided InstallPlan for the relevant CatalogSource reference,
// matches that reference against the supplied CatalogSource slice, and retrieves
// the image field which represents the container image index to be used.
// If no matching catalog source is found or if any error occurs while accessing
// the fields, it returns an empty string along with an error describing the issue.
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

// getOperatorTargetNamespaces retrieves the namespaces targeted by an Operator in a given namespace.
//
// It takes a namespace name as input, queries the cluster for operator groups and operators
// within that namespace, extracts their target namespaces, and returns them as a slice of strings.
// If any step fails it returns an error describing the failure.
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

// GetAllOperatorGroups retrieves all OperatorGroup resources from the cluster.
//
// It returns a slice of pointers to olmv1.OperatorGroup and an error if the
// operation fails. The function uses the provider's client holder to list
// OperatorGroup objects via the OperatorsV1 API. If no groups are found, it
// returns an empty slice without error. Any other errors encountered during
// the listing process are propagated back to the caller.
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

// searchPodInSlice searches a slice of Pod pointers for a pod that matches the given name and namespace.
// It iterates over the provided []*Pod slice and returns the first Pod whose Name and Namespace fields match the
// supplied arguments. If no matching pod is found, it returns nil. This helper is used internally by operator
// logic to locate specific pods within a collection.
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

// addOperatorPodsToTestPods adds operator pods to the test environment pod list.
//
// It scans the current test pod slice and appends any operator-related pods that are not already present.
// The function uses the TestEnvironment context to determine which pods qualify as operators,
// checks for duplicates via searchPodInSlice, logs actions with Info, and updates the pod slice in place.
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

// addOperandPodsToTestPods appends operand pods to the test environment pod list.
//
// It iterates over the provided slice of pod references and checks if each
// pod is already present in the TestEnvironment's collection by calling
// searchPodInSlice. If a pod is not found, it logs an informational message
// and appends the pod to the environment's internal slice.
// The function does not return a value; it mutates the TestEnvironment passed
// as its second argument.
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
