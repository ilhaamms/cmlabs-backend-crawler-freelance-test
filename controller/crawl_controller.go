package controller

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/ilhaamms/crawler-website/helper"
	"github.com/ilhaamms/crawler-website/repository"
	"github.com/ilhaamms/crawler-website/request"
	"github.com/ilhaamms/crawler-website/response"
	"github.com/ilhaamms/crawler-website/service"
)

type CrawlController struct {
	crawlService service.CrawlService
	crawlRepo    repository.CrawlRepository
	crawledDir   string
}

func NewCrawlController(crawlService service.CrawlService, crawlRepo repository.CrawlRepository, crawledDir string) *CrawlController {
	return &CrawlController{
		crawlService: crawlService,
		crawlRepo:    crawlRepo,
		crawledDir:   crawledDir,
	}
}

func (ctrl *CrawlController) CrawlSingle(c *gin.Context) {
	var req request.CrawlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ApiResponse{
			Status:  "error",
			Message: "Request tidak valid: " + err.Error(),
		})
		return
	}

	req.URL = helper.EnsureScheme(req.URL)

	if !helper.ValidateURL(req.URL) {
		c.JSON(http.StatusBadRequest, response.ApiResponse{
			Status:  "error",
			Message: "URL tidak valid: " + req.URL,
		})
		return
	}

	result, err := ctrl.crawlService.CrawlURL(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ApiResponse{
			Status:  "error",
			Message: "Gagal melakukan crawl: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.ApiResponse{
		Status:  "success",
		Message: "Website berhasil di-crawl",
		Data: response.CrawlResponse{
			URL:           result.URL,
			Title:         result.Title,
			FilePath:      result.FilePath,
			CrawledAt:     result.CrawledAt,
			StatusCode:    result.StatusCode,
			ContentLength: result.ContentLength,
			CrawlMethod:   result.CrawlMethod,
		},
	})
}

func (ctrl *CrawlController) CrawlBatch(c *gin.Context) {
	var req request.CrawlBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ApiResponse{
			Status:  "error",
			Message: "Request tidak valid: " + err.Error(),
		})
		return
	}

	if len(req.URLs) == 0 {
		c.JSON(http.StatusBadRequest, response.ApiResponse{
			Status:  "error",
			Message: "Minimal satu URL harus disertakan",
		})
		return
	}

	for i := range req.URLs {
		req.URLs[i].URL = helper.EnsureScheme(req.URLs[i].URL)
	}

	results, crawlErrors := ctrl.crawlService.CrawlBatch(&req)

	var responseResults []response.CrawlResponse
	for _, r := range results {
		responseResults = append(responseResults, response.CrawlResponse{
			URL:           r.URL,
			Title:         r.Title,
			FilePath:      r.FilePath,
			CrawledAt:     r.CrawledAt,
			StatusCode:    r.StatusCode,
			ContentLength: r.ContentLength,
			CrawlMethod:   r.CrawlMethod,
		})
	}

	var responseErrors []response.CrawlError
	for _, e := range crawlErrors {
		responseErrors = append(responseErrors, response.CrawlError{
			URL:   e.URL,
			Error: e.Error,
		})
	}

	batchResp := response.CrawlBatchResponse{
		TotalRequested: len(req.URLs),
		TotalSuccess:   len(results),
		TotalFailed:    len(crawlErrors),
		Results:        responseResults,
		Errors:         responseErrors,
	}

	statusCode := http.StatusOK
	message := "Semua website berhasil di-crawl"
	if len(crawlErrors) > 0 && len(results) > 0 {
		statusCode = http.StatusPartialContent
		message = "Beberapa website gagal di-crawl"
	} else if len(crawlErrors) > 0 && len(results) == 0 {
		statusCode = http.StatusInternalServerError
		message = "Semua website gagal di-crawl"
	}

	c.JSON(statusCode, response.ApiResponse{
		Status:  "success",
		Message: message,
		Data:    batchResp,
	})
}

func (ctrl *CrawlController) ListFiles(c *gin.Context) {
	files, err := ctrl.crawlRepo.ListCrawledFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ApiResponse{
			Status:  "error",
			Message: "Gagal membaca daftar file: " + err.Error(),
		})
		return
	}

	var fileInfos []response.FileInfo
	for _, f := range files {
		fileInfos = append(fileInfos, response.FileInfo{
			Filename: f.Filename,
			Size:     f.Size,
			Path:     f.Path,
		})
	}

	c.JSON(http.StatusOK, response.ApiResponse{
		Status:  "success",
		Message: "Daftar file hasil crawl",
		Data:    fileInfos,
	})
}

func (ctrl *CrawlController) GetFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, response.ApiResponse{
			Status:  "error",
			Message: "Nama file harus disertakan",
		})
		return
	}

	filename = filepath.Base(filename)

	c.File(filepath.Join(ctrl.crawledDir, filename))
}
