package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ilhaamms/crawler-website/config"
	"github.com/ilhaamms/crawler-website/helper"
	"github.com/ilhaamms/crawler-website/model"
)

type CrawlRepository interface {
	SaveHTML(result *model.CrawlResult) (string, error)
	GetHTML(filename string) (string, error)
	ListCrawledFiles() ([]FileDetail, error)
}

type FileDetail struct {
	Filename string
	Size     int64
	Path     string
}

type CrawlRepositoryImpl struct {
	baseDir string
}

func NewCrawlRepository(cfg *config.Config) CrawlRepository {
	if err := os.MkdirAll(cfg.CrawledPagesDir, os.ModePerm); err != nil {
		panic(fmt.Sprintf("gagal membuat direktori %s: %v", cfg.CrawledPagesDir, err))
	}
	return &CrawlRepositoryImpl{
		baseDir: cfg.CrawledPagesDir,
	}
}

func (r *CrawlRepositoryImpl) SaveHTML(result *model.CrawlResult) (string, error) {
	filename := helper.SanitizeURLToFilename(result.URL) + ".html"
	filePath := filepath.Join(r.baseDir, filename)

	err := os.WriteFile(filePath, []byte(result.HTMLContent), 0644)
	if err != nil {
		return "", fmt.Errorf("gagal menyimpan file %s: %w", filePath, err)
	}

	result.FilePath = filePath
	return filePath, nil
}

func (r *CrawlRepositoryImpl) GetHTML(filename string) (string, error) {
	filename = filepath.Base(filename)
	if !strings.HasSuffix(filename, ".html") {
		filename += ".html"
	}

	filePath := filepath.Join(r.baseDir, filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file %s tidak ditemukan", filename)
		}
		return "", fmt.Errorf("gagal membaca file %s: %w", filename, err)
	}

	return string(data), nil
}

func (r *CrawlRepositoryImpl) ListCrawledFiles() ([]FileDetail, error) {
	entries, err := os.ReadDir(r.baseDir)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca direktori %s: %w", r.baseDir, err)
	}

	var files []FileDetail
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, FileDetail{
			Filename: entry.Name(),
			Size:     info.Size(),
			Path:     filepath.Join(r.baseDir, entry.Name()),
		})
	}

	return files, nil
}
