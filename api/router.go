package api

import (
	"github.com/gin-gonic/gin"

	"github.com/ilhaamms/crawler-website/controller"
)

func SetupRouter(crawlController *controller.CrawlController) *gin.Engine {
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "crawler-website",
		})
	})

	api := router.Group("/api")
	{
		crawl := api.Group("/crawl")
		{
			crawl.POST("", crawlController.CrawlSingle)

			crawl.POST("/batch", crawlController.CrawlBatch)

			crawl.GET("/files", crawlController.ListFiles)

			crawl.GET("/files/:filename", crawlController.GetFile)
		}
	}

	return router
}
