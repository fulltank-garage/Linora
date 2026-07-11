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
	OwnerID   string
	Pages     []models.FacebookPage
}

type facebookOAuthState struct {
	ExpiresAt time.Time
	OwnerID   string
}

type FacebookAPIError struct {
	Code       int
	Message    string
	StatusCode int
}

func (e *FacebookAPIError) Error() string {
	return fmt.Sprintf("Facebook API request failed (%d, code %d): %s", e.StatusCode, e.Code, e.Message)
}

func IsFacebookAccessTokenError(err error) bool {
	var facebookError *FacebookAPIError
	return errors.As(err, &facebookError) && (facebookError.Code == 190 || facebookError.StatusCode == http.StatusUnauthorized)
}

type FacebookService struct {
	config   config.FacebookConfig
	handoffs map[string]facebookHandoff
	oauth    map[string]facebookOAuthState
	http     *http.Client
	mu       sync.Mutex
}

func NewFacebookService(cfg config.FacebookConfig) *FacebookService {
	if cfg.GraphVersion == "" {
		cfg.GraphVersion = "v24.0"
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{"pages_show_list", "pages_read_engagement", "pages_read_user_content", "read_insights"}
	}
	return &FacebookService{
		config:   cfg,
		handoffs: make(map[string]facebookHandoff),
		oauth:    make(map[string]facebookOAuthState),
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
	query.Set("scope", strings.Join(s.config.Scopes, ","))
	query.Set("state", state)
	authorizeURL.RawQuery = query.Encode()
	return authorizeURL.String()
}

func (s *FacebookService) StartAuthorization(ownerID string) (string, error) {
	state, err := SecureToken()
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for key, entry := range s.oauth {
		if now.After(entry.ExpiresAt) {
			delete(s.oauth, key)
		}
	}
	s.oauth[state] = facebookOAuthState{ExpiresAt: now.Add(FacebookHandoffLifetime), OwnerID: ownerID}
	return s.AuthorizationURL(state), nil
}

func (s *FacebookService) ConsumeAuthorizationState(state string) (string, error) {
	s.mu.Lock()
	entry, ok := s.oauth[state]
	if ok {
		delete(s.oauth, state)
	}
	s.mu.Unlock()
	if !ok || time.Now().After(entry.ExpiresAt) {
		return "", errors.New("Facebook Login session expired")
	}
	return entry.OwnerID, nil
}

func (s *FacebookService) CompleteLogin(ctx context.Context, code string, ownerID string) (string, error) {
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
	s.storeHandoff(handoffCode, ownerID, pages)
	return handoffCode, nil
}

func (s *FacebookService) RedeemHandoff(code string, ownerID string) ([]models.FacebookPage, error) {
	s.mu.Lock()
	entry, ok := s.handoffs[code]
	s.mu.Unlock()
	if !ok || entry.OwnerID != ownerID || time.Now().After(entry.ExpiresAt) {
		return nil, errors.New("session Facebook Login หมดอายุแล้ว กรุณาเข้าสู่ระบบอีกครั้ง")
	}
	return entry.Pages, nil
}

func (s *FacebookService) ConsumePage(code string, ownerID string, pageID string) (models.FacebookPage, error) {
	page, _, err := s.ConsumePages(code, ownerID, pageID)
	return page, err
}

func (s *FacebookService) ConsumePages(code string, ownerID string, pageID string) (models.FacebookPage, []models.FacebookPage, error) {
	s.mu.Lock()
	entry, ok := s.handoffs[code]
	if ok {
		delete(s.handoffs, code)
	}
	s.mu.Unlock()
	if !ok || entry.OwnerID != ownerID || time.Now().After(entry.ExpiresAt) {
		return models.FacebookPage{}, nil, errors.New("session Facebook Login หมดอายุแล้ว กรุณาเข้าสู่ระบบอีกครั้ง")
	}
	for _, page := range entry.Pages {
		if page.PageID == pageID && page.AccessToken != "" {
			return page, entry.Pages, nil
		}
	}
	return models.FacebookPage{}, nil, errors.New("ไม่พบสิทธิ์เข้าถึงเพจที่เลือก")
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
		return "", facebookResponseError(response)
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
	query.Set("fields", "id,name,category,access_token")
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
		return nil, facebookResponseError(response)
	}

	var payload struct {
		Data []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Category    string `json:"category"`
			AccessToken string `json:"access_token"`
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
		pages = append(pages, models.FacebookPage{AccessToken: page.AccessToken, PageID: page.ID, PageName: page.Name, Category: category, IsActive: true})
	}
	return pages, nil
}

func (s *FacebookService) FetchPageSnapshot(ctx context.Context, pageID string, accessToken string) (models.PageSnapshot, error) {
	postsURL := url.URL{Scheme: "https", Host: "graph.facebook.com", Path: "/" + s.config.GraphVersion + "/" + pageID + "/posts"}
	query := postsURL.Query()
	query.Set("fields", "id,message,created_time,shares,reactions.limit(0).summary(true),comments.limit(0).summary(true)")
	query.Set("limit", "25")
	query.Set("access_token", accessToken)
	postsURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, postsURL.String(), nil)
	if err != nil {
		return models.PageSnapshot{}, err
	}
	response, err := s.http.Do(request)
	if err != nil {
		return models.PageSnapshot{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return models.PageSnapshot{}, facebookResponseError(response)
	}

	var payload struct {
		Data []struct {
			ID          string `json:"id"`
			Message     string `json:"message"`
			CreatedTime string `json:"created_time"`
			Shares      struct {
				Count int64 `json:"count"`
			} `json:"shares"`
			Reactions struct {
				Summary struct {
					TotalCount int64 `json:"total_count"`
				} `json:"summary"`
			} `json:"reactions"`
			Comments struct {
				Summary struct {
					TotalCount int64 `json:"total_count"`
				} `json:"summary"`
			} `json:"comments"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return models.PageSnapshot{}, err
	}

	snapshot := models.PageSnapshot{Posts: make([]models.FacebookPost, 0, len(payload.Data))}
	for _, post := range payload.Data {
		if post.ID == "" {
			continue
		}
		item := models.FacebookPost{
			Comments:  post.Comments.Summary.TotalCount,
			CreatedAt: post.CreatedTime,
			ID:        post.ID,
			Message:   post.Message,
			Reactions: post.Reactions.Summary.TotalCount,
			Shares:    post.Shares.Count,
		}
		snapshot.Posts = append(snapshot.Posts, item)
		snapshot.Metrics.Engagements += item.Comments + item.Reactions + item.Shares
	}

	insights, _ := s.fetchInsights(ctx, pageID, accessToken)
	snapshot.Metrics.Impressions = insights.Impressions
	snapshot.Metrics.Reach = insights.Reach
	if snapshot.Metrics.Reach == 0 {
		snapshot.Metrics.Reach = snapshot.Metrics.Engagements
	}
	if snapshot.Metrics.Impressions == 0 {
		snapshot.Metrics.Impressions = snapshot.Metrics.Reach
	}

	for _, post := range snapshot.Posts[:min(len(snapshot.Posts), 5)] {
		comments, err := s.fetchImportantComments(ctx, post.ID, accessToken)
		if err == nil {
			snapshot.Comments = append(snapshot.Comments, comments...)
		}
	}
	return snapshot, nil
}

func (s *FacebookService) fetchInsights(ctx context.Context, pageID string, accessToken string) (models.PageMetrics, error) {
	insightsURL := url.URL{Scheme: "https", Host: "graph.facebook.com", Path: "/" + s.config.GraphVersion + "/" + pageID + "/insights"}
	query := insightsURL.Query()
	query.Set("metric", "page_impressions,page_post_engagements")
	query.Set("period", "day")
	query.Set("access_token", accessToken)
	insightsURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, insightsURL.String(), nil)
	if err != nil {
		return models.PageMetrics{}, err
	}
	response, err := s.http.Do(request)
	if err != nil {
		return models.PageMetrics{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return models.PageMetrics{}, facebookResponseError(response)
	}

	var payload struct {
		Data []struct {
			Name   string `json:"name"`
			Values []struct {
				Value json.RawMessage `json:"value"`
			} `json:"values"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return models.PageMetrics{}, err
	}
	metrics := models.PageMetrics{}
	for _, metric := range payload.Data {
		if len(metric.Values) == 0 {
			continue
		}
		var value int64
		if json.Unmarshal(metric.Values[len(metric.Values)-1].Value, &value) != nil {
			continue
		}
		switch metric.Name {
		case "page_impressions":
			metrics.Impressions = value
			metrics.Reach = value
		case "page_post_engagements":
			metrics.Engagements = value
		}
	}
	return metrics, nil
}

func (s *FacebookService) fetchImportantComments(ctx context.Context, postID string, accessToken string) ([]models.ImportantComment, error) {
	commentsURL := url.URL{Scheme: "https", Host: "graph.facebook.com", Path: "/" + s.config.GraphVersion + "/" + postID + "/comments"}
	query := commentsURL.Query()
	query.Set("fields", "id,message")
	query.Set("limit", "25")
	query.Set("access_token", accessToken)
	commentsURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, commentsURL.String(), nil)
	if err != nil {
		return nil, err
	}
	response, err := s.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, facebookResponseError(response)
	}

	var payload struct {
		Data []struct {
			ID      string `json:"id"`
			Message string `json:"message"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	comments := make([]models.ImportantComment, 0, 3)
	for _, comment := range payload.Data {
		if len(comments) == 3 || !isImportantComment(comment.Message) {
			continue
		}
		comments = append(comments, models.ImportantComment{
			CommentID:      comment.ID,
			Message:        comment.Message,
			Reason:         "เป็นคำถามหรือสัญญาณที่ควรตอบเร็ว",
			SuggestedReply: "ขอบคุณที่สนใจครับ รบกวนแจ้งรายละเอียดเพิ่มเติมทาง LINE ได้เลยครับ",
		})
	}
	return comments, nil
}

func isImportantComment(message string) bool {
	value := strings.ToLower(strings.TrimSpace(message))
	for _, keyword := range []string{"ราคา", "จอง", "ซื้อ", "สั่ง", "how much", "price", "order", "booking"} {
		if strings.Contains(value, keyword) {
			return true
		}
	}
	return false
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func facebookResponseError(response *http.Response) error {
	var payload struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.NewDecoder(response.Body).Decode(&payload)
	message := strings.TrimSpace(payload.Error.Message)
	if message == "" {
		message = response.Status
	}
	return &FacebookAPIError{Code: payload.Error.Code, Message: message, StatusCode: response.StatusCode}
}

func (s *FacebookService) storeHandoff(code string, ownerID string, pages []models.FacebookPage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for key, entry := range s.handoffs {
		if now.After(entry.ExpiresAt) {
			delete(s.handoffs, key)
		}
	}
	s.handoffs[code] = facebookHandoff{ExpiresAt: now.Add(FacebookHandoffLifetime), OwnerID: ownerID, Pages: pages}
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
