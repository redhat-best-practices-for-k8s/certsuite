package webserver

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"net/http"

	yaml "gopkg.in/yaml.v2"

	"github.com/gorilla/websocket"
	"github.com/robert-nix/ansihtml"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
)

//go:embed index.html
var indexHTML []byte

//go:embed submit.js
var submit []byte

//go:embed logs.js
var logs []byte

//go:embed toast.js
var toast []byte
var Buf *bytes.Buffer

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func logStreamHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()
	// Create a scanner to read the log file line by line
	for {
		scanner := bufio.NewScanner(Buf)
		for scanner.Scan() {
			line := scanner.Bytes()
			line = append(ansihtml.ConvertToHTML(line), []byte("<br>")...)

			// Send each log line to the client
			if err := conn.WriteMessage(websocket.TextMessage, line); err != nil {
				fmt.Println(err)
				return
			}
		}
		if err := scanner.Err(); err != nil {
			logrus.Printf("Error reading log file: %v", err)
			return
		}
	}
}

type RequestedData struct {
	SelectedOptions                      []string `json:"selectedOptions"`
	TargetNameSpaces                     []string `json:"targetNameSpaces"`
	PodsUnderTestLabels                  []string `json:"podsUnderTestLabels"`
	OperatorsUnderTestLabels             []string `json:"operatorsUnderTestLabels"`
	ManagedDeployments                   []string `json:"managedDeployments"`
	ManagedStatefulsets                  []string `json:"managedStatefulsets"`
	SkipScalingTestDeploymentsnamespace  []string `json:"skipScalingTestDeploymentsnamespace"`
	SkipScalingTestDeploymentsname       []string `json:"skipScalingTestDeploymentsname"`
	SkipScalingTestStatefulsetsnamespace []string `json:"skipScalingTestStatefulsetsnamespace"`
	SkipScalingTestStatefulsetsname      []string `json:"skipScalingTestStatefulsetsname"`
	TargetCrdFiltersnameSuffix           []string `json:"targetCrdFiltersnameSuffix"`
	TargetCrdFiltersscalable             []string `json:"targetCrdFiltersscalable"`
	AcceptedKernelTaints                 []string `json:"acceptedKernelTaints"`
	SkipHelmChartList                    []string `json:"skipHelmChartList"`
	Servicesignorelist                   []string `json:"servicesignorelist"`
	ValidProtocolNames                   []string `json:"ValidProtocolNames"`
	DebugDaemonSetNamespace              []string `json:"DebugDaemonSetNamespace"`
	CollectorAppEndPoint                 []string `json:"CollectorAppEndPoint"`
	ExecutedBy                           []string `json:"executedBy"`
	CollectorAppPassword                 []string `json:"CollectorAppPassword"`
	PartnerName                          []string `json:"PartnerName"`
}
type ResponseData struct {
	Message string `json:"message"`
}

func FlattenData(data interface{}, result []string) []string {
	switch v := data.(type) {
	case string:
		result = append(result, v)
	case []interface{}:
		for _, item := range v {
			result = FlattenData(item, result)
		}
	case map[string]interface{}:
		for key, item := range v {
			if key == "selectedOptions" {
				result = FlattenData(item, result)
			}
			result = FlattenData(item, result)
		}
	}
	return result
}
func HandlereqFunc() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to "text/html".
		w.Header().Set("Content-Type", "text/html")
		// Write the embedded HTML content to the response.
		_, err := w.Write(indexHTML)
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/submit.js", func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to "application/javascript".
		w.Header().Set("Content-Type", "application/javascript")
		// Write the embedded JavaScript content to the response.
		_, err := w.Write(submit)
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/logs.js", func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to "application/javascript".
		w.Header().Set("Content-Type", "application/javascript")
		// Write the embedded JavaScript content to the response.
		_, err := w.Write(logs)
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/toast.js", func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to "application/javascript".
		w.Header().Set("Content-Type", "application/javascript")
		// Write the embedded JavaScript content to the response.
		_, err := w.Write(toast)
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})
	// Serve the static HTML file
	http.HandleFunc("/logstream", logStreamHandler)
}

func UpdateTnf(tnf_config []byte, data RequestedData) []byte {
	// Unmarshal the YAML data into a Config struct
	var config configuration.TestConfiguration

	err := yaml.Unmarshal(tnf_config, &config)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML: %v", err)
	}

	// Modify the configuration
	var namespace []configuration.Namespace
	for _, tnamespace := range data.TargetNameSpaces {
		namespace = append(namespace, configuration.Namespace{Name: tnamespace})
	}
	config.TargetNameSpaces = namespace

	config.PodsUnderTestLabels = data.PodsUnderTestLabels

	config.OperatorsUnderTestLabels = data.OperatorsUnderTestLabels

	var managedDeployments []configuration.ManagedDeploymentsStatefulsets
	for _, val := range data.ManagedDeployments {
		managedDeployments = append(managedDeployments, configuration.ManagedDeploymentsStatefulsets{Name: val})
	}
	config.ManagedDeployments = managedDeployments

	var managedStatefulsets []configuration.ManagedDeploymentsStatefulsets
	for _, val := range data.ManagedDeployments {
		managedStatefulsets = append(managedStatefulsets, configuration.ManagedDeploymentsStatefulsets{Name: val})
	}
	config.ManagedStatefulsets = managedStatefulsets

	var crdFilter []configuration.CrdFilter
	for i := range data.TargetCrdFiltersnameSuffix {
		val := true
		if data.TargetCrdFiltersscalable[i] == "false" {
			val = false
		}
		crdFilter = append(crdFilter, configuration.CrdFilter{NameSuffix: data.TargetCrdFiltersnameSuffix[i],
			Scalable: val})
	}
	config.CrdFilters = crdFilter

	var acceptedKernelTaints []configuration.AcceptedKernelTaintsInfo
	for _, val := range data.AcceptedKernelTaints {
		acceptedKernelTaints = append(acceptedKernelTaints, configuration.AcceptedKernelTaintsInfo{Module: val})
	}
	config.AcceptedKernelTaints = acceptedKernelTaints

	var skipHelmChartList []configuration.SkipHelmChartList
	for _, val := range data.AcceptedKernelTaints {
		skipHelmChartList = append(skipHelmChartList, configuration.SkipHelmChartList{Name: val})
	}
	config.SkipHelmChartList = skipHelmChartList

	var skipScalingTestDeployments []configuration.SkipScalingTestDeploymentsInfo
	for i := range data.SkipScalingTestDeploymentsname {
		skipScalingTestDeployments = append(skipScalingTestDeployments, configuration.SkipScalingTestDeploymentsInfo{Name: data.SkipScalingTestDeploymentsname[i],
			Namespace: data.SkipScalingTestDeploymentsnamespace[i]})
	}
	config.SkipScalingTestDeployments = skipScalingTestDeployments

	var skipScalingTestStatefulSets []configuration.SkipScalingTestStatefulSetsInfo
	for i := range data.SkipScalingTestStatefulsetsname {
		skipScalingTestStatefulSets = append(skipScalingTestStatefulSets, configuration.SkipScalingTestStatefulSetsInfo{Name: data.SkipScalingTestStatefulsetsname[i],
			Namespace: data.SkipScalingTestStatefulsetsnamespace[i]})
	}
	config.SkipScalingTestStatefulSets = skipScalingTestStatefulSets

	config.ServicesIgnoreList = data.Servicesignorelist
	config.ValidProtocolNames = data.ValidProtocolNames
	if len(data.CollectorAppEndPoint) > 0 {
		config.CollectorAppEndPoint = data.CollectorAppEndPoint[0]
	}
	if len(data.CollectorAppPassword) > 0 {
		config.CollectorAppPassword = data.CollectorAppPassword[0]
	}
	if len(data.ExecutedBy) > 0 {
		config.ExecutedBy = data.ExecutedBy[0]
	}
	if len(data.PartnerName) > 0 {
		config.PartnerName = data.PartnerName[0]
	}
	if len(data.DebugDaemonSetNamespace) > 0 {
		config.DebugDaemonSetNamespace = data.DebugDaemonSetNamespace[0]
	}

	// Serialize the modified config back to YAML format
	newData, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatalf("Error marshaling YAML: %v", err)
	}
	return newData
}
