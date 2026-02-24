// Copyright (C) 2023-2026 Red Hat, Inc.
package webserver

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/arrayhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/certsuite"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
	"github.com/robert-nix/ansihtml"

	yaml "gopkg.in/yaml.v3"
)

type webServerContextKey string

const (
	logTimeout = 1000

	readTimeoutSeconds = 10
)

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

var buf *bytes.Buffer

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func logStreamHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Info("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()
	// Create a scanner to read the log file line by line
	for {
		scanner := bufio.NewScanner(buf)
		for scanner.Scan() {
			line := scanner.Bytes()
			fmt.Println(string(line))
			line = append(ansihtml.ConvertToHTML(line), []byte("<br>")...)

			// Send each log line to the client
			if err := conn.WriteMessage(websocket.TextMessage, line); err != nil {
				fmt.Println(err)
				return
			}
			time.Sleep(logTimeout)
		}
		if err := scanner.Err(); err != nil {
			log.Info("Error reading log file: %v", err)
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
	ProbeDaemonSetNamespace              []string `json:"ProbeDaemonSetNamespace"`
	CollectorAppEndPoint                 []string `json:"CollectorAppEndPoint"`
	ExecutedBy                           []string `json:"executedBy"`
	CollectorAppPassword                 []string `json:"CollectorAppPassword"`
	PartnerName                          []string `json:"PartnerName"`
	ConnectAPIKey                        []string `json:"key,omitempty"`
	ConnectProjectID                     []string `json:"projectID,omitempty"`
	ConnectAPIBaseURL                    []string `json:"baseURL,omitempty"`
	ConnectAPIProxyURL                   []string `json:"proxyURL,omitempty"`
	ConnectAPIProxyPort                  []string `json:"proxyPort,omitempty"`
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
	http.HandleFunc("/classification.js", func(w http.ResponseWriter, r *http.Request) {
		classification := outputTestCases()

		// Set the content type to "application/javascript".
		w.Header().Set("Content-Type", "application/javascript")
		// Write the embedded JavaScript content to the response.
		_, err := w.Write([]byte(classification))
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})

	// Serve the static HTML file
	http.HandleFunc("/logstream", logStreamHandler)
}

func StartServer(outputFolder string) {
	ctx := context.TODO()
	server := &http.Server{
		Addr:        ":8084",                          // Server address
		ReadTimeout: readTimeoutSeconds * time.Second, // Maximum duration for reading the entire request
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, outputFolderCtxKey, outputFolder)
			return ctx
		},
	}

	installReqHandlers()

	http.HandleFunc("/runFunction", runHandler)

	log.Info("Server is running on :8084...")
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

// Define an HTTP handler that triggers CERTSUITE tests
//
//nolint:funlen
func runHandler(w http.ResponseWriter, r *http.Request) {
	buf = bytes.NewBufferString("")
	// The log output will be written to the log file and to this buffer buf
	log.SetLogger(log.GetMultiLogger(buf))

	jsonData := r.FormValue("jsonData") // "jsonData" is the name of the JSON input field
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

	log.Info("Kubeconfig file name received: %s", fileHeader.Filename)
	kubeconfigTempFile, err := os.CreateTemp("", "webserver-kubeconfig-*")
	if err != nil {
		http.Error(w, "Failed to create temp file to store the kubeconfig content.", http.StatusBadRequest)
		return
	}

	defer func() {
		log.Info("Removing temporary kubeconfig file %s", kubeconfigTempFile.Name())
		err = os.Remove(kubeconfigTempFile.Name())
		if err != nil {
			log.Error("Failed to remove temp kubeconfig file %s", kubeconfigTempFile.Name())
		}
	}()

	_, err = io.Copy(kubeconfigTempFile, file)
	if err != nil {
		http.Error(w, "Unable to copy file", http.StatusInternalServerError)
		return
	}

	_ = kubeconfigTempFile.Close()

	log.Info("Web Server kubeconfig file : %v (copied into %v)", fileHeader.Filename, kubeconfigTempFile.Name())
	log.Info("Web Server Labels filter   : %v", flattenedOptions)

	tnfConfig, err := os.ReadFile("config/certsuite_config.yml")
	if err != nil {
		log.Fatal("Error reading YAML file: %v", err) //nolint:gocritic // exitAfterDefer
	}

	newData := updateTnf(tnfConfig, &data)

	// Write the modified YAML data back to the file
	var filePerm fs.FileMode = 0o644 // owner can read/write, group and others can only read
	err = os.WriteFile("config/certsuite_config.yml", newData, filePerm)
	if err != nil {
		log.Fatal("Error writing YAML file: %v", err)
	}
	labelsFilter := strings.Join(flattenedOptions, ",")

	_ = clientsholder.GetNewClientsHolder(kubeconfigTempFile.Name())
	certsuite.LoadChecksDB(labelsFilter)

	outputFolder := r.Context().Value(outputFolderCtxKey).(string)

	if err := checksdb.InitLabelsExprEvaluator(labelsFilter); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize a test case label evaluator, err: %v", err)
		os.Exit(1)
	}

	if err := log.CreateGlobalLogFile(outputFolder, "debug"); err != nil {
		fmt.Fprintf(os.Stderr, "Could not create the log file, err: %v\n", err)
		os.Exit(1)
	}

	log.Info("Running CNF Cert Suite (web-mode). Labels filter: %s, outputFolder: %s", labelsFilter, outputFolder)
	err = certsuite.Run(labelsFilter, outputFolder)
	if err != nil {
		log.Error("Failed to run CNF Cert Suite: %v", err)
	}

	// Return the result as JSON
	response := struct {
		Message string `json:"Message"`
	}{
		Message: fmt.Sprintf("Succeeded to run %s", strings.Join(flattenedOptions, " ")),
	}
	// Serialize the response data to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Error("Failed to marshal jsonResponse: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to specify that the response is JSON
	w.Header().Set("Content-Type", "application/json")
	// Write the JSON response to the client
	log.Info("Sending web response: %v", response)
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Error("Failed to write jsonResponse: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

//nolint:funlen,gocyclo
func updateTnf(tnfConfig []byte, data *RequestedData) []byte {
	// Unmarshal the YAML data into a Config struct
	var config configuration.TestConfiguration

	err := yaml.Unmarshal(tnfConfig, &config)
	if err != nil {
		log.Fatal("Error unmarshalling YAML: %v", err)
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
	if len(data.CollectorAppPassword) > 0 {
		config.CollectorAppPassword = data.CollectorAppPassword[0]
	}
	if len(data.ExecutedBy) > 0 {
		config.ExecutedBy = data.ExecutedBy[0]
	}
	if len(data.PartnerName) > 0 {
		config.PartnerName = data.PartnerName[0]
	}
	if len(data.ProbeDaemonSetNamespace) > 0 {
		config.ProbeDaemonSetNamespace = data.ProbeDaemonSetNamespace[0]
	}
	if len(data.ConnectAPIKey) > 0 {
		config.ConnectAPIConfig.APIKey = data.ConnectAPIKey[0]
	}
	if len(data.ConnectProjectID) > 0 {
		config.ConnectAPIConfig.ProjectID = data.ConnectProjectID[0]
	}
	if len(data.ConnectAPIBaseURL) > 0 {
		config.ConnectAPIConfig.BaseURL = data.ConnectAPIBaseURL[0]
	}
	if len(data.ConnectAPIProxyURL) > 0 {
		config.ConnectAPIConfig.ProxyURL = data.ConnectAPIProxyURL[0]
	}
	if len(data.ConnectAPIProxyPort) > 0 {
		config.ConnectAPIConfig.ProxyPort = data.ConnectAPIProxyPort[0]
	}

	// Serialize the modified config back to YAML format
	newData, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatal("Error marshaling YAML: %v", err)
	}
	return newData
}

// outputTestCases outputs the Markdown representation for test cases from the catalog to stdout.
func outputTestCases() (outString string) {
	// Building a separate data structure to store the key order for the map
	keys := make([]claim.Identifier, 0, len(identifiers.Catalog))
	for k := range identifiers.Catalog {
		keys = append(keys, k)
	}

	// Sorting the map by identifier ID
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Id < keys[j].Id
	})

	catalog := CreatePrintableCatalogFromIdentifiers(keys)
	if catalog == nil {
		return
	}
	// we need the list of suite's names
	suites := GetSuitesFromIdentifiers(keys)

	// Sort the list of suite names
	sort.Strings(suites)

	// Iterating the map by test and suite names
	outString = "classification= {\n"
	for _, suite := range suites {
		for _, k := range catalog[suite] {
			classificationString := "\"categoryClassification\": "
			// Every paragraph starts with a new line.

			outString += fmt.Sprintf("%q: [\n{\n", k.identifier.Id)
			outString += fmt.Sprintf("\"description\": %q,\n", strings.ReplaceAll(strings.ReplaceAll(identifiers.Catalog[k.identifier].Description, "\n", " "), "\"", " "))
			outString += fmt.Sprintf("\"remediation\": %q,\n", strings.ReplaceAll(strings.ReplaceAll(identifiers.Catalog[k.identifier].Remediation, "\n", " "), "\"", " "))
			outString += fmt.Sprintf("\"bestPracticeReference\": %q,\n", strings.ReplaceAll(identifiers.Catalog[k.identifier].BestPracticeReference, "\n", " "))
			outString += classificationString + toJSONString(identifiers.Catalog[k.identifier].CategoryClassification) + ",\n}\n]\n,"
		}
	}
	outString += "}"
	return outString
}
func toJSONString(data map[string]string) string {
	// Convert the map to a JSON-like string
	jsonbytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return ""
	}

	return string(jsonbytes)
}
func GetSuitesFromIdentifiers(keys []claim.Identifier) []string {
	var suites []string
	for _, i := range keys {
		suites = append(suites, i.Suite)
	}
	return arrayhelper.Unique(suites)
}

type Entry struct {
	testName   string
	identifier claim.Identifier // {url and version}
}

func CreatePrintableCatalogFromIdentifiers(keys []claim.Identifier) map[string][]Entry {
	catalog := make(map[string][]Entry)
	// we need the list of suite's names
	for _, i := range keys {
		catalog[i.Suite] = append(catalog[i.Suite], Entry{
			testName:   i.Id,
			identifier: i,
		})
	}
	return catalog
}
