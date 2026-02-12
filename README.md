# yfinance-go

A Go implementation of Yahoo Finance API client, based on the Python [yfinance](https://github.com/ranaroussi/yfinance) library.

## Features

- Stock quotes and price data
- Historical price data (OHLCV) with customizable periods and intervals
- Company information and fundamentals
- News and analyst recommendations
- Symbol search functionality
- Multi-ticker download support
- Cookie and crumb authentication handling

## Installation

```bash
go get github.com/ranaroussi/yfinance-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "time"

    yf "github.com/ranaroussi/yfinance-go"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Create a ticker
    ticker := yf.NewTicker("AAPL")

    // Get current quote
    quote, err := ticker.GetQuote(ctx)
    if err != nil {
        panic(err)
    }
    fmt.Printf("%s: $%.2f\n", quote.Symbol, quote.RegularMarketPrice)

    // Get historical data
    history, err := ticker.History(ctx, &yf.HistoryOptions{
        Period:   "1mo",
        Interval: "1d",
    })
    if err != nil {
        panic(err)
    }

    for _, data := range history.Data {
        fmt.Printf("%s: Close=%.2f\n", data.Date.Format("2006-01-02"), data.Close)
    }
}
```

## API Reference

### Ticker

```go
// Create a new ticker
ticker := yf.NewTicker("AAPL")

// Create ticker with Market Identifier Code
ticker, err := yf.NewTickerWithMIC("OR", "XPAR") // Returns OR.PA

// Get current quote
quote, err := ticker.GetQuote(ctx)

// Get fast info (essential ticker info)
info, err := ticker.GetFastInfo(ctx)

// Get detailed company info
info, err := ticker.GetInfo(ctx)

// Get historical data
history, err := ticker.History(ctx, &yf.HistoryOptions{
    Period:   "1mo",      // 1d, 5d, 1mo, 3mo, 6mo, 1y, 2y, 5y, 10y, ytd, max
    Interval: "1d",       // 1m, 2m, 5m, 15m, 30m, 60m, 90m, 1h, 1d, 5d, 1wk, 1mo, 3mo
    PrePost:  true,       // Include pre/post market data
})

// Get news
news, err := ticker.GetNews(ctx, 10)

// Get analyst recommendations
recs, err := ticker.GetRecommendations(ctx)

// Get calendar events
calendar, err := ticker.GetCalendar(ctx)
```

### Historical Data Options

```go
// Using period
options := &yf.HistoryOptions{
    Period:   "1y",
    Interval: "1d",
}

// Using date range
start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
end := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)

options := &yf.HistoryOptions{
    Start:    &start,
    End:      &end,
    Interval: "1d",
}
```

### Multiple Tickers

```go
// Get quotes for multiple symbols
quotes, err := yf.GetQuotes(ctx, []string{"AAPL", "MSFT", "GOOGL"})

// Download historical data for multiple tickers
result, err := yf.Download(ctx, &yf.DownloadOptions{
    Tickers:  []string{"AAPL", "MSFT", "GOOGL"},
    Period:   "3mo",
    Interval: "1d",
})

for symbol, history := range result.Data {
    fmt.Printf("%s: %d data points\n", symbol, len(history.Data))
}
```

### Search

```go
// Search for symbols
quotes, err := yf.SearchSymbols(ctx, "Apple",
    yf.WithMaxResults(10),
    yf.WithFuzzyQuery(true),
)

for _, quote := range quotes {
    fmt.Printf("%s - %s\n", quote.Symbol, quote.ShortName)
}
```

### Configuration

```go
// Global configuration
yf.SetConfig("", 3, false, 30) // proxy, retries, hideExceptions, timeout

// Or use individual setters
cfg := yf.GlobalConfig
cfg.SetProxy("http://proxy:8080")
cfg.SetRetries(5)
cfg.SetTimeout(60)
```

## License

Apache License 2.0

## Credits

Based on the Python [yfinance](https://github.com/ranaroussi/yfinance) library by Ran Aroussi.
