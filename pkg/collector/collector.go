// Copyright (C) 2023-2026 Red Hat, Inc.
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
