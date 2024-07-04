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

package feedback

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	feedbackJSONFilePath string
	feedbackOutputPath   string

	// generateCmd is the root of the "catalog generate" CLI program.
	generateFeedbackJsFile = &cobra.Command{
		Use:   "feedbackjs",
		Short: "Generates a javascript file called feedback.js from a feedback.json that was downloaded from the results html viewer.",
		RunE:  runGenerateFeedbackJsFile,
	}
)

func runGenerateFeedbackJsFile(_ *cobra.Command, _ []string) error {
	dat, err := os.ReadFile(feedbackJSONFilePath)
	if err != nil {
		return fmt.Errorf("failed to read json feedback file: %v", err)
	}
	var obj map[string]interface{}
	err = json.Unmarshal(dat, &obj)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json feedback file %s: %v", feedbackJSONFilePath, err)
	}

	// Print the JSON content
	jsonBytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal feedback js content: %v", err)
	}
	feedbackJsFilePath := filepath.Join(feedbackOutputPath, "feedback.js")
	file, err := os.Create(feedbackJsFilePath)
	if err != nil {
		return fmt.Errorf("failed to create javascript feedback file: %v", err)
	}
	feedbackjs := "feedback="
	_, err = file.WriteString(feedbackjs + string(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to write javascript feedback file: %v", err)
	}

	fmt.Println(feedbackjs + string(jsonBytes))
	return nil
}

// Execute executes the "catalog" CLI.
func NewCommand() *cobra.Command {
	generateFeedbackJsFile.Flags().StringVarP(
		&feedbackJSONFilePath, "feedback", "f", "",
		"path to the feedback.json file")

	err := generateFeedbackJsFile.MarkFlagRequired("feedback")
	if err != nil {
		log.Fatalf("failed to mark feedback flag as required:  :%v", err)
		return nil
	}
	generateFeedbackJsFile.Flags().StringVarP(
		&feedbackOutputPath, "outputPath", "o", "",
		"path to create on it the feedback.js file")

	err = generateFeedbackJsFile.MarkFlagRequired("outputPath")
	if err != nil {
		log.Fatalf("failed to mark outputPath flag as required:  :%v", err)
		return nil
	}
	return generateFeedbackJsFile
}
