package yfinance_test

import (
	"context"
	"fmt"
	"time"

	yf "github.com/FFengIll/yfinance-go"
)

func ExampleNewTicker() {
	// Create a new ticker
	ticker := yf.NewTicker("AAPL")
	fmt.Printf("Ticker: %s\n", ticker.Symbol)
}

func ExampleTicker_History() {
	// Create a ticker
	ticker := yf.NewTicker("AAPL")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch 1 month of daily history
	options := yf.DefaultHistoryOptions()
	options.Period = "1mo"
	options.Interval = "1d"

	history, err := ticker.History(ctx, options)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print history data
	fmt.Printf("Symbol: %s\n", history.Meta.Symbol)
	fmt.Printf("Timezone: %s\n", history.Timezone)
	fmt.Printf("Data points: %d\n", len(history.Data))

	for i, data := range history.Data {
		if i >= 5 {
			break // Only print first 5
		}
		fmt.Printf("%s: Open=%.2f, High=%.2f, Low=%.2f, Close=%.2f, Volume=%d\n",
			data.Date.Format("2006-01-02"),
			data.Open,
			data.High,
			data.Low,
			data.Close,
			data.Volume,
		)
	}
}

func ExampleTicker_GetQuote() {
	// Create a ticker
	ticker := yf.NewTicker("AAPL")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	quote, err := ticker.GetQuote(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Symbol: %s\n", quote.Symbol)
	fmt.Printf("Name: %s\n", quote.ShortName)
	fmt.Printf("Price: %.2f %s\n", quote.RegularMarketPrice, quote.Currency)
	fmt.Printf("Change: %.2f (%.2f%%)\n", quote.RegularMarketChange, quote.RegularMarketChangePercent)
	fmt.Printf("Volume: %d\n", quote.RegularMarketVolume)
	fmt.Printf("Market Cap: %d\n", quote.MarketCap)
}

func ExampleTicker_GetFastInfo() {
	ticker := yf.NewTicker("MSFT")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := ticker.GetFastInfo(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Symbol: %s\n", info.Symbol)
	fmt.Printf("Name: %s\n", info.ShortName)
	fmt.Printf("Exchange: %s\n", info.Exchange)
	fmt.Printf("Market Price: %.2f\n", info.MarketPrice)
	fmt.Printf("52 Week High: %.2f\n", info.YearHigh)
	fmt.Printf("52 Week Low: %.2f\n", info.YearLow)
}

func ExampleTicker_GetInfo() {
	ticker := yf.NewTicker("GOOGL")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := ticker.GetInfo(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Symbol: %s\n", info.Symbol)
	fmt.Printf("Name: %s\n", info.LongName)
	fmt.Printf("Sector: %s\n", info.Sector)
	fmt.Printf("Industry: %s\n", info.Industry)
	fmt.Printf("Website: %s\n", info.Website)
	fmt.Printf("Employees: %d\n", info.FullTimeEmployees)
}

func ExampleDownload() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Download historical data for multiple tickers
	result, err := yf.Download(ctx, &yf.DownloadOptions{
		Tickers:    []string{"AAPL", "MSFT", "GOOGL"},
		Period:     "3mo",
		Interval:   "1d",
		AutoAdjust: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Successfully downloaded: %v\n", result.Succeeded)
	fmt.Printf("Failed: %v\n", result.Failed)

	for symbol, history := range result.Data {
		fmt.Printf("\n%s:\n", symbol)
		if len(history.Data) > 0 {
			latest := history.Data[len(history.Data)-1]
			fmt.Printf("  Latest Close: %.2f\n", latest.Close)
		}
	}
}

func ExampleSearchSymbols() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Search for symbols
	quotes, err := yf.SearchSymbols(ctx, "Apple",
		yf.WithMaxResults(10),
		yf.WithFuzzyQuery(true),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, quote := range quotes {
		fmt.Printf("%s - %s (%s)\n", quote.Symbol, quote.ShortName, quote.Exchange)
	}
}

func ExampleGetQuotes() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get quotes for multiple symbols at once
	quotes, err := yf.GetQuotes(ctx, []string{"AAPL", "MSFT", "GOOGL"})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, quote := range quotes {
		fmt.Printf("%s: %.2f %s\n", quote.Symbol, quote.RegularMarketPrice, quote.Currency)
	}
}

func ExampleNewTickerWithMIC() {
	// Create a ticker using Market Identifier Code
	ticker, err := yf.NewTickerWithMIC("OR", "XPAR")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(ticker.Symbol) // Output: OR.PA
}

func ExampleConfig() {
	// Configure global settings
	yf.SetConfig("", 3, false, 30) // proxy, retries, hideExceptions, timeout

	// Or use individual setters
	cfg := yf.GlobalConfig
	cfg.SetProxy("http://proxy:8080")
	cfg.SetRetries(5)
	cfg.SetTimeout(60)
}

func ExampleTicker_GetNews() {
	ticker := yf.NewTicker("TSLA")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	news, err := ticker.GetNews(ctx, 5)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, n := range news {
		fmt.Printf("%s - %s\n", n.Title, n.Publisher)
		fmt.Printf("  %s\n", n.Link)
	}
}

func ExampleTicker_GetRecommendations() {
	ticker := yf.NewTicker("AAPL")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	recs, err := ticker.GetRecommendations(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, r := range recs {
		fmt.Printf("%s: Buy=%d, Hold=%d, Sell=%d\n",
			r.Period, r.Buy+r.StrongBuy, r.Hold, r.Sell+r.StrongSell)
	}
}

func ExampleHistoryOptions_customRange() {
	ticker := yf.NewTicker("AAPL")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Custom date range
	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)

	options := &yf.HistoryOptions{
		Start:      &start,
		End:        &end,
		Interval:   "1d",
		AutoAdjust: true,
	}

	history, err := ticker.History(ctx, options)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Fetched %d days of data\n", len(history.Data))
}
