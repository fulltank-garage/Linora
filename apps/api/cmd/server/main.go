package main

import (
	"log"

	"github.com/fulltank-garage/linora/apps/api/internal/config"
	"github.com/fulltank-garage/linora/apps/api/internal/routes"
	"github.com/fulltank-garage/linora/apps/api/internal/services"
)

func main() {
	cfg := config.Load()
	router := routes.NewRouter(
		cfg,
		services.NewAnalysisService(),
		services.NewFacebookService(cfg.Facebook),
	)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
