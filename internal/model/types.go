package model

// Query captures the inputs to `cnki search`.
type Query struct {
	Q       string   `json:"q"`
	Field   string   `json:"field"`
	From    int      `json:"from,omitempty"`
	To      int      `json:"to,omitempty"`
	Types   []string `json:"types,omitempty"`
	Sources []string `json:"sources,omitempty"`
	Sort    string   `json:"sort"`
	Size    int      `json:"size"`
}

// Paper is a single result row on the search page.
type Paper struct {
	Seq       int      `json:"seq"`
	Title     string   `json:"title"`
	URL       string   `json:"url"`
	Authors   []string `json:"authors,omitempty"`
	Source    string   `json:"source,omitempty"`
	Year      int      `json:"year,omitempty"`
	Issue     string   `json:"issue,omitempty"`
	Cited     int      `json:"cited"`
	Downloads int      `json:"downloads"`
}

// SearchResult is the top-level payload for `cnki search`.
type SearchResult struct {
	Query     Query   `json:"query"`
	TotalHits int     `json:"total_hits"`
	Fetched   int     `json:"fetched"`
	Results   []Paper `json:"results"`
}

// Detail is the payload for `cnki detail`.
type Detail struct {
	URL          string      `json:"url"`
	Title        string      `json:"title,omitempty"`
	Authors      []string    `json:"authors,omitempty"`
	Institutions []string    `json:"institutions,omitempty"`
	Abstract     string      `json:"abstract,omitempty"`
	Keywords     []string    `json:"keywords,omitempty"`
	DOI          string      `json:"doi,omitempty"`
	CLC          string      `json:"clc,omitempty"`
	Source       string      `json:"source,omitempty"`
	Issue        string      `json:"issue,omitempty"`
	Year         int         `json:"year,omitempty"`
	Fund         string      `json:"fund,omitempty"`
	Cited        int         `json:"cited"`
	Downloads    int         `json:"downloads"`
	References   []Reference `json:"references,omitempty"`
}

// Reference is one item from a paper's 参考文献 list.
type Reference struct {
	Seq  int    `json:"seq"`
	Text string `json:"text"`
}
