// Copyright (C) 2020-2024 Red Hat, Inc.
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
	"fmt"
	"os"

	"github.com/test-network-function/cnf-certification-test/pkg/certsuite"
	"github.com/test-network-function/cnf-certification-test/pkg/flags"
	"github.com/test-network-function/cnf-certification-test/pkg/versions"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/webserver"

	"github.com/test-network-function/cnf-certification-test/internal/cli"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
)

const (
	CnfCertificationTestSuiteName = "CNF Certification Test Suite"
	defaultCliArgValue            = ""
	junitFlagKey                  = "junit"
	TNFReportKey                  = "cnf-certification-test"
	extraInfoKey                  = "testsExtraInfo"
)

func init() {
	flags.InitFlags()
}

func createLogFile(outputDir string) (*os.File, error) {
	logFilePath := outputDir + "/" + log.LogFileName
	err := os.Remove(logFilePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("could not delete old log file, err: %v", err)
	}

	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE, log.LogFilePermissions)
	if err != nil {
		return nil, fmt.Errorf("could not open a new log file, err: %v", err)
	}

	return logFile, nil
}

func setupLogger(logFile *os.File) {
	logLevel := configuration.GetTestParameters().LogLevel
	log.SetupLogger(logFile, logLevel)
	log.Info("Log file: %s (level=%s)", log.LogFileName, logLevel)
}

func main() {
	err := configuration.LoadEnvironmentVariables()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not load the environment variables, err: %v", err)
		os.Exit(1)
	}

	logFile, err := createLogFile(*flags.OutputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create the log file, err: %v", err)
		os.Exit(1)
	}
	defer logFile.Close()

	setupLogger(logFile)

	log.Info("TNF Version         : %v", versions.GitVersion())
	log.Info("Claim Format Version: %s", versions.ClaimFormatVersion)
	log.Info("Labels filter       : %v", *flags.LabelsFlag)

	log.Debug("Test environment variables: %#v", *configuration.GetTestParameters())

	cli.PrintBanner()

	fmt.Printf("CNFCERT version: %s\n", versions.GitVersion())
	fmt.Printf("Claim file version: %s\n", versions.ClaimFormatVersion)
	fmt.Printf("Checks filter: %s\n", *flags.LabelsFlag)
	fmt.Printf("Output folder: %s\n", *flags.OutputDir)
	fmt.Printf("Log file: %s\n", log.LogFileName)
	fmt.Printf("\n")

	// Set clientsholder singleton with the filenames from the env vars.
	log.Info("Output folder for the claim file: %s", *flags.OutputDir)
	if *flags.ServerModeFlag {
		log.Info("Running CNF Certification Suite in web server mode")
		webserver.StartServer(*flags.OutputDir)
	} else {
		log.Info("Running CNF Certification Suite in stand-alone mode")
		err = certsuite.Run(*flags.LabelsFlag, *flags.OutputDir)
		if err != nil {
			log.Error("Failed to run CNF Certification Suite: %v", err)
			os.Exit(1) //nolint:gocritic // exitAfterDefer
		}
	}
}
