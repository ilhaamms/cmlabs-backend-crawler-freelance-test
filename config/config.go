package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	CrawledPagesDir string
	ChromeTimeout   time.Duration
}

func NewConfig() *Config {
	port := getEnv("PORT", "8080")
	crawledDir := getEnv("CRAWLED_PAGES_DIR", "crawled_pages")
	timeoutSec := getEnvAsInt("CHROME_TIMEOUT", 60)

	return &Config{
		Port:            port,
		CrawledPagesDir: crawledDir,
		ChromeTimeout:   time.Duration(timeoutSec) * time.Second,
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}
