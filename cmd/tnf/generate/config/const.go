package config

// Menu names
const (
	// Main menu
	create = "Create"
	show   = "Show"
	save   = "Save"
	close  = "Exit"
	// Create
	cnfResources = "CNF resources"
	exceptions   = "Exceptions"
	collector    = "Collector"
	settings     = "Settings"
	previousMenu = "\U0001F878"
	// CNF resources
	namespaces          = "Namespaces"
	pods                = "Pods"
	operators           = "Operators"
	crdFilters          = "CRD filters"
	managedDeployments  = "Managed Deployments"
	managedStatefulSets = "Managed StatefulSets"
	// Exceptions
	kernelTaints            = "Kernel taints"
	helmCharts              = "Helm charts"
	protocolNames           = "Protocol names"
	services                = "Services"
	nonScalableDeployments  = "Non-scalable Deployments"
	nonScalableStatefulSets = "Non-scalable StatefulSets"
	// Collector
	appEndPoint = "Application end point"
	executedBy  = "Certification executor"
	partnerName = "Partner name"
	appPassword = "Application password"
	// Settings
	debugDaemonSet = "Debug DaemonSet namespace"
)

// Menu help
const (
	// Main menu
	createConfigHelp = "Create a configuration for the CNF Certification Suite"
	showConfigHelp   = "Show the current configuration in YAML format"
	saveConfigHelp   = `Save the current configuration to a YAML file (default "tnf_config.yaml")`
	exitHelp         = "Exit the tool (changes not saved will be lost)"
	backHelp         = "Move to previous menu"
	// Create
	cnfResourcesHelp = `Configure the workload resources of the CNF to be verified.
Only the resources that the CNF uses are required to be configured. The rest can be left empty.
Usually a basic configuration includes "Namespaces" and "Pods" at least.`
	exceptionsdHelp = `Allow adding exceptions to skip several checks for different resources.
The exceptions must be justified in order to pass the CNF Certification. Feedback
regarding the exceptions configured can be provided in an HTML page after loading
the claim.json file with the results.`
	collectordHelp = `Parameters required to send the CNF Certification Suite results to a data collector.`
	settingsHelp   = `Configure various settings for the CNF Certification Suite.`
	// CNF resources
	namespacesHelp = `The namespaces in which the CNF under test will be deployed.`

	podLabelsHelp = `The labels that each Pod of the CNF under test must have to be verified
by the CNF Certification Suite.
If a new label is used for this purpose make sure it is added to the CNF's Pods,
ideally in the pod's definition as the on-the-fly labels are lost if the Pod gets
rescheduled.
For Pods own by a Deployment, the same label as the one defined in the
"spec.selector.matchLabels" section of the Deployment can be used.`

	operatorLabelsHelp = `The labels that each operator's CSV of the CNF under test must have to be verified
by the CNF Certification Suite.
If a new label is used for this purpose make sure it is added to the CNF operator's CSVs.`

	crdFiltersHelp = `The CRD name suffix used to filter the CNF's CRDs among all the CRDs present in the cluster.
It must also be specified if the resources own by the CRD are scalable or not in order to avoid
some lifecycle test cases.`
	// Exceptions
	kernelTaintsHelp = `The list of kernel modules loaded by the CNF that make the Linux kernel mark itself
as "tainted" but that should skip verification.
Test cases affected: platform-alteration-tainted-node-kernel.`
	helmChartsHelp = `The list of Helm charts that the CNF uses whose certification status will not be verified.
If no exception is configured, the certification status for all Helm charts will be checked
in the OpenShift Helms Charts repository (see https://charts.openshift.io/).
Test cases affected: affiliated-certification-helmchart-is-certified`
	protocolNamesHelp = `The list of allowed protocol names to be used for container port names.
The name field of a container port must be of the form <protocol>[-<suffix>] where <protocol> must
be allowed by default or added to this list. The optional <suffix> can be chosen by the application.
Protocol names allowed by default: "grpc", "grpc-web", "http", "http2", "tcp", "udp".
Test cases affected: manageability-container-port-name-format.`
	// Collector (TODO)
	// Settings
	debugDaemonSetHelp = `Set the namespace where the debug DaemonSet will be deployed.
The namespace will be created in case it does not exist. If not set, the default namespace
is "cnf-suite".
This DaemonSet, called "tnf-debug" is deployed and used internally by the CNF Certification Suite
to issue some shell commands that are needed in certain test cases. Some of these test cases might
fail or be skipped in case it is not deployed correctly.`
)

// Prompts, syxtax, examples
const (
	// CNF resources
	namespacePrompt  = "Write a comma-separated list of the namespaces in which the CNF is deploying its workload."
	namespaceSyntax  = "ns1[, <ns2>]..."
	namespaceExample = "cnf, cnf-workload"
	podsPrompt       = "Write a comma-separated list of labels to identify the CNF's Pods under test."
	podsSyntax       = "pod-label-1[, pod-label-2]..."
	podsExample      = "test-network-function.com/generic: target"
	operatorsPrompt  = "Write a comma-separated list of labels to identify the CNF's operators under test."
	operatorsSyntax  = "operator-label-1[, operator-label-2]..."
	operatorsExample = "test-network-function.com/operator1: target"
	crdFiltersPrompt = "Write a comma-separated list of the CRD's name suffixes that the CNF contains. Also, specify if the\n" +
		"resources managed by those CRDs are scalable."
	crdFiltersSyntax  = "crd-name-suffix/{true|false}[,crd-name-suffix/{true|false}]..."
	crdFiltersExample = "group1.test.com/false"
	// Exceptions
	kernelTaintsPrompt   = "Write a comma-separated list of kernel taints (modules)"
	kernelTaintsSyntax   = "mod1[,mod2]..."
	kernelTaintsExample  = "vboxsf, vboxguest"
	helmChartsPrompt     = "Enter a comma-separated list of Helm charts that will skip verification."
	helmChartsSyntax     = "chart1[,chart2]..."
	helmChartsExample    = "coredns"
	protocolNamesPrompt  = "Enter a comma-separated list of protocol names"
	protocolNamesSyntax  = "proto1[,proto2]..."
	protocolNamesExample = "http3, sctp"
	// Collector (TODO)
	// Settings
	debugDaemonSetPrompt  = "Enter the namespace in which de debug DaemonSet will be deployed."
	debugDaemonSetSyntax  = "ds-namespace"
	debugDaemonSetExample = "cnf-cert"
)
