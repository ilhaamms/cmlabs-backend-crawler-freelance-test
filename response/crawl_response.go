package response

import "time"

type ApiResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CrawlResponse struct {
	URL           string    `json:"url"`
	Title         string    `json:"title"`
	FilePath      string    `json:"file_path"`
	CrawledAt     time.Time `json:"crawled_at"`
	StatusCode    int       `json:"status_code"`
	ContentLength int64     `json:"content_length"`
	CrawlMethod   string    `json:"crawl_method"`
}

type CrawlBatchResponse struct {
	TotalRequested int             `json:"total_requested"`
	TotalSuccess   int             `json:"total_success"`
	TotalFailed    int             `json:"total_failed"`
	Results        []CrawlResponse `json:"results"`
	Errors         []CrawlError    `json:"errors,omitempty"`
}

type CrawlError struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

type FileInfo struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Path     string `json:"path"`
}
