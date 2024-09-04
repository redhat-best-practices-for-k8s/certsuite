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

type configOption struct {
	Option string
	Help   string
}

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

func loadNamespaces(namespaces []string) {
	certsuiteConfig.TargetNameSpaces = nil
	for _, namespace := range namespaces {
		certsuiteNamespace := configuration.Namespace{Name: namespace}
		certsuiteConfig.TargetNameSpaces = append(certsuiteConfig.TargetNameSpaces, certsuiteNamespace)
	}
}

func loadPodLabels(podLabels []string) {
	certsuiteConfig.PodsUnderTestLabels = nil
	certsuiteConfig.PodsUnderTestLabels = podLabels
}

func loadOperatorLabels(operatorLabels []string) {
	certsuiteConfig.OperatorsUnderTestLabels = nil
	certsuiteConfig.OperatorsUnderTestLabels = operatorLabels
}

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

func loadManagedDeployments(deployments []string) {
	certsuiteConfig.ManagedDeployments = nil
	for _, deployment := range deployments {
		certsuiteManagedDeployment := configuration.ManagedDeploymentsStatefulsets{Name: deployment}
		certsuiteConfig.ManagedDeployments = append(certsuiteConfig.ManagedDeployments, certsuiteManagedDeployment)
	}
}

func loadManagedStatefulSets(statefulSets []string) {
	certsuiteConfig.ManagedStatefulsets = nil
	for _, statefulSet := range statefulSets {
		certsuiteManagedStatefulSet := configuration.ManagedDeploymentsStatefulsets{Name: statefulSet}
		certsuiteConfig.ManagedStatefulsets = append(certsuiteConfig.ManagedStatefulsets, certsuiteManagedStatefulSet)
	}
}

func loadAcceptedKernelTaints(taints []string) {
	certsuiteConfig.AcceptedKernelTaints = nil
	for _, taint := range taints {
		certsuiteKernelTaint := configuration.AcceptedKernelTaintsInfo{Module: taint}
		certsuiteConfig.AcceptedKernelTaints = append(certsuiteConfig.AcceptedKernelTaints, certsuiteKernelTaint)
	}
}

func loadHelmCharts(helmCharts []string) {
	certsuiteConfig.SkipHelmChartList = nil
	for _, chart := range helmCharts {
		certsuiteHelmChart := configuration.SkipHelmChartList{Name: chart}
		certsuiteConfig.SkipHelmChartList = append(certsuiteConfig.SkipHelmChartList, certsuiteHelmChart)
	}
}

func loadProtocolNames(protocolNames []string) {
	certsuiteConfig.ValidProtocolNames = nil
	certsuiteConfig.ValidProtocolNames = protocolNames
}

func loadServices(services []string) {
	certsuiteConfig.ServicesIgnoreList = nil
	certsuiteConfig.ServicesIgnoreList = services
}

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

func loadProbeDaemonSetNamespace(namespace []string) {
	certsuiteConfig.ProbeDaemonSetNamespace = namespace[0]
}
