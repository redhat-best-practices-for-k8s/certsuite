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

package main

import (
	_ "embed"
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/pkg/checksdb"
	"github.com/test-network-function/cnf-certification-test/pkg/claimhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/collector"
	"github.com/test-network-function/cnf-certification-test/pkg/loghelper"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"

	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/observability"

	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/diagnostics"
)

const (
	claimPathFlagKey              = "claimloc"
	CnfCertificationTestSuiteName = "CNF Certification Test Suite"
	defaultClaimPath              = "."
	defaultCliArgValue            = ""
	junitFlagKey                  = "junit"
	TNFReportKey                  = "cnf-certification-test"
	extraInfoKey                  = "testsExtraInfo"

	defaultGlobalTimeout = 2 * time.Hour
)

const (
	labelsFlagName         = "labels"
	labelsFlagDefaultValue = "common"

	labelsFlagUsage = "--labels <expression>     e.g. --labels 'access-control && !access-control-sys-admin-capability'"
)

var (
	claimPath *string
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
	// labelsFlag holds the labels expression to filter the checks to run.
	labelsFlag *string
)

func init() {
	claimPath = flag.String(claimPathFlagKey, defaultClaimPath,
		"the path where the claimfile will be output")
	labelsFlag = flag.String(labelsFlagName, labelsFlagDefaultValue, labelsFlagUsage)

	flag.Parse()
	if *labelsFlag == "" {
		*labelsFlag = "none"
	}
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

func main() {
	err := configuration.LoadEnvironmentVariables()
	if err != nil {
		log.Fatalf("could not load the environment variables, error: %v", err)
	}

	// Set up logging params for logrus
	loghelper.SetLogFormat()
	setLogLevel()

	log.Infof("TNF Version         : %v", getGitVersion())
	log.Infof("Claim Format Version: %s", ClaimFormatVersion)
	log.Infof("Labels filter       : %v", *labelsFlag)

	// Diagnostic functions will run when no labels are provided.
	var diagnosticMode bool

	// Set clientsholder singleton with the filenames from the env vars.
	_ = clientsholder.GetClientsHolder(getK8sClientsConfigFileNames()...)

	// Initialize the claim with the start time, tnf version, etc.
	claimRoot := claimhelper.CreateClaimRoot()
	claimData := claimRoot.Claim
	claimData.Configurations = make(map[string]interface{})
	claimData.Nodes = make(map[string]interface{})
	incorporateVersions(claimData)

	configurations, err := claimhelper.MarshalConfigurations()
	if err != nil {
		log.Fatalf("Configuration node missing because of: %s", err)
	}

	claimData.Nodes = claimhelper.GenerateNodes()
	claimhelper.UnmarshalConfigurations(configurations, claimData.Configurations)

	// initialize abort flag
	testhelper.AbortTrigger = ""

	// Run tests specs only if not in diagnostic mode, otherwise all TSs would run.
	var env provider.TestEnvironment
	if !diagnosticMode {
		env.SetNeedsRefresh()
		env = provider.GetTestEnvironment()
		checksdb.RunChecks(*labelsFlag, defaultGlobalTimeout)
	}

	endTime := time.Now()
	claimData.Metadata.EndTime = endTime.UTC().Format(claimhelper.DateTimeFormatDirective)

	claimData.Results = checksdb.GetReconciledResults()

	// Marshal the claim and output to file
	payload := claimhelper.MarshalClaimOutput(claimRoot)
	claimOutputFile := filepath.Join(*claimPath, results.ClaimFileName)
	claimhelper.WriteClaimOutput(claimOutputFile, payload)

	logrus.Infof("Claim file created at %s", claimOutputFile)

	// Send claim file to the collector if specified by env var
	if configuration.GetTestParameters().EnableDataCollection {
		err = collector.SendClaimFileToCollector(env.CollectorAppEndPoint, claimOutputFile, env.ExecutedBy, env.PartnerName, env.CollectorAppPassword)
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
