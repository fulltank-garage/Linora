package services

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/fulltank-garage/linora/apps/api/internal/config"
	"github.com/fulltank-garage/linora/apps/api/internal/models"
)

const FacebookHandoffLifetime = 5 * time.Minute

type facebookHandoff struct {
	ExpiresAt time.Time
	Pages     []models.FacebookPage
}

type FacebookService struct {
	config   config.FacebookConfig
	handoffs map[string]facebookHandoff
	http     *http.Client
	mu       sync.Mutex
}

func NewFacebookService(cfg config.FacebookConfig) *FacebookService {
	if cfg.GraphVersion == "" {
		cfg.GraphVersion = "v24.0"
	}
	return &FacebookService{
		config:   cfg,
		handoffs: make(map[string]facebookHandoff),
		http:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *FacebookService) Configured() bool {
	return s.config.AppID != "" && s.config.AppSecret != "" && s.config.AppURL != "" && s.config.RedirectURI != ""
}

func (s *FacebookService) AppURL() string {
	return s.config.AppURL
}

func (s *FacebookService) RedirectURI() string {
	return s.config.RedirectURI
}

func (s *FacebookService) AuthorizationURL(state string) string {
	authorizeURL := url.URL{Scheme: "https", Host: "www.facebook.com", Path: "/" + s.config.GraphVersion + "/dialog/oauth"}
	query := authorizeURL.Query()
	query.Set("client_id", s.config.AppID)
	query.Set("redirect_uri", s.config.RedirectURI)
	query.Set("response_type", "code")
	query.Set("scope", "pages_show_list,pages_read_engagement")
	query.Set("state", state)
	authorizeURL.RawQuery = query.Encode()
	return authorizeURL.String()
}

func (s *FacebookService) CompleteLogin(ctx context.Context, code string) (string, error) {
	accessToken, err := s.exchangeCode(ctx, code)
	if err != nil {
		return "", err
	}
	pages, err := s.fetchPages(ctx, accessToken)
	if err != nil {
		return "", err
	}
	handoffCode, err := SecureToken()
	if err != nil {
		return "", err
	}
	s.storeHandoff(handoffCode, pages)
	return handoffCode, nil
}

func (s *FacebookService) RedeemHandoff(code string) ([]models.FacebookPage, error) {
	s.mu.Lock()
	entry, ok := s.handoffs[code]
	delete(s.handoffs, code)
	s.mu.Unlock()
	if !ok || time.Now().After(entry.ExpiresAt) {
		return nil, errors.New("session Facebook Login หมดอายุแล้ว กรุณาเข้าสู่ระบบอีกครั้ง")
	}
	return entry.Pages, nil
}

func (s *FacebookService) exchangeCode(ctx context.Context, code string) (string, error) {
	tokenURL := url.URL{Scheme: "https", Host: "graph.facebook.com", Path: "/" + s.config.GraphVersion + "/oauth/access_token"}
	query := tokenURL.Query()
	query.Set("client_id", s.config.AppID)
	query.Set("client_secret", s.config.AppSecret)
	query.Set("redirect_uri", s.config.RedirectURI)
	query.Set("code", code)
	tokenURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, tokenURL.String(), nil)
	if err != nil {
		return "", err
	}
	response, err := s.http.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("facebook token exchange returned %s", response.Status)
	}

	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.AccessToken == "" {
		return "", errors.New("facebook response did not include an access token")
	}
	return payload.AccessToken, nil
}

func (s *FacebookService) fetchPages(ctx context.Context, accessToken string) ([]models.FacebookPage, error) {
	pagesURL := url.URL{Scheme: "https", Host: "graph.facebook.com", Path: "/" + s.config.GraphVersion + "/me/accounts"}
	query := pagesURL.Query()
	query.Set("fields", "id,name,category")
	query.Set("access_token", accessToken)
	pagesURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, pagesURL.String(), nil)
	if err != nil {
		return nil, err
	}
	response, err := s.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("facebook pages request returned %s", response.Status)
	}

	var payload struct {
		Data []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Category string `json:"category"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}

	pages := make([]models.FacebookPage, 0, len(payload.Data))
	for _, page := range payload.Data {
		if page.ID == "" || page.Name == "" {
			continue
		}
		category := page.Category
		if category == "" {
			category = "Facebook Page"
		}
		pages = append(pages, models.FacebookPage{PageID: page.ID, PageName: page.Name, Category: category, IsActive: true})
	}
	return pages, nil
}

func (s *FacebookService) storeHandoff(code string, pages []models.FacebookPage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for key, entry := range s.handoffs {
		if now.After(entry.ExpiresAt) {
			delete(s.handoffs, key)
		}
	}
	s.handoffs[code] = facebookHandoff{ExpiresAt: now.Add(FacebookHandoffLifetime), Pages: pages}
}

func SecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", bytes), nil
}

func IsHTTPS(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	return err == nil && strings.EqualFold(parsed.Scheme, "https")
}
