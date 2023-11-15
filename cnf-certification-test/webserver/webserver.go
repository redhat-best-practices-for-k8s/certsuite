package webserver

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	rlog "log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/robert-nix/ansihtml"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/pkg/certsuite"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
)

type webServerContextKey string

const (
	defaultTimeout = 24 * time.Hour
)

var (
	outputFolderCtxKey webServerContextKey = "output-folder"
)

//go:embed index.html
var indexHTML []byte

//go:embed submit.js
var submit []byte

//go:embed logs.js
var logs []byte

//go:embed toast.js
var toast []byte
var Buf *bytes.Buffer

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func logStreamHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()
	// Create a scanner to read the log file line by line
	for {
		scanner := bufio.NewScanner(Buf)
		for scanner.Scan() {
			line := scanner.Bytes()
			fmt.Println(string(line))
			line = append(ansihtml.ConvertToHTML(line), []byte("<br>")...)

			// Send each log line to the client
			if err := conn.WriteMessage(websocket.TextMessage, line); err != nil {
				fmt.Println(err)
				return
			}
		}
		if err := scanner.Err(); err != nil {
			logrus.Printf("Error reading log file: %v", err)
			return
		}
	}
}

type RequestedData struct {
	SelectedOptions interface{} `json:"selectedOptions"`
}
type ResponseData struct {
	Message string `json:"message"`
}

func FlattenData(data interface{}, result []string) []string {
	switch v := data.(type) {
	case string:
		result = append(result, v)
	case []interface{}:
		for _, item := range v {
			result = FlattenData(item, result)
		}
	case map[string]interface{}:
		for key, item := range v {
			if key == "selectedOptions" {
				result = FlattenData(item, result)
			}
			result = FlattenData(item, result)
		}
	}
	return result
}

func installReqHandlers() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to "text/html".
		w.Header().Set("Content-Type", "text/html")
		// Write the embedded HTML content to the response.
		_, err := w.Write(indexHTML)
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/submit.js", func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to "application/javascript".
		w.Header().Set("Content-Type", "application/javascript")
		// Write the embedded JavaScript content to the response.
		_, err := w.Write(submit)
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/logs.js", func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to "application/javascript".
		w.Header().Set("Content-Type", "application/javascript")
		// Write the embedded JavaScript content to the response.
		_, err := w.Write(logs)
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/toast.js", func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to "application/javascript".
		w.Header().Set("Content-Type", "application/javascript")
		// Write the embedded JavaScript content to the response.
		_, err := w.Write(toast)
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})
	// Serve the static HTML file
	http.HandleFunc("/logstream", logStreamHandler)
}

func StartServer(outputFolder string) {
	ctx := context.Background()
	server := &http.Server{
		Addr:         ":8084",           // Server address
		ReadTimeout:  10 * time.Second,  // Maximum duration for reading the entire request
		WriteTimeout: 10 * time.Second,  // Maximum duration for writing the entire response
		IdleTimeout:  120 * time.Second, // Maximum idle duration before closing the connection
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, outputFolderCtxKey, outputFolder)
			return ctx
		},
	}

	installReqHandlers()

	http.HandleFunc("/runFunction", runHandler)

	logrus.Infof("Server is running on :8084...")
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

// Define an HTTP handler that triggers Ginkgo tests
//
//nolint:funlen
func runHandler(w http.ResponseWriter, r *http.Request) {
	Buf = bytes.NewBufferString("")
	logrus.SetOutput(Buf)
	rlog.SetOutput(Buf)

	jsonData := r.FormValue("jsonData") // "jsonData" is the name of the JSON input field
	logrus.Info(jsonData)
	var data RequestedData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		fmt.Println("Error:", err)
	}
	var flattenedOptions []string
	flattenedOptions = FlattenData(data.SelectedOptions, flattenedOptions)

	// Get the file from the request
	file, fileHeader, err := r.FormFile("kubeConfigPath") // "fileInput" is the name of the file input field
	if err != nil {
		http.Error(w, "Unable to retrieve file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create a new file on the server to store the uploaded content
	uploadedFile, err := os.Create(fileHeader.Filename)
	if err != nil {
		http.Error(w, "Unable to create file for writing", http.StatusInternalServerError)
		return
	}
	defer uploadedFile.Close()

	// Copy the uploaded file's content to the new file
	_, err = io.Copy(uploadedFile, file)
	if err != nil {
		http.Error(w, "Unable to copy file", http.StatusInternalServerError)
		return
	}

	// Copy the uploaded file to the server file
	os.Setenv("KUBECONFIG", fileHeader.Filename)
	logrus.Infof("Web Server KUBECONFIG    : %v", fileHeader.Filename)
	logrus.Infof("Web Server Labels filter : %v", flattenedOptions)

	var env provider.TestEnvironment
	env.SetNeedsRefresh()
	env = provider.GetTestEnvironment()

	labelsFilter := strings.Join(flattenedOptions, "")
	outputFolder := r.Context().Value(outputFolderCtxKey).(string)

	logrus.Infof("Running CNF Cert Suite (web-mode). Labels filter: %s, outputFolder: %s", labelsFilter, outputFolder)
	certsuite.Run(labelsFilter, outputFolder, defaultTimeout)

	// Return the result as JSON
	response := struct {
		Message string `json:"Message"`
	}{
		Message: fmt.Sprintf("Succeeded to run %s", strings.Join(flattenedOptions, "")),
	}
	// Serialize the response data to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Set the Content-Type header to specify that the response is JSON
	w.Header().Set("Content-Type", "application/json")
	// Write the JSON response to the client
	_, err = w.Write(jsonResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
