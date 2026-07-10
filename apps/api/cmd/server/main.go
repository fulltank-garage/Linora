package main

import (
	"log"
	"os"

	"github.com/fulltank-garage/linora/apps/api/internal/analysis"
	"github.com/fulltank-garage/linora/apps/api/internal/httpapi"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router := httpapi.NewRouter(analysis.NewService(), httpapi.FacebookConfig{
		AppID:        os.Getenv("FACEBOOK_APP_ID"),
		AppSecret:    os.Getenv("FACEBOOK_APP_SECRET"),
		AppURL:       os.Getenv("APP_URL"),
		GraphVersion: os.Getenv("FACEBOOK_GRAPH_VERSION"),
		RedirectURI:  os.Getenv("FACEBOOK_REDIRECT_URI"),
	})
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
