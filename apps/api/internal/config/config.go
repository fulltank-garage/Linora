package config

import (
	"os"

	"github.com/joho/godotenv"
)

type FacebookConfig struct {
	AppID        string
	AppSecret    string
	AppURL       string
	GraphVersion string
	RedirectURI  string
}

type Config struct {
	Facebook FacebookConfig
	Port     string
}

func Load() Config {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		Port: port,
		Facebook: FacebookConfig{
			AppID:        os.Getenv("FACEBOOK_APP_ID"),
			AppSecret:    os.Getenv("FACEBOOK_APP_SECRET"),
			AppURL:       os.Getenv("APP_URL"),
			GraphVersion: os.Getenv("FACEBOOK_GRAPH_VERSION"),
			RedirectURI:  os.Getenv("FACEBOOK_REDIRECT_URI"),
		},
	}
}
