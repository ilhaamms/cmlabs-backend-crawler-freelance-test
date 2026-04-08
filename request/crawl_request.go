package request

type CrawlRequest struct {
	URL          string `json:"url" binding:"required"`
	WaitSelector string `json:"wait_selector,omitempty"`
	Timeout      int    `json:"timeout,omitempty"`
}

type CrawlURLItem struct {
	URL          string `json:"url" binding:"required"`
	WaitSelector string `json:"wait_selector,omitempty"`
}

type CrawlBatchRequest struct {
	URLs    []CrawlURLItem `json:"urls" binding:"required,dive"`
	Timeout int            `json:"timeout,omitempty"`
}
