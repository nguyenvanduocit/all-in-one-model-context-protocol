package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/nguyenvanduocit/all-in-one-model-context-protocol/services"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func main() {
	// Define command line flags
	credentialsPath := flag.String("credentials", "", "Path to Google credentials JSON file")
	tokenPath := flag.String("token", "", "Path to save/load Google token JSON file")
	flag.Parse()

	// Validate required flags
	if *credentialsPath == "" || *tokenPath == "" {
		flag.PrintDefaults()
		log.Fatal("Both -credentials and -token flags are required")
	}

	// try to delete the token file if it exists
	if _, err := os.Stat(*tokenPath); err == nil {
		os.Remove(*tokenPath)
	}

	ctx := context.Background()
	b, err := os.ReadFile(*credentialsPath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, services.ListGoogleScopes()...)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(config, *tokenPath)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	// Test the connection
	user := "me"
	_, err = srv.Users.Labels.List(user).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve labels: %v", err)
	}

	tokenFileAbsPath, err := filepath.Abs(*tokenPath)
	if err != nil {
		log.Fatalf("Unable to get absolute path of token.json: %v", err)
	}

	fmt.Println("It works! Token file is saved at:")
	fmt.Println(tokenFileAbsPath)
}

// Update getClient to accept tokenPath parameter
func getClient(config *oauth2.Config, tokenPath string) *http.Client {
	tok, err := tokenFromFile(tokenPath)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenPath, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	// Create a channel to receive the authorization code
	codeChan := make(chan string)

	// Start a local HTTP server to handle the redirect
	http.HandleFunc("/oauth2/callback", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		code := r.Form.Get("code")
		if code == "" {
			http.Error(w, "Authorization code not found", http.StatusBadRequest)
			return
		}

		// Send the code to the channel
		codeChan <- code

		// Inform the user that the process is complete
		fmt.Fprintln(w, "<h1>Authentication successful!</h1><p>You can close this window.</p>")
	})

	// Determine the port for the local server
	port := "8080"
	redirectURL := fmt.Sprintf("http://localhost:%s/oauth2/callback", port)

	// Update the configuration with the redirect URL
	config.RedirectURL = redirectURL

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	
	// Open the URL in the default browser
	err := openBrowser(authURL)
	if err != nil {
		log.Printf("Could not open browser automatically: %v", err)
		fmt.Printf("Please open the following URL in your browser:\n\n%v\n\n", authURL)
	} else {
		fmt.Println("Opening your browser to authenticate...")
	}

	// Start the HTTP server in a goroutine
	server := &http.Server{Addr: ":" + port}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for the authorization code
	code := <-codeChan

	// Shut down the server
	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("Failed to shut down server: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
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
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func openBrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("cmd", "/c", "start", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	return err
}
