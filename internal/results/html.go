package results

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

const (
	htmlResultsFileName = "results.html"
	jsClaimVarFileName  = "claimjson.js"

	writeFilePerms = 0o644
)

//go:embed html/results.html
var htmlResultsFileContent []byte

// createClaimJSFile Creates a JavaScript file containing the claim JSON data
//
// The function reads the contents of a specified claim.json file, prefixes it
// with a JavaScript variable declaration, and writes this combined string to a
// new file in the given output directory. It returns the path to the newly
// created file or an error if reading or writing fails.
func createClaimJSFile(claimFilePath, outputDir string) (filePath string, err error) {
	// Read claim.json content.
	claimContent, err := os.ReadFile(claimFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read claim file %s content in %s: %v", claimFilePath, outputDir, err)
	}

	// Add the content as the value for the js variable.
	jsClaimContent := "var initialjson = " + string(claimContent)

	filePath = filepath.Join(outputDir, jsClaimVarFileName)
	err = os.WriteFile(filePath, []byte(jsClaimContent), writeFilePerms)
	if err != nil {
		return "", fmt.Errorf("failed to write file %s: %v", filePath, err)
	}

	return filePath, nil
}

// CreateResultsWebFiles Creates HTML web assets for claim data
//
// The function generates the JavaScript file that exposes the claim JSON
// content, writes a static results page, and returns their paths. It accepts an
// output directory and a claim file name, constructs the necessary files,
// handles any I/O errors, and collects the created file locations in a slice.
// The returned slice contains the paths to all web artifacts for later use.
func CreateResultsWebFiles(outputDir, claimFileName string) (filePaths []string, err error) {
	type file struct {
		Path    string
		Content []byte
	}

	staticFiles := []file{
		{
			Path:    filepath.Join(outputDir, htmlResultsFileName),
			Content: htmlResultsFileContent,
		},
	}

	claimFilePath := filepath.Join(outputDir, claimFileName)
	claimJSFilePath, err := createClaimJSFile(claimFilePath, outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %v", jsClaimVarFileName, err)
	}

	filePaths = []string{claimJSFilePath}
	for _, f := range staticFiles {
		err := os.WriteFile(f.Path, f.Content, writeFilePerms)
		if err != nil {
			return nil, fmt.Errorf("failed to create file %s: %v", f.Path, err)
		}

		// Add this file path to the slice.
		filePaths = append(filePaths, f.Path)
	}

	return filePaths, nil
}
