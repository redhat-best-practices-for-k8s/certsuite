package webserver

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	rlog "log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/robert-nix/ansihtml"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/certsuite"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	yaml "gopkg.in/yaml.v2"
)

type webServerContextKey string

var (
	outputFolderCtxKey webServerContextKey = "output-folder"
)

//go:embed index.html
var indexHTML []byte

//go:embed submit.js
var submit []byte

//go:embed logs.js
var logs []byte

//go:embed toast.js
var toast []byte

//go:embed index.js
var index []byte

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
			fmt.Println(string(line))
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

func installReqHandlers() {
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

	http.HandleFunc("/index.js", func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to "application/javascript".
		w.Header().Set("Content-Type", "application/javascript")
		// Write the embedded JavaScript content to the response.
		_, err := w.Write(index)
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})

	// Serve the static HTML file
	http.HandleFunc("/logstream", logStreamHandler)
}

func StartServer(outputFolder string) {
	ctx := context.Background()
	server := &http.Server{
		Addr:        ":8084",          // Server address
		ReadTimeout: 10 * time.Second, // Maximum duration for reading the entire request
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, outputFolderCtxKey, outputFolder)
			return ctx
		},
	}

	installReqHandlers()

	http.HandleFunc("/runFunction", runHandler)

	logrus.Infof("Server is running on :8084...")
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

// Define an HTTP handler that triggers Ginkgo tests
//
//nolint:funlen
func runHandler(w http.ResponseWriter, r *http.Request) {
	Buf = bytes.NewBufferString("")
	logrus.SetOutput(Buf)
	rlog.SetOutput(Buf)

	jsonData := r.FormValue("jsonData") // "jsonData" is the name of the JSON input field
	logrus.Info(jsonData)
	var data RequestedData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		fmt.Println("Error:", err)
	}
	flattenedOptions := data.SelectedOptions

	// Get the file from the request
	file, fileHeader, err := r.FormFile("kubeConfigPath") // "fileInput" is the name of the file input field
	if err != nil {
		http.Error(w, "Unable to retrieve file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	logrus.Infof("Kubeconfig file name received: %s", fileHeader.Filename)
	kubeconfigTempFile, err := os.CreateTemp("", "webserver-kubeconfig-*")
	if err != nil {
		http.Error(w, "Failed to create temp file to store the kubeconfig content.", http.StatusBadRequest)
		return
	}

	defer func() {
		logrus.Infof("Removing temporary kubeconfig file %s", kubeconfigTempFile.Name())
		err = os.Remove(kubeconfigTempFile.Name())
		if err != nil {
			logrus.Errorf("Failed to remove temp kubeconfig file %s", kubeconfigTempFile.Name())
		}
	}()

	_, err = io.Copy(kubeconfigTempFile, file)
	if err != nil {
		http.Error(w, "Unable to copy file", http.StatusInternalServerError)
		return
	}

	_ = kubeconfigTempFile.Close()

	logrus.Infof("Web Server kubeconfig file : %v (copied into %v)", fileHeader.Filename, kubeconfigTempFile.Name())
	logrus.Infof("Web Server Labels filter   : %v", flattenedOptions)

	tnfConfig, err := os.ReadFile("tnf_config.yml")
	if err != nil {
		logrus.Fatalf("Error reading YAML file: %v", err)
	}

	newData := updateTnf(tnfConfig, &data)

	// Write the modified YAML data back to the file
	err = os.WriteFile("tnf_config.yml", newData, os.ModePerm)
	if err != nil {
		logrus.Fatalf("Error writing YAML file: %v", err)
	}
	_ = clientsholder.GetNewClientsHolder(kubeconfigTempFile.Name())

	var env provider.TestEnvironment
	env.SetNeedsRefresh()
	env = provider.GetTestEnvironment()

	labelsFilter := strings.Join(flattenedOptions, "")
	outputFolder := r.Context().Value(outputFolderCtxKey).(string)

	logrus.Infof("Running CNF Cert Suite (web-mode). Labels filter: %s, outputFolder: %s", labelsFilter, outputFolder)
	certsuite.Run(labelsFilter, outputFolder)

	// Return the result as JSON
	response := struct {
		Message string `json:"Message"`
	}{
		Message: fmt.Sprintf("Succeeded to run %s", strings.Join(flattenedOptions, "")),
	}
	// Serialize the response data to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		logrus.Errorf("Failed to marshal jsonResponse: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to specify that the response is JSON
	w.Header().Set("Content-Type", "application/json")
	// Write the JSON response to the client
	logrus.Infof("Sending web response: %v", response)
	_, err = w.Write(jsonResponse)
	if err != nil {
		logrus.Errorf("Failed to write jsonResponse: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

//nolint:funlen
func updateTnf(tnfConfig []byte, data *RequestedData) []byte {
	// Unmarshal the YAML data into a Config struct
	var config configuration.TestConfiguration

	err := yaml.Unmarshal(tnfConfig, &config)
	if err != nil {
		logrus.Fatalf("Error unmarshalling YAML: %v", err)
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
	for _, val := range data.SkipHelmChartList {
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
		logrus.Fatalf("Error marshaling YAML: %v", err)
	}
	return newData
}
