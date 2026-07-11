package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fulltank-garage/linora/apps/api/internal/config"
)

var ErrInvalidLineIdentity = errors.New("LINE identity could not be verified")

type LineIdentityService struct {
	channelID string
	http      *http.Client
}

func NewLineIdentityService(cfg config.LineConfig) *LineIdentityService {
	return &LineIdentityService{
		channelID: strings.TrimSpace(cfg.ChannelID),
		http:      &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *LineIdentityService) Configured() bool {
	return s.channelID != ""
}

func (s *LineIdentityService) VerifyIDToken(ctx context.Context, idToken string) (string, error) {
	if !s.Configured() || strings.TrimSpace(idToken) == "" {
		return "", ErrInvalidLineIdentity
	}

	form := url.Values{}
	form.Set("id_token", idToken)
	form.Set("client_id", s.channelID)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.line.me/oauth2/v2.1/verify", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := s.http.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", ErrInvalidLineIdentity
	}

	var payload struct {
		Subject string `json:"sub"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil || strings.TrimSpace(payload.Subject) == "" {
		return "", ErrInvalidLineIdentity
	}
	return payload.Subject, nil
}
