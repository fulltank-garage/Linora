package main

import (
	"context"
	"log"
	"strings"

	"github.com/fulltank-garage/linora/apps/api/internal/config"
	"github.com/fulltank-garage/linora/apps/api/internal/repositories"
	"github.com/fulltank-garage/linora/apps/api/internal/routes"
	"github.com/fulltank-garage/linora/apps/api/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("load configuration: ", err)
	}
	ctx := context.Background()
	store, err := repositories.NewPostgresStore(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Fatal("connect PostgreSQL: ", err)
	}
	defer store.Close()
	if err := store.Migrate(ctx); err != nil {
		log.Fatal("migrate PostgreSQL: ", err)
	}
	var reportCache repositories.ReportCache
	if strings.TrimSpace(cfg.RedisURL) != "" {
		cache, err := repositories.NewRedisReportCache(ctx, cfg.RedisURL)
		if err != nil {
			log.Fatal("connect Redis: ", err)
		}
		defer cache.Close()
		reportCache = cache
	}
	cipher, err := services.NewTokenCipher(cfg.EncryptionKey)
	if err != nil {
		log.Fatal("configure token encryption: ", err)
	}
	analysisService := services.NewAnalysisService()
	facebookService := services.NewFacebookService(cfg.Facebook)
	aiService := services.NewAIService(cfg.AI)
	pageService := services.NewPageService(store, reportCache, cipher, facebookService, analysisService, aiService)
	lineService := services.NewLineService(store, aiService, cfg.Line)
	lineIdentityService := services.NewLineIdentityService(cfg.Line)
	router := routes.NewRouter(
		cfg,
		analysisService,
		facebookService,
		pageService,
		lineService,
		lineIdentityService,
	)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
