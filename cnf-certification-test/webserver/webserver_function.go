package webserver

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/robert-nix/ansihtml"
	"github.com/sirupsen/logrus"
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
func HandlereqFunc() {
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
