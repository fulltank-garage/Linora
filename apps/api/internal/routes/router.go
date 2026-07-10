package routes

import (
	"net/http"

	"github.com/fulltank-garage/linora/apps/api/internal/config"
	"github.com/fulltank-garage/linora/apps/api/internal/controllers"
	"github.com/fulltank-garage/linora/apps/api/internal/middleware"
	"github.com/fulltank-garage/linora/apps/api/internal/services"
	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, analysisService *services.AnalysisService, facebookService *services.FacebookService) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), middleware.CORS(cfg.Facebook.AppURL))

	analysisController := controllers.NewAnalysisController(analysisService)
	facebookController := controllers.NewFacebookController(facebookService)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "linora-api"})
	})
	router.POST("/api/analysis/manual", analysisController.Manual)
	router.GET("/api/facebook/login", facebookController.Begin)
	router.GET("/api/facebook/callback", facebookController.Callback)
	router.GET("/api/facebook/session", facebookController.Session)

	return router
}
