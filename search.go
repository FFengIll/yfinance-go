package yfinance

import (
	"context"
)

// Search performs a search on Yahoo Finance
type Search struct {
	Query        string
	MaxResults   int
	NewsCount    int
	ListsCount   int
	IncludeCB    bool
	EnableFuzzy  bool
	Recommended  int

	data        *YfData
	response    *searchResponse
}

// SearchOption is a functional option for Search
type SearchOption func(*Search)

// WithMaxResults sets the maximum number of results
func WithMaxResults(n int) SearchOption {
	return func(s *Search) {
		s.MaxResults = n
	}
}

// WithNewsCount sets the number of news articles
func WithNewsCount(n int) SearchOption {
	return func(s *Search) {
		s.NewsCount = n
	}
}

// WithListsCount sets the number of lists
func WithListsCount(n int) SearchOption {
	return func(s *Search) {
		s.ListsCount = n
	}
}

// WithFuzzyQuery enables fuzzy search
func WithFuzzyQuery(enable bool) SearchOption {
	return func(s *Search) {
		s.EnableFuzzy = enable
	}
}

// WithRecommended sets the number of recommended results
func WithRecommended(n int) SearchOption {
	return func(s *Search) {
		s.Recommended = n
	}
}

// NewSearch creates a new Search instance
func NewSearch(query string, opts ...SearchOption) *Search {
	s := &Search{
		Query:      query,
		MaxResults: 8,
		NewsCount:  8,
		ListsCount: 8,
		IncludeCB:  true,
		EnableFuzzy: false,
		Recommended: 8,
		data:       NewYfData(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Do executes the search
func (s *Search) Do(ctx context.Context) error {
	params := map[string]string{
		"q":                    s.Query,
		"quotesCount":          string(rune(s.MaxResults)),
		"newsCount":            string(rune(s.NewsCount)),
		"listsCount":           string(rune(s.ListsCount)),
		"enableCb":             "true",
		"enableFuzzyQuery":     boolToString(s.EnableFuzzy),
		"recommendedCount":     string(rune(s.Recommended)),
		"quotesQueryId":        "tss_match_phrase_query",
		"newsQueryId":          "news_cie_vespa",
	}

	endpoint := BaseURL + "/v1/finance/search"

	var result searchResponse
	if err := s.data.GetRawJSON(ctx, endpoint, params, &result); err != nil {
		return err
	}

	s.response = &result
	return nil
}

// SearchQuote represents a quote result from search
type SearchQuote struct {
	Symbol        string `json:"symbol"`
	ShortName     string `json:"shortname"`
	LongName      string `json:"longname"`
	Exchange      string `json:"exchange"`
	QuoteType     string `json:"quoteType"`
	Score         float64 `json:"score"`
	TypeDisp      string `json:"typeDisp"`
}

// SearchNews represents a news result from search
type SearchNews struct {
	UUID        string `json:"uuid"`
	Title       string `json:"title"`
	Publisher   string `json:"publisher"`
	Link        string `json:"link"`
	ProviderPublishTime int64 `json:"providerPublishTime"`
	Type        string `json:"type"`
	Thumbnail   *struct {
		Resolutions []struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
			Tag    string `json:"tag"`
		} `json:"resolutions"`
	} `json:"thumbnail,omitempty"`
}

// searchResponse represents the search API response
type searchResponse struct {
	Quotes []SearchQuote `json:"quotes"`
	News   []SearchNews  `json:"news"`
	Lists  []interface{} `json:"lists"`
}

// Quotes returns the quote results from the search
func (s *Search) Quotes() []SearchQuote {
	if s.response == nil {
		return nil
	}
	return s.response.Quotes
}

// News returns the news results from the search
func (s *Search) News() []SearchNews {
	if s.response == nil {
		return nil
	}
	return s.response.News
}

// Lists returns the list results from the search
func (s *Search) Lists() []interface{} {
	if s.response == nil {
		return nil
	}
	return s.response.Lists
}

// Search performs a search and returns the results
func SearchSymbols(ctx context.Context, query string, opts ...SearchOption) ([]SearchQuote, error) {
	s := NewSearch(query, opts...)
	if err := s.Do(ctx); err != nil {
		return nil, err
	}
	return s.Quotes(), nil
}

// Lookup performs a lookup for a specific symbol
type Lookup struct {
	Query    string
	Type     string // "all", "equity", "mutualfund", "etf"
	data     *YfData
}

// NewLookup creates a new Lookup instance
func NewLookup(query string) *Lookup {
	return &Lookup{
		Query: query,
		Type:  "all",
		data:  NewYfData(),
	}
}

// Do executes the lookup
func (l *Lookup) Do(ctx context.Context) ([]SearchQuote, error) {
	s := NewSearch(l.Query, WithMaxResults(20))
	if err := s.Do(ctx); err != nil {
		return nil, err
	}

	quotes := s.Quotes()
	if l.Type == "all" {
		return quotes, nil
	}

	// Filter by type
	filtered := make([]SearchQuote, 0)
	for _, q := range quotes {
		if q.QuoteType == l.Type {
			filtered = append(filtered, q)
		}
	}
	return filtered, nil
}

// Helper function to convert bool to string
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
