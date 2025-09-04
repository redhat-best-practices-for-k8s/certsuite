package results

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
)

const (
	// redHatConnectAPITimeout is the timeout for Red Hat Connect API calls
	redHatConnectAPITimeout = 60 * time.Second
)

// createFormField Creates a single form field in a multipart payload
//
// The function accepts a multipart writer, a field name, and its value. It uses
// the writer to create the field and writes the provided string into it. Errors
// during creation or writing are wrapped with context and returned.
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

// CertIDResponse Represents a certification case response from RHConnect
//
// This struct holds information returned by the RHConnect API for a
// certification request, including the internal ID, external case number,
// status, level, URL, and whether the partner initiated it. It also embeds a
// nested type providing the certification category's identifier and name.
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

// GetCertIDFromConnectAPI Retrieves a certification identifier from the Red Hat Connect service
//
// The function builds a JSON payload containing a project ID, sends it as a
// POST request to the Connect API endpoint for certifications, and decodes the
// returned JSON to extract the numeric certification ID. It supports optional
// proxy configuration, sanitizes input strings, and applies a timeout to the
// HTTP client. The resulting ID is returned as a string; errors are reported if
// any step fails.
func GetCertIDFromConnectAPI(apiKey, projectID, connectAPIBaseURL, proxyURL, proxyPort string) (string, error) {
	log.Info("Getting certification ID from Red Hat Connect API")

	// sanitize the incoming variables, remove the double quotes if any
	apiKey = strings.ReplaceAll(apiKey, "\"", "")
	projectID = strings.ReplaceAll(projectID, "\"", "")
	proxyURL = strings.ReplaceAll(proxyURL, "\"", "")
	proxyPort = strings.ReplaceAll(proxyPort, "\"", "")
	connectAPIBaseURL = strings.ReplaceAll(connectAPIBaseURL, "\"", "")

	// remove quotes from projectID
	projectIDJSON := fmt.Sprintf(`{ "projectId": %q }`, projectID)

	// Convert JSON to bytes
	projectIDJSONBytes := []byte(projectIDJSON)

	// Create the URL
	certIDURL := fmt.Sprintf("%s/projects/certifications", connectAPIBaseURL)

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

	client := &http.Client{
		Timeout: redHatConnectAPITimeout, // 60 second timeout for Red Hat Connect API
	}
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

// UploadResult Details the outcome of a file upload operation
//
// This structure holds information about an uploaded artifact, including its
// unique identifier, type, name, size, MIME type, description, download link,
// uploader, upload timestamp, and related certificate ID. It is used to convey
// all relevant metadata back to clients or services that need to reference the
// stored file.
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

// SendResultsToConnectAPI Uploads a ZIP file of test results to the Red Hat Connect API
//
// The function takes a zip file path, an API key, base URL, certification ID,
// and optional proxy settings. It builds a multipart/form‑data POST request
// containing the file and metadata fields, then sends it with a
// timeout‑limited HTTP client that can use a proxy if configured. On success
// it logs the returned download URL and upload date; otherwise it returns an
// error describing what failed.
//
//nolint:funlen
func SendResultsToConnectAPI(zipFile, apiKey, connectBaseURL, certID, proxyURL, proxyPort string) error {
	log.Info("Sending results to Red Hat Connect")

	// sanitize the incoming variables, remove the double quotes if any
	apiKey = strings.ReplaceAll(apiKey, "\"", "")
	certID = strings.ReplaceAll(certID, "\"", "")
	proxyURL = strings.ReplaceAll(proxyURL, "\"", "")
	proxyPort = strings.ReplaceAll(proxyPort, "\"", "")
	connectBaseURL = strings.ReplaceAll(connectBaseURL, "\"", "")

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

	// Create the URL
	connectAPIURL := fmt.Sprintf("%s/attachments/upload", connectBaseURL)

	// Create a new request
	req, err := http.NewRequest("POST", connectAPIURL, &buffer)
	if err != nil {
		return fmt.Errorf("failed to create new request: %v", err)
	}

	// Set the content type
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("x-api-key", apiKey)

	// Create a client
	client := &http.Client{
		Timeout: redHatConnectAPITimeout, // 60 second timeout for Red Hat Connect API upload
	}
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

// sendRequest Sends an HTTP request using a client and checks for success
//
// This function logs the target URL, executes the request with the provided
// http.Client, and returns the response if the status code is 200 OK. If an
// error occurs during execution or the status code differs from OK, it returns
// a formatted error describing the failure.
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

// setProxy configures an HTTP client to use a proxy when provided
//
// When both the proxy address and port are supplied, the function builds a
// proxy URL, parses it, logs the configuration, and assigns a transport with
// that proxy to the client. If parsing fails, an error is logged but no panic
// occurs. The client remains unchanged if either value is empty.
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
