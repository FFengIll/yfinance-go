package yfinance

import (
	"context"
	"fmt"
	"time"
)

// Info contains detailed information about a ticker
type Info struct {
	Symbol          string                 `json:"symbol"`
	ShortName       string                 `json:"shortName"`
	LongName        string                 `json:"longName"`
	Exchange        string                 `json:"exchange"`
	Market          string                 `json:"market"`
	QuoteType       string                 `json:"quoteType"`
	Currency        string                 `json:"currency"`
	Sector          string                 `json:"sector"`
	Industry        string                 `json:"industry"`
	Country         string                 `json:"country"`
	State           string                 `json:"state"`
	City            string                 `json:"city"`
	Address         string                 `json:"address"`
	Website         string                 `json:"website"`
	Phone           string                 `json:"phone"`
	FullTimeEmployees int64                `json:"fullTimeEmployees"`
	Overview        string                 `json:"longBusinessSummary"`

	// Officers and management
	CompanyOfficers []CompanyOfficer       `json:"companyOfficers"`

	// Price data
	CurrentPrice    float64                `json:"currentPrice"`
	PreviousClose   float64                `json:"previousClose"`
	Open            float64                `json:"open"`
	DayLow          float64                `json:"dayLow"`
	DayHigh         float64                `json:"dayHigh"`
	Volume          int64                  `json:"volume"`
	AvgVolume       int64                  `json:"averageVolume"`

	// Market data
	MarketCap       int64                  `json:"marketCap"`
	SharesOutstanding int64                `json:"sharesOutstanding"`
	FloatShares     int64                  `json:"floatShares"`
	Beta            float64                `json:"beta"`

	// Valuation metrics
	TrailingPE      float64                `json:"trailingPE"`
	ForwardPE       float64                `json:"forwardPE"`
	PEGRatio        float64                `json:"pegRatio"`
	PriceToBook     float64                `json:"priceToBook"`
	PriceToSales    float64                `json:"priceToSalesTrailing12Months"`
	EnterpriseValue int64                  `json:"enterpriseValue"`
	EnterpriseToRevenue float64            `json:"enterpriseToRevenue"`
	EnterpriseToEbitda float64             `json:"enterpriseToEbitda"`

	// Profitability
	ProfitMargin    float64                `json:"profitMargins"`
	OperatingMargin float64                `json:"operatingMargins"`
	ROE             float64                `json:"returnOnEquity"`
	ROA             float64                `json:"returnOnAssets"`
	GrossMargin     float64                `json:"grossMargins"`
	EBITDA          int64                  `json:"ebitda"`
	Revenue         int64                  `json:"totalRevenue"`
	RevenueGrowth   float64                `json:"revenueGrowth"`
	NetIncome       int64                  `json:"netIncomeToCommon"`

	// Dividend data
	DividendRate    float64                `json:"dividendRate"`
	DividendYield   float64                `json:"dividendYield"`
	ExDividendDate  time.Time              `json:"exDividendDate"`
	PayoutRatio     float64                `json:"payoutRatio"`

	// 52-week data
	FiftyTwoWeekLow  float64               `json:"fiftyTwoWeekLow"`
	FiftyTwoWeekHigh float64               `json:"fiftyTwoWeekHigh"`
	FiftyDayAverage  float64               `json:"fiftyDayAverage"`
	TwoHundredDayAverage float64           `json:"twoHundredDayAverage"`

	// Additional fields
	BookValue       float64                `json:"bookValue"`
	EPS             float64                `json:"trailingEps"`
	ForwardEPS      float64                `json:"forwardEps"`
	TargetHighPrice float64                `json:"targetHighPrice"`
	TargetLowPrice  float64                `json:"targetLowPrice"`
	TargetMeanPrice float64                `json:"targetMeanPrice"`
	NumberOfAnalystOpinions int64          `json:"numberOfAnalystOpinions"`

	// Raw response
	Raw             map[string]interface{} `json:"-"`
}

// CompanyOfficer represents a company officer/director
type CompanyOfficer struct {
	Name        string `json:"name"`
	Age         int    `json:"age"`
	Title       string `json:"title"`
	YearBorn    int    `json:"yearBorn"`
	TotalPay    string `json:"totalPay"`
	ExercisedValue string `json:"exercisedValue"`
	UnexercisedValue string `json:"unexercisedValue"`
}

// GetInfo fetches detailed information about the ticker
func (t *Ticker) GetInfo(ctx context.Context) (*Info, error) {
	// Fetch modules from quote summary
	modules := []string{
		"summaryProfile",
		"summaryDetail",
		"assetProfile",
		"price",
		"quoteType",
		"defaultKeyStatistics",
		"financialData",
	}

	endpoint := fmt.Sprintf("%s/v10/finance/quoteSummary/%s", BaseURL, t.Symbol)
	params := map[string]string{
		"modules": joinModules(modules),
	}

	var result quoteSummaryResponse
	if err := t.data.GetRawJSON(ctx, endpoint, params, &result); err != nil {
		return nil, err
	}

	if result.QuoteSummary.Error != nil {
		return nil, fmt.Errorf("quote summary error: %s", result.QuoteSummary.Error.Description)
	}

	return parseInfo(&result.QuoteSummary.Result[0]), nil
}

// quoteSummaryResponse represents the quote summary API response
type quoteSummaryResponse struct {
	QuoteSummary struct {
		Result []quoteSummaryResult `json:"result"`
		Error  *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error,omitempty"`
	} `json:"quoteSummary"`
}

// quoteSummaryResult represents a single quote summary result
type quoteSummaryResult struct {
	SummaryProfile   interface{} `json:"summaryProfile"`
	SummaryDetail    interface{} `json:"summaryDetail"`
	AssetProfile     interface{} `json:"assetProfile"`
	Price            interface{} `json:"price"`
	QuoteType        interface{} `json:"quoteType"`
	DefaultKeyStatistics interface{} `json:"defaultKeyStatistics"`
	FinancialData    interface{} `json:"financialData"`
}

// parseInfo parses the quote summary result into Info
func parseInfo(result *quoteSummaryResult) *Info {
	info := &Info{
		Raw: make(map[string]interface{}),
	}

	// Parse summary profile
	if sp, ok := result.SummaryProfile.(map[string]interface{}); ok {
		info.Symbol = getString(sp, "symbol")
		info.ShortName = getString(sp, "shortName")
		info.LongName = getString(sp, "longName")
		info.Sector = getString(sp, "sector")
		info.Industry = getString(sp, "industry")
		info.Country = getString(sp, "country")
		info.State = getString(sp, "state")
		info.City = getString(sp, "city")
		info.Address = getString(sp, "address1")
		info.Website = getString(sp, "website")
		info.Phone = getString(sp, "phone")
		info.Overview = getString(sp, "longBusinessSummary")
		info.FullTimeEmployees = getInt64(sp, "fullTimeEmployees")

		// Parse company officers
		if officers, ok := sp["companyOfficers"].([]interface{}); ok {
			info.CompanyOfficers = make([]CompanyOfficer, 0, len(officers))
			for _, o := range officers {
				if officer, ok := o.(map[string]interface{}); ok {
					info.CompanyOfficers = append(info.CompanyOfficers, CompanyOfficer{
						Name:     getString(officer, "name"),
						Title:    getString(officer, "title"),
						Age:      getInt(officer, "age"),
						YearBorn: getInt(officer, "yearBorn"),
					})
				}
			}
		}
	}

	// Parse summary detail
	if sd, ok := result.SummaryDetail.(map[string]interface{}); ok {
		info.Currency = getString(sd, "currency")
		info.PreviousClose = getFloat64(sd, "previousClose")
		info.Open = getFloat64(sd, "open")
		info.DayLow = getFloat64(sd, "dayLow")
		info.DayHigh = getFloat64(sd, "dayHigh")
		info.Volume = getInt64(sd, "volume")
		info.AvgVolume = getInt64(sd, "averageVolume")
		info.MarketCap = getInt64(sd, "marketCap")
		info.FiftyTwoWeekLow = getFloat64(sd, "fiftyTwoWeekLow")
		info.FiftyTwoWeekHigh = getFloat64(sd, "fiftyTwoWeekHigh")
		info.FiftyDayAverage = getFloat64(sd, "fiftyDayAverage")
		info.TwoHundredDayAverage = getFloat64(sd, "twoHundredDayAverage")
		info.DividendRate = getFloat64(sd, "dividendRate")
		info.DividendYield = getFloat64(sd, "dividendYield")
		info.ExDividendDate = getTime(sd, "exDividendDate")
		info.PayoutRatio = getFloat64(sd, "payoutRatio")
		info.Beta = getFloat64(sd, "beta")
		info.TrailingPE = getFloat64(sd, "trailingPE")
		info.ForwardPE = getFloat64(sd, "forwardPE")
		info.PriceToSales = getFloat64(sd, "priceToSalesTrailing12Months")
	}

	// Parse price
	if pr, ok := result.Price.(map[string]interface{}); ok {
		info.Symbol = getString(pr, "symbol")
		info.ShortName = getString(pr, "shortName")
		info.LongName = getString(pr, "longName")
		info.Exchange = getString(pr, "exchangeName")
		info.Market = getString(pr, "market")
		info.QuoteType = getString(pr, "quoteType")
		info.Currency = getString(pr, "currency")
		info.CurrentPrice = getFloat64(pr, "regularMarketPrice")
		info.MarketCap = getInt64(pr, "marketCap")
	}

	// Parse financial data
	if fd, ok := result.FinancialData.(map[string]interface{}); ok {
		info.CurrentPrice = getFloat64(fd, "currentPrice")
		info.TargetHighPrice = getFloat64(fd, "targetHighPrice")
		info.TargetLowPrice = getFloat64(fd, "targetLowPrice")
		info.TargetMeanPrice = getFloat64(fd, "targetMeanPrice")
		info.NumberOfAnalystOpinions = getInt64(fd, "numberOfAnalystOpinions")
		info.Revenue = getInt64(fd, "totalRevenue")
		info.RevenueGrowth = getFloat64(fd, "revenueGrowth")
		info.GrossMargin = getFloat64(fd, "grossMargins")
		info.EBITDA = getInt64(fd, "ebitda")
		info.OperatingMargin = getFloat64(fd, "operatingMargins")
		info.ProfitMargin = getFloat64(fd, "profitMargins")
		info.ROE = getFloat64(fd, "returnOnEquity")
		info.ROA = getFloat64(fd, "returnOnAssets")
		info.Revenue = getInt64(fd, "totalRevenue")
		info.NetIncome = getInt64(fd, "netIncomeToCommon")
		info.EnterpriseValue = getInt64(fd, "enterpriseValue")
		info.EnterpriseToRevenue = getFloat64(fd, "enterpriseToRevenue")
		info.EnterpriseToEbitda = getFloat64(fd, "enterpriseToEbitda")
	}

	// Parse default key statistics
	if ks, ok := result.DefaultKeyStatistics.(map[string]interface{}); ok {
		info.EnterpriseValue = getInt64(ks, "enterpriseValue")
		info.ProfitMargin = getFloat64(ks, "profitMargins")
		info.FloatShares = getInt64(ks, "floatShares")
		info.SharesOutstanding = getInt64(ks, "sharesOutstanding")
		info.Beta = getFloat64(ks, "beta")
		info.BookValue = getFloat64(ks, "bookValue")
		info.PriceToBook = getFloat64(ks, "priceToBook")
		info.EPS = getFloat64(ks, "trailingEps")
		info.ForwardEPS = getFloat64(ks, "forwardEps")
		info.PEGRatio = getFloat64(ks, "pegRatio")
	}

	return info
}

// News represents a news article
type News struct {
	UUID             string    `json:"uuid"`
	Title            string    `json:"title"`
	Publisher        string    `json:"publisher"`
	Link             string    `json:"link"`
	ProviderPublishTime time.Time `json:"providerPublishTime"`
	Type             string    `json:"type"`
	ThumbnailURL     string    `json:"thumbnailUrl"`
	Summary          string    `json:"summary"`
}

// GetNews fetches news for the ticker
func (t *Ticker) GetNews(ctx context.Context, count int) ([]News, error) {
	if count <= 0 {
		count = 10
	}

	// Use the XHR endpoint for news
	endpoint := fmt.Sprintf("%s/xhr/ncp", RootURL)
	params := map[string]string{
		"queryRef":   "latestNews",
		"serviceKey": "ncp_fin",
	}

	body := map[string]interface{}{
		"serviceConfig": map[string]interface{}{
			"snippetCount": count,
			"s":           []string{t.Symbol},
		},
	}

	var result newsResponse
	resp, err := t.data.Post(ctx, endpoint, params, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := parseJSONResponse(resp.Body, &result); err != nil {
		return nil, err
	}

	// Parse news
	news := make([]News, 0)
	if result.Data != nil && result.Data.TickerStream != nil {
		for _, item := range result.Data.TickerStream.Stream {
			// Skip ads
			if len(item.Ad) > 0 {
				continue
			}

			n := News{
				UUID:              item.UUID,
				Title:             item.Title,
				Publisher:         item.Publisher,
				Link:              item.LinkURL,
				Type:              item.ContentType,
				Summary:           item.Summary,
			}

			if item.PubTime > 0 {
				n.ProviderPublishTime = time.Unix(item.PubTime, 0)
			}

			if len(item.Thumbnails) > 0 && len(item.Thumbnails[0].URL) > 0 {
				n.ThumbnailURL = item.Thumbnails[0].URL[0].URL
			}

			news = append(news, n)
		}
	}

	return news, nil
}

// newsResponse represents the news API response
type newsResponse struct {
	Data *struct {
		TickerStream *struct {
			Stream []newsItem `json:"stream"`
		} `json:"tickerStream"`
	} `json:"data"`
}

type newsItem struct {
	UUID        string `json:"uuid"`
	Title       string `json:"title"`
	Publisher   string `json:"publisher"`
	LinkURL     string `json:"linkUrl"`
	PubTime     int64  `json:"pubTime"`
	ContentType string `json:"contentType"`
	Summary     string `json:"summary"`
	Ad          []interface{} `json:"ad"`
	Thumbnails  []struct {
		URL []struct {
			URL string `json:"url"`
		} `json:"url"`
	} `json:"thumbnails"`
}

// Calendar represents calendar events for a ticker
type Calendar struct {
	Earnings struct {
		Date      time.Time `json:"date"`
		EpsEstimate float64 `json:"epsEstimate"`
	} `json:"earnings"`
	Dividends []DividendData `json:"dividends"`
	Splits    []SplitData    `json:"splits"`
}

// GetCalendar fetches calendar events for the ticker
func (t *Ticker) GetCalendar(ctx context.Context) (*Calendar, error) {
	endpoint := fmt.Sprintf("%s/v10/finance/quoteSummary/%s", BaseURL, t.Symbol)
	params := map[string]string{
		"modules": "calendarEvents",
	}

	var result struct {
		QuoteSummary struct {
			Result []struct {
				CalendarEvents struct {
					Earnings struct {
						Date      []int64 `json:"earningsDate"`
						EpsEstimate float64 `json:"epsAverage"`
					} `json:"earnings"`
					Dividends struct {
						Rows []struct {
							Date   string  `json:"date"`
							Amount float64 `json:"amount"`
						} `json:"rows"`
					} `json:"dividends"`
					Splits struct {
						Rows []struct {
							Date  string `json:"date"`
							Ratio string `json:"ratio"`
						} `json:"rows"`
					} `json:"splits"`
				} `json:"calendarEvents"`
			} `json:"result"`
		} `json:"quoteSummary"`
	}

	if err := t.data.GetRawJSON(ctx, endpoint, params, &result); err != nil {
		return nil, err
	}

	if len(result.QuoteSummary.Result) == 0 {
		return &Calendar{}, nil
	}

	calendar := &Calendar{}
	ce := result.QuoteSummary.Result[0].CalendarEvents

	// Parse earnings
	if len(ce.Earnings.Date) > 0 {
		calendar.Earnings.Date = time.Unix(ce.Earnings.Date[0], 0)
		calendar.Earnings.EpsEstimate = ce.Earnings.EpsEstimate
	}

	// Parse dividends
	for _, row := range ce.Dividends.Rows {
		calendar.Dividends = append(calendar.Dividends, DividendData{
			Amount: row.Amount,
		})
	}

	// Parse splits
	for _, row := range ce.Splits.Rows {
		calendar.Splits = append(calendar.Splits, SplitData{
			Ratio: row.Ratio,
		})
	}

	return calendar, nil
}

// Recommendation represents analyst recommendations
type Recommendation struct {
	Period      string  `json:"period"`
	StrongBuy   int     `json:"strongBuy"`
	Buy         int     `json:"buy"`
	Hold        int     `json:"hold"`
	Sell        int     `json:"sell"`
	StrongSell  int     `json:"strongSell"`
}

// GetRecommendations fetches analyst recommendations for the ticker
func (t *Ticker) GetRecommendations(ctx context.Context) ([]Recommendation, error) {
	endpoint := fmt.Sprintf("%s/v10/finance/quoteSummary/%s", BaseURL, t.Symbol)
	params := map[string]string{
		"modules": "recommendationTrend",
	}

	var result struct {
		QuoteSummary struct {
			Result []struct {
				RecommendationTrend struct {
					Trend []struct {
						Period     string `json:"period"`
						StrongBuy  int    `json:"strongBuy"`
						Buy        int    `json:"buy"`
						Hold       int    `json:"hold"`
						Sell       int    `json:"sell"`
						StrongSell int    `json:"strongSell"`
					} `json:"trend"`
				} `json:"recommendationTrend"`
			} `json:"result"`
		} `json:"quoteSummary"`
	}

	if err := t.data.GetRawJSON(ctx, endpoint, params, &result); err != nil {
		return nil, err
	}

	if len(result.QuoteSummary.Result) == 0 {
		return []Recommendation{}, nil
	}

	trend := result.QuoteSummary.Result[0].RecommendationTrend.Trend
	recommendations := make([]Recommendation, 0, len(trend))
	for _, t := range trend {
		recommendations = append(recommendations, Recommendation{
			Period:     t.Period,
			StrongBuy:  t.StrongBuy,
			Buy:        t.Buy,
			Hold:       t.Hold,
			Sell:       t.Sell,
			StrongSell: t.StrongSell,
		})
	}

	return recommendations, nil
}
