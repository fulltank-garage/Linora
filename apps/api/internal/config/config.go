package config

import (
	"fmt"
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
	ChannelAccessToken  string
	ChannelID           string
	ChannelSecret       string
	ConnectRichMenuID   string
	DashboardRichMenuID string
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

func Load() (Config, error) {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cfg := Config{
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
			ChannelAccessToken:  os.Getenv("LINE_CHANNEL_ACCESS_TOKEN"),
			ChannelID:           os.Getenv("LINE_CHANNEL_ID"),
			ChannelSecret:       os.Getenv("LINE_CHANNEL_SECRET"),
			ConnectRichMenuID:   os.Getenv("LINE_RICH_MENU_CONNECT_ID"),
			DashboardRichMenuID: os.Getenv("LINE_RICH_MENU_DASHBOARD_ID"),
		},
	}
	if err := cfg.AI.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c AIConfig) Validate() error {
	missing := make([]string, 0, 4)
	if strings.TrimSpace(c.Provider) == "" {
		missing = append(missing, "AI_PROVIDER")
	}
	if strings.TrimSpace(c.BaseURL) == "" {
		missing = append(missing, "AI_BASE_URL")
	}
	if strings.TrimSpace(c.Model) == "" {
		missing = append(missing, "AI_MODEL")
	}
	if strings.TrimSpace(c.APIKey) == "" {
		missing = append(missing, "AI_API_KEY")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required AI environment variables: %s", strings.Join(missing, ", "))
	}

	switch strings.ToLower(strings.TrimSpace(c.Provider)) {
	case "anthropic", "deepseek", "gemini", "openai", "openai-compatible":
		return nil
	default:
		return fmt.Errorf("unsupported AI_PROVIDER: %s", c.Provider)
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
