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

// Operator represents an installed operator within a cluster
//
// This data structure holds metadata about an operator, including its name,
// namespace, deployment phase, subscription details, package information, and
// any associated install plans. It also tracks whether the operator is
// cluster‑wide or scoped to specific namespaces and stores preflight test
// results for validation. The fields provide a comprehensive view of an
// operator’s state and configuration used by the certification framework.
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

// CsvInstallPlan Describes an operator's install plan details
//
// This structure holds the name of the install plan along with URLs for both
// the bundle image and the index image used in the installation process. It is
// primarily utilized to convey necessary information when creating or managing
// operator deployments, ensuring that the correct images are referenced during
// installation.
type CsvInstallPlan struct {
	// Operator's installPlan name
	Name string `yaml:"name" json:"name"`
	// BundleImage is the URL referencing the bundle image
	BundleImage string `yaml:"bundleImage" json:"bundleImage"`
	// IndexImage is the URL referencing the index image
	IndexImage string `yaml:"indexImage" json:"indexImage"`
}

// Operator.String Provides a human-readable representation of the operator
//
// This method formats key fields such as the operator name, namespace,
// subscription name, and target namespaces into a single string. It uses a
// standard formatting function to create the output and returns it for display
// or logging purposes.
func (op *Operator) String() string {
	return fmt.Sprintf("csv: %s ns:%s subscription:%s targetNamespaces=%v", op.Name, op.Namespace, op.SubscriptionName, op.TargetNamespaces)
}

// Operator.SetPreflightResults Collects and stores Preflight test outcomes for an operator
//
// The function runs the OpenShift Preflight checks against the operator's
// bundle image, capturing passed, failed, and error results. It writes all
// check logs to a buffer and attaches them to the global log output. After
// processing, it removes temporary artifacts and assigns the collected results
// to the operator’s PreflightResults field.
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

// getUniqueCsvListByName filters a list to include only one instance per CSV name
//
// The function receives a slice of ClusterServiceVersion objects, removes any
// duplicates by keeping the last occurrence for each unique name, logs how many
// unique entries were found, and then returns the deduplicated slice sorted
// alphabetically by CSV name. It uses an internal map to track seen names and
// sort.Slice for deterministic ordering.
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

// createOperators Creates a list of operator objects from CSV data
//
// The function iterates over unique cluster service versions, filters out
// failed ones if required, and builds an Operator struct for each. It extracts
// package and version information from the CSV name, associates at least one
// subscription to determine target namespaces, and gathers install plans linked
// to the CSV. The resulting slice contains operators enriched with phase,
// namespace, and optional CSV details.
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

// getAtLeastOneSubscription Finds a subscription linked to the given CSV and updates the operator record
//
// The function scans through all subscriptions, matching one whose installed
// CSV name equals that of the provided CSV. When found, it populates the
// operator with subscription details such as name, namespace, package, catalog
// source, and channel. If the channel is missing, it retrieves the default
// channel from the related package manifest; otherwise it logs an error.
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

// getPackageManifestWithSubscription Finds a matching package manifest for a subscription
//
// The function iterates over the provided package manifests, checking whether
// each one matches the subscription’s package name, catalog source namespace,
// and catalog source. If a match is found, that package manifest is returned;
// otherwise the function returns nil. This lookup assists in determining
// default channel information when it is not explicitly set in the
// subscription.
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

// getAtLeastOneCsv Determines if an install plan includes a specific CSV
//
// The function iterates through the names listed in the install plan’s
// specification to see if it matches the provided CSV. If a match is found, it
// verifies that the install plan contains bundle lookup information; otherwise
// it logs a warning and skips that plan. It returns true when a matching CSV
// with valid bundle lookups exists, false otherwise.
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

// getAtLeastOneInstallPlan retrieves at least one install plan for an operator
//
// This function iterates through all available install plans, filtering by
// namespace and ensuring the plan includes the specified CSV. For each
// qualifying plan it extracts bundle and index image information from catalog
// sources. The install plan details are appended to the operator’s
// InstallPlans slice and a true flag is returned when at least one plan has
// been added.
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

// CsvToString Formats a CSV name and namespace into a readable string
//
// The function receives a pointer to a ClusterServiceVersion object and returns
// a string that includes the object's name followed by its namespace. It uses
// formatting to produce a concise representation suitable for logging or
// debugging purposes.
func CsvToString(csv *olmv1Alpha.ClusterServiceVersion) string {
	return fmt.Sprintf("operator csv: %s ns: %s",
		csv.Name,
		csv.Namespace,
	)
}

// getSummaryAllOperators Creates a sorted list of unique operator status strings
//
// This function iterates over a slice of operators, building a key that
// includes the phase, package name, version and namespace information. It
// stores each distinct key in a map to avoid duplicates, then collects the keys
// into a slice, sorts them alphabetically, and returns the result.
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

// getCatalogSourceImageIndexFromInstallPlan retrieves the image index of a catalog source referenced by an install plan
//
// The function takes an install plan and a list of catalog sources, finds the
// catalog source referenced in the first bundle lookup, and returns its image
// field. If no matching catalog source is found it reports an error. The
// returned string is used elsewhere to identify the index image for a CSV.
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

// getOperatorTargetNamespaces Retrieves the list of namespaces an operator targets
//
// The function queries the Operator Group resource within a specified namespace
// to determine which namespaces the operator is allowed to operate in. It
// returns a slice of target namespace names and an error if no OperatorGroup
// exists or if the API call fails.
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

// GetAllOperatorGroups Retrieves all OperatorGroup resources from the cluster
//
// This function queries the OpenShift Operator Lifecycle Manager for
// OperatorGroup objects across all namespaces. It returns a slice of pointers
// to each group found, or nil if none exist, while logging warnings when the
// API resource is missing or empty. Errors unrelated to a missing resource are
// propagated back to the caller.
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

// searchPodInSlice Finds a pod in a list by name and namespace
//
// The function receives a pod name, its namespace, and a slice of pod objects.
// It builds an index map keyed on the namespaced name and looks up the
// requested key, returning the matching pod if found or nil otherwise.
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

// addOperatorPodsToTestPods Adds operator pods to the test pod list
//
// This function iterates over a slice of operator pods, checking each one
// against the current environment's pod collection. If an operator pod is
// already present, it marks that existing pod as an operator; otherwise, it
// appends the new pod to the test list. Logging statements provide visibility
// into whether pods were added or already discovered.
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

// addOperandPodsToTestPods Adds discovered operand pods to the test environment
//
// This routine iterates over a list of operand pods, checking each against the
// current set of test pods in the environment. If a pod is already present, it
// logs that fact and marks the existing entry as an operand; otherwise it
// appends the new pod to the environment's pod list. The function ensures no
// duplicate entries while guaranteeing all operand pods are available for
// subsequent tests.
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
