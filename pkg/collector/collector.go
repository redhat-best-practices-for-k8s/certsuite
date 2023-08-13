package collector

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func createSendToCollectorPostRequest(endPoint, claimFilePath, executedBy, partnerName string) (*http.Request, error) {
	// Create a new buffer to hold the form-data
	var buffer bytes.Buffer
	w := multipart.NewWriter(&buffer)

	// Add the claim file to the request
	claimFile, err := os.Open(claimFilePath)
	if err != nil {
		return nil, err
	}
	defer claimFile.Close()
	fw, err := w.CreateFormFile("claimFile", claimFilePath)
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(fw, claimFile); err != nil {
		return nil, err
	}

	// Add the executed by and partner name to the request
	if fw, err = w.CreateFormField("executed_by"); err != nil {
		return nil, err
	}
	if _, err = fw.Write([]byte(executedBy)); err != nil {
		return nil, err
	}

	if fw, err = w.CreateFormField("partner_name"); err != nil {
		return nil, err
	}
	if _, err = fw.Write([]byte(partnerName)); err != nil {
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

func SendClaimFileToCollector(endPoint, claimFilePath, executedBy, partnerName string) error {
	// Temporary end point
	postReq, err := createSendToCollectorPostRequest(endPoint, claimFilePath, executedBy, partnerName)
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
