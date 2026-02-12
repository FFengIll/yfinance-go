package yfinance

import (
	"context"
	"testing"
	"time"
)

func TestNewTicker(t *testing.T) {
	ticker := NewTicker("AAPL")
	if ticker.Symbol != "AAPL" {
		t.Errorf("Expected symbol AAPL, got %s", ticker.Symbol)
	}
}

func TestNewTickerLowercase(t *testing.T) {
	ticker := NewTicker("aapl")
	if ticker.Symbol != "AAPL" {
		t.Errorf("Expected symbol AAPL, got %s", ticker.Symbol)
	}
}

func TestNewTickerWithMIC(t *testing.T) {
	ticker, err := NewTickerWithMIC("OR", "XPAR")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if ticker.Symbol != "OR.PA" {
		t.Errorf("Expected symbol OR.PA, got %s", ticker.Symbol)
	}
}

func TestNewTickerWithInvalidMIC(t *testing.T) {
	_, err := NewTickerWithMIC("TEST", "INVALID")
	if err == nil {
		t.Error("Expected error for invalid MIC code")
	}
}

func TestHistoryOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		options *HistoryOptions
		wantErr bool
	}{
		{
			name:    "default options",
			options: DefaultHistoryOptions(),
			wantErr: false,
		},
		{
			name: "invalid period",
			options: &HistoryOptions{
				Period:   "invalid",
				Interval: "1d",
			},
			wantErr: true,
		},
		{
			name: "invalid interval",
			options: &HistoryOptions{
				Period:   "1mo",
				Interval: "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid period and interval",
			options: &HistoryOptions{
				Period:   "1y",
				Interval: "1d",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHistoryOptionsToParams(t *testing.T) {
	// Test with period only (no start/end)
	optionsWithPeriod := &HistoryOptions{
		Period:   "1y",
		Interval: "1d",
		PrePost:  true,
	}

	params := optionsWithPeriod.ToParams()

	if params["range"] != "1y" {
		t.Errorf("Expected range 1y, got %s", params["range"])
	}
	if params["interval"] != "1d" {
		t.Errorf("Expected interval 1d, got %s", params["interval"])
	}
	if params["includePrePost"] != "true" {
		t.Errorf("Expected includePrePost true, got %s", params["includePrePost"])
	}

	// Test with start/end dates
	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)

	optionsWithDates := &HistoryOptions{
		Interval: "1d",
		Start:    &start,
		End:      &end,
	}

	paramsWithDates := optionsWithDates.ToParams()

	if _, ok := paramsWithDates["period1"]; !ok {
		t.Error("Expected period1 parameter")
	}
	if _, ok := paramsWithDates["period2"]; !ok {
		t.Error("Expected period2 parameter")
	}
	// When start is set, range should not be used
	if _, ok := paramsWithDates["range"]; ok {
		t.Error("Did not expect range parameter when start is set")
	}
}

func TestDefaultDownloadOptions(t *testing.T) {
	options := DefaultDownloadOptions()
	if options.Period != "1mo" {
		t.Errorf("Expected period 1mo, got %s", options.Period)
	}
	if options.Interval != "1d" {
		t.Errorf("Expected interval 1d, got %s", options.Interval)
	}
	if options.AutoAdjust != true {
		t.Error("Expected AutoAdjust to be true")
	}
}

func TestValidPeriods(t *testing.T) {
	periods := ValidPeriods
	expected := []string{"1d", "5d", "1mo", "3mo", "6mo", "1y", "2y", "5y", "10y", "ytd", "max"}
	if len(periods) != len(expected) {
		t.Errorf("Expected %d periods, got %d", len(expected), len(periods))
	}
}

func TestValidIntervals(t *testing.T) {
	intervals := ValidIntervals
	expected := []string{"1m", "2m", "5m", "15m", "30m", "60m", "90m", "1h", "1d", "5d", "1wk", "1mo", "3mo"}
	if len(intervals) != len(expected) {
		t.Errorf("Expected %d intervals, got %d", len(expected), len(intervals))
	}
}

func TestIntradayIntervals(t *testing.T) {
	if !IntradayIntervals["1m"] {
		t.Error("Expected 1m to be an intraday interval")
	}
	if !IntradayIntervals["1h"] {
		t.Error("Expected 1h to be an intraday interval")
	}
	if IntradayIntervals["1d"] {
		t.Error("Expected 1d to NOT be an intraday interval")
	}
}

func TestMICMapping(t *testing.T) {
	tests := []struct {
		mic      string
		expected string
	}{
		{"XNYS", ""},
		{"XNAS", ""},
		{"XPAR", "PA"},
		{"XHKG", "HK"},
		{"XTKS", "T"},
	}

	for _, tt := range tests {
		t.Run(tt.mic, func(t *testing.T) {
			suffix, ok := MICToYahooSuffix[tt.mic]
			if !ok {
				t.Errorf("MIC %s not found in mapping", tt.mic)
			}
			if suffix != tt.expected {
				t.Errorf("Expected suffix %s, got %s", tt.expected, suffix)
			}
		})
	}
}

// Integration tests (require network)
// These tests are skipped by default, use -tags=integration to run

func TestIntegrationGetQuote(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ticker := NewTicker("AAPL")
	quote, err := ticker.GetQuote(ctx)
	if err != nil {
		t.Fatalf("Failed to get quote: %v", err)
	}

	if quote.Symbol != "AAPL" {
		t.Errorf("Expected symbol AAPL, got %s", quote.Symbol)
	}
	if quote.RegularMarketPrice <= 0 {
		t.Errorf("Expected positive price, got %f", quote.RegularMarketPrice)
	}
}

func TestIntegrationGetHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ticker := NewTicker("AAPL")
	history, err := ticker.History(ctx, &HistoryOptions{
		Period:   "1mo",
		Interval: "1d",
	})
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(history.Data) == 0 {
		t.Error("Expected some history data")
	}
	if history.Meta.Symbol != "AAPL" {
		t.Errorf("Expected symbol AAPL, got %s", history.Meta.Symbol)
	}
}

func TestIntegrationSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	quotes, err := SearchSymbols(ctx, "Apple", WithMaxResults(5))
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(quotes) == 0 {
		t.Error("Expected some search results")
	}
}

func TestIntegrationDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := Download(ctx, &DownloadOptions{
		Tickers:  []string{"AAPL", "MSFT"},
		Period:   "1mo",
		Interval: "1d",
	})
	if err != nil {
		t.Fatalf("Failed to download: %v", err)
	}

	if len(result.Succeeded) != 2 {
		t.Errorf("Expected 2 successful downloads, got %d", len(result.Succeeded))
	}
}
