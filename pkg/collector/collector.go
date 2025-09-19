package collector

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

const (
	// collectorUploadTimeout is the timeout for collector uploads
	collectorUploadTimeout = 30 * time.Second
)

// addClaimFileToPostRequest Adds a claim file as multipart form data
//
// The function opens the specified file, creates a new part in the multipart
// writer using that file's name, copies the file contents into the part, and
// then returns any error encountered during these steps. It closes the file
// automatically with defer to avoid resource leaks. The result is ready for
// inclusion in an HTTP POST request.
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

// addVarFieldsToPostRequest Adds form fields for execution details
//
// This function writes three key-value pairs into a multipart request: the user
// who executed the operation, the partner name, and the decoded password. It
// creates each field using the writer's CreateFormField method and then writes
// the corresponding string value. If any step fails it returns an error;
// otherwise it completes silently.
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

// createSendToCollectorPostRequest Creates a multipart POST request to upload a claim file
//
// This function builds an HTTP POST request with form-data that includes the
// specified claim file and several text fields: executed_by, partner_name, and
// decoded_password. It writes these parts into a buffer using a multipart
// writer, sets the appropriate content type header, and returns the constructed
// request or an error if any step fails.
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

// SendClaimFileToCollector Sends a claim file to a collector endpoint
//
// The function builds an HTTP POST request that includes the claim file and
// authentication fields, then executes it with a timeout. It returns any error
// encountered during request creation or execution; successful completion
// results in nil.
func SendClaimFileToCollector(endPoint, claimFilePath, executedBy, partnerName, password string) error {
	// Temporary end point
	postReq, err := createSendToCollectorPostRequest(endPoint, claimFilePath, executedBy, partnerName, password)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: collectorUploadTimeout, // 30 second timeout for collector uploads
	}
	resp, err := client.Do(postReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
