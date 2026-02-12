package yfinance

import (
	"os"
	"sync"
)

// Config holds the yfinance configuration
type Config struct {
	mu sync.RWMutex

	// Network configuration
	Proxy   string
	Retries int

	// Debug configuration
	HideExceptions bool
	Logging        bool

	// Request timeout in seconds
	Timeout int
}

// GlobalConfig is the global configuration instance
var GlobalConfig = &Config{
	Proxy:          "",
	Retries:        0,
	HideExceptions: true,
	Logging:        false,
	Timeout:        30,
}

// SetProxy sets the proxy for HTTP requests
func (c *Config) SetProxy(proxy string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Proxy = proxy
}

// GetProxy gets the current proxy setting
// Returns environment variable YFINANCE_PROXY if set, otherwise returns configured proxy
func (c *Config) GetProxy() string {
	// First check environment variable
	if proxy := os.Getenv("YFINANCE_PROXY"); proxy != "" {
		return proxy
	}
	// Then check HTTP_PROXY/HTTPS_PROXY
	if proxy := os.Getenv("HTTPS_PROXY"); proxy != "" {
		return proxy
	}
	if proxy := os.Getenv("HTTP_PROXY"); proxy != "" {
		return proxy
	}
	// Finally return configured proxy
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Proxy
}

// SetRetries sets the number of retries for failed requests
func (c *Config) SetRetries(retries int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Retries = retries
}

// GetRetries gets the current retry count
func (c *Config) GetRetries() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Retries
}

// SetHideExceptions sets whether to hide exceptions
func (c *Config) SetHideExceptions(hide bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.HideExceptions = hide
}

// GetHideExceptions gets the current hide exceptions setting
func (c *Config) GetHideExceptions() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.HideExceptions
}

// SetTimeout sets the request timeout in seconds
func (c *Config) SetTimeout(timeout int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Timeout = timeout
}

// GetTimeout gets the current timeout setting
func (c *Config) GetTimeout() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Timeout
}

// SetConfig is a convenience function to set multiple config options
func SetConfig(proxy string, retries int, hideExceptions bool, timeout int) {
	cfg := GlobalConfig
	if proxy != "" {
		cfg.SetProxy(proxy)
	}
	if retries >= 0 {
		cfg.SetRetries(retries)
	}
	cfg.SetHideExceptions(hideExceptions)
	if timeout > 0 {
		cfg.SetTimeout(timeout)
	}
}
