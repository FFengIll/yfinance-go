package yfinance

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Quote represents stock quote information
type Quote struct {
	Symbol                   string    `json:"symbol"`
	ShortName                string    `json:"shortName"`
	LongName                 string    `json:"longName"`
	Exchange                 string    `json:"exchange"`
	Market                   string    `json:"market"`
	QuoteType                string    `json:"quoteType"`
	Currency                 string    `json:"currency"`

	// Price information
	RegularMarketPrice       float64   `json:"regularMarketPrice"`
	RegularMarketChange      float64   `json:"regularMarketChange"`
	RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
	RegularMarketOpen        float64   `json:"regularMarketOpen"`
	RegularMarketDayHigh     float64   `json:"regularMarketDayHigh"`
	RegularMarketDayLow      float64   `json:"regularMarketDayLow"`
	RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
	RegularMarketVolume      int64     `json:"regularMarketVolume"`
	RegularMarketTime        time.Time `json:"regularMarketTime"`

	// Pre/Post market
	PreMarketPrice           float64   `json:"preMarketPrice"`
	PreMarketChange          float64   `json:"preMarketChange"`
	PreMarketChangePercent   float64   `json:"preMarketChangePercent"`
	PostMarketPrice          float64   `json:"postMarketPrice"`
	PostMarketChange         float64   `json:"postMarketChange"`
	PostMarketChangePercent  float64   `json:"postMarketChangePercent"`

	// 52 week range
	FiftyTwoWeekLow          float64   `json:"fiftyTwoWeekLow"`
	FiftyTwoWeekHigh         float64   `json:"fiftyTwoWeekHigh"`
	FiftyDayAverage          float64   `json:"fiftyDayAverage"`
	TwoHundredDayAverage     float64   `json:"twoHundredDayAverage"`

	// Market data
	MarketCap                int64     `json:"marketCap"`
	SharesOutstanding        int64     `json:"sharesOutstanding"`
	FloatShares              int64     `json:"floatShares"`
	Beta                     float64   `json:"beta"`
	PE                       float64   `json:"trailingPE"`
	ForwardPE                float64   `json:"forwardPE"`
	DividendYield            float64   `json:"dividendYield"`
	EPS                      float64   `json:"trailingEps"`
	ForwardEPS               float64   `json:"forwardEps"`

	// Additional info
	BookValue                float64   `json:"bookValue"`
	PriceToBook              float64   `json:"priceToBook"`
	Revenue                  int64     `json:"totalRevenue"`
	EBITDA                   int64     `json:"ebitda"`
	ProfitMargin             float64   `json:"profitMargins"`
	OperatingMargin          float64   `json:"operatingMargins"`

	// Timestamps
	FirstTradeDate           time.Time `json:"firstTradeDate"`
}

// FastInfo provides quick access to essential ticker information
type FastInfo struct {
	Symbol         string    `json:"symbol"`
	ShortName      string    `json:"shortName"`
	Exchange       string    `json:"exchange"`
	Currency       string    `json:"currency"`
	MarketPrice    float64   `json:"marketPrice"`
	MarketCap      int64     `json:"marketCap"`
	Shares         int64     `json:"shares"`
	LastPrice      float64   `json:"lastPrice"`
	LastVolume     int64     `json:"lastVolume"`
	PreviousClose  float64   `json:"previousClose"`
	Open           float64   `json:"open"`
	DayHigh        float64   `json:"dayHigh"`
	DayLow         float64   `json:"dayLow"`
	FiftyDayAvg    float64   `json:"fiftyDayAverage"`
	TwoHundredDayAvg float64 `json:"twoHundredDayAverage"`
	YearHigh       float64   `json:"yearHigh"`
	YearLow        float64   `json:"yearLow"`
	YearChange     float64   `json:"yearChange"`
	Timezone       string    `json:"timezone"`
	QuoteType      string    `json:"quoteType"`
}

// GetQuote fetches the current quote for the ticker
func (t *Ticker) GetQuote(ctx context.Context) (*Quote, error) {
	params := map[string]string{
		"symbols": t.Symbol,
		"fields": strings.Join([]string{
			"symbol",
			"shortName",
			"longName",
			"exchangeName",
			"market",
			"quoteType",
			"currency",
			"regularMarketPrice",
			"regularMarketChange",
			"regularMarketChangePercent",
			"regularMarketOpen",
			"regularMarketDayHigh",
			"regularMarketDayLow",
			"regularMarketPreviousClose",
			"regularMarketVolume",
			"regularMarketTime",
			"preMarketPrice",
			"preMarketChange",
			"preMarketChangePercent",
			"postMarketPrice",
			"postMarketChange",
			"postMarketChangePercent",
			"fiftyTwoWeekLow",
			"fiftyTwoWeekHigh",
			"fiftyDayAverage",
			"twoHundredDayAverage",
			"marketCap",
			"sharesOutstanding",
			"floatShares",
			"beta",
			"trailingPE",
			"forwardPE",
			"dividendYield",
			"trailingEps",
			"forwardEps",
			"bookValue",
			"priceToBook",
			"totalRevenue",
			"ebitda",
			"profitMargins",
			"operatingMargins",
		}, ","),
	}

	endpoint := fmt.Sprintf("%s/v7/finance/quote", Query1URL)

	var result quoteResponse
	if err := t.data.GetRawJSON(ctx, endpoint, params, &result); err != nil {
		return nil, err
	}

	if result.QuoteResponse.Error != nil {
		return nil, fmt.Errorf("quote error: %v", result.QuoteResponse.Error)
	}

	if len(result.QuoteResponse.Result) == 0 {
		return nil, NewYFTickerMissingError(t.Symbol, "no quote data found")
	}

	return parseQuote(result.QuoteResponse.Result[0]), nil
}

// quoteResponse represents the quote API response
type quoteResponse struct {
	QuoteResponse struct {
		Result []quoteResult `json:"result"`
		Error  interface{}  `json:"error"`
	} `json:"quoteResponse"`
}

// quoteResult represents a single quote result
type quoteResult struct {
	Symbol                     string      `json:"symbol"`
	ShortName                  string      `json:"shortName"`
	LongName                   string      `json:"longName"`
	ExchangeName               string      `json:"exchangeName"`
	Market                     string      `json:"market"`
	QuoteType                  string      `json:"quoteType"`
	Currency                   string      `json:"currency"`
	RegularMarketPrice         float64     `json:"regularMarketPrice"`
	RegularMarketChange        float64     `json:"regularMarketChange"`
	RegularMarketChangePercent float64     `json:"regularMarketChangePercent"`
	RegularMarketOpen          float64     `json:"regularMarketOpen"`
	RegularMarketDayHigh       float64     `json:"regularMarketDayHigh"`
	RegularMarketDayLow        float64     `json:"regularMarketDayLow"`
	RegularMarketPreviousClose float64     `json:"regularMarketPreviousClose"`
	RegularMarketVolume        int64       `json:"regularMarketVolume"`
	RegularMarketTime          interface{} `json:"regularMarketTime"` // Can be int or int64
	PreMarketPrice             float64     `json:"preMarketPrice"`
	PreMarketChange            float64     `json:"preMarketChange"`
	PreMarketChangePercent     float64     `json:"preMarketChangePercent"`
	PostMarketPrice            float64     `json:"postMarketPrice"`
	PostMarketChange           float64     `json:"postMarketChange"`
	PostMarketChangePercent    float64     `json:"postMarketChangePercent"`
	FiftyTwoWeekLow            float64     `json:"fiftyTwoWeekLow"`
	FiftyTwoWeekHigh           float64     `json:"fiftyTwoWeekHigh"`
	FiftyDayAverage            float64     `json:"fiftyDayAverage"`
	TwoHundredDayAverage       float64     `json:"twoHundredDayAverage"`
	MarketCap                  int64       `json:"marketCap"`
	SharesOutstanding          int64       `json:"sharesOutstanding"`
	FloatShares                int64       `json:"floatShares"`
	Beta                       float64     `json:"beta"`
	TrailingPE                 float64     `json:"trailingPE"`
	ForwardPE                  float64     `json:"forwardPE"`
	TrailingAnnualDividendYield float64    `json:"trailingAnnualDividendYield"`
	TrailingEps                float64     `json:"trailingEps"`
	ForwardEps                 float64     `json:"forwardEps"`
	BookValue                  float64     `json:"bookValue"`
	PriceToBook                float64     `json:"priceToBook"`
	TotalRevenue               int64       `json:"totalRevenue"`
	EBITDA                     int64       `json:"ebitda"`
	ProfitMargins              float64     `json:"profitMargins"`
	OperatingMargins           float64     `json:"operatingMargins"`
	FirstTradeDateMilliseconds int64       `json:"firstTradeDateMilliseconds"`
}

// parseQuote converts quoteResult to Quote
func parseQuote(qr quoteResult) *Quote {
	quote := &Quote{
		Symbol:                     qr.Symbol,
		ShortName:                  qr.ShortName,
		LongName:                   qr.LongName,
		Exchange:                   qr.ExchangeName,
		Market:                     qr.Market,
		QuoteType:                  qr.QuoteType,
		Currency:                   qr.Currency,
		RegularMarketPrice:         qr.RegularMarketPrice,
		RegularMarketChange:        qr.RegularMarketChange,
		RegularMarketChangePercent: qr.RegularMarketChangePercent,
		RegularMarketOpen:          qr.RegularMarketOpen,
		RegularMarketDayHigh:       qr.RegularMarketDayHigh,
		RegularMarketDayLow:        qr.RegularMarketDayLow,
		RegularMarketPreviousClose: qr.RegularMarketPreviousClose,
		RegularMarketVolume:        qr.RegularMarketVolume,
		PreMarketPrice:             qr.PreMarketPrice,
		PreMarketChange:            qr.PreMarketChange,
		PreMarketChangePercent:     qr.PreMarketChangePercent,
		PostMarketPrice:            qr.PostMarketPrice,
		PostMarketChange:           qr.PostMarketChange,
		PostMarketChangePercent:    qr.PostMarketChangePercent,
		FiftyTwoWeekLow:            qr.FiftyTwoWeekLow,
		FiftyTwoWeekHigh:           qr.FiftyTwoWeekHigh,
		FiftyDayAverage:            qr.FiftyDayAverage,
		TwoHundredDayAverage:       qr.TwoHundredDayAverage,
		MarketCap:                  qr.MarketCap,
		SharesOutstanding:          qr.SharesOutstanding,
		FloatShares:                qr.FloatShares,
		Beta:                       qr.Beta,
		PE:                         qr.TrailingPE,
		ForwardPE:                  qr.ForwardPE,
		DividendYield:              qr.TrailingAnnualDividendYield,
		EPS:                        qr.TrailingEps,
		ForwardEPS:                 qr.ForwardEps,
		BookValue:                  qr.BookValue,
		PriceToBook:                qr.PriceToBook,
		Revenue:                    qr.TotalRevenue,
		EBITDA:                     qr.EBITDA,
		ProfitMargin:               qr.ProfitMargins,
		OperatingMargin:            qr.OperatingMargins,
	}

	// Parse timestamps
	switch v := qr.RegularMarketTime.(type) {
	case int64:
		quote.RegularMarketTime = time.Unix(v, 0)
	case int:
		quote.RegularMarketTime = time.Unix(int64(v), 0)
	case float64:
		quote.RegularMarketTime = time.Unix(int64(v), 0)
	}

	if qr.FirstTradeDateMilliseconds > 0 {
		quote.FirstTradeDate = time.Unix(qr.FirstTradeDateMilliseconds/1000, 0)
	}

	return quote
}

// GetFastInfo gets quick essential information for the ticker
func (t *Ticker) GetFastInfo(ctx context.Context) (*FastInfo, error) {
	// Get quote first
	quote, err := t.GetQuote(ctx)
	if err != nil {
		return nil, err
	}

	// Get history for additional data
	history, err := t.History(ctx, &HistoryOptions{
		Period:   "1y",
		Interval: "1d",
	})
	if err != nil {
		// Still return quote data even if history fails
		return &FastInfo{
			Symbol:        quote.Symbol,
			ShortName:     quote.ShortName,
			Exchange:      quote.Exchange,
			Currency:      quote.Currency,
			MarketPrice:   quote.RegularMarketPrice,
			MarketCap:     quote.MarketCap,
			Shares:        quote.SharesOutstanding,
			LastPrice:     quote.RegularMarketPrice,
			PreviousClose: quote.RegularMarketPreviousClose,
			Open:          quote.RegularMarketOpen,
			DayHigh:       quote.RegularMarketDayHigh,
			DayLow:        quote.RegularMarketDayLow,
			FiftyDayAvg:   quote.FiftyDayAverage,
			TwoHundredDayAvg: quote.TwoHundredDayAverage,
			YearHigh:      quote.FiftyTwoWeekHigh,
			YearLow:       quote.FiftyTwoWeekLow,
			QuoteType:     quote.QuoteType,
		}, nil
	}

	// Calculate year change
	var yearChange float64
	if len(history.Data) > 0 {
		firstPrice := history.Data[0].Close
		lastPrice := history.Data[len(history.Data)-1].Close
		if firstPrice > 0 {
			yearChange = (lastPrice - firstPrice) / firstPrice * 100
		}
	}

	return &FastInfo{
		Symbol:           quote.Symbol,
		ShortName:        quote.ShortName,
		Exchange:         quote.Exchange,
		Currency:         quote.Currency,
		MarketPrice:      quote.RegularMarketPrice,
		MarketCap:        quote.MarketCap,
		Shares:           quote.SharesOutstanding,
		LastPrice:        quote.RegularMarketPrice,
		LastVolume:       quote.RegularMarketVolume,
		PreviousClose:    quote.RegularMarketPreviousClose,
		Open:             quote.RegularMarketOpen,
		DayHigh:          quote.RegularMarketDayHigh,
		DayLow:           quote.RegularMarketDayLow,
		FiftyDayAvg:      quote.FiftyDayAverage,
		TwoHundredDayAvg: quote.TwoHundredDayAverage,
		YearHigh:         quote.FiftyTwoWeekHigh,
		YearLow:          quote.FiftyTwoWeekLow,
		YearChange:       yearChange,
		Timezone:         history.Timezone,
		QuoteType:        quote.QuoteType,
	}, nil
}

// GetQuotes fetches quotes for multiple tickers
func GetQuotes(ctx context.Context, symbols []string) ([]*Quote, error) {
	if len(symbols) == 0 {
		return []*Quote{}, nil
	}

	params := map[string]string{
		"symbols": strings.Join(symbols, ","),
		"fields": strings.Join([]string{
			"symbol",
			"shortName",
			"longName",
			"exchangeName",
			"market",
			"quoteType",
			"currency",
			"regularMarketPrice",
			"regularMarketChange",
			"regularMarketChangePercent",
			"regularMarketOpen",
			"regularMarketDayHigh",
			"regularMarketDayLow",
			"regularMarketPreviousClose",
			"regularMarketVolume",
			"regularMarketTime",
			"fiftyTwoWeekLow",
			"fiftyTwoWeekHigh",
			"marketCap",
			"sharesOutstanding",
		}, ","),
	}

	endpoint := fmt.Sprintf("%s/v7/finance/quote", Query1URL)
	data := NewYfData()

	var result quoteResponse
	if err := data.GetRawJSON(ctx, endpoint, params, &result); err != nil {
		return nil, err
	}

	if result.QuoteResponse.Error != nil {
		return nil, fmt.Errorf("quote error: %v", result.QuoteResponse.Error)
	}

	quotes := make([]*Quote, 0, len(result.QuoteResponse.Result))
	for _, qr := range result.QuoteResponse.Result {
		quotes = append(quotes, parseQuote(qr))
	}

	return quotes, nil
}
