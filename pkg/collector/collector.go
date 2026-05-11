// Copyright (C) 2023-2026 Red Hat, Inc.
package collector

import (
	"bytes"
	"fmt"
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
		return fmt.Errorf("failed to open claim file %s: %w", claimFilePath, err)
	}
	defer claimFile.Close()
	fw, err := w.CreateFormFile("claimFile", claimFilePath)
	if err != nil {
		return fmt.Errorf("failed to create form file field for claim: %w", err)
	}
	if _, err = io.Copy(fw, claimFile); err != nil {
		return fmt.Errorf("failed to copy claim file to multipart writer: %w", err)
	}
	return nil
}

func addVarFieldsToPostRequest(w *multipart.Writer, executedBy, partnerName, password string) error {
	fw, err := w.CreateFormField("executed_by")
	if err != nil {
		return fmt.Errorf("failed to create form field 'executed_by': %w", err)
	}
	if _, err = fw.Write([]byte(executedBy)); err != nil {
		return fmt.Errorf("failed to write 'executed_by' field: %w", err)
	}

	fw, err = w.CreateFormField("partner_name")
	if err != nil {
		return fmt.Errorf("failed to create form field 'partner_name': %w", err)
	}
	if _, err = fw.Write([]byte(partnerName)); err != nil {
		return fmt.Errorf("failed to write 'partner_name' field: %w", err)
	}

	fw, err = w.CreateFormField("decoded_password")
	if err != nil {
		return fmt.Errorf("failed to create form field 'decoded_password': %w", err)
	}
	if _, err = fw.Write([]byte(password)); err != nil {
		return fmt.Errorf("failed to write 'decoded_password' field: %w", err)
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
		return nil, fmt.Errorf("failed to add claim file to post request: %w", err)
	}

	// Add the executed by, partner name and password fields to the request
	err = addVarFieldsToPostRequest(w, executedBy, partnerName, password)
	if err != nil {
		return nil, fmt.Errorf("failed to add variable fields to post request: %w", err)
	}

	w.Close()

	// Create POST request with the form-data as body
	req, err := http.NewRequest("POST", endPoint, &buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request to %s: %w", endPoint, err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	return req, nil
}

func SendClaimFileToCollector(endPoint, claimFilePath, executedBy, partnerName, password string) error {
	// Temporary end point
	postReq, err := createSendToCollectorPostRequest(endPoint, claimFilePath, executedBy, partnerName, password)
	if err != nil {
		return fmt.Errorf("failed to create collector POST request: %w", err)
	}

	client := &http.Client{
		Timeout: collectorUploadTimeout, // 30 second timeout for collector uploads
	}
	resp, err := client.Do(postReq)
	if err != nil {
		return fmt.Errorf("failed to send claim file to collector at %s: %w", endPoint, err)
	}
	defer resp.Body.Close()
	return nil
}
