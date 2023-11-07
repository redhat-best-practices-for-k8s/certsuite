// Copyright (C) 2020-2023 Red Hat, Inc.
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

package suite

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	rlog "log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/pkg/claimhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/collector"
	"github.com/test-network-function/cnf-certification-test/pkg/loghelper"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"

	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/certification"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/chaostesting"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/manageability"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/observability"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/operator"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/performance"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/preflight"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/webserver"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/diagnostics"
)

const (
	claimPathFlagKey              = "claimloc"
	CnfCertificationTestSuiteName = "CNF Certification Test Suite"
	defaultClaimPath              = ".."
	defaultCliArgValue            = ""
	junitFlagKey                  = "junit"
	serverModeFlag                = "serverMode"
	TNFReportKey                  = "cnf-certification-test"
	extraInfoKey                  = "testsExtraInfo"
	defaultServerMode             = false
)

var (
	claimPath *string
	junitPath *string
	// GitCommit is the latest commit in the current git branch
	GitCommit string
	// GitRelease is the list of tags (if any) applied to the latest commit
	// in the current branch
	GitRelease string
	// GitPreviousRelease is the last release at the date of the latest commit
	// in the current branch
	GitPreviousRelease string
	// gitDisplayRelease is a string used to hold the text to display
	// the version on screen and in the claim file
	gitDisplayRelease string
	// ClaimFormat is the current version for the claim file format to be produced by the TNF test suite.
	// A client decoding this claim file must support decoding its specific version.
	ClaimFormatVersion string
	serverMode         *bool
)

func init() {
	claimPath = flag.String(claimPathFlagKey, defaultClaimPath,
		"the path where the claimfile will be output")
	junitPath = flag.String(junitFlagKey, defaultCliArgValue,
		"the path for the junit format report")
	serverMode = flag.Bool("serverMode", defaultServerMode, "run test with webserver")
}

// setLogLevel sets the log level for logrus based on the "TNF_LOG_LEVEL" environment variable
func setLogLevel() {
	params := configuration.GetTestParameters()

	var logLevel, err = log.ParseLevel(params.LogLevel)
	if err != nil {
		log.Error("TNF_LOG_LEVEL environment set with an invalid value, defaulting to DEBUG \n Valid values are:  trace, debug, info, warn, error, fatal, panic")
		logLevel = log.DebugLevel
	}

	log.Info("Log level set to: ", logLevel)
	log.SetLevel(logLevel)
}

func getK8sClientsConfigFileNames() []string {
	params := configuration.GetTestParameters()
	fileNames := []string{}
	if params.Kubeconfig != "" {
		fileNames = append(fileNames, params.Kubeconfig)
	}
	if params.Home != "" {
		kubeConfigFilePath := filepath.Join(params.Home, ".kube", "config")
		fileNames = append(fileNames, kubeConfigFilePath)
	}

	return fileNames
}

// getGitVersion returns the git display version: the latest previously released
// build in case this build is not released. Otherwise display the build version
func getGitVersion() string {
	if GitRelease == "" {
		gitDisplayRelease = "Unreleased build post " + GitPreviousRelease
	} else {
		gitDisplayRelease = GitRelease
	}

	return gitDisplayRelease + " ( " + GitCommit + " )"
}

func PreRun(t *testing.T) (claimData *claim.Claim, claimRoot *claim.Root) {
	// When running unit tests, skip the suite
	if os.Getenv("UNIT_TEST") != "" {
		t.Skip("Skipping test suite when running unit tests")
	}

	err := configuration.LoadEnvironmentVariables()
	if err != nil {
		log.Fatalf("could not load the environment variables, error: %v", err)
	}

	// Set up logging params for logrus
	loghelper.SetLogFormat()
	setLogLevel()

	ginkgoConfig, _ := ginkgo.GinkgoConfiguration()
	log.Infof("TNF Version         : %v", getGitVersion())
	log.Infof("Claim Format Version: %s", ClaimFormatVersion)
	log.Infof("Ginkgo Version      : %v", ginkgo.GINKGO_VERSION)
	log.Infof("Labels filter       : %v", ginkgoConfig.LabelFilter)
	log.Infof("run test with webserver      : %v", *serverMode)
	// Set clientsholder singleton with the filenames from the env vars.
	_ = clientsholder.GetClientsHolder(getK8sClientsConfigFileNames()...)

	// Initialize the claim with the start time, tnf version, etc.
	claimRoot = claimhelper.CreateClaimRoot()
	claimData = claimRoot.Claim
	claimData.Configurations = make(map[string]interface{})
	claimData.Nodes = make(map[string]interface{})
	incorporateVersions(claimData)

	configurations, err := claimhelper.MarshalConfigurations()
	if err != nil {
		log.Errorf("Configuration node missing because of: %s", err)
		t.FailNow()
	}
	claimData.Nodes = claimhelper.GenerateNodes()
	claimhelper.UnmarshalConfigurations(configurations, claimData.Configurations)
	return claimData, claimRoot
}

// TestTest invokes the CNF Certification Test Suite.
func TestTest(t *testing.T) {
	ginkgoConfig, _ := ginkgo.GinkgoConfiguration()
	if !*serverMode {
		claimData, claimRoot := PreRun(t)
		var diagnosticMode bool
		// Diagnostic functions will run when no labels are provided.

		if ginkgoConfig.LabelFilter == "" {
			log.Infof("TNF will run in diagnostic mode so no test case will be launched.")
			diagnosticMode = true
		}

		// initialize abort flag
		testhelper.AbortTrigger = ""

		// Run tests specs only if not in diagnostic mode, otherwise all TSs would run.
		var env provider.TestEnvironment
		if !diagnosticMode {
			env.SetNeedsRefresh()
			env = provider.GetTestEnvironment()
			ginkgo.RunSpecs(t, CnfCertificationTestSuiteName)
		}
		ContinueRun(diagnosticMode, &env, claimData, claimRoot)
	} else {
		go StartServer()
		select {}
	}
}

func ContinueRun(diagnosticMode bool, env *provider.TestEnvironment, claimData *claim.Claim, claimRoot *claim.Root) {
	endTime := time.Now()
	claimData.Metadata.EndTime = endTime.UTC().Format(claimhelper.DateTimeFormatDirective)

	// Process the test results from the suites, the cnf-features-deploy test suite,
	// and any extra informational messages.
	junitMap := make(map[string]interface{})
	cnfCertificationJUnitFilename := filepath.Join(*junitPath, results.JunitXMLFileName)

	if !diagnosticMode {
		claimhelper.LoadJUnitXMLIntoMap(junitMap, cnfCertificationJUnitFilename, TNFReportKey)
		claimhelper.AppendCNFFeatureValidationReportResults(junitPath, junitMap)
	}

	junitMap[extraInfoKey] = "" // tnf.TestsExtraInfo

	// Append results to claim file data.
	claimData.RawResults = junitMap
	claimData.Results = results.GetReconciledResults()

	// Marshal the claim and output to file
	payload := claimhelper.MarshalClaimOutput(claimRoot)
	claimOutputFile := filepath.Join(*claimPath, results.ClaimFileName)
	claimhelper.WriteClaimOutput(claimOutputFile, payload)

	// Send claim file to the collector if specified by env var
	if configuration.GetTestParameters().EnableDataCollection {
		err := collector.SendClaimFileToCollector(env.CollectorAppEndPoint, claimOutputFile, env.ExecutedBy, env.PartnerName, env.CollectorAppPassword)
		if err != nil {
			log.Errorf("Failed to send post request to the collector: %v", err)
		}
	}

	// Create HTML artifacts for the web results viewer/parser.
	resultsOutputDir := *claimPath
	webFilePaths, err := results.CreateResultsWebFiles(resultsOutputDir)
	if err != nil {
		log.Errorf("Failed to create results web files: %v", err)
	}

	allArtifactsFilePaths := []string{filepath.Join(*claimPath, results.ClaimFileName)}

	// Add the junit xml file only if we're not in diagnostic mode.
	if !diagnosticMode {
		allArtifactsFilePaths = append(allArtifactsFilePaths, filepath.Join(*junitPath, results.JunitXMLFileName))
	}

	// Add all the web artifacts file paths.
	allArtifactsFilePaths = append(allArtifactsFilePaths, webFilePaths...)

	// tar.gz file creation with results and html artifacts, unless omitted by env var.
	if !configuration.GetTestParameters().OmitArtifactsZipFile {
		err = results.CompressResultsArtifacts(resultsOutputDir, allArtifactsFilePaths)
		if err != nil {
			log.Fatalf("Failed to compress results artifacts: %v", err)
		}
	}

	// Remove web artifacts if user does not want them.
	if !configuration.GetTestParameters().IncludeWebFilesInOutputFolder {
		for _, file := range webFilePaths {
			err := os.Remove(file)
			if err != nil {
				log.Fatalf("failed to remove web file %s: %v", file, err)
			}
		}
	}
}

// incorporateTNFVersion adds the TNF version to the claim.
func incorporateVersions(claimData *claim.Claim) {
	claimData.Versions = &claim.Versions{
		Tnf:          gitDisplayRelease,
		TnfGitCommit: GitCommit,
		OcClient:     diagnostics.GetVersionOcClient(),
		Ocp:          diagnostics.GetVersionOcp(),
		K8s:          diagnostics.GetVersionK8s(),
		ClaimFormat:  ClaimFormatVersion,
	}
}
func StartServer() {
	server := &http.Server{
		Addr:         ":8084",           // Server address
		ReadTimeout:  10 * time.Second,  // Maximum duration for reading the entire request
		WriteTimeout: 10 * time.Second,  // Maximum duration for writing the entire response
		IdleTimeout:  120 * time.Second, // Maximum idle duration before closing the connection
	}
	webserver.HandlereqFunc()

	http.HandleFunc("/runFunction", RunHandler)

	fmt.Println("Server is running on :8084...")
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

// Define an HTTP handler that triggers Ginkgo tests
func RunHandler(w http.ResponseWriter, r *http.Request) {
	webserver.Buf = bytes.NewBufferString("")
	log.SetOutput(webserver.Buf)
	rlog.SetOutput(webserver.Buf)

	jsonData := r.FormValue("jsonData") // "jsonData" is the name of the JSON input field
	log.Info(jsonData)
	var data webserver.RequestedData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		fmt.Println("Error:", err)
	}
	var flattenedOptions []string
	flattenedOptions = webserver.FlattenData(data.SelectedOptions, flattenedOptions)

	// Get the file from the request
	file, handler, err := r.FormFile("kubeConfigPath") // "fileInput" is the name of the file input field
	if err != nil {
		http.Error(w, "Unable to retrieve file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create a new file on the server to store the uploaded content
	uploadedFile, err := os.Create(handler.Filename)
	if err != nil {
		http.Error(w, "Unable to create file for writing", http.StatusInternalServerError)
		return
	}
	defer uploadedFile.Close()

	// Copy the uploaded file's content to the new file
	_, err = io.Copy(uploadedFile, file)
	if err != nil {
		http.Error(w, "Unable to copy file", http.StatusInternalServerError)
		return
	}

	// Copy the uploaded file to the server file

	os.Setenv("KUBECONFIG", handler.Filename)
	log.Infof("KUBECONFIG      : %v", handler.Filename)

	log.Infof("Labels filter       : %v", flattenedOptions)
	t := testing.T{}
	claimData, claimRoot := PreRun(&t)
	var env provider.TestEnvironment
	env.SetNeedsRefresh()
	env = provider.GetTestEnvironment()

	// fetch the current config
	suiteConfig, reporterConfig := ginkgo.GinkgoConfiguration()
	// adjust it
	suiteConfig.SkipStrings = []string{"NEVER-RUN"}
	reporterConfig.FullTrace = true
	reporterConfig.JUnitReport = "cnf-certification-tests_junit.xml"
	// pass it in to RunSpecs
	suiteConfig.LabelFilter = strings.Join(flattenedOptions, "")
	ginkgo.RunSpecs(&t, CnfCertificationTestSuiteName, suiteConfig, reporterConfig)

	ContinueRun(false, &env, claimData, claimRoot)
	// Return the result as JSON
	response := struct {
		Message string `json:"Message"`
	}{
		Message: fmt.Sprintf("Sucsses to run %s", strings.Join(flattenedOptions, "")),
	}
	// Serialize the response data to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Set the Content-Type header to specify that the response is JSON
	w.Header().Set("Content-Type", "application/json")
	// Write the JSON response to the client
	_, err = w.Write(jsonResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
