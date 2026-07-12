package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/fulltank-garage/linora/apps/api/internal/config"
	"github.com/fulltank-garage/linora/apps/api/internal/repositories"
)

type LineService struct {
	apiBaseURL string
	ai         *AIService
	config     config.LineConfig
	http       *http.Client
	store      repositories.Store
}

func NewLineService(store repositories.Store, ai *AIService, cfg config.LineConfig) *LineService {
	return &LineService{apiBaseURL: "https://api.line.me", ai: ai, config: cfg, http: &http.Client{}, store: store}
}

func (s *LineService) LinkDashboardRichMenu(ctx context.Context, lineUserID string) error {
	if s.config.ChannelAccessToken == "" || s.config.DashboardRichMenuID == "" || strings.TrimSpace(lineUserID) == "" {
		return nil
	}

	baseURL := s.apiBaseURL
	if baseURL == "" {
		baseURL = "https://api.line.me"
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/v2/bot/user/" + url.PathEscape(lineUserID) + "/richmenu/" + url.PathEscape(s.config.DashboardRichMenuID)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+s.config.ChannelAccessToken)
	response, err := s.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return errors.New("LINE rich menu link request failed")
	}
	return nil
}

func (s *LineService) Link(ctx context.Context, lineUserID string, code string) (string, error) {
	return s.store.UseLinkCode(ctx, strings.TrimSpace(code), lineUserID)
}

func (s *LineService) Chat(ctx context.Context, lineUserID string, message string) (string, error) {
	pageID, err := s.store.GetLinkedPage(ctx, lineUserID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return "ยังไม่ได้เชื่อม LINE กับเพจ กรุณาสร้างรหัสเชื่อมต่อจาก Linora ก่อนครับ", nil
		}
		return "", err
	}
	report, err := s.store.GetLatestReport(ctx, lineUserID, pageID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return "เพจนี้ยังไม่มีรายงานล่าสุด กรุณากดวิเคราะห์เพจก่อนครับ", nil
		}
		return "", err
	}
	return s.ai.Answer(ctx, report, message), nil
}

func (s *LineService) Reply(ctx context.Context, replyToken string, message string) error {
	if s.config.ChannelAccessToken == "" || replyToken == "" {
		return nil
	}
	body := []byte(`{"replyToken":` + quoteJSON(replyToken) + `,"messages":[{"type":"text","text":` + quoteJSON(message) + `}]}`)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.line.me/v2/bot/message/reply", bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+s.config.ChannelAccessToken)
	request.Header.Set("Content-Type", "application/json")
	response, err := s.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return errors.New("LINE reply request failed")
	}
	return nil
}

func quoteJSON(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func VerifyLineSignature(channelSecret string, body []byte, signature string) bool {
	if channelSecret == "" || signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(channelSecret))
	mac.Write(body)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
