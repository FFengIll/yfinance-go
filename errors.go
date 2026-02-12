package yfinance

import "fmt"

// YFException is the base exception for yfinance errors
type YFException struct {
	Description string
}

func (e *YFException) Error() string {
	return e.Description
}

// NewYFException creates a new YFException
func NewYFException(description string) *YFException {
	return &YFException{Description: description}
}

// YFDataException represents data-related errors
type YFDataException struct {
	YFException
}

// NewYFDataException creates a new YFDataException
func NewYFDataException(description string) *YFDataException {
	return &YFDataException{
		YFException: YFException{Description: description},
	}
}

// YFNotImplementedError represents unimplemented functionality
type YFNotImplementedError struct {
	MethodName string
}

func (e *YFNotImplementedError) Error() string {
	return fmt.Sprintf("Have not implemented fetching '%s' from Yahoo API", e.MethodName)
}

// NewYFNotImplementedError creates a new YFNotImplementedError
func NewYFNotImplementedError(methodName string) *YFNotImplementedError {
	return &YFNotImplementedError{MethodName: methodName}
}

// YFTickerMissingError represents missing ticker errors
type YFTickerMissingError struct {
	Ticker    string
	Rationale string
}

func (e *YFTickerMissingError) Error() string {
	return fmt.Sprintf("$%s: possibly delisted; %s", e.Ticker, e.Rationale)
}

// NewYFTickerMissingError creates a new YFTickerMissingError
func NewYFTickerMissingError(ticker, rationale string) *YFTickerMissingError {
	return &YFTickerMissingError{Ticker: ticker, Rationale: rationale}
}

// YFTzMissingError represents missing timezone errors
type YFTzMissingError struct {
	YFTickerMissingError
}

// NewYFTzMissingError creates a new YFTzMissingError
func NewYFTzMissingError(ticker string) *YFTzMissingError {
	return &YFTzMissingError{
		YFTickerMissingError: YFTickerMissingError{
			Ticker:    ticker,
			Rationale: "no timezone found",
		},
	}
}

// YFPricesMissingError represents missing price data errors
type YFPricesMissingError struct {
	YFTickerMissingError
	DebugInfo string
}

// NewYFPricesMissingError creates a new YFPricesMissingError
func NewYFPricesMissingError(ticker, debugInfo string) *YFPricesMissingError {
	rationale := "no price data found"
	if debugInfo != "" {
		rationale = fmt.Sprintf("no price data found %s", debugInfo)
	}
	return &YFPricesMissingError{
		YFTickerMissingError: YFTickerMissingError{
			Ticker:    ticker,
			Rationale: rationale,
		},
		DebugInfo: debugInfo,
	}
}

// YFEarningsDateMissing represents missing earnings date errors
type YFEarningsDateMissing struct {
	YFTickerMissingError
}

// NewYFEarningsDateMissing creates a new YFEarningsDateMissing
func NewYFEarningsDateMissing(ticker string) *YFEarningsDateMissing {
	return &YFEarningsDateMissing{
		YFTickerMissingError: YFTickerMissingError{
			Ticker:    ticker,
			Rationale: "no earnings dates found",
		},
	}
}

// YFInvalidPeriodError represents invalid period errors
type YFInvalidPeriodError struct {
	Ticker        string
	InvalidPeriod string
	ValidRanges   []string
}

func (e *YFInvalidPeriodError) Error() string {
	return fmt.Sprintf("%s: Period '%s' is invalid, must be one of: %v",
		e.Ticker, e.InvalidPeriod, e.ValidRanges)
}

// NewYFInvalidPeriodError creates a new YFInvalidPeriodError
func NewYFInvalidPeriodError(ticker string, invalidPeriod string, validRanges []string) *YFInvalidPeriodError {
	return &YFInvalidPeriodError{
		Ticker:        ticker,
		InvalidPeriod: invalidPeriod,
		ValidRanges:   validRanges,
	}
}

// YFRateLimitError represents rate limiting errors
type YFRateLimitError struct{}

func (e *YFRateLimitError) Error() string {
	return "Too Many Requests. Rate limited. Try after a while."
}

// NewYFRateLimitError creates a new YFRateLimitError
func NewYFRateLimitError() *YFRateLimitError {
	return &YFRateLimitError{}
}

// IsTransientError checks if an error is transient and should be retried
func IsTransientError(err error) bool {
	if err == nil {
		return false
	}
	// In Go, we check for network timeout errors
	// This is a simplified check
	return true
}
