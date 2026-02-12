package yfinance

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// DownloadOptions contains options for downloading multiple tickers
type DownloadOptions struct {
	Tickers       []string
	Period        string
	Interval      string
	Start         *time.Time
	End           *time.Time
	GroupBy       string // "column" or "ticker"
	PrePost       bool
	AutoAdjust    bool
	BackAdjust    bool
	Repair        bool
	KeepNaN       bool
	Threads       int
	Progress      bool
	ShowErrors    bool
	Timeout       int
}

// DefaultDownloadOptions returns default download options
func DefaultDownloadOptions() *DownloadOptions {
	return &DownloadOptions{
		Period:     "1mo",
		Interval:   "1d",
		GroupBy:    "column",
		PrePost:    false,
		AutoAdjust: true,
		BackAdjust: false,
		Repair:     false,
		KeepNaN:    false,
		Threads:    4,
		Progress:   false,
		ShowErrors: false,
		Timeout:    10,
	}
}

// DownloadResult contains the result of downloading multiple tickers
type DownloadResult struct {
	Data       map[string]*HistoryResult
	Errors     map[string]error
	Failed     []string
	Succeeded  []string
}

// Download downloads historical data for multiple tickers
func Download(ctx context.Context, options *DownloadOptions) (*DownloadResult, error) {
	if options == nil {
		options = DefaultDownloadOptions()
	}

	if len(options.Tickers) == 0 {
		return &DownloadResult{
			Data:      make(map[string]*HistoryResult),
			Errors:    make(map[string]error),
			Failed:    []string{},
			Succeeded: []string{},
		}, nil
	}

	// Normalize tickers
	tickers := make([]string, 0, len(options.Tickers))
	seen := make(map[string]bool)
	for _, t := range options.Tickers {
		// Handle comma-separated tickers
		parts := strings.Split(t, ",")
		for _, p := range parts {
			p = strings.ToUpper(strings.TrimSpace(p))
			if p != "" && !seen[p] {
				tickers = append(tickers, p)
				seen[p] = true
			}
		}
	}

	result := &DownloadResult{
		Data:      make(map[string]*HistoryResult),
		Errors:    make(map[string]error),
		Failed:    []string{},
		Succeeded: []string{},
	}

	// Create shared YfData for cookie/crumb sharing
	sharedData := NewYfData()

	// Use mutex for thread-safe map access
	var mu sync.Mutex

	// Determine number of workers
	workers := options.Threads
	if workers <= 0 {
		workers = 1
	}
	if workers > len(tickers) {
		workers = len(tickers)
	}

	// Create ticker channel
	tickerChan := make(chan string, len(tickers))
	for _, t := range tickers {
		tickerChan <- t
	}
	close(tickerChan)

	// Create wait group
	var wg sync.WaitGroup
	wg.Add(workers)

	// Start workers
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for ticker := range tickerChan {
				// Create ticker with shared data
				t := NewTickerWithData(ticker, sharedData)

				// Create history options
				histOpts := &HistoryOptions{
					Period:      options.Period,
					Interval:    options.Interval,
					Start:       options.Start,
					End:         options.End,
					PrePost:     options.PrePost,
					AutoAdjust:  options.AutoAdjust,
					BackAdjust:  options.BackAdjust,
					Repair:      options.Repair,
					KeepNaN:     options.KeepNaN,
					Timeout:     options.Timeout,
				}

				// Fetch history
				history, err := t.History(ctx, histOpts)

				mu.Lock()
				if err != nil {
					result.Errors[ticker] = err
					result.Failed = append(result.Failed, ticker)
				} else {
					result.Data[ticker] = history
					result.Succeeded = append(result.Succeeded, ticker)
				}
				mu.Unlock()
			}
		}()
	}

	// Wait for all workers to complete
	wg.Wait()

	return result, nil
}

// DownloadSimple is a simplified download function for common use cases
func DownloadSimple(ctx context.Context, tickers []string, period, interval string) (map[string]*HistoryResult, error) {
	result, err := Download(ctx, &DownloadOptions{
		Tickers:    tickers,
		Period:     period,
		Interval:   interval,
		AutoAdjust: true,
	})
	if err != nil {
		return nil, err
	}

	if len(result.Failed) > 0 {
		return result.Data, fmt.Errorf("failed to download: %v", result.Failed)
	}

	return result.Data, nil
}

// Tickers represents multiple tickers
type Tickers struct {
	Symbols []string
	data    *YfData
}

// NewTickers creates a new Tickers instance
func NewTickers(symbols []string) *Tickers {
	// Normalize symbols
	normalized := make([]string, 0, len(symbols))
	seen := make(map[string]bool)
	for _, s := range symbols {
		s = strings.ToUpper(strings.TrimSpace(s))
		if s != "" && !seen[s] {
			normalized = append(normalized, s)
			seen[s] = true
		}
	}

	return &Tickers{
		Symbols: normalized,
		data:    NewYfData(),
	}
}

// History fetches historical data for all tickers
func (t *Tickers) History(ctx context.Context, options *HistoryOptions) (map[string]*HistoryResult, error) {
	result := make(map[string]*HistoryResult)
	errors := make(map[string]error)

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, symbol := range t.Symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			ticker := NewTickerWithData(sym, t.data)
			history, err := ticker.History(ctx, options)

			mu.Lock()
			if err != nil {
				errors[sym] = err
			} else {
				result[sym] = history
			}
			mu.Unlock()
		}(symbol)
	}

	wg.Wait()

	if len(errors) > 0 {
		return result, fmt.Errorf("some tickers failed: %v", errors)
	}

	return result, nil
}

// Quotes fetches quotes for all tickers
func (t *Tickers) Quotes(ctx context.Context) ([]*Quote, error) {
	return GetQuotes(ctx, t.Symbols)
}

// String returns the string representation
func (t *Tickers) String() string {
	return strings.Join(t.Symbols, ",")
}
