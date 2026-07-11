package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type FacebookConfig struct {
	AppID        string
	AppSecret    string
	AppURL       string
	GraphVersion string
	RedirectURI  string
	Scopes       []string
}

type AIConfig struct {
	APIKey   string
	BaseURL  string
	Model    string
	Provider string
}

type LineConfig struct {
	ChannelAccessToken string
	ChannelID          string
	ChannelSecret      string
}

type Config struct {
	AI            AIConfig
	DatabaseDSN   string
	EncryptionKey string
	Environment   string
	Facebook      FacebookConfig
	Line          LineConfig
	Port          string
	RedisURL      string
}

func Load() Config {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		AI: AIConfig{
			APIKey:   os.Getenv("AI_API_KEY"),
			BaseURL:  os.Getenv("AI_BASE_URL"),
			Model:    os.Getenv("AI_MODEL"),
			Provider: os.Getenv("AI_PROVIDER"),
		},
		DatabaseDSN:   os.Getenv("DB_DSN"),
		EncryptionKey: os.Getenv("TOKEN_ENCRYPTION_KEY"),
		Environment:   os.Getenv("APP_ENV"),
		Port:          port,
		RedisURL:      os.Getenv("REDIS_URL"),
		Facebook: FacebookConfig{
			AppID:        os.Getenv("FACEBOOK_APP_ID"),
			AppSecret:    os.Getenv("FACEBOOK_APP_SECRET"),
			AppURL:       os.Getenv("APP_URL"),
			GraphVersion: os.Getenv("FACEBOOK_GRAPH_VERSION"),
			RedirectURI:  os.Getenv("FACEBOOK_REDIRECT_URI"),
			Scopes:       splitScopes(os.Getenv("FACEBOOK_SCOPES")),
		},
		Line: LineConfig{
			ChannelAccessToken: os.Getenv("LINE_CHANNEL_ACCESS_TOKEN"),
			ChannelID:          os.Getenv("LINE_CHANNEL_ID"),
			ChannelSecret:      os.Getenv("LINE_CHANNEL_SECRET"),
		},
	}
}

func splitScopes(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{"pages_show_list", "pages_read_engagement", "pages_read_user_content", "read_insights"}
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			values = append(values, value)
		}
	}
	return values
}
