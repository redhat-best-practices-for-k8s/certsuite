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

// createClaimJSFile generates a claimjson.js file from a given claim.json.
//
// It reads the specified claim.json file, transforms its contents into
// JavaScript syntax, writes the result to a new file with the same base
// name but a .js extension, and returns the path of the created file.
// On failure it returns an empty string and an error describing the issue.
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

// CreateResultsWebFiles generates the web assets required to display test results.
//
// It creates a JavaScript file containing the claim JSON, an HTML page that
// renders the results, and a classification script. The function writes these
// files into outputDir using default permissions and returns a slice of the
// paths to each created file. If any write fails, it returns an error.
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
