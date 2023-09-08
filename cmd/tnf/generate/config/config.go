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
	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
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
	// generateConfiCmd is a helper tool to generate a CNF config YAML file
	generateConfigCmd = &cobra.Command{
		Use:   "config",
		Short: "Generates a CNF config YAML file with user input.",
		Run: func(cmd *cobra.Command, args []string) {
			generateConfig()
		},
	}
)

var tnfConfig = configuration.TestConfiguration{}

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
		{Option: close, Help: exitHelp},
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
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
		switch mainMenu[opt].Option {
		case create:
			createConfiguration()
		case show:
			showConfiguration(&tnfConfig)
		case save:
			saveConfiguration(&tnfConfig)
		case close:
			exit = true
		}
	}
}

func createConfiguration() {
	createMenu := []configOption{
		{Option: cnfResources, Help: cnfResourcesHelp},
		{Option: exceptions, Help: exceptionsdHelp},
		{Option: collector, Help: collectordHelp},
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
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
		switch createMenu[i].Option {
		case cnfResources:
			createCnfResourcesConfiguration()
		case exceptions:
			createExceptionsConfiguration()
		case collector:
			createCollectorConfiguration()
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
	fmt.Println("================= CNF CONFIGURATION =================")
	fmt.Println(string(configYaml))
	fmt.Println("=====================================================")
}

func saveConfiguration(config *configuration.TestConfiguration) {
	configYaml, err := yaml.Marshal(config)
	if err != nil {
		log.Printf("could not marshal the YAML file, err: %v", err)
		return
	}

	const filePermissions = 0644
	err = os.WriteFile("tnf_config.yml", configYaml, filePermissions) // TODO: make the file configurable
	if err != nil {
		log.Printf("could not write file, err: %v", err)
		return
	}

	fmt.Println(color.GreenString("Configuration saved"))
}

func createCnfResourcesConfiguration() {
	cnfResourcesOptions := []configOption{
		{Option: namespaces, Help: namespacesHelp},
		{Option: pods, Help: podLabelsHelp},
		{Option: operators, Help: operatorLabelsHelp},
		{Option: crdFilters, Help: crdFiltersHelp},
		{Option: managedDeployments, Help: ""},
		{Option: managedStatefulSets, Help: ""},
		{Option: previousMenu, Help: backHelp},
	}
	cnfResourcesSearcher := func(input string, index int) bool {
		basicOption := cnfResourcesOptions[index]
		name := strings.Replace(strings.ToLower(basicOption.Option), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}
	cnfResourcesPrompt := promptui.Select{
		Label:        "",
		Items:        cnfResourcesOptions,
		Templates:    templates,
		Size:         7,
		Searcher:     cnfResourcesSearcher,
		HideSelected: true,
	}
	var exit bool
	for !exit {
		i, _, err := cnfResourcesPrompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
		switch cnfResourcesOptions[i].Option {
		case namespaces:
			loadNamespaces(getAnswer(namespacePrompt, namespaceSyntax, namespaceExample))
		case pods:
			loadPodLabels(getAnswer(podsPrompt, podsSyntax, podsExample))
		case operators:
			loadOperatorLabels(getAnswer(operatorsPrompt, operatorsSyntax, operatorsExample))
		case crdFilters:
			loadCRDfilters(getAnswer(crdFiltersPrompt, crdFiltersSyntax, crdFiltersExample))
		case managedDeployments:
			// TODO
		case managedStatefulSets:
			// TODO
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
		{Option: services, Help: ""},
		{Option: nonScalableDeployments, Help: ""},
		{Option: nonScalableStatefulSets, Help: ""},
		{Option: previousMenu, Help: backHelp},
	}
	exceptionsSearcher := func(input string, index int) bool {
		exceptionOption := exceptionsOptions[index]
		name := strings.Replace(strings.ToLower(exceptionOption.Option), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

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
			fmt.Printf("Prompt failed %v\n", err)
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
			// TODO
		case nonScalableDeployments:
			// TODO
		case nonScalableStatefulSets:
			// TODO
		case previousMenu:
			exit = true
		}
	}
}

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
		name := strings.Replace(strings.ToLower(collectorOption.Option), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

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
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
		switch collectorOptions[i].Option {
		case appEndPoint:
			// TODO
		case executedBy:
			// TODO
		case partnerName:
			// TODO
		case appPassword:
			// TODO
		case previousMenu:
			// TODO
			exit = true
		}
	}
}

func createSettingsConfiguration() {
	settingsOptions := []configOption{
		{Option: debugDaemonSet, Help: debugDaemonSetHelp},
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
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
		switch settingsOptions[i].Option {
		case debugDaemonSet:
			loadDebugDaemonSetNamespace(getAnswer(debugDaemonSetPrompt, debugDaemonSetSyntax, debugDaemonSetExample))
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
	tnfConfig.TargetNameSpaces = nil
	for _, namespace := range namespaces {
		tnfNamespace := configuration.Namespace{Name: namespace}
		tnfConfig.TargetNameSpaces = append(tnfConfig.TargetNameSpaces, tnfNamespace)
	}
}

func loadPodLabels(podLabels []string) {
	tnfConfig.PodsUnderTestLabels = nil
	tnfConfig.PodsUnderTestLabels = podLabels
}

func loadOperatorLabels(operatorLabels []string) {
	tnfConfig.OperatorsUnderTestLabels = nil
	tnfConfig.OperatorsUnderTestLabels = operatorLabels
}

func loadCRDfilters(crdFilters []string) {
	tnfConfig.CrdFilters = nil
	for _, crdFilterStr := range crdFilters {
		crdFilter := strings.Split(crdFilterStr, "/")
		crdFilterName := crdFilter[0]
		crdFilterScalable, err := strconv.ParseBool(crdFilter[1])
		if err != nil {
			log.Printf("could not parse CRD filter, err: %v", err)
			return
		}
		tnfCrdFilter := configuration.CrdFilter{NameSuffix: crdFilterName, Scalable: crdFilterScalable}
		tnfConfig.CrdFilters = append(tnfConfig.CrdFilters, tnfCrdFilter)
	}
}

func loadAcceptedKernelTaints(taints []string) {
	tnfConfig.AcceptedKernelTaints = nil
	for _, taint := range taints {
		tnfKernelTaint := configuration.AcceptedKernelTaintsInfo{Module: taint}
		tnfConfig.AcceptedKernelTaints = append(tnfConfig.AcceptedKernelTaints, tnfKernelTaint)
	}
}

func loadHelmCharts(helmCharts []string) {
	tnfConfig.SkipHelmChartList = nil
	for _, chart := range helmCharts {
		tnfHelmChart := configuration.SkipHelmChartList{Name: chart}
		tnfConfig.SkipHelmChartList = append(tnfConfig.SkipHelmChartList, tnfHelmChart)
	}
}

func loadProtocolNames(protocolNames []string) {
	tnfConfig.ValidProtocolNames = nil
	tnfConfig.ValidProtocolNames = protocolNames
}

func loadDebugDaemonSetNamespace(namespace []string) {
	tnfConfig.DebugDaemonSetNamespace = namespace[0]
}
