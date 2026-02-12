// Package yfinance provides a Go implementation of Yahoo Finance API client
// based on the Python yfinance library
package yfinance

// Base URLs for Yahoo Finance API
const (
	Query1URL = "https://query1.finance.yahoo.com"
	BaseURL   = "https://query2.finance.yahoo.com"
	RootURL   = "https://finance.yahoo.com"
)

// Valid periods for history
var ValidPeriods = []string{
	"1d", "5d", "1mo", "3mo", "6mo", "1y", "2y", "5y", "10y", "ytd", "max",
}

// Valid intervals for history
var ValidIntervals = []string{
	"1m", "2m", "5m", "15m", "30m", "60m", "90m", "1h", "1d", "5d", "1wk", "1mo", "3mo",
}

// Intraday intervals (cannot extend past 60 days)
var IntradayIntervals = map[string]bool{
	"1m":  true,
	"2m":  true,
	"5m":  true,
	"15m": true,
	"30m": true,
	"60m": true,
	"90m": true,
	"1h":  true,
}

// PriceColumnNames for history data
var PriceColumnNames = []string{
	"Open", "High", "Low", "Close", "Adj Close", "Volume",
}

// QuoteSummaryValidModules defines valid modules for quote summary
var QuoteSummaryValidModules = []string{
	"summaryProfile",
	"summaryDetail",
	"assetProfile",
	"fundProfile",
	"price",
	"quoteType",
	"esgScores",
	"incomeStatementHistory",
	"incomeStatementHistoryQuarterly",
	"balanceSheetHistory",
	"balanceSheetHistoryQuarterly",
	"cashFlowStatementHistory",
	"cashFlowStatementHistoryQuarterly",
	"defaultKeyStatistics",
	"financialData",
	"calendarEvents",
	"secFilings",
	"upgradeDowngradeHistory",
	"institutionOwnership",
	"fundOwnership",
	"majorDirectHolders",
	"majorHoldersBreakdown",
	"insiderTransactions",
	"insiderHolders",
	"netSharePurchaseActivity",
	"earnings",
	"earningsHistory",
	"earningsTrend",
	"industryTrend",
	"indexTrend",
	"sectorTrend",
	"recommendationTrend",
	"futuresChain",
}

// MICToYahooSuffix maps Market Identifier Codes to Yahoo Finance suffixes
var MICToYahooSuffix = map[string]string{
	"XCBT": "CBT", "XCME": "CME", "IFUS": "NYB", "CECS": "CMX", "XNYM": "NYM", "XNYS": "", "XNAS": "", // United States
	"XBUE": "BA",   // Argentina
	"XVIE": "VI",   // Austria
	"XASX": "AX",   // Australia
	"XAUS": "XA",   // Australia
	"XBRU": "BR",   // Belgium
	"BVMF": "SA",   // Brazil
	"CNSX": "CN",   // Canada
	"NEOE": "NE",   // Canada
	"XTSE": "TO",   // Canada
	"XTSX": "V",    // Canada
	"XSGO": "SN",   // Chile
	"XSHG": "SS",   // China
	"XSHE": "SZ",   // China
	"XBOG": "CL",   // Colombia
	"XPRA": "PR",   // Czech Republic
	"XCSE": "CO",   // Denmark
	"XCAI": "CA",   // Egypt
	"XTAL": "TL",   // Estonia
	"CEUX": "XD",   // Europe
	"XEUR": "NX",   // Europe
	"XHEL": "HE",   // Finland
	"XPAR": "PA",   // France
	"XBER": "BE",   // Germany
	"XBMS": "BM",   // Germany
	"XDUS": "DU",   // Germany
	"XFRA": "F",    // Germany
	"XHAM": "HM",   // Germany
	"XHAN": "HA",   // Germany
	"XMUN": "MU",   // Germany
	"XSTU": "SG",   // Germany
	"XETR": "DE",   // Germany
	"XATH": "AT",   // Greece
	"XHKG": "HK",   // Hong Kong
	"XBUD": "BD",   // Hungary
	"XICE": "IC",   // Iceland
	"XBOM": "BO",   // India
	"XNSE": "NS",   // India
	"XIDX": "JK",   // Indonesia
	"XDUB": "IR",   // Ireland
	"XTAE": "TA",   // Israel
	"MTAA": "MI",   // Italy
	"EUTL": "TI",   // Italy
	"XTKS": "T",    // Japan
	"XKFE": "KW",   // Kuwait
	"XRIS": "RG",   // Latvia
	"XVIL": "VS",   // Lithuania
	"XKLS": "KL",   // Malaysia
	"XMEX": "MX",   // Mexico
	"XAMS": "AS",   // Netherlands
	"XNZE": "NZ",   // New Zealand
	"XOSL": "OL",   // Norway
	"XPHS": "PS",   // Philippines
	"XWAR": "WA",   // Poland
	"XLIS": "LS",   // Portugal
	"XQAT": "QA",   // Qatar
	"XBSE": "RO",   // Romania
	"XSES": "SI",   // Singapore
	"XJSE": "JO",   // South Africa
	"XKRX": "KS",   // South Korea
	"KQKS": "KQ",   // South Korea
	"BMEX": "MC",   // Spain
	"XSAU": "SR",   // Saudi Arabia
	"XSTO": "ST",   // Sweden
	"XSWX": "SW",   // Switzerland
	"ROCO": "TWO",  // Taiwan
	"XTAI": "TW",   // Taiwan
	"XBKK": "BK",   // Thailand
	"XIST": "IS",   // Turkey
	"XDFM": "AE",   // UAE
	"AQXE": "AQ",   // UK
	"XCHI": "XC",   // UK
	"XLON": "L",    // UK
	"ILSE": "IL",   // UK
	"XCAR": "CR",   // Venezuela
	"XSTC": "VN",   // Vietnam
}

// UserAgents for HTTP requests
var UserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14.7; rv:135.0) Gecko/20100101 Firefox/135.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_7_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3 Safari/605.1.15",
}
