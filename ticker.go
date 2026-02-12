package yfinance

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Ticker represents a Yahoo Finance ticker/symbol
type Ticker struct {
	Symbol string
	data   *YfData
	tz     string
}

// NewTicker creates a new Ticker instance
func NewTicker(symbol string) *Ticker {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	return &Ticker{
		Symbol: symbol,
		data:   NewYfData(),
	}
}

// NewTickerWithData creates a new Ticker with a custom YfData instance
func NewTickerWithData(symbol string, data *YfData) *Ticker {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	return &Ticker{
		Symbol: symbol,
		data:   data,
	}
}

// NewTickerWithMIC creates a new Ticker with Market Identifier Code
func NewTickerWithMIC(symbol, micCode string) (*Ticker, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))

	// Clean MIC code
	if strings.HasPrefix(micCode, ".") {
		micCode = micCode[1:]
	}

	micCode = strings.ToUpper(micCode)
	suffix, ok := MICToYahooSuffix[micCode]
	if !ok {
		return nil, fmt.Errorf("unknown MIC code: '%s'", micCode)
	}

	if suffix != "" {
		symbol = fmt.Sprintf("%s.%s", symbol, suffix)
	}

	return NewTicker(symbol), nil
}

// String returns the ticker symbol
func (t *Ticker) String() string {
	return t.Symbol
}

// History fetches historical price data for the ticker
func (t *Ticker) History(ctx context.Context, options *HistoryOptions) (*HistoryResult, error) {
	if options == nil {
		options = DefaultHistoryOptions()
	}

	// Validate options
	if err := options.Validate(); err != nil {
		return nil, err
	}

	// Get timezone if needed
	if t.tz == "" {
		tz, err := t.GetTimezone(ctx)
		if err == nil {
			t.tz = tz
		}
	}

	params := options.ToParams()
	endpoint := fmt.Sprintf("%s/v8/finance/chart/%s", BaseURL, t.Symbol)

	var result chartResponse
	if err := t.data.GetRawJSON(ctx, endpoint, params, &result); err != nil {
		return nil, err
	}

	if result.Chart.Error != nil {
		return nil, fmt.Errorf("chart error: %s", result.Chart.Error.Description)
	}

	if len(result.Chart.Result) == 0 {
		return nil, NewYFPricesMissingError(t.Symbol, "")
	}

	chartResult := result.Chart.Result[0]
	return t.parseChartResult(chartResult, options)
}

// HistoryOptions defines options for fetching historical data
type HistoryOptions struct {
	Period      string     // 1d, 5d, 1mo, 3mo, 6mo, 1y, 2y, 5y, 10y, ytd, max
	Interval    string     // 1m, 2m, 5m, 15m, 30m, 60m, 90m, 1h, 1d, 5d, 1wk, 1mo, 3mo
	Start       *time.Time // Start date
	End         *time.Time // End date
	PrePost     bool       // Include pre/post market data
	AutoAdjust  bool       // Auto-adjust prices for splits/dividends
	BackAdjust  bool       // Back-adjust prices
	Repair      bool       // Detect and repair price errors
	KeepNaN     bool       // Keep NaN rows
	Rounding    bool       // Round to 2 decimal places
	Timeout     int        // Request timeout in seconds
	ShowErrors  bool       // Show errors in response
}

// DefaultHistoryOptions returns default history options
func DefaultHistoryOptions() *HistoryOptions {
	return &HistoryOptions{
		Period:     "1mo",
		Interval:   "1d",
		PrePost:    false,
		AutoAdjust: true,
		BackAdjust: false,
		Repair:     false,
		KeepNaN:    false,
		Rounding:   false,
		Timeout:    10,
		ShowErrors: false,
	}
}

// Validate validates the history options
func (o *HistoryOptions) Validate() error {
	if o.Period != "" {
		valid := false
		for _, p := range ValidPeriods {
			if p == o.Period {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid period: %s, must be one of: %v", o.Period, ValidPeriods)
		}
	}

	if o.Interval != "" {
		valid := false
		for _, i := range ValidIntervals {
			if i == o.Interval {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid interval: %s, must be one of: %v", o.Interval, ValidIntervals)
		}
	}

	return nil
}

// ToParams converts options to URL parameters
func (o *HistoryOptions) ToParams() map[string]string {
	params := make(map[string]string)

	if o.Period != "" && o.Start == nil {
		params["range"] = o.Period
	}

	if o.Interval != "" {
		params["interval"] = o.Interval
	}

	if o.Start != nil {
		params["period1"] = fmt.Sprintf("%d", o.Start.Unix())
	}

	if o.End != nil {
		params["period2"] = fmt.Sprintf("%d", o.End.Unix())
	}

	if o.PrePost {
		params["includePrePost"] = "true"
	}

	if o.ShowErrors {
		params["events"] = "div,split"
	}

	return params
}

// HistoryResult contains historical price data
type HistoryResult struct {
	Meta       HistoryMeta
	Data       []PriceData
	Dividends  []DividendData
	Splits     []SplitData
	Timezone   string
	Currency   string
	Exchange   string
}

// HistoryMeta contains metadata about the historical data
type HistoryMeta struct {
	Symbol            string
	Currency          string
	ExchangeName      string
	InstrumentType    string
	FirstTradeDate    time.Time
	RegularMarketTime time.Time
	Gmtoffset         int
	Timezone          string
	ExchangeTimezone  string
	RegularMarketPrice float64
	ChartPreviousClose float64
	PreviousClose      float64
}

// PriceData represents a single price data point
type PriceData struct {
	Date          time.Time
	Open          float64
	High          float64
	Low           float64
	Close         float64
	AdjClose      float64
	Volume        int64
}

// DividendData represents dividend information
type DividendData struct {
	Date   time.Time
	Amount float64
}

// SplitData represents stock split information
type SplitData struct {
	Date     time.Time
	Ratio    string
	Numerator   float64
	Denominator float64
}

// chartResponse represents the Yahoo Finance chart API response
type chartResponse struct {
	Chart struct {
		Result []chartResult `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

// chartResult represents a single chart result
type chartResult struct {
	Meta       chartMeta        `json:"meta"`
	Timestamp  []int64          `json:"timestamp"`
	Indicators chartIndicators  `json:"indicators"`
	Events     *chartEvents     `json:"events,omitempty"`
}

// chartMeta represents chart metadata
type chartMeta struct {
	Currency             string  `json:"currency"`
	Symbol               string  `json:"symbol"`
	ExchangeName         string  `json:"exchangeName"`
	InstrumentType       string  `json:"instrumentType"`
	FirstTradeDate       int64   `json:"firstTradeDate"`
	RegularMarketTime    int64   `json:"regularMarketTime"`
	Gmtoffset            int     `json:"gmtoffset"`
	Timezone             string  `json:"timezone"`
	ExchangeTimezoneName string  `json:"exchangeTimezoneName"`
	RegularMarketPrice   float64 `json:"regularMarketPrice"`
	ChartPreviousClose   float64 `json:"chartPreviousClose"`
	PreviousClose        float64 `json:"previousClose,omitempty"`
}

// chartIndicators contains price indicators
type chartIndicators struct {
	Quote []chartQuote `json:"quote"`
	Adjclose []struct {
		Adjclose []float64 `json:"adjclose"`
	} `json:"adjclose,omitempty"`
}

// chartQuote contains quote data
type chartQuote struct {
	Open   []float64 `json:"open"`
	High   []float64 `json:"high"`
	Low    []float64 `json:"low"`
	Close  []float64 `json:"close"`
	Volume []int64   `json:"volume"`
}

// chartEvents contains event data (dividends, splits)
type chartEvents struct {
	Dividends map[string]struct {
		Amount float64 `json:"amount"`
	} `json:"dividends"`
	Splits map[string]struct {
		Numerator   float64 `json:"numerator"`
		Denominator float64 `json:"denominator"`
	} `json:"splits"`
}

// parseChartResult parses the chart result into HistoryResult
func (t *Ticker) parseChartResult(result chartResult, options *HistoryOptions) (*HistoryResult, error) {
	hr := &HistoryResult{
		Timezone: result.Meta.Timezone,
		Currency: result.Meta.Currency,
		Exchange: result.Meta.ExchangeName,
	}

	// Parse metadata
	hr.Meta = HistoryMeta{
		Symbol:             result.Meta.Symbol,
		Currency:           result.Meta.Currency,
		ExchangeName:       result.Meta.ExchangeName,
		InstrumentType:     result.Meta.InstrumentType,
		Gmtoffset:          result.Meta.Gmtoffset,
		Timezone:           result.Meta.Timezone,
		ExchangeTimezone:   result.Meta.ExchangeTimezoneName,
		RegularMarketPrice: result.Meta.RegularMarketPrice,
		ChartPreviousClose: result.Meta.ChartPreviousClose,
		PreviousClose:      result.Meta.PreviousClose,
	}

	if result.Meta.FirstTradeDate > 0 {
		hr.Meta.FirstTradeDate = time.Unix(result.Meta.FirstTradeDate, 0)
	}
	if result.Meta.RegularMarketTime > 0 {
		hr.Meta.RegularMarketTime = time.Unix(result.Meta.RegularMarketTime, 0)
	}

	// Get timezone for parsing
	loc := time.UTC
	if result.Meta.Timezone != "" {
		if parsed, err := time.LoadLocation(result.Meta.Timezone); err == nil {
			loc = parsed
		}
	}

	// Parse price data
	if len(result.Timestamp) > 0 && len(result.Indicators.Quote) > 0 {
		quote := result.Indicators.Quote[0]
		adjClose := []float64(nil)
		if len(result.Indicators.Adjclose) > 0 {
			adjClose = result.Indicators.Adjclose[0].Adjclose
		}

		hr.Data = make([]PriceData, 0, len(result.Timestamp))
		for i, ts := range result.Timestamp {
			if i >= len(quote.Open) {
				break
			}

			pd := PriceData{
				Date:   time.Unix(ts, 0).In(loc),
				Open:   quote.Open[i],
				High:   quote.High[i],
				Low:    quote.Low[i],
				Close:  quote.Close[i],
				Volume: quote.Volume[i],
			}

			if adjClose != nil && i < len(adjClose) {
				pd.AdjClose = adjClose[i]
			} else {
				pd.AdjClose = pd.Close
			}

			hr.Data = append(hr.Data, pd)
		}
	}

	// Parse dividends
	if result.Events != nil && result.Events.Dividends != nil {
		hr.Dividends = make([]DividendData, 0, len(result.Events.Dividends))
		for tsStr, div := range result.Events.Dividends {
			var ts int64
			fmt.Sscanf(tsStr, "%d", &ts)
			hr.Dividends = append(hr.Dividends, DividendData{
				Date:   time.Unix(ts, 0).In(loc),
				Amount: div.Amount,
			})
		}
	}

	// Parse splits
	if result.Events != nil && result.Events.Splits != nil {
		hr.Splits = make([]SplitData, 0, len(result.Events.Splits))
		for tsStr, split := range result.Events.Splits {
			var ts int64
			fmt.Sscanf(tsStr, "%d", &ts)
			hr.Splits = append(hr.Splits, SplitData{
				Date:        time.Unix(ts, 0).In(loc),
				Numerator:   split.Numerator,
				Denominator: split.Denominator,
				Ratio:       fmt.Sprintf("%.0f:%.0f", split.Numerator, split.Denominator),
			})
		}
	}

	// Auto-adjust prices if requested
	if options.AutoAdjust && len(hr.Data) > 0 {
		hr.AutoAdjustPrices()
	}

	return hr, nil
}

// AutoAdjustPrices adjusts historical prices for splits and dividends
func (hr *HistoryResult) AutoAdjustPrices() {
	if len(hr.Data) == 0 {
		return
	}

	// Get the last close and adjClose
	lastClose := hr.Data[len(hr.Data)-1].Close
	lastAdjClose := hr.Data[len(hr.Data)-1].AdjClose

	if lastClose == 0 || lastAdjClose == 0 {
		return
	}

	adjustFactor := lastAdjClose / lastClose

	for i := range hr.Data {
		hr.Data[i].AdjClose = hr.Data[i].Close * adjustFactor
	}
}

// GetTimezone fetches the timezone for the ticker
func (t *Ticker) GetTimezone(ctx context.Context) (string, error) {
	if t.tz != "" {
		return t.tz, nil
	}

	params := map[string]string{
		"range":    "1d",
		"interval": "1d",
	}
	endpoint := fmt.Sprintf("%s/v8/finance/chart/%s", BaseURL, t.Symbol)

	var result chartResponse
	if err := t.data.GetRawJSON(ctx, endpoint, params, &result); err != nil {
		return "", err
	}

	if len(result.Chart.Result) == 0 {
		return "", NewYFTzMissingError(t.Symbol)
	}

	tz := result.Chart.Result[0].Meta.Timezone
	t.tz = tz
	return tz, nil
}
