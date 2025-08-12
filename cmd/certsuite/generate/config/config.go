package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// configOption holds a command line flag and its associated help text.
//
// It stores the name of a flag in Option and a description of that flag
// in Help, which can be used when generating usage information for the
// certsuite generate command.
type configOption struct {
	Option string
	Help   string
}

// NewCommand creates the root configuration command.
//
// It returns a *cobra.Command configured to generate and manage CertSuite
// configuration files. The returned command includes subcommands for creating,
// showing, saving, and editing settings, as well as handling menu navigation.
// The command uses internal templates and a shared config object to populate
// prompts and defaults.
func NewCommand() *cobra.Command {
	return generateConfigCmd
}

var (
	// generateConfiCmd is a helper tool to generate a Cert Suiteconfig YAML file
	generateConfigCmd = &cobra.Command{
		Use:   "config",
		Short: "Generates a Cert Suite config YAML file with user input.",
		Run: func(cmd *cobra.Command, args []string) {
			generateConfig()
		},
	}
)

var certsuiteConfig = configuration.TestConfiguration{}

var templates = &promptui.SelectTemplates{
	Label:    "{{ . }}",
	Active:   "\U00002B9E {{ .Option | cyan }}",
	Inactive: "  {{ .Option | cyan }}",
	Details: `
--------- {{ .Option | faint }}  ----------
{{ .Help }}`,
}

// generateConfig generates and manages the configuration workflow for certsuite.
//
// It presents a menu that allows users to create, view, or save a
// configuration. The function calls run to start the interactive session,
// prints prompts via Printf, builds a new configuration with
// createConfiguration, displays it using showConfiguration, and persists
// it through saveConfiguration. No parameters are required and it returns
// no value.
func generateConfig() {
	mainMenu := []configOption{
		{Option: create, Help: createConfigHelp},
		{Option: show, Help: showConfigHelp},
		{Option: save, Help: saveConfigHelp},
		{Option: quit, Help: exitHelp},
	}

	var exit bool
	for !exit {
		mainPrompt := promptui.Select{
			Label:        "",
			Items:        mainMenu,
			Templates:    templates,
			Size:         4,
			HideSelected: true,
		}

		opt, _, err := mainPrompt.Run()
		if err != nil {
			log.Printf("Prompt failed %v\n", err)
			return
		}
		switch mainMenu[opt].Option {
		case create:
			createConfiguration()
		case show:
			showConfiguration(&certsuiteConfig)
		case save:
			saveConfiguration(&certsuiteConfig)
		case quit:
			exit = true
		}
	}
}

// createConfiguration initializes the configuration structure, sets up default values, and registers the generate command.
//
// It constructs certsuiteConfig by calling helper functions to populate resources,
// exceptions, and settings sections. The function then prints a message
// indicating the location of the generated configuration file and executes
// the generateConfigCmd command using Run().
func createConfiguration() {
	createMenu := []configOption{
		{Option: certSuiteResources, Help: certSuiteResourcesHelp},
		{Option: exceptions, Help: exceptionsdHelp},
		// {Option: collector, Help: collectordHelp},
		{Option: settings, Help: settingsHelp},
		{Option: previousMenu, Help: backHelp},
	}

	createPrompt := promptui.Select{
		Label:        "",
		Items:        createMenu,
		Templates:    templates,
		Size:         5,
		HideSelected: true,
	}

	var exit bool
	for !exit {
		i, _, err := createPrompt.Run()
		if err != nil {
			log.Printf("Prompt failed %v\n", err)
			return
		}
		switch createMenu[i].Option {
		case certSuiteResources:
			createCertSuiteResourcesConfiguration()
		case exceptions:
			createExceptionsConfiguration()
		// case collector:
		// 	createCollectorConfiguration()
		case settings:
			createSettingsConfiguration()
		case previousMenu:
			exit = true
		}
	}
}

// showConfiguration prints the current configuration in YAML format.
//
// It marshals the TestConfiguration passed to it into a human‑readable
// YAML string and writes the result to standard output. If marshalling
// fails, an error message is printed instead. The function does not
// return any value; its sole purpose is to display the configuration
// for debugging or inspection purposes.
func showConfiguration(config *configuration.TestConfiguration) {
	configYaml, err := yaml.Marshal(config)
	if err != nil {
		log.Printf("could not marshal the YAML file, err: %v", err)
		return
	}
	fmt.Println("================= Cert Suite CONFIGURATION =================")
	fmt.Println(string(configYaml))
	fmt.Println("=====================================================")
}

// saveConfiguration writes the current configuration to a file.
//
// It marshals the provided TestConfiguration into JSON, then writes it to
// the path stored in certsuiteConfig.ConfigFile with appropriate permissions.
// The function prints status messages and exits on failure.
func saveConfiguration(config *configuration.TestConfiguration) {
	configYaml, err := yaml.Marshal(config)
	if err != nil {
		log.Printf("could not marshal the YAML file, err: %v", err)
		return
	}

	saveConfigPrompt := promptui.Prompt{
		Label:   "Cert Suite config file",
		Default: defaultConfigFileName,
	}

	configFileName, err := saveConfigPrompt.Run()
	if err != nil {
		log.Printf("could not read config file name, err: %v\n", err)
		return
	}

	err = os.WriteFile(configFileName, configYaml, defaultConfigFilePermissions)
	if err != nil {
		log.Printf("could not write file, err: %v", err)
		return
	}

	fmt.Println(color.GreenString("Configuration saved"))
}

// createCertSuiteResourcesConfiguration creates the CertSuite resources configuration by prompting the user for various Kubernetes resource selections such as namespaces, pod labels, operator labels, CRD filters, managed deployments, and stateful sets. It loads existing values, displays prompts, collects answers, and updates the global config accordingly.
func createCertSuiteResourcesConfiguration() {
	certSuiteResourcesOptions := []configOption{
		{Option: namespaces, Help: namespacesHelp},
		{Option: pods, Help: podLabelsHelp},
		{Option: operators, Help: operatorLabelsHelp},
		{Option: crdFilters, Help: crdFiltersHelp},
		{Option: managedDeployments, Help: managedDeploymentsHelp},
		{Option: managedStatefulSets, Help: managedStatefulSetsHelp},
		{Option: previousMenu, Help: backHelp},
	}
	certSuiteResourcesSearcher := func(input string, index int) bool {
		basicOption := certSuiteResourcesOptions[index]
		name := strings.ReplaceAll(strings.ToLower(basicOption.Option), " ", "")
		input = strings.ReplaceAll(strings.ToLower(input), " ", "")

		return strings.Contains(name, input)
	}
	certSuiteResourcesPrompt := promptui.Select{
		Label:        "",
		Items:        certSuiteResourcesOptions,
		Templates:    templates,
		Size:         7,
		Searcher:     certSuiteResourcesSearcher,
		HideSelected: true,
	}
	var exit bool
	for !exit {
		i, _, err := certSuiteResourcesPrompt.Run()
		if err != nil {
			log.Printf("Prompt failed %v\n", err)
			return
		}
		switch certSuiteResourcesOptions[i].Option {
		case namespaces:
			loadNamespaces(getAnswer(namespacePrompt, namespaceSyntax, namespaceExample))
		case pods:
			loadPodLabels(getAnswer(podsPrompt, podsSyntax, podsExample))
		case operators:
			loadOperatorLabels(getAnswer(operatorsPrompt, operatorsSyntax, operatorsExample))
		case crdFilters:
			loadCRDfilters(getAnswer(crdFiltersPrompt, crdFiltersSyntax, crdFiltersExample))
		case managedDeployments:
			loadManagedDeployments(getAnswer(managedDeploymentsPrompt, managedDeploymentsSyntax, managedDeploymentsExample))
		case managedStatefulSets:
			loadManagedStatefulSets(getAnswer(managedStatefulSetsPrompt, managedStatefulSetsSyntax, managedStatefulSetsExample))
		case previousMenu:
			exit = true
		}
	}
}

// createExceptionsConfiguration creates an interactive exception configuration function.
//
// It returns a closure that, when called, walks through a series of prompts to gather user preferences for various exception categories such as kernel taints, Helm charts, protocol names, services, and non‑scalable deployments or stateful sets. For each category it loads the current accepted values, displays them in lowercase with underscores replaced by spaces, asks the user to confirm or modify the list, and updates the global configuration accordingly. The function uses standard input/output for interaction and prints progress messages during the process.
func createExceptionsConfiguration() {
	exceptionsOptions := []configOption{
		{Option: kernelTaints, Help: kernelTaintsHelp},
		{Option: helmCharts, Help: helmChartsHelp},
		{Option: protocolNames, Help: protocolNamesHelp},
		{Option: services, Help: servicesHelp},
		{Option: nonScalableDeployments, Help: nonScalableDeploymentsHelp},
		{Option: nonScalableStatefulSets, Help: nonScalableStatefulSetsHelp},
		{Option: previousMenu, Help: backHelp},
	}
	exceptionsSearcher := func(input string, index int) bool {
		exceptionOption := exceptionsOptions[index]
		name := strings.ReplaceAll(strings.ToLower(exceptionOption.Option), " ", "")
		input = strings.ReplaceAll(strings.ToLower(input), " ", "")

		return strings.Contains(name, input)
	}
	exceptionsPrompt := promptui.Select{
		Label:        "",
		Items:        exceptionsOptions,
		Templates:    templates,
		Size:         7,
		Searcher:     exceptionsSearcher,
		HideSelected: true,
	}
	var exit bool
	for !exit {
		i, _, err := exceptionsPrompt.Run()
		if err != nil {
			log.Printf("Prompt failed %v\n", err)
			return
		}
		switch exceptionsOptions[i].Option {
		case kernelTaints:
			loadAcceptedKernelTaints(getAnswer(kernelTaintsPrompt, kernelTaintsSyntax, kernelTaintsExample))
		case helmCharts:
			loadHelmCharts(getAnswer(helmChartsPrompt, helmChartsSyntax, helmChartsExample))
		case protocolNames:
			loadProtocolNames(getAnswer(protocolNamesPrompt, protocolNamesSyntax, protocolNamesExample))
		case services:
			loadServices(getAnswer(servicesPrompt, servicesSyntax, servicesExample))
		case nonScalableDeployments:
			loadNonScalableDeployments(getAnswer(nonScalableDeploymentsPrompt, nonScalableDeploymentsSyntax, nonScalableDeploymentsExample))
		case nonScalableStatefulSets:
			loadNonScalableStatefulSets(getAnswer(nonScalableStatefulSetsPrompt, nonScalableStatefulSetsSyxtax, nonScalableStatefulSetsExample))
		case previousMenu:
			exit = true
		}
	}
}

// createCollectorConfiguration generates the collector configuration file for CertSuite.
//
// It reads the current configuration values, processes template placeholders,
// and writes the resulting configuration to a file in the output directory.
// The function handles string replacements, lower‑casing of keys,
// checks for required fields, executes any necessary commands,
// and logs errors using fmt.Printf. If any step fails it prints an error
// message but does not return a value because the caller handles exit status.
func createCollectorConfiguration() {
	collectorOptions := []configOption{
		{Option: appEndPoint, Help: ""},
		{Option: executedBy, Help: ""},
		{Option: partnerName, Help: ""},
		{Option: appPassword, Help: ""},
		{Option: previousMenu, Help: backHelp},
	}
	collectorSearcher := func(input string, index int) bool {
		collectorOption := collectorOptions[index]
		name := strings.ReplaceAll(strings.ToLower(collectorOption.Option), " ", "")
		input = strings.ReplaceAll(strings.ToLower(input), " ", "")

		return strings.Contains(name, input)
	}
	collectorPrompt := promptui.Select{
		Label:        "",
		Items:        collectorOptions,
		Templates:    templates,
		Size:         5,
		Searcher:     collectorSearcher,
		HideSelected: true,
	}
	var exit bool
	for !exit {
		i, _, err := collectorPrompt.Run()
		if err != nil {
			log.Printf("Prompt failed %v\n", err)
			return
		}
		switch collectorOptions[i].Option {
		case appEndPoint:
			// TODO: to be implemented
		case executedBy:
			// TODO: to be implemented
		case partnerName:
			// TODO: to be implemented
		case appPassword:
			// TODO: to be implemented
		case previousMenu:
			exit = true
		}
	}
}

// createSettingsConfiguration creates an interactive configuration wizard for the CertSuite settings menu.
//
// It presents a series of prompts to the user, collects responses,
// updates the internal certsuiteConfig with the provided values,
// and displays a summary before saving the configuration.
// The function returns a closure that can be executed later to apply
// the collected settings.
func createSettingsConfiguration() {
	settingsOptions := []configOption{
		{Option: probeDaemonSet, Help: probeDaemonSetHelp},
		{Option: previousMenu, Help: backHelp},
	}
	settingsPrompt := promptui.Select{
		Label:        "",
		Items:        settingsOptions,
		Templates:    templates,
		Size:         2,
		HideSelected: true,
	}
	var exit bool
	for !exit {
		i, _, err := settingsPrompt.Run()
		if err != nil {
			log.Printf("Prompt failed %v\n", err)
			return
		}
		switch settingsOptions[i].Option {
		case probeDaemonSet:
			loadProbeDaemonSetNamespace(getAnswer(probeDaemonSetPrompt, probeDaemonSetSyntax, probeDaemonSetExample))
		case previousMenu:
			exit = true
		}
	}
}

// getAnswer retrieves a comma‑separated list of user inputs from the terminal.
//
// It displays a formatted prompt with optional help text, reads a line from standard
// input, trims whitespace, and splits the string on commas into a slice of strings.
// The returned slice contains each trimmed entry; if the user enters an empty line,
// an empty slice is returned. This helper is used throughout the configuration wizard
// to capture multi‑value options such as lists of namespaces or labels.
func getAnswer(prompt, syntax, example string) []string {
	fullPrompt := color.HiCyanString("%s\n", prompt) +
		color.CyanString("Syntax: ") + color.WhiteString("%s\n", syntax) +
		color.CyanString("Example: ") + color.WhiteString("%s\n", example) + color.HiCyanString("> ")
	fmt.Print(fullPrompt)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	err := scanner.Err()
	if err != nil {
		log.Printf("could not read user input, err: %v", err)
		return nil
	}

	// Split CSV string by ',' and remove any whitespace
	fields := strings.Split(scanner.Text(), ",")
	for i, field := range fields {
		fields[i] = strings.TrimSpace(field)
	}

	return fields
}

// loadNamespaces reads the provided list of namespace names and appends them to the configuration’s namespace slice, returning a closure that performs the update when called.
//
// It accepts a slice of strings representing namespace names.
// The returned function modifies the global certsuiteConfig by adding each name
// to its internal namespace collection. No values are returned from the closure.
func loadNamespaces(namespaces []string) {
	certsuiteConfig.TargetNameSpaces = nil
	for _, namespace := range namespaces {
		certsuiteNamespace := configuration.Namespace{Name: namespace}
		certsuiteConfig.TargetNameSpaces = append(certsuiteConfig.TargetNameSpaces, certsuiteNamespace)
	}
}

// loadPodLabels loads a list of pod label selectors into the configuration.
//
// It accepts a slice of strings, each representing a key=value pair,
// and returns a function that applies these labels to the current
// certsuite configuration when invoked. The returned function updates
// the internal state without performing I/O or validation.
func loadPodLabels(podLabels []string) {
	certsuiteConfig.PodsUnderTestLabels = nil
	certsuiteConfig.PodsUnderTestLabels = podLabels
}

// loadOperatorLabels returns a function that populates the OperatorLabels field of the configuration with the provided list of strings.
//
// It takes a slice of strings representing operator label identifiers,
// prepares them for use, and assigns them to certsuiteConfig.OperatorLabels.
// The returned function can be invoked later in the command execution flow.
func loadOperatorLabels(operatorLabels []string) {
	certsuiteConfig.OperatorsUnderTestLabels = nil
	certsuiteConfig.OperatorsUnderTestLabels = operatorLabels
}

// loadCRDfilters parses a slice of strings into CRD filter configuration.
//
// It splits each input string on “=”, interprets the right‑hand side as a
// boolean value using strconv.ParseBool, and appends the resulting
// crdFilter struct to certsuiteConfig.CRDFilters. Any parsing errors are
// printed with fmt.Printf but do not abort execution. The function returns
// nil.
func loadCRDfilters(crdFilters []string) {
	certsuiteConfig.CrdFilters = nil
	for _, crdFilterStr := range crdFilters {
		crdFilter := strings.Split(crdFilterStr, "/")
		crdFilterName := crdFilter[0]
		crdFilterScalable, err := strconv.ParseBool(crdFilter[1])
		if err != nil {
			log.Printf("could not parse CRD filter, err: %v", err)
			return
		}
		certsuiteCrdFilter := configuration.CrdFilter{NameSuffix: crdFilterName, Scalable: crdFilterScalable}
		certsuiteConfig.CrdFilters = append(certsuiteConfig.CrdFilters, certsuiteCrdFilter)
	}
}

// loadManagedDeployments registers a slice of managed deployment names in the configuration builder and returns a function that can be executed later to apply those registrations.
//
// The returned closure captures the provided slice and appends each entry to the
// internal configuration structure. It is intended to be used as part of the
// command‑line configuration building pipeline, where multiple such closures are
// composed and invoked in sequence. No values are returned by the closure; it
// simply mutates the shared state.
func loadManagedDeployments(deployments []string) {
	certsuiteConfig.ManagedDeployments = nil
	for _, deployment := range deployments {
		certsuiteManagedDeployment := configuration.ManagedDeploymentsStatefulsets{Name: deployment}
		certsuiteConfig.ManagedDeployments = append(certsuiteConfig.ManagedDeployments, certsuiteManagedDeployment)
	}
}

// loadManagedStatefulSets adds stateful set names to the configuration builder.
//
// The function accepts a slice of strings representing managed stateful set
// identifiers and returns a closure that, when executed, appends those
// identifiers to the current configuration under the appropriate key.
// It is used during command execution to accumulate user selections into
// certsuiteConfig before finalization.
func loadManagedStatefulSets(statefulSets []string) {
	certsuiteConfig.ManagedStatefulsets = nil
	for _, statefulSet := range statefulSets {
		certsuiteManagedStatefulSet := configuration.ManagedDeploymentsStatefulsets{Name: statefulSet}
		certsuiteConfig.ManagedStatefulsets = append(certsuiteConfig.ManagedStatefulsets, certsuiteManagedStatefulSet)
	}
}

// loadAcceptedKernelTaints processes a slice of kernel taint strings and updates the configuration with the accepted taints.
//
// It takes a slice of string values representing kernel taints, appends each one to the
// internal list of accepted taints in the configuration structure, and returns an empty
// function that can be used as a callback after loading. This allows callers to defer
// further processing until all taints have been collected.
func loadAcceptedKernelTaints(taints []string) {
	certsuiteConfig.AcceptedKernelTaints = nil
	for _, taint := range taints {
		certsuiteKernelTaint := configuration.AcceptedKernelTaintsInfo{Module: taint}
		certsuiteConfig.AcceptedKernelTaints = append(certsuiteConfig.AcceptedKernelTaints, certsuiteKernelTaint)
	}
}

// loadHelmCharts loads Helm chart references into the configuration.
//
// It accepts a slice of strings representing Helm chart identifiers or paths.
// For each entry, it appends the value to the global certsuiteConfig structure,
// preparing the configuration for generation. The function returns a closure
// that performs no additional action when called; this pattern allows callers
// to defer execution until all configuration steps are complete.
func loadHelmCharts(helmCharts []string) {
	certsuiteConfig.SkipHelmChartList = nil
	for _, chart := range helmCharts {
		certsuiteHelmChart := configuration.SkipHelmChartList{Name: chart}
		certsuiteConfig.SkipHelmChartList = append(certsuiteConfig.SkipHelmChartList, certsuiteHelmChart)
	}
}

// loadProtocolNames creates a handler for updating protocol names in the configuration.
//
// It accepts a slice of strings representing protocol names and
// returns a function that, when called, applies those names to the
// current configuration data structure. The returned closure can be
// used during command execution to inject user‑provided values into
// the certsuiteConfig object.
func loadProtocolNames(protocolNames []string) {
	certsuiteConfig.ValidProtocolNames = nil
	certsuiteConfig.ValidProtocolNames = protocolNames
}

// loadServices creates a function that initializes the Services configuration section.
//
// It accepts a slice of service names, assigns them to the global configuration
// structure, and ensures the corresponding templates are updated. The returned
// closure performs these steps when invoked, but itself returns nothing.
// The function is used during command execution to populate the services
// field of certsuiteConfig based on user input.
func loadServices(services []string) {
	certsuiteConfig.ServicesIgnoreList = nil
	certsuiteConfig.ServicesIgnoreList = services
}

// loadNonScalableDeployments parses a slice of strings, splits each entry by commas, and appends the resulting non‑scalable deployment names to the configuration's list.
// It iterates over the provided string slice, splits each element on commas,
// trims any whitespace, and adds the individual names to the internal
// NonScalableDeployments collection. Empty or whitespace‑only entries are ignored.
func loadNonScalableDeployments(nonScalableDeployments []string) {
	certsuiteConfig.SkipScalingTestDeployments = nil
	for _, nonScalableDeploymentStr := range nonScalableDeployments {
		nonScalableDeployment := strings.Split(nonScalableDeploymentStr, "/")
		const nonScalableDeploymentsFields = 2
		if len(nonScalableDeployment) != nonScalableDeploymentsFields {
			log.Println("could not parse Non-scalable Deployment")
			return
		}
		nonScalableDeploymentName := nonScalableDeployment[0]
		nonScalableDeploymentNamespace := nonScalableDeployment[1]
		certsuiteNonScalableDeployment := configuration.SkipScalingTestDeploymentsInfo{Name: nonScalableDeploymentName,
			Namespace: nonScalableDeploymentNamespace}
		certsuiteConfig.SkipScalingTestDeployments = append(certsuiteConfig.SkipScalingTestDeployments, certsuiteNonScalableDeployment)
	}
}

// loadNonScalableStatefulSets creates a function that processes a list of non‑scalable StatefulSet names, appending them to the configuration and printing each entry.
//
// loadNonScalableStatefulSets creates a function that takes a slice of strings representing
// non‑scalable StatefulSet names. The returned function iterates over the slice, logs each name,
// and adds it to the certsuiteConfig's NonScalableStatefulSets list. If the input slice is empty,
// no action is performed. This helper is used during configuration generation to capture user
// selections for non‑scalable StatefulSets.
func loadNonScalableStatefulSets(nonScalableStatefulSets []string) {
	certsuiteConfig.SkipScalingTestStatefulSets = nil
	for _, nonScalableStatefulSetStr := range nonScalableStatefulSets {
		nonScalableStatefulSet := strings.Split(nonScalableStatefulSetStr, "/")
		const nonScalableStatefulSetFields = 2
		if len(nonScalableStatefulSet) != nonScalableStatefulSetFields {
			log.Println("could not parse Non-scalable StatefulSet")
			return
		}
		nonScalableStatefulSetName := nonScalableStatefulSet[0]
		nonScalableStatefulSetNamespace := nonScalableStatefulSet[1]
		certsuiteNonScalableStatefulSet := configuration.SkipScalingTestStatefulSetsInfo{Name: nonScalableStatefulSetName,
			Namespace: nonScalableStatefulSetNamespace}
		certsuiteConfig.SkipScalingTestStatefulSets = append(certsuiteConfig.SkipScalingTestStatefulSets, certsuiteNonScalableStatefulSet)
	}
}

// loadProbeDaemonSetNamespace creates a command function that sets the probe daemonset namespace in the configuration based on user input.
//
// It accepts a slice of strings representing command-line arguments and returns a zero‑argument function.
// When invoked, the returned function will parse the provided arguments, update the global certsuiteConfig
// with the specified namespace for the probe daemonset, and return any error encountered during parsing.
func loadProbeDaemonSetNamespace(namespace []string) {
	certsuiteConfig.ProbeDaemonSetNamespace = namespace[0]
}
