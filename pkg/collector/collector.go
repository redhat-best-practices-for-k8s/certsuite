package collector

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// addClaimFileToPostRequest(writer *multipart.Writer, filePath string) error {
// Adds a claim file to the provided multipart writer.
//
// It opens the file located at filePath, creates a form field in the
// multipart writer with the appropriate filename, copies the file's
// contents into that part, and then closes the file. The function
// returns an error if any step fails.}
func addClaimFileToPostRequest(w *multipart.Writer, claimFilePath string) error {
	claimFile, err := os.Open(claimFilePath)
	if err != nil {
		return err
	}
	defer claimFile.Close()
	fw, err := w.CreateFormFile("claimFile", claimFilePath)
	if err != nil {
		return err
	}
	if _, err = io.Copy(fw, claimFile); err != nil {
		return err
	}
	return nil
}

// addVarFieldsToPostRequest writes variable fields to a multipart form request.
//
// It creates three form fields with the given names and writes the provided values
// into each field using the supplied multipart writer.
// The function returns an error if any step in creating or writing the fields fails.
func addVarFieldsToPostRequest(w *multipart.Writer, executedBy, partnerName, password string) error {
	fw, err := w.CreateFormField("executed_by")
	if err != nil {
		return err
	}
	if _, err = fw.Write([]byte(executedBy)); err != nil {
		return err
	}

	fw, err = w.CreateFormField("partner_name")
	if err != nil {
		return err
	}
	if _, err = fw.Write([]byte(partnerName)); err != nil {
		return err
	}

	fw, err = w.CreateFormField("decoded_password")
	if err != nil {
		return err
	}
	if _, err = fw.Write([]byte(password)); err != nil {
		return err
	}
	return nil
}

// createSendToCollectorPostRequest(url string, token string, username string, password string, certPath string) (*http.Request, error)
//
// createSendToCollectorPostRequest constructs an HTTP POST request for sending a certificate to the collector service.
// It creates a multipart/form-data body containing the certificate file and user credentials, then returns the
// prepared *http.Request. The function sets the Authorization header using the provided token and ensures the
// Content-Type is set appropriately. It may return an error if any step in request creation or writing fails.
func createSendToCollectorPostRequest(endPoint, claimFilePath, executedBy, partnerName, password string) (*http.Request, error) {
	// Create a new buffer to hold the form-data
	var buffer bytes.Buffer
	w := multipart.NewWriter(&buffer)

	// Add the claim file to the request
	err := addClaimFileToPostRequest(w, claimFilePath)
	if err != nil {
		return nil, err
	}

	// Add the executed by, partner name and password fields to the request
	err = addVarFieldsToPostRequest(w, executedBy, partnerName, password)
	if err != nil {
		return nil, err
	}

	w.Close()

	// Create POST request with the form-data as body
	req, err := http.NewRequest("POST", endPoint, &buffer)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	return req, nil
}

// SendClaimFileToCollector sends a claim file to the collector service.
//
// It takes five string arguments: the collector URL, the path to the
// claim file, the target directory on the collector, an API key,
// and a request ID. The function creates a POST request with the
// provided data, executes it, closes the response body, and returns
// any error that occurs during these steps.
func SendClaimFileToCollector(endPoint, claimFilePath, executedBy, partnerName, password string) error {
	// Temporary end point
	postReq, err := createSendToCollectorPostRequest(endPoint, claimFilePath, executedBy, partnerName, password)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(postReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
