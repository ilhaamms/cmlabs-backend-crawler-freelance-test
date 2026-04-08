package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"

	"github.com/ilhaamms/crawler-website/config"
	"github.com/ilhaamms/crawler-website/model"
	"github.com/ilhaamms/crawler-website/repository"
	"github.com/ilhaamms/crawler-website/request"
)

type CrawlService interface {
	CrawlURL(req *request.CrawlRequest) (*model.CrawlResult, error)
	CrawlBatch(req *request.CrawlBatchRequest) ([]*model.CrawlResult, []CrawlErr)
}

type CrawlErr struct {
	URL   string
	Error string
}

type CrawlServiceImpl struct {
	repo   repository.CrawlRepository
	config *config.Config
}

func NewCrawlService(repo repository.CrawlRepository, cfg *config.Config) CrawlService {
	return &CrawlServiceImpl{
		repo:   repo,
		config: cfg,
	}
}

func (s *CrawlServiceImpl) CrawlURL(req *request.CrawlRequest) (*model.CrawlResult, error) {
	timeout := s.config.ChromeTimeout
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	var htmlContent string
	var title string

	waitSelector := req.WaitSelector
	if waitSelector == "" {
		waitSelector = "body"
	}

	log.Printf("[CRAWL] Mulai crawling URL: %s (method: chromedp, timeout: %v)", req.URL, timeout)

	err := chromedp.Run(ctx,
		chromedp.Navigate(req.URL),

		chromedp.WaitReady(waitSelector, chromedp.ByQuery),

		chromedp.Sleep(3*time.Second),

		chromedp.Title(&title),

		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(`document.documentElement.outerHTML`, &htmlContent).Do(ctx)
		}),
	)

	if err != nil {
		log.Printf("[CRAWL] Chromedp gagal untuk %s: %v, mencoba fallback HTTP...", req.URL, err)
		return s.crawlWithHTTP(req.URL)
	}

	if !strings.HasPrefix(strings.TrimSpace(strings.ToLower(htmlContent)), "<!doctype") {
		htmlContent = "<!DOCTYPE html>\n" + htmlContent
	}

	result := &model.CrawlResult{
		URL:           req.URL,
		Title:         title,
		HTMLContent:   htmlContent,
		CrawledAt:     time.Now(),
		StatusCode:    200,
		ContentLength: int64(len(htmlContent)),
		CrawlMethod:   "chromedp",
	}

	filePath, err := s.repo.SaveHTML(result)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan hasil crawl: %w", err)
	}
	result.FilePath = filePath

	log.Printf("[CRAWL] Berhasil crawl URL: %s (title: %s, size: %d bytes, file: %s)",
		req.URL, title, result.ContentLength, filePath)

	return result, nil
}

func (s *CrawlServiceImpl) crawlWithHTTP(targetURL string) (*model.CrawlResult, error) {
	log.Printf("[CRAWL] Fallback HTTP untuk URL: %s", targetURL)

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("terlalu banyak redirect")
			}
			return nil
		},
	}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal melakukan request ke %s: %w", targetURL, err)
	}
	defer resp.Body.Close()

	bodyBytes := make([]byte, 0)
	buf := make([]byte, 32*1024)
	maxSize := int64(50 * 1024 * 1024)
	var totalRead int64

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			totalRead += int64(n)
			if totalRead > maxSize {
				break
			}
			bodyBytes = append(bodyBytes, buf[:n]...)
		}
		if readErr != nil {
			break
		}
	}

	htmlContent := string(bodyBytes)

	title := extractTitle(htmlContent)

	result := &model.CrawlResult{
		URL:           targetURL,
		Title:         title,
		HTMLContent:   htmlContent,
		CrawledAt:     time.Now(),
		StatusCode:    resp.StatusCode,
		ContentLength: int64(len(htmlContent)),
		CrawlMethod:   "http",
	}

	filePath, err := s.repo.SaveHTML(result)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan hasil crawl: %w", err)
	}
	result.FilePath = filePath

	log.Printf("[CRAWL] Berhasil crawl URL via HTTP: %s (title: %s, size: %d bytes)",
		targetURL, title, result.ContentLength)

	return result, nil
}

func (s *CrawlServiceImpl) CrawlBatch(req *request.CrawlBatchRequest) ([]*model.CrawlResult, []CrawlErr) {
	var (
		results []*model.CrawlResult
		errors  []CrawlErr
		mu      sync.Mutex
		wg      sync.WaitGroup
	)

	semaphore := make(chan struct{}, 3)

	for _, item := range req.URLs {
		wg.Add(1)

		go func(urlItem request.CrawlURLItem) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			crawlReq := &request.CrawlRequest{
				URL:          urlItem.URL,
				WaitSelector: urlItem.WaitSelector,
				Timeout:      req.Timeout,
			}

			result, err := s.CrawlURL(crawlReq)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errors = append(errors, CrawlErr{
					URL:   urlItem.URL,
					Error: err.Error(),
				})
			} else {
				results = append(results, result)
			}
		}(item)
	}

	wg.Wait()
	return results, errors
}

func extractTitle(html string) string {
	lower := strings.ToLower(html)

	startIdx := strings.Index(lower, "<title")
	if startIdx == -1 {
		return ""
	}

	closeOpenTag := strings.Index(lower[startIdx:], ">")
	if closeOpenTag == -1 {
		return ""
	}
	contentStart := startIdx + closeOpenTag + 1

	endIdx := strings.Index(lower[contentStart:], "</title>")
	if endIdx == -1 {
		return ""
	}

	title := strings.TrimSpace(html[contentStart : contentStart+endIdx])
	return title
}
