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

// configOption Represents a configuration setting with its description
//
// This structure holds two text fields: one that specifies the name of an
// option, and another that provides explanatory help for that option. It is
// used internally to map command-line flags or configuration keys to
// user-facing descriptions.
type configOption struct {
	Option string
	Help   string
}

// NewCommand Creates the configuration subcommand
//
// This function returns a preconfigured cobra.Command that provides options for
// generating or managing configuration files within the application. It does
// not take any arguments and simply returns the command instance that has been
// set up elsewhere in the package.
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

// generateConfig Launches an interactive menu for managing configuration
//
// When invoked, this routine displays a prompt with options to create, view,
// save, or exit the configuration workflow. It loops until the user selects
// quit, calling helper functions to handle each action. Errors during prompt
// execution are logged and cause an early return.
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

// createConfiguration Starts the interactive configuration menu
//
// The function presents a list of configuration categories such as resources,
// exceptions, settings, and an option to return to the previous menu. It uses a
// prompt loop that displays the choices and handles user selection by invoking
// dedicated sub‑configuration functions. Errors during the prompt are logged
// and cause an early exit from the routine.
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

// showConfiguration Displays the current configuration in YAML format
//
// The function serializes a TestConfiguration object into YAML and prints it to
// standard output, surrounded by header and footer lines for readability. If
// marshaling fails, it logs an error message and exits without printing
// anything.
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

// saveConfiguration Saves the current configuration to a YAML file
//
// The function converts a TestConfiguration struct into YAML, prompts the user
// for a filename with a default suggestion, writes the data to that file with
// appropriate permissions, and prints a success message. If any step
// fails—marshalling, prompting, or writing—it logs an error and aborts
// without returning a value.
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

// createCertSuiteResourcesConfiguration Presents an interactive menu to configure resource selections
//
// The function displays a list of configuration options such as namespaces, pod
// labels, operator labels, CRD filters, and managed deployments. Users can
// select each option to provide input via prompts, which is then parsed and
// stored in the global configuration. Selecting "previousMenu" exits back to
// the higher‑level menu.
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

// createExceptionsConfiguration Presents an interactive menu to configure exception lists
//
// The routine builds a selection list of exception categories such as kernel
// taints, Helm charts, protocol names, services, and non‑scalable
// deployments. It uses promptui to allow the user to search and choose an
// option; upon selection it calls helper functions that read comma‑separated
// input from the terminal and populate global configuration slices. The process
// repeats until the user chooses to return to the previous menu.
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

// createCollectorConfiguration prompts the user to select a collector configuration option
//
// The function presents an interactive menu of configuration options such as
// endpoint, executor identity, partner name, password, and an exit choice. It
// uses a searcher that filters options by matching input text ignoring case and
// spaces. When the user selects an item, the corresponding action is handled in
// a switch; currently only the exit option terminates the loop while other
// cases are placeholders for future implementation.
//
//nolint:unused
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

// createSettingsConfiguration Prompts user to configure Probe DaemonSet namespace
//
// The function presents a menu with an option to set the Probe DaemonSet
// namespace or return to the previous menu. When selected, it asks the user for
// a comma‑separated list of namespaces, parses the input, and assigns the
// first value to the global configuration. The loop continues until the user
// chooses to exit.
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

// getAnswer Collects a comma‑separated list of items from the user
//
// The function displays a prompt with syntax and example guidance, then reads a
// single line of text from standard input. It splits the entered string on
// commas, trims surrounding whitespace from each element, and returns the
// resulting slice of strings. If reading fails, it logs an error and returns
// nil.
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

// loadNamespaces Stores selected namespaces in the configuration
//
// This routine receives a slice of namespace names, clears any previously
// stored target namespaces, and then appends each provided name as a Namespace
// struct to the global configuration list. It modifies the config in place
// without returning a value.
func loadNamespaces(namespaces []string) {
	certsuiteConfig.TargetNameSpaces = nil
	for _, namespace := range namespaces {
		certsuiteNamespace := configuration.Namespace{Name: namespace}
		certsuiteConfig.TargetNameSpaces = append(certsuiteConfig.TargetNameSpaces, certsuiteNamespace)
	}
}

// loadPodLabels Stores user-specified pod labels for later configuration
//
// This routine clears any existing pod label settings and then assigns the
// supplied slice of strings to the global configuration structure. It is
// invoked after the user selects pod labels from an interactive prompt,
// ensuring that only the chosen labels are retained. No value is returned; the
// effect is visible through the updated configuration state.
func loadPodLabels(podLabels []string) {
	certsuiteConfig.PodsUnderTestLabels = nil
	certsuiteConfig.PodsUnderTestLabels = podLabels
}

// loadOperatorLabels Updates the configuration with new operator labels
//
// This function replaces any previously stored operator labels in the global
// configuration with a fresh list provided as input. It first resets the
// current label collection to an empty state and then assigns the supplied
// slice, ensuring that subsequent operations use only the latest set of labels.
func loadOperatorLabels(operatorLabels []string) {
	certsuiteConfig.OperatorsUnderTestLabels = nil
	certsuiteConfig.OperatorsUnderTestLabels = operatorLabels
}

// loadCRDfilters parses CRD filter strings into configuration objects
//
// The function clears the existing list of CRD filters, then iterates over each
// supplied string. Each string is split on a slash to extract a name suffix and
// a boolean flag indicating scalability; it converts the second part to a bool,
// logs an error if conversion fails, and appends a new filter structure to the
// global configuration.
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

// loadManagedDeployments Populates the list of deployments to be managed
//
// The function receives a slice of deployment names, clears any previously
// stored deployments in the global configuration, and then iterates over each
// name. For every entry it creates a new ManagedDeploymentsStatefulsets object
// with the name field set, appending this object to the configuration’s
// ManagedDeployments list. This prepares the configuration for subsequent
// resource generation.
func loadManagedDeployments(deployments []string) {
	certsuiteConfig.ManagedDeployments = nil
	for _, deployment := range deployments {
		certsuiteManagedDeployment := configuration.ManagedDeploymentsStatefulsets{Name: deployment}
		certsuiteConfig.ManagedDeployments = append(certsuiteConfig.ManagedDeployments, certsuiteManagedDeployment)
	}
}

// loadManagedStatefulSets Stores user-selected StatefulSet names for later configuration
//
// This routine clears any previously stored StatefulSet entries in the global
// configuration, then iterates over each supplied name. For every name it
// creates a lightweight structure containing that name and appends it to the
// list of managed StatefulSets maintained by the application. The function has
// no return value but updates shared state used by subsequent setup steps.
func loadManagedStatefulSets(statefulSets []string) {
	certsuiteConfig.ManagedStatefulsets = nil
	for _, statefulSet := range statefulSets {
		certsuiteManagedStatefulSet := configuration.ManagedDeploymentsStatefulsets{Name: statefulSet}
		certsuiteConfig.ManagedStatefulsets = append(certsuiteConfig.ManagedStatefulsets, certsuiteManagedStatefulSet)
	}
}

// loadAcceptedKernelTaints stores a list of accepted kernel taints in the configuration
//
// The function clears any previously stored taint entries, then iterates over
// the supplied slice. For each taint string it creates a new struct containing
// the module name and appends it to the global configuration slice. The
// resulting list is used by the tool when evaluating cluster readiness.
func loadAcceptedKernelTaints(taints []string) {
	certsuiteConfig.AcceptedKernelTaints = nil
	for _, taint := range taints {
		certsuiteKernelTaint := configuration.AcceptedKernelTaintsInfo{Module: taint}
		certsuiteConfig.AcceptedKernelTaints = append(certsuiteConfig.AcceptedKernelTaints, certsuiteKernelTaint)
	}
}

// loadHelmCharts Stores specified Helm chart names to skip during configuration
//
// The function receives a slice of chart identifiers and resets the global skip
// list before adding each entry as a new configuration object. Each name is
// wrapped in a struct that represents an item to be excluded from processing.
// The resulting list is used elsewhere to avoid handling those Helm charts.
func loadHelmCharts(helmCharts []string) {
	certsuiteConfig.SkipHelmChartList = nil
	for _, chart := range helmCharts {
		certsuiteHelmChart := configuration.SkipHelmChartList{Name: chart}
		certsuiteConfig.SkipHelmChartList = append(certsuiteConfig.SkipHelmChartList, certsuiteHelmChart)
	}
}

// loadProtocolNames stores a list of acceptable protocol names
//
// This function replaces the current collection of valid protocol identifiers
// in the configuration with a new slice supplied by the caller. It first clears
// any previously stored values to avoid residual data, then assigns the
// provided slice directly to the global configuration variable. No return value
// is produced.
func loadProtocolNames(protocolNames []string) {
	certsuiteConfig.ValidProtocolNames = nil
	certsuiteConfig.ValidProtocolNames = protocolNames
}

// loadServices sets the list of services to ignore
//
// The function replaces any existing ignored service entries with a new slice
// provided as input. It first clears the current configuration's ignore list,
// then assigns the supplied list directly. The resulting configuration is used
// elsewhere to skip checks for these services.
func loadServices(services []string) {
	certsuiteConfig.ServicesIgnoreList = nil
	certsuiteConfig.ServicesIgnoreList = services
}

// loadNonScalableDeployments parses a list of non-scalable deployments to skip scaling tests
//
// The function receives an array of strings where each entry contains a
// deployment name and namespace separated by a slash. It clears any previously
// stored entries, then splits each string into its two parts; if the format is
// invalid it logs an error and aborts. Valid pairs are converted into
// configuration objects that are appended to the global skip list.
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

// loadNonScalableStatefulSets Parses a list of non-scalable StatefulSet identifiers to skip scaling tests
//
// The function takes an array of strings, each expected in the form
// "name/namespace", splits them into name and namespace components, validates
// the format, and appends the parsed information to a global configuration
// slice. If any string does not contain exactly two parts separated by a slash,
// it logs an error and aborts further processing.
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

// loadProbeDaemonSetNamespace Sets the Probe DaemonSet namespace in the configuration
//
// The function receives a list of strings and assigns the first element to the
// ProbeDaemonSetNamespace field of the shared configuration object. It assumes
// that the slice contains at least one entry and uses it directly without
// validation or conversion.
func loadProbeDaemonSetNamespace(namespace []string) {
	certsuiteConfig.ProbeDaemonSetNamespace = namespace[0]
}
