package model

import "time"

type CrawlResult struct {
	URL           string    `json:"url"`
	Title         string    `json:"title"`
	HTMLContent   string    `json:"-"`
	FilePath      string    `json:"file_path"`
	CrawledAt     time.Time `json:"crawled_at"`
	StatusCode    int       `json:"status_code"`
	ContentLength int64     `json:"content_length"`
	CrawlMethod   string    `json:"crawl_method"`
}
