// Package yfinance provides a Go implementation of Yahoo Finance API client
// based on the Python yfinance library.
//
// This library provides access to Yahoo Finance data including:
//   - Stock quotes and price data
//   - Historical price data (OHLCV)
//   - Company information and fundamentals
//   - News and analyst recommendations
//   - Symbol search functionality
//
// Basic Usage:
//
//	// Create a ticker
//	ticker := yfinance.NewTicker("AAPL")
//
//	// Get current quote
//	ctx := context.Background()
//	quote, err := ticker.GetQuote(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("%s: $%.2f\n", quote.Symbol, quote.RegularMarketPrice)
//
//	// Get historical data
//	history, err := ticker.History(ctx, &yfinance.HistoryOptions{
//	    Period:   "1mo",
//	    Interval: "1d",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, data := range history.Data {
//	    fmt.Printf("%s: Close=%.2f\n", data.Date.Format("2006-01-02"), data.Close)
//	}
//
// Multiple Tickers:
//
//	// Download data for multiple tickers
//	result, err := yfinance.Download(ctx, &yfinance.DownloadOptions{
//	    Tickers:  []string{"AAPL", "MSFT", "GOOGL"},
//	    Period:   "3mo",
//	    Interval: "1d",
//	})
//
//	// Get quotes for multiple symbols
//	quotes, err := yfinance.GetQuotes(ctx, []string{"AAPL", "MSFT", "GOOGL"})
//
// Search:
//
//	// Search for symbols
//	quotes, err := yfinance.SearchSymbols(ctx, "Apple", yfinance.WithMaxResults(10))
//
// Configuration:
//
//	// Configure global settings
//	yfinance.SetConfig("", 3, false, 30) // proxy, retries, hideExceptions, timeout
//
//	// Or use individual setters
//	cfg := yfinance.GlobalConfig
//	cfg.SetProxy("http://proxy:8080")
//	cfg.SetRetries(5)
//	cfg.SetTimeout(60)
package yfinance
