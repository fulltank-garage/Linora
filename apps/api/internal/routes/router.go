package routes

import (
	"net/http"

	"github.com/fulltank-garage/linora/apps/api/internal/config"
	"github.com/fulltank-garage/linora/apps/api/internal/controllers"
	"github.com/fulltank-garage/linora/apps/api/internal/middleware"
	"github.com/fulltank-garage/linora/apps/api/internal/services"
	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, analysisService *services.AnalysisService, facebookService *services.FacebookService, pageService *services.PageService, lineService *services.LineService, lineIdentity *services.LineIdentityService) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), middleware.CORS(cfg.Facebook.AppURL))

	facebookController := controllers.NewFacebookController(facebookService, pageService)
	requireLineIdentity := middleware.RequireLineIdentity(lineIdentity, cfg.Environment)
	if pageService != nil && lineService != nil {
		integrationController := controllers.NewIntegrationController(pageService, lineService, cfg.Line.ChannelSecret)
		facebook := router.Group("/api/facebook", requireLineIdentity)
		facebook.POST("/connections", integrationController.ConnectPage)
		facebook.GET("/dashboard", integrationController.Dashboard)
		facebook.GET("/pages", integrationController.ListPages)
		facebook.POST("/pages/:pageID/select", integrationController.SelectPage)
		facebook.POST("/pages/:pageID/sync", integrationController.SyncPage)
		facebook.GET("/pages/:pageID/report", integrationController.LatestReport)
		facebook.GET("/pages/:pageID/weekly-report", integrationController.WeeklyReport)
		facebook.DELETE("/pages/:pageID/connection", integrationController.DisconnectPage)
		facebook.DELETE("/pages/:pageID", integrationController.DeletePage)
		facebook.POST("/pages/:pageID/line-link-code", integrationController.CreateLineLinkCode)
		router.POST("/api/line/link", integrationController.LinkLine)
		router.POST("/api/line/chat", integrationController.LocalLineChat)
		router.POST("/api/line/rich-menu/connect", requireLineIdentity, integrationController.ActivateConnectRichMenu)
		router.POST("/api/line/rich-menu/dashboard", requireLineIdentity, integrationController.ActivateDashboardRichMenu)
		router.POST("/api/line/webhook", integrationController.LineWebhook)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "linora-api"})
	})
	router.POST("/api/facebook/login", requireLineIdentity, facebookController.Begin)
	router.GET("/api/facebook/callback", facebookController.Callback)
	router.GET("/api/facebook/session", requireLineIdentity, facebookController.Session)
	router.POST("/api/facebook/data-deletion", facebookController.DataDeletion)

	return router
}
