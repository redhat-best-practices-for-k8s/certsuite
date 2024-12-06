package results

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
)

const (
	// tarGz file prefix layout format: YearMonthDay-HourMinSec
	tarGzFileNamePrefixLayout = "20060102-150405"
	tarGzFileNameSuffix       = "cnf-test-results.tar.gz"

	// Connect API Information
	// certIDURL     = "https://access.qa.redhat.com/hydra/cwe/rest/v1.0/projects/certifications"
	// connectAPIURL = "https://access.qa.redhat.com/hydra/cwe/rest/v1.0/attachments/upload"

	certIDURL     = "https://access.redhat.com/hydra/cwe/rest/v1.0/projects/certifications"
	connectAPIURL = "https://access.redhat.com/hydra/cwe/rest/v1.0/attachments/upload"
)

func generateZipFileName() string {
	return fmt.Sprintf(time.Now().Format(tarGzFileNamePrefixLayout) + "-" + tarGzFileNameSuffix)
}

// Helper function to get the tar file header from a file.
func getFileTarHeader(file string) (*tar.Header, error) {
	info, err := os.Stat(file)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info from %s: %v", file, err)
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to get file info header for %s: %v", file, err)
	}

	return header, nil
}

// Creates a zip file in the outputDir containing each file in the filePaths slice.
func CompressResultsArtifacts(outputDir string, filePaths []string) (string, error) {
	zipFileName := generateZipFileName()
	zipFilePath := filepath.Join(outputDir, zipFileName)

	log.Info("Compressing results artifacts into %s", zipFilePath)
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return "", fmt.Errorf("failed creating tar.gz file %s in dir %s (filepath=%s): %v",
			zipFileName, outputDir, zipFilePath, err)
	}

	zipWriter := gzip.NewWriter(zipFile)
	defer zipWriter.Close()

	tarWriter := tar.NewWriter(zipWriter)
	defer tarWriter.Close()

	for _, file := range filePaths {
		log.Debug("Zipping file %s", file)

		tarHeader, err := getFileTarHeader(file)
		if err != nil {
			return "", err
		}

		err = tarWriter.WriteHeader(tarHeader)
		if err != nil {
			return "", fmt.Errorf("failed to write tar header for %s: %v", file, err)
		}

		f, err := os.Open(file)
		if err != nil {
			return "", fmt.Errorf("failed to open file %s: %v", file, err)
		}

		if _, err = io.Copy(tarWriter, f); err != nil {
			return "", fmt.Errorf("failed to tar file %s: %v", file, err)
		}

		f.Close()
	}

	// Create fully qualified path to the zip file
	zipFilePath, err = filepath.Abs(zipFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %v", zipFilePath, err)
	}

	// Return the entire path to the zip file
	return zipFilePath, nil
}

func createFormField(w *multipart.Writer, field, value string) error {
	fw, err := w.CreateFormField(field)
	if err != nil {
		return fmt.Errorf("failed to create form field: %v", err)
	}

	_, err = fw.Write([]byte(value))
	if err != nil {
		return fmt.Errorf("failed to write field %s: %v", field, err)
	}

	return nil
}

type CertIDResponse struct {
	ID                  int    `json:"id"`
	CaseNumber          string `json:"caseNumber"`
	Status              string `json:"status"`
	CertificationLevel  string `json:"certificationLevel"`
	RhcertURL           string `json:"rhcertUrl"`
	HasStartedByPartner bool   `json:"hasStartedByPartner"`
	CertificationType   struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"certificationType"`
}

// GetCertIDFromConnectAPI gets the certification ID from the Red Hat Connect API
func GetCertIDFromConnectAPI(apiKey, projectID, proxyURL, proxyPort string) (string, error) {
	log.Info("Getting certification ID from Red Hat Connect API")

	// sanitize the incoming variables, remove the double quotes if any
	apiKey = strings.ReplaceAll(apiKey, "\"", "")
	projectID = strings.ReplaceAll(projectID, "\"", "")
	proxyURL = strings.ReplaceAll(proxyURL, "\"", "")
	proxyPort = strings.ReplaceAll(proxyPort, "\"", "")

	// remove quotes from projectID
	projectIDJSON := fmt.Sprintf(`{ "projectId": %q }`, projectID)

	// Convert JSON to bytes
	projectIDJSONBytes := []byte(projectIDJSON)

	// Create a new request
	req, err := http.NewRequest("POST", certIDURL, bytes.NewBuffer(projectIDJSONBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create new request: %v", err)
	}

	log.Debug("Request Body: %s", req.Body)

	// Set the content type
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", apiKey)

	// print the request
	log.Debug("Sending request to %s", certIDURL)

	client := &http.Client{}
	setProxy(client, proxyURL, proxyPort)
	res, err := sendRequest(req, client)
	if err != nil {
		return "", fmt.Errorf("failed to send post request to the endpoint: %v", err)
	}
	defer res.Body.Close()

	// Parse the response
	var certIDResponse CertIDResponse
	err = json.NewDecoder(res.Body).Decode(&certIDResponse)
	if err != nil {
		return "", fmt.Errorf("failed to decode response body: %v", err)
	}

	log.Info("Certification ID retrieved from the API: %d", certIDResponse.ID)

	// Return the certification ID
	return fmt.Sprintf("%d", certIDResponse.ID), nil
}

type UploadResult struct {
	UUID         string    `json:"uuid"`
	Type         string    `json:"type"`
	Name         string    `json:"name"`
	Size         int       `json:"size"`
	ContentType  string    `json:"contentType"`
	Desc         string    `json:"desc"`
	DownloadURL  string    `json:"downloadUrl"`
	UploadedBy   string    `json:"uploadedBy"`
	UploadedDate time.Time `json:"uploadedDate"`
	CertID       int       `json:"certId"`
}

// SendResultsToConnectAPI sends the results to the Red Hat Connect API
//
//nolint:funlen
func SendResultsToConnectAPI(zipFile, apiKey, certID, proxyURL, proxyPort string) error {
	log.Info("Sending results to Red Hat Connect")

	// sanitize the incoming variables, remove the double quotes if any
	apiKey = strings.ReplaceAll(apiKey, "\"", "")
	certID = strings.ReplaceAll(certID, "\"", "")
	proxyURL = strings.ReplaceAll(proxyURL, "\"", "")
	proxyPort = strings.ReplaceAll(proxyPort, "\"", "")

	var buffer bytes.Buffer

	// Create a new multipart writer
	w := multipart.NewWriter(&buffer)

	log.Debug("Creating form file for %s", zipFile)

	claimFile, err := os.Open(zipFile)
	if err != nil {
		return err
	}
	defer claimFile.Close()

	fw, err := w.CreateFormFile("attachment", zipFile)
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}

	if _, err = io.Copy(fw, claimFile); err != nil {
		return err
	}

	// Create a form field
	err = createFormField(w, "type", "RhocpBestPracticeTestResult")
	if err != nil {
		return err
	}

	// Create a form field
	err = createFormField(w, "certId", certID)
	if err != nil {
		return err
	}

	// Create a form field
	err = createFormField(w, "description", "CNF Test Results")
	if err != nil {
		return err
	}

	w.Close()

	// Create a new request
	req, err := http.NewRequest("POST", connectAPIURL, &buffer)
	if err != nil {
		return fmt.Errorf("failed to create new request: %v", err)
	}

	// Set the content type
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("x-api-key", apiKey)

	// Create a client
	client := &http.Client{}
	setProxy(client, proxyURL, proxyPort)
	response, err := sendRequest(req, client)
	if err != nil {
		return fmt.Errorf("failed to send post request to the endpoint: %v", err)
	}
	defer response.Body.Close()

	// Parse the result of the request
	var uploadResult UploadResult
	err = json.NewDecoder(response.Body).Decode(&uploadResult)
	if err != nil {
		return fmt.Errorf("failed to decode response body: %v", err)
	}

	log.Info("Download URL: %s", uploadResult.DownloadURL)
	log.Info("Upload Date: %s", uploadResult.UploadedDate)
	return nil
}

func sendRequest(req *http.Request, client *http.Client) (*http.Response, error) {
	// print the request
	log.Debug("Sending request to %s", req.URL)

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send post request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		log.Debug("Response: %v", res)
		return nil, fmt.Errorf("failed to send post request to the endpoint: %v", res.Status)
	}

	return res, nil
}

func setProxy(client *http.Client, proxyURL, proxyPort string) {
	if proxyURL != "" && proxyPort != "" {
		log.Debug("Proxy is set. Using proxy %s:%s", proxyURL, proxyPort)
		proxyURL := fmt.Sprintf("%s:%s", proxyURL, proxyPort)
		parsedURL, err := url.Parse(proxyURL)
		if err != nil {
			log.Error("Failed to parse proxy URL: %v", err)
		}
		log.Debug("Proxy URL: %s", parsedURL)
		client.Transport = &http.Transport{Proxy: http.ProxyURL(parsedURL)}
	}
}
