package yfinance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// YfData handles HTTP communication with Yahoo Finance API
type YfData struct {
	client         *http.Client
	crumb          string
	cookie         string
	cookieStrategy string
	mu             sync.Mutex
	userAgent      string
}

// NewYfData creates a new YfData instance
func NewYfData() *YfData {
	yd := &YfData{
		client:         &http.Client{Timeout: 30 * time.Second},
		cookieStrategy: "basic",
		userAgent:      UserAgents[rand.Intn(len(UserAgents))],
	}
	return yd
}

// NewYfDataWithClient creates a new YfData instance with a custom HTTP client
func NewYfDataWithClient(client *http.Client) *YfData {
	yd := &YfData{
		client:         client,
		cookieStrategy: "basic",
		userAgent:      UserAgents[rand.Intn(len(UserAgents))],
	}
	return yd
}

// SetUserAgent sets a custom user agent
func (yd *YfData) SetUserAgent(ua string) {
	yd.mu.Lock()
	defer yd.mu.Unlock()
	yd.userAgent = ua
}

// Get performs a GET request to Yahoo Finance API
func (yd *YfData) Get(ctx context.Context, endpoint string, params map[string]string) (*http.Response, error) {
	return yd.makeRequest(ctx, "GET", endpoint, params, nil)
}

// Post performs a POST request to Yahoo Finance API
func (yd *YfData) Post(ctx context.Context, endpoint string, params map[string]string, body interface{}) (*http.Response, error) {
	return yd.makeRequest(ctx, "POST", endpoint, params, body)
}

// makeRequest creates and executes an HTTP request with retry logic
func (yd *YfData) makeRequest(ctx context.Context, method, endpoint string, params map[string]string, body interface{}) (*http.Response, error) {
	var lastErr error
	retries := GlobalConfig.GetRetries()

	for attempt := 0; attempt <= retries; attempt++ {
		resp, err := yd.doRequest(ctx, method, endpoint, params, body)
		if err != nil {
			lastErr = err
			if IsTransientError(err) && attempt < retries {
				time.Sleep(time.Duration(1<<uint(attempt)) * time.Second)
				continue
			}
			return nil, err
		}

		// Handle rate limiting
		if resp.StatusCode == 429 {
			resp.Body.Close()
			lastErr = NewYFRateLimitError()
			if attempt < retries {
				time.Sleep(time.Duration(1<<uint(attempt)) * time.Second)
				continue
			}
			return nil, lastErr
		}

		// Handle cookie consent redirect
		if yd.isConsentURL(resp.Request.URL.String()) {
			resp.Body.Close()
			if err := yd.acceptConsent(ctx, endpoint); err != nil {
				return nil, err
			}
			continue
		}

		return resp, nil
	}

	return nil, lastErr
}

// doRequest executes a single HTTP request
func (yd *YfData) doRequest(ctx context.Context, method, endpoint string, params map[string]string, body interface{}) (*http.Response, error) {
	// Build URL with params
	reqURL := endpoint
	if len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Set(k, v)
		}
		reqURL = fmt.Sprintf("%s?%s", endpoint, values.Encode())
	}

	// Ensure we have crumb
	crumb, err := yd.getCookieAndCrumb(ctx)
	if err != nil {
		// Try without crumb
		crumb = ""
	}

	// Add crumb to params
	if crumb != "" {
		if strings.Contains(reqURL, "?") {
			reqURL = fmt.Sprintf("%s&crumb=%s", reqURL, url.QueryEscape(crumb))
		} else {
			reqURL = fmt.Sprintf("%s?crumb=%s", reqURL, url.QueryEscape(crumb))
		}
	}

	// Create request
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	yd.mu.Lock()
	ua := yd.userAgent
	yd.mu.Unlock()

	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set proxy if configured
	proxy := GlobalConfig.GetProxy()
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err == nil {
			yd.mu.Lock()
			yd.client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
			yd.mu.Unlock()
		}
	}

	return yd.client.Do(req)
}

// getCookieAndCrumb fetches and caches the cookie and crumb for authentication
func (yd *YfData) getCookieAndCrumb(ctx context.Context) (string, error) {
	yd.mu.Lock()
	defer yd.mu.Unlock()

	if yd.crumb != "" {
		return yd.crumb, nil
	}

	// First get cookie (internal method, no lock needed)
	if err := yd.getCookieBasicInternal(ctx); err != nil {
		// Try CSRF strategy
		if err := yd.getCookieCSRF(ctx); err != nil {
			return "", err
		}
	}

	// Then get crumb (internal method, no lock needed)
	crumb, err := yd.getCrumbBasicInternal(ctx)
	if err != nil {
		return "", err
	}

	yd.crumb = crumb
	return crumb, nil
}

// getUserAgent safely returns the user agent
func (yd *YfData) getUserAgent() string {
	yd.mu.Lock()
	defer yd.mu.Unlock()
	return yd.userAgent
}

// getCookieBasicInternal fetches cookie using basic strategy (must be called with lock held)
func (yd *YfData) getCookieBasicInternal(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://fc.yahoo.com", nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", yd.userAgent)
	req.Header.Set("Accept", "*/*")

	resp, err := yd.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Extract cookie from response
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "A3" {
			yd.cookie = cookie.Value
			return nil
		}
	}

	return nil
}

// getCookieBasic fetches cookie using basic strategy
func (yd *YfData) getCookieBasic(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://fc.yahoo.com", nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", yd.getUserAgent())
	req.Header.Set("Accept", "*/*")

	resp, err := yd.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Extract cookie from response
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "A3" {
			yd.mu.Lock()
			yd.cookie = cookie.Value
			yd.mu.Unlock()
			return nil
		}
	}

	return nil
}

// getCookieCSRF fetches cookie using CSRF strategy (fallback)
func (yd *YfData) getCookieCSRF(ctx context.Context) error {
	// This is a simplified implementation
	// Full implementation would parse the consent form
	return fmt.Errorf("CSRF cookie strategy not implemented")
}

// getCrumbBasicInternal fetches the crumb token (must be called with lock held)
func (yd *YfData) getCrumbBasicInternal(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://query1.finance.yahoo.com/v1/test/getcrumb", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", yd.userAgent)
	req.Header.Set("Accept", "*/*")

	resp, err := yd.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return "", NewYFRateLimitError()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	crumb := string(body)
	if crumb == "" || strings.Contains(crumb, "<html>") {
		return "", fmt.Errorf("failed to get crumb")
	}

	return crumb, nil
}

// getCrumbBasic fetches the crumb token
func (yd *YfData) getCrumbBasic(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://query1.finance.yahoo.com/v1/test/getcrumb", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", yd.getUserAgent())
	req.Header.Set("Accept", "*/*")

	resp, err := yd.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return "", NewYFRateLimitError()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	crumb := string(body)
	if crumb == "" || strings.Contains(crumb, "<html>") {
		return "", fmt.Errorf("failed to get crumb")
	}

	return crumb, nil
}

// isConsentURL checks if the URL is a consent page
func (yd *YfData) isConsentURL(urlStr string) bool {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return strings.HasSuffix(parsed.Hostname(), "consent.yahoo.com")
}

// acceptConsent handles the consent form
func (yd *YfData) acceptConsent(ctx context.Context, originalURL string) error {
	// Simplified implementation - just reset and try again
	yd.mu.Lock()
	yd.crumb = ""
	yd.cookie = ""
	yd.mu.Unlock()
	return nil
}

// GetRawJSON fetches and parses JSON from a URL
func (yd *YfData) GetRawJSON(ctx context.Context, endpoint string, params map[string]string, v interface{}) error {
	resp, err := yd.Get(ctx, endpoint, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Check for Yahoo downtime message
	if strings.Contains(string(body), "Will be right back") {
		return NewYFDataException("*** YAHOO! FINANCE IS CURRENTLY DOWN! ***")
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	return nil
}

// ResetCrumb clears the cached crumb (useful when getting auth errors)
func (yd *YfData) ResetCrumb() {
	yd.mu.Lock()
	defer yd.mu.Unlock()
	yd.crumb = ""
	yd.cookie = ""
}
