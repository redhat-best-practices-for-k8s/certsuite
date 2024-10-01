package resultsspreadsheet

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

const (
	tokenPermissions   = 0o600
	readHeadersTimeout = 10 * time.Second
)

var authCode string
var ctxShutdown, cancel = context.WithCancel(context.Background())

// OpenBrowser opens up the provided URL in a browser
func OpenBrowser(u string) error {
	if _, err := url.ParseRequestURI(u); err != nil {
		return err
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "openbsd":
		cmd = exec.Command("xdg-open", u)
	case "darwin":
		cmd = exec.Command("open", u)
	case "windows":
		r := strings.NewReplacer("&", "^&")
		cmd = exec.Command("cmd", "/c", "start", r.Replace(u))
	}
	if cmd != nil {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			return err
		}
		err = cmd.Wait()
		if err != nil {
			return err
		}
		return nil
	} else {
		return fmt.Errorf("unsupported platform")
	}
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) (*http.Client, error) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		log.Println("Auth token not found, retrieving token from web.")
		tok, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		if err := saveToken(tokFile, tok); err != nil {
			return nil, err
		}
	}
	return config.Client(context.Background(), tok), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authDone := &sync.WaitGroup{}
	authDone.Add(1)

	startAuthServer(config.RedirectURL, authDone)
	if err := OpenBrowser(config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)); err != nil {
		return nil, fmt.Errorf("failed to open browser for authentication: %v", err)
	}

	authDone.Wait()

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %v", err)
	}

	return tok, nil
}

// startAuthServer starts a local service waiting for an authcode needed to continue authentication.
func startAuthServer(serverURL string, wg *sync.WaitGroup) {
	server := &http.Server{
		Addr:              strings.TrimPrefix(serverURL, "http://"),
		ReadHeaderTimeout: readHeadersTimeout,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-ctxShutdown.Done():
			return
		default:
		}

		authCode = r.URL.Query().Get("code")
		if authCode == "" {
			log.Fatalf("Auth code has not been provided")
			return
		}

		cancel()
		err := server.Shutdown(context.Background())
		if err != nil {
			log.Fatalf("Error has accured while shuting down the auth server: %v", err)
		}
	})

	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	log.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, tokenPermissions)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return fmt.Errorf("unable to encode token: %v", err)
	}
	return nil
}
