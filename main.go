package main

import (
	"fmt"
	"log"

	"github.com/ilhaamms/crawler-website/api"
	"github.com/ilhaamms/crawler-website/config"
	"github.com/ilhaamms/crawler-website/controller"
	"github.com/ilhaamms/crawler-website/repository"
	"github.com/ilhaamms/crawler-website/request"
	"github.com/ilhaamms/crawler-website/service"
)

func main() {
	cfg := config.NewConfig()
	log.Println("=== Website Crawler API ===")
	log.Printf("Port: %s", cfg.Port)
	log.Printf("Crawled Pages Dir: %s", cfg.CrawledPagesDir)
	log.Printf("Chrome Timeout: %v", cfg.ChromeTimeout)

	crawlRepo := repository.NewCrawlRepository(cfg)

	crawlService := service.NewCrawlService(crawlRepo, cfg)

	autoCrawl(crawlService)

	crawlController := controller.NewCrawlController(crawlService, crawlRepo, cfg.CrawledPagesDir)

	router := api.SetupRouter(crawlController)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server berjalan di http://localhost%s", addr)
	log.Printf("API Endpoints:")
	log.Printf("  POST /api/crawl         - Crawl satu URL")
	log.Printf("  POST /api/crawl/batch   - Crawl banyak URL")
	log.Printf("  GET  /api/crawl/files   - List file hasil crawl")
	log.Printf("  GET  /api/crawl/files/:filename - Download file HTML")

	if err := router.Run(addr); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}

func autoCrawl(crawlService service.CrawlService) {
	urls := []struct {
		URL          string
		WaitSelector string
	}{
		{URL: "https://cmlabs.co", WaitSelector: "body"},
		{URL: "https://sequence.day", WaitSelector: "body"},
		{URL: "https://tokopedia.com", WaitSelector: "body"},
	}

	log.Println("")
	log.Println("=== Auto-Crawling 3 Website ===")

	for i, u := range urls {
		log.Printf("[%d/3] Crawling %s ...", i+1, u.URL)

		req := &request.CrawlRequest{
			URL:          u.URL,
			WaitSelector: u.WaitSelector,
			Timeout:      60,
		}

		result, err := crawlService.CrawlURL(req)
		if err != nil {
			log.Printf("[%d/3] GAGAL crawl %s: %v", i+1, u.URL, err)
			continue
		}

		log.Printf("[%d/3] BERHASIL crawl %s", i+1, u.URL)
		log.Printf("       Title: %s", result.Title)
		log.Printf("       File: %s", result.FilePath)
		log.Printf("       Size: %d bytes", result.ContentLength)
		log.Printf("       Method: %s", result.CrawlMethod)
	}

	log.Println("=== Auto-Crawling Selesai ===")
	log.Println("")
}
