package models

type AnalysisResult struct {
	URL               string         `json:"url"`
	StatusCode        int            `json:"status_code"`
	ErrorMessage      string         `json:"error_message,omitempty"`
	HTMLVersion       string         `json:"html_version"`
	PageTitle         string         `json:"page_title"`
	Headings          map[string]int `json:"headings"`
	InternalLinks     int            `json:"internal_links"`
	ExternalLinks     int            `json:"external_links"`
	InaccessibleLinks int            `json:"inaccessible_links"`
	HasLoginForm      bool           `json:"has_login_form"`
	Links             []LinkInfo     `json:"links,omitempty"`
}

type LinkInfo struct {
	URL        string `json:"url"`
	Accessible bool   `json:"accessible"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error,omitempty"`
}

type AnalysisRequest struct {
	URL string `json:"url"`
}
