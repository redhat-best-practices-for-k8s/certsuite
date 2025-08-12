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

// createFormField writes a form field into a multipart writer.
//
// It creates a new form field with the given name and value, then writes
// the provided string content to that field. The function returns an error
// if creating the form field or writing its contents fails.
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

// CertIDResponse represents a response containing certification information.
//
// It holds details about a specific certification case, including its number,
// level, type, status, and related URLs. The CertificationType field contains
// an identifier and name of the certification category. The HasStartedByPartner
// flag indicates whether the partner has initiated the process. This struct
// is used to unmarshal JSON responses from the RH Connect API.
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

// GetCertIDFromConnectAPI retrieves the certification ID from Red Hat Connect.
//
// It sends an authenticated request to the specified Red Hat Connect API endpoint
// using the provided client credentials and certificate name. The function
// returns the extracted certification ID as a string, or an error if any step
// of the process fails. Parameters: the base URL, username, password,
// organization ID, and certificate name. Returns: the certification ID and
// an error value.
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

// UploadResult represents the metadata of a file uploaded to the RHConnect service.
//
// It contains information about the certificate or artifact that was successfully stored,
// including its unique identifiers, descriptive fields, and timestamps.
//
// Fields:
// CertID          – numeric identifier assigned by the system.
// ContentType     – MIME type of the uploaded content.
// Desc            – short description provided at upload time.
// DownloadURL     – URL to retrieve the file.
// Name            – original filename.
// Size            – size in bytes.
// Type            – classification or category of the upload.
// UUID            – globally unique identifier for the resource.
// UploadedBy      – user who performed the upload.
// UploadedDate    – timestamp when the upload completed.
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

// SendResultsToConnectAPI uploads test results to the Red Hat Connect API.
//
// It accepts the Connect API URL, authentication token, test suite name,
// organization ID, test run ID, and a file path containing the result data.
// The function constructs a multipart/form-data request with the result
// payload and sends it over HTTPS. Upon success it returns nil; on failure
// it returns an error describing what went wrong during the upload or
// response handling.
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

// sendRequest sends an HTTP request and returns the response or an error.
//
// It logs debug information before and after performing the request using the provided http.Client.
// The function accepts a pointer to http.Request and a pointer to http.Client,
// executes the request via client.Do, and returns the resulting *http.Response along with any error encountered.
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

// setProxy configures the HTTP client to use a specified proxy URL and user‑agent header, returning a cleanup function.
//
// It accepts an http.Client pointer, a proxy URL string, and a user agent string.
// The function parses the proxy URL, assigns it to the client's Transport.Proxy,
// and sets the User-Agent header for subsequent requests.
// A no‑op cleanup function is returned that restores the original Proxy setting
// when called. This allows temporary proxy configuration within tests or
// specific operations without affecting global client state.
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
