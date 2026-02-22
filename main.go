package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Handler is the entry point for Vercel.
func Handler(w http.ResponseWriter, r *http.Request) {
	// 1. "Try-Catch" equivalent: Recovery from unexpected panics
	defer func() {
		if r := recover(); r != nil {
			http.Error(w, fmt.Sprintf("Recovered from panic: %v", r), http.StatusInternalServerError)
		}
	}()

	// 2. Setup Proxy Configuration
	targetURL := "https://api.github.com/zen" // Example target
	
	// Execute the proxy logic
	err := proxyRequest(w, r, targetURL)

	// 3. Explicit Error Handling (The Go "Catch")
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "Proxy Error: %v", err)
		return
	}
}

// proxyRequest handles the actual logic of fetching from another server
func proxyRequest(w http.ResponseWriter, r *http.Request, target string) error {
	// Parse the target URL
	remote, err := url.Parse(target)
	if err != nil {
		return fmt.Errorf("invalid target URL: %w", err)
	}

	// Create a custom client with a timeout (Best practice)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create the outbound request
	proxyReq, err := http.NewRequest(r.Method, remote.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Copy headers from original request (optional, e.g., Auth tokens)
	proxyReq.Header.Set("User-Agent", "Vercel-Go-Proxy")

	// Execute the request
	resp, err := client.Do(proxyReq)
	if err != nil {
		return fmt.Errorf("external request failed: %w", err)
	}
	defer resp.Body.Close()

	// Copy the remote response headers and body back to the client
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	
	return err
}
