package httpapi

import (
	"net/http"

	"github.com/fulltank-garage/linora/apps/api/internal/analysis"
	"github.com/gin-gonic/gin"
)

func NewRouter(analysisService *analysis.Service) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(localDevCors())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "linora-api",
		})
	})

	router.POST("/api/analysis/manual", func(c *gin.Context) {
		var input analysis.ManualInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
			return
		}

		report, err := analysisService.AnalyzeManualInput(input)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"report": report})
	})

	return router
}

func localDevCors() gin.HandlerFunc {
	allowedOrigins := map[string]bool{
		"http://localhost:5173": true,
		"http://127.0.0.1:5173": true,
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
