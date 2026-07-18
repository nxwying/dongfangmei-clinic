package main

import (
	"embed"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"

	"clinic-mgmt/internal/config"
	"clinic-mgmt/internal/database"
	"clinic-mgmt/internal/backup"
	"clinic-mgmt/internal/handler"
	"clinic-mgmt/internal/seed"

	"github.com/gin-gonic/gin"
)

//go:embed web/dist
var frontend embed.FS

func serveFrontend(c *gin.Context) {
	// Only serve non-API routes
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	subFS, err := fs.Sub(frontend, "web/dist")
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Strip /v2/ prefix for the new base URL
	filePath := strings.TrimPrefix(c.Request.URL.Path, "/v2/")
	filePath = strings.TrimPrefix(filePath, "/")
	if filePath == "" {
		filePath = "index.html"
	}

	data, err := fs.ReadFile(subFS, filePath)
	if err != nil {
		// SPA fallback: return index.html for all unknown paths
		data, err = fs.ReadFile(subFS, "index.html")
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		filePath = "index.html"
	}

	// Set content type based on file extension
	ext := path.Ext(filePath)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// No caching (development mode - force fresh load)
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	c.Data(http.StatusOK, contentType, data)
	c.Abort()
}

func main() {
	cfg := config.Load()
	db := database.InitDB(cfg.DBDsn)

	seed.Seed(db)

	r := gin.Default()

	// API routes
	handler.RegisterRoutes(r, db, cfg)

	// Start backup scheduler
	go backup.StartScheduler(db, cfg)

	// Serve frontend (SPA with fallback to index.html)
	r.Use(serveFrontend)

	log.Printf("Server starting on :%s", cfg.Port)
	log.Printf("Open http://localhost:%s in your browser", cfg.Port)
	log.Printf("Login with admin / admin123")
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
