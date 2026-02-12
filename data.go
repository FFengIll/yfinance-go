package yfinance

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	mrand "math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/publicsuffix"
)

// CookieCache holds cached cookie data
type CookieCache struct {
	Cookie   string    `json:"cookie"`
	Crumb    string    `json:"crumb"`
	Expiry   time.Time `json:"expiry"`
	Strategy string    `json:"strategy"`
}

// YfData handles HTTP communication with Yahoo Finance API
type YfData struct {
	client         *http.Client
	jar            *cookiejar.Jar
	crumb          string
	cookie         string
	cookieStrategy string
	mu             sync.Mutex
	userAgent      string
	cacheDir       string
	sessionID      string
	transport      *utlsTransport
}

// utlsTransport is a custom transport that uses uTLS for TLS fingerprinting
type utlsTransport struct {
	originalTransport *http.Transport
	proxyURL          *url.URL
}

// NewUtlsTransport creates a new uTLS transport
func NewUtlsTransport() *utlsTransport {
	return NewUtlsTransportWithProxy("")
}

// NewUtlsTransportWithProxy creates a new uTLS transport with proxy support
func NewUtlsTransportWithProxy(proxy string) *utlsTransport {
	var proxyURL *url.URL
	if proxy != "" {
		proxyURL, _ = url.Parse(proxy)
	}

	return &utlsTransport{
		proxyURL: proxyURL,
		originalTransport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  false,
			TLSHandshakeTimeout: 10 * time.Second,
			ForceAttemptHTTP2:   false, // Disable HTTP/2
			Proxy: func(req *http.Request) (*url.URL, error) {
				if proxyURL != nil {
					return proxyURL, nil
				}
				return nil, nil
			},
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// Extract hostname for SNI
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					host = addr
					port = "443"
				}

				var tcpConn net.Conn

				// If using proxy, connect through it
				if proxyURL != nil {
					// Connect to proxy
					proxyHost := proxyURL.Host
					tcpConn, err = net.DialTimeout(network, proxyHost, 10*time.Second)
					if err != nil {
						return nil, fmt.Errorf("proxy dial error: %w", err)
					}

					// Send CONNECT request for HTTPS
					connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", net.JoinHostPort(host, port), host)
					_, err = tcpConn.Write([]byte(connectReq))
					if err != nil {
						tcpConn.Close()
						return nil, fmt.Errorf("proxy connect error: %w", err)
					}

					// Read proxy response
					buf := make([]byte, 1024)
					n, err := tcpConn.Read(buf)
					if err != nil {
						tcpConn.Close()
						return nil, fmt.Errorf("proxy response error: %w", err)
					}

					// Check for 200 Connection established
					if !bytes.Contains(buf[:n], []byte("200")) {
						tcpConn.Close()
						return nil, fmt.Errorf("proxy connection failed: %s", string(buf[:n]))
					}
				} else {
					// Direct connection
					tcpConn, err = net.DialTimeout(network, net.JoinHostPort(host, port), 10*time.Second)
					if err != nil {
						return nil, fmt.Errorf("dial error: %w", err)
					}
				}

				// Create uTLS config - no ALPN means HTTP/1.1 only
				config := &utls.Config{
					ServerName:         host,
					InsecureSkipVerify: false,
				}

				// Use randomized fingerprint without ALPN to force HTTP/1.1
				tlsConn := utls.UClient(tcpConn, config, utls.HelloRandomizedNoALPN)

				// Handshake
				if err := tlsConn.Handshake(); err != nil {
					tcpConn.Close()
					return nil, fmt.Errorf("TLS handshake error: %w", err)
				}

				return tlsConn, nil
			},
		},
	}
}

// RoundTrip implements http.RoundTripper
func (t *utlsTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.originalTransport.RoundTrip(req)
}

// CloseIdleConnections closes idle connections
func (t *utlsTransport) CloseIdleConnections() {
	t.originalTransport.CloseIdleConnections()
}

// NewYfData creates a new YfData instance
func NewYfData() *YfData {
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		jar, _ = cookiejar.New(nil)
	}

	// Generate session ID
	b := make([]byte, 8)
	rand.Read(b)
	sessionID := hex.EncodeToString(b)

	// Get proxy from config (which checks environment variables)
	proxy := GlobalConfig.GetProxy()
	transport := NewUtlsTransportWithProxy(proxy)

	yd := &YfData{
		jar:            jar,
		cookieStrategy: "basic",
		userAgent:      UserAgents[mrand.Intn(len(UserAgents))],
		sessionID:      sessionID,
		cacheDir:       getCacheDir(),
		transport:      transport,
	}

	yd.client = &http.Client{
		Timeout:   30 * time.Second,
		Jar:       jar,
		Transport: transport,
	}

	// Try to load cached cookie
	yd.loadCookieCache()

	return yd
}

// NewYfDataWithClient creates a new YfData instance with a custom HTTP client
func NewYfDataWithClient(client *http.Client) *YfData {
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		jar, _ = cookiejar.New(nil)
	}
	client.Jar = jar

	// Generate session ID
	b := make([]byte, 8)
	rand.Read(b)

	// Get proxy from config (which checks environment variables)
	proxy := GlobalConfig.GetProxy()
	transport := NewUtlsTransportWithProxy(proxy)

	yd := &YfData{
		client:         client,
		jar:            jar,
		cookieStrategy: "basic",
		userAgent:      UserAgents[mrand.Intn(len(UserAgents))],
		sessionID:      hex.EncodeToString(b),
		cacheDir:       getCacheDir(),
		transport:      transport,
	}

	// Try to load cached cookie
	yd.loadCookieCache()

	return yd
}

// getCacheDir returns the cache directory path
func getCacheDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	cacheDir := filepath.Join(homeDir, ".cache", "yfinance-go")
	os.MkdirAll(cacheDir, 0755)
	return cacheDir
}

// isValidCrumb checks if the crumb is valid (not an error page)
func isValidCrumb(crumb string) bool {
	if crumb == "" {
		return false
	}
	// Check for common error responses
	if strings.Contains(crumb, "<html>") ||
		strings.Contains(crumb, "<!DOCTYPE") ||
		strings.Contains(crumb, "Too Many Requests") ||
		strings.Contains(crumb, "Yahoo") {
		return false
	}
	return true
}

// loadCookieCache loads cached cookie from disk
func (yd *YfData) loadCookieCache() bool {
	if yd.cacheDir == "" {
		return false
	}

	cacheFile := filepath.Join(yd.cacheDir, "cookie_cache.json")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return false
	}

	var cache CookieCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return false
	}

	// Check if cache is expired
	if time.Now().After(cache.Expiry) {
		return false
	}

	// Validate crumb is not an error page
	if !isValidCrumb(cache.Crumb) {
		return false
	}

	yd.cookie = cache.Cookie
	yd.crumb = cache.Crumb
	yd.cookieStrategy = cache.Strategy
	return true
}

// saveCookieCache saves cookie to disk
func (yd *YfData) saveCookieCache() error {
	if yd.cacheDir == "" {
		return nil
	}

	cache := CookieCache{
		Cookie:   yd.cookie,
		Crumb:    yd.crumb,
		Expiry:   time.Now().Add(24 * time.Hour), // Cache for 24 hours
		Strategy: yd.cookieStrategy,
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	cacheFile := filepath.Join(yd.cacheDir, "cookie_cache.json")
	return os.WriteFile(cacheFile, data, 0644)
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
	if retries == 0 {
		retries = 3
	}

	for attempt := 0; attempt <= retries; attempt++ {
		resp, err := yd.doRequest(ctx, method, endpoint, params, body)
		if err != nil {
			lastErr = err
			if IsTransientError(err) && attempt < retries {
				backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
				if backoff > 30*time.Second {
					backoff = 30 * time.Second
				}
				time.Sleep(backoff)
				continue
			}
			return nil, err
		}

		// Handle rate limiting with strategy switch
		if resp.StatusCode == 429 {
			resp.Body.Close()
			lastErr = NewYFRateLimitError()

			// Switch cookie strategy and retry
			yd.switchCookieStrategy()

			if attempt < retries {
				backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
				if backoff > 30*time.Second {
					backoff = 30 * time.Second
				}
				time.Sleep(backoff)
				continue
			}
			return nil, lastErr
		}

		// Handle cookie consent redirect
		if yd.isConsentURL(resp.Request.URL.String()) {
			resp.Body.Close()
			if err := yd.acceptConsent(ctx); err != nil {
				return nil, err
			}
			continue
		}

		// Handle 401/403 - might need new cookie
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			resp.Body.Close()
			yd.ResetCrumb()
			if attempt < retries {
				continue
			}
			return nil, fmt.Errorf("authentication failed: %d", resp.StatusCode)
		}

		return resp, nil
	}

	return nil, lastErr
}

// switchCookieStrategy toggles between basic and csrf strategies
func (yd *YfData) switchCookieStrategy() {
	yd.mu.Lock()
	defer yd.mu.Unlock()

	if yd.cookieStrategy == "basic" {
		yd.cookieStrategy = "csrf"
	} else {
		yd.cookieStrategy = "basic"
	}
	yd.crumb = ""
	yd.cookie = ""
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

	// Set browser-like headers (thread-safe)
	yd.setBrowserHeadersSafe(req)

	return yd.client.Do(req)
}

// setBrowserHeadersSafe sets headers with lock protection
func (yd *YfData) setBrowserHeadersSafe(req *http.Request) {
	yd.mu.Lock()
	ua := yd.userAgent
	yd.mu.Unlock()

	yd.setBrowserHeadersWithUA(req, ua)
}

// setBrowserHeaders sets realistic browser headers (must be called with lock held)
func (yd *YfData) setBrowserHeaders(req *http.Request) {
	yd.setBrowserHeadersWithUA(req, yd.userAgent)
}

// setBrowserHeadersWithUA sets headers with provided user agent
func (yd *YfData) setBrowserHeadersWithUA(req *http.Request, ua string) {
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Cache-Control", "max-age=0")

	if body := req.Body; body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
}

// getCookieAndCrumb fetches and caches the cookie and crumb for authentication
func (yd *YfData) getCookieAndCrumb(ctx context.Context) (string, error) {
	yd.mu.Lock()
	defer yd.mu.Unlock()

	// Return cached crumb if available
	if yd.crumb != "" {
		return yd.crumb, nil
	}

	var crumb string
	var err error

	if yd.cookieStrategy == "csrf" {
		crumb, err = yd.getCookieAndCrumbCSRFInternal(ctx)
		if err != nil {
			// Fall back to basic
			yd.cookieStrategy = "basic"
			crumb, err = yd.getCookieAndCrumbBasicInternal(ctx)
		}
	} else {
		crumb, err = yd.getCookieAndCrumbBasicInternal(ctx)
		if err != nil {
			// Try CSRF as fallback
			yd.cookieStrategy = "csrf"
			crumb, err = yd.getCookieAndCrumbCSRFInternal(ctx)
		}
	}

	if err != nil {
		return "", err
	}

	// Validate crumb before saving
	if !isValidCrumb(crumb) {
		return "", fmt.Errorf("invalid crumb received: %s", crumb)
	}

	yd.crumb = crumb
	yd.saveCookieCache()
	return crumb, nil
}

// getCookieAndCrumbBasicInternal gets cookie and crumb using basic strategy (must be called with lock held)
func (yd *YfData) getCookieAndCrumbBasicInternal(ctx context.Context) (string, error) {
	// Get cookie first
	if err := yd.getCookieBasicInternal(ctx); err != nil {
		return "", fmt.Errorf("failed to get cookie: %w", err)
	}

	// Then get crumb
	return yd.getCrumbBasicInternal(ctx)
}

// getCookieBasicInternal fetches cookie using basic strategy (must be called with lock held)
func (yd *YfData) getCookieBasicInternal(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://fc.yahoo.com", nil)
	if err != nil {
		return err
	}

	yd.setBrowserHeadersWithUA(req, yd.userAgent)

	resp, err := yd.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Extract A3 cookie from response or jar
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "A3" {
			yd.cookie = cookie.Value
			return nil
		}
	}

	// Check jar for yahoo.com cookies
	yahooURL, _ := url.Parse("https://yahoo.com")
	for _, cookie := range yd.jar.Cookies(yahooURL) {
		if cookie.Name == "A3" {
			yd.cookie = cookie.Value
			return nil
		}
	}

	return nil
}

// getCrumbBasicInternal fetches the crumb token (must be called with lock held)
func (yd *YfData) getCrumbBasicInternal(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://query1.finance.yahoo.com/v1/test/getcrumb", nil)
	if err != nil {
		return "", err
	}

	yd.setBrowserHeadersWithUA(req, yd.userAgent)

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
	if crumb == "" || strings.Contains(crumb, "<html>") || strings.Contains(crumb, "Too Many Requests") {
		return "", fmt.Errorf("failed to get crumb: %s", crumb)
	}

	return crumb, nil
}

// getCookieAndCrumbCSRFInternal gets cookie and crumb using CSRF strategy (must be called with lock held)
func (yd *YfData) getCookieAndCrumbCSRFInternal(ctx context.Context) (string, error) {
	// Get cookie via consent flow
	if err := yd.getCookieCSRFInternal(ctx); err != nil {
		return "", err
	}

	// Get crumb from query2
	return yd.getCrumbCSRFInternal(ctx)
}

// getCookieCSRFInternal fetches cookie using CSRF/consent strategy (must be called with lock held)
func (yd *YfData) getCookieCSRFInternal(ctx context.Context) error {
	// Step 1: Get consent page
	req, err := http.NewRequestWithContext(ctx, "GET", "https://guce.yahoo.com/consent", nil)
	if err != nil {
		return err
	}
	yd.setBrowserHeadersWithUA(req, yd.userAgent)

	resp, err := yd.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	htmlBody := string(body)

	// Extract csrfToken
	csrfToken := extractInputValue(htmlBody, "csrfToken")
	if csrfToken == "" {
		return fmt.Errorf("failed to find csrfToken")
	}

	// Extract sessionId
	sessionId := extractInputValue(htmlBody, "sessionId")
	if sessionId == "" {
		// Use our generated session ID
		sessionId = yd.sessionID
	}

	// Step 2: Submit consent form
	formData := url.Values{}
	formData.Set("agree", "agree")
	formData.Set("consentUUID", "default")
	formData.Set("sessionId", sessionId)
	formData.Set("csrfToken", csrfToken)
	formData.Set("originalDoneUrl", "https://finance.yahoo.com/")
	formData.Set("namespace", "yahoo")

	consentURL := fmt.Sprintf("https://consent.yahoo.com/v2/collectConsent?sessionId=%s", sessionId)
	req2, err := http.NewRequestWithContext(ctx, "POST", consentURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	yd.setBrowserHeadersWithUA(req2, yd.userAgent)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp2, err := yd.client.Do(req2)
	if err != nil {
		return err
	}
	resp2.Body.Close()

	// Step 3: Copy consent
	copyURL := fmt.Sprintf("https://guce.yahoo.com/copyConsent?sessionId=%s", sessionId)
	req3, err := http.NewRequestWithContext(ctx, "GET", copyURL, nil)
	if err != nil {
		return err
	}
	yd.setBrowserHeadersWithUA(req3, yd.userAgent)

	resp3, err := yd.client.Do(req3)
	if err != nil {
		return err
	}
	resp3.Body.Close()

	yd.cookie = "csrf-obtained"
	return nil
}

// getCrumbCSRFInternal fetches crumb using query2 endpoint (must be called with lock held)
func (yd *YfData) getCrumbCSRFInternal(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://query2.finance.yahoo.com/v1/test/getcrumb", nil)
	if err != nil {
		return "", err
	}

	yd.setBrowserHeadersWithUA(req, yd.userAgent)

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
	if crumb == "" || strings.Contains(crumb, "<html>") || strings.Contains(crumb, "Too Many Requests") {
		return "", fmt.Errorf("failed to get crumb via CSRF: %s", crumb)
	}

	return crumb, nil
}

// extractInputValue extracts value from HTML input by name
func extractInputValue(html, name string) string {
	// Simple extraction - find input with given name
	pattern := fmt.Sprintf(`name="%s"`, name)
	idx := strings.Index(html, pattern)
	if idx == -1 {
		pattern = fmt.Sprintf(`name='%s'`, name)
		idx = strings.Index(html, pattern)
		if idx == -1 {
			return ""
		}
	}

	// Find value attribute
	valueStart := strings.Index(html[idx:], `value="`)
	if valueStart == -1 {
		valueStart = strings.Index(html[idx:], `value='`)
		if valueStart == -1 {
			return ""
		}
	}
	valueStart += idx + 7

	valueEnd := strings.Index(html[valueStart:], `"`)
	if valueEnd == -1 {
		valueEnd = strings.Index(html[valueStart:], `'`)
		if valueEnd == -1 {
			return ""
		}
	}

	return html[valueStart : valueStart+valueEnd]
}

// isConsentURL checks if the URL is a consent page
func (yd *YfData) isConsentURL(urlStr string) bool {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return strings.HasSuffix(parsed.Hostname(), "consent.yahoo.com")
}

// acceptConsent handles the consent form when redirected
func (yd *YfData) acceptConsent(ctx context.Context) error {
	yd.mu.Lock()
	defer yd.mu.Unlock()

	// Reset and use CSRF strategy
	yd.crumb = ""
	yd.cookie = ""
	yd.cookieStrategy = "csrf"

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

// ClearCache clears the cookie cache file
func (yd *YfData) ClearCache() error {
	yd.mu.Lock()
	defer yd.mu.Unlock()

	yd.crumb = ""
	yd.cookie = ""

	if yd.cacheDir == "" {
		return nil
	}

	cacheFile := filepath.Join(yd.cacheDir, "cookie_cache.json")
	return os.Remove(cacheFile)
}
