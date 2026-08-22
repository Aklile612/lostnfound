package handler

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Router(h *Handler, origin, uploadDir string) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.MaxMultipartMemory = 32 << 20
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{origin},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{"Content-Type"},
	}))
	r.GET("/health", h.Health)
	r.GET("/api/meta", h.Meta)
	r.POST("/api/reports/lost", h.CreateLost)
	r.POST("/api/reports/found", h.CreateFound)
	r.GET("/api/reports/:id", h.GetReport)
	r.POST("/api/reports/:id/matches", h.RefreshMatches)
	r.POST("/api/reports/:id/claim", h.Claim)
	r.Static("/uploads", uploadDir)
	return r
}

func EnsureUploadDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

func NormalizeOrigin(origin string) string {
	return strings.TrimRight(origin, "/")
}
