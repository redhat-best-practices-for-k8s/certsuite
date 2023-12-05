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
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/pkg/certsuite"
	"github.com/test-network-function/cnf-certification-test/pkg/flags"
	"github.com/test-network-function/cnf-certification-test/pkg/loghelper"
	"github.com/test-network-function/cnf-certification-test/pkg/versions"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/webserver"

	"github.com/test-network-function/cnf-certification-test/internal/cli"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
)

const (
	claimPathFlagKey              = "claimloc"
	CnfCertificationTestSuiteName = "CNF Certification Test Suite"
	defaultClaimPath              = "."
	defaultCliArgValue            = ""
	junitFlagKey                  = "junit"
	TNFReportKey                  = "cnf-certification-test"
	extraInfoKey                  = "testsExtraInfo"
	logFileName                   = "cnf-certsuite.log"
	logFilePermissions            = 0o644
)

var (
	claimPath *string
)

func init() {
	claimPath = flag.String(claimPathFlagKey, defaultClaimPath,
		"the path where the claimfile will be output")

	flags.InitFlags()
}

// setLogLevel sets the log level for logrus based on the "TNF_LOG_LEVEL" environment variable
func setLogLevel() {
	params := configuration.GetTestParameters()

	var logLevel, err = logrus.ParseLevel(params.LogLevel)
	if err != nil {
		logrus.Error("TNF_LOG_LEVEL environment set with an invalid value, defaulting to DEBUG \n Valid values are:  trace, debug, info, warn, error, fatal, panic")
		logLevel = logrus.DebugLevel
	}

	logrus.SetLevel(logLevel)
}

//nolint:funlen
func main() {
	err := configuration.LoadEnvironmentVariables()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not load the environment variables, err: %v", err)
		os.Exit(1)
	}

	// Set up logging params for logrus
	loghelper.SetLogFormat()
	setLogLevel()

	logrusLogFile, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE, logFilePermissions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create log file, err: %v", err)
		os.Exit(1)
	}
	defer logrusLogFile.Close()

	logrus.SetOutput(logrusLogFile)

	// Set up logger
	err = os.Remove("test_log") // TODO: use proper file when logrus is removed
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "could not delete old log file, err: %v", err)
		os.Exit(1) //nolint:gocritic // the error will not happen after logrus is removed
	}

	logFile, err := os.OpenFile("test_log", os.O_RDWR|os.O_CREATE, logFilePermissions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create log file, err: %v", err)
		os.Exit(1)
	}
	defer logFile.Close()

	log.SetupLogger(logFile)
	log.Info("Log file: %s", logFileName)

	logrus.Infof("TNF Version         : %v", versions.GitVersion())
	logrus.Infof("Claim Format Version: %s", versions.ClaimFormatVersion)
	logrus.Infof("Labels filter       : %v", *flags.LabelsFlag)

	cli.PrintBanner()

	fmt.Printf("CNFCERT version: %s\n", versions.GitVersion())
	fmt.Printf("Claim file version: %s\n", versions.ClaimFormatVersion)
	fmt.Printf("Checks filter: %s\n", *flags.LabelsFlag)
	fmt.Printf("Output folder: %s\n", *claimPath)
	fmt.Printf("Log file: %s\n", logFileName)
	fmt.Printf("\n")

	if *flags.ListFlag {
		// ToDo: List all the available checks, filtered with --labels.
		logrus.Errorf("Not implemented yet.")
		os.Exit(1)
	}

	// Set clientsholder singleton with the filenames from the env vars.
	logrus.Infof("Output folder for the claim file: %s", *claimPath)
	if *flags.ServerModeFlag {
		logrus.Info("Running CNF Certification Suite in web server mode.")
		webserver.StartServer(*claimPath)
	} else {
		log.Info("Running CNF Certification Suite in stand-alone mode.")
		certsuite.Run(*flags.LabelsFlag, *claimPath)
	}
}
