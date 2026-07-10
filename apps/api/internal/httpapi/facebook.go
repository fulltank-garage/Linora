package httpapi

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

	"github.com/gin-gonic/gin"
)

const (
	facebookStateCookie = "linora_facebook_oauth_state"
	handoffLifetime     = 5 * time.Minute
)

type FacebookConfig struct {
	AppID        string
	AppSecret    string
	AppURL       string
	GraphVersion string
	RedirectURI  string
}

type FacebookPage struct {
	PageID   string `json:"pageId"`
	PageName string `json:"pageName"`
	Category string `json:"category"`
	IsActive bool   `json:"isActive"`
}

type facebookHandoff struct {
	ExpiresAt time.Time
	Pages     []FacebookPage
}

type facebookOAuth struct {
	config   FacebookConfig
	handoffs map[string]facebookHandoff
	mu       sync.Mutex
	http     *http.Client
}

func newFacebookOAuth(config FacebookConfig) *facebookOAuth {
	if config.GraphVersion == "" {
		config.GraphVersion = "v24.0"
	}
	return &facebookOAuth{
		config:   config,
		handoffs: make(map[string]facebookHandoff),
		http:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (f *facebookOAuth) configured() bool {
	return f.config.AppID != "" && f.config.AppSecret != "" && f.config.AppURL != "" && f.config.RedirectURI != ""
}

func (f *facebookOAuth) begin(c *gin.Context) {
	if !f.configured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Facebook Login ยังไม่ได้ตั้งค่าในระบบ"})
		return
	}

	state, err := secureToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเริ่ม Facebook Login ได้"})
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(facebookStateCookie, state, int(handoffLifetime.Seconds()), "/api/facebook", "", isHTTPS(f.config.RedirectURI), true)

	authorizeURL := url.URL{Scheme: "https", Host: "www.facebook.com", Path: "/" + f.config.GraphVersion + "/dialog/oauth"}
	query := authorizeURL.Query()
	query.Set("client_id", f.config.AppID)
	query.Set("redirect_uri", f.config.RedirectURI)
	query.Set("response_type", "code")
	query.Set("scope", "pages_show_list,pages_read_engagement")
	query.Set("state", state)
	authorizeURL.RawQuery = query.Encode()
	c.Redirect(http.StatusFound, authorizeURL.String())
}

func (f *facebookOAuth) callback(c *gin.Context) {
	if c.Query("error") != "" {
		f.redirectWithError(c, "Facebook Login ถูกยกเลิกหรือไม่สำเร็จ")
		return
	}

	state, err := c.Cookie(facebookStateCookie)
	if err != nil || state == "" || state != c.Query("state") {
		f.redirectWithError(c, "ไม่สามารถยืนยัน Facebook Login ได้")
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(facebookStateCookie, "", -1, "/api/facebook", "", isHTTPS(f.config.RedirectURI), true)

	code := c.Query("code")
	if code == "" {
		f.redirectWithError(c, "Facebook ไม่ส่งรหัสยืนยันกลับมา")
		return
	}

	accessToken, err := f.exchangeCode(c.Request.Context(), code)
	if err != nil {
		f.redirectWithError(c, "ไม่สามารถเชื่อมต่อ Facebook ได้")
		return
	}
	pages, err := f.fetchPages(c.Request.Context(), accessToken)
	if err != nil {
		f.redirectWithError(c, "ไม่สามารถอ่านรายการ Facebook Page ได้")
		return
	}

	handoff, err := secureToken()
	if err != nil {
		f.redirectWithError(c, "ไม่สามารถสร้าง session สำหรับ Facebook Login ได้")
		return
	}
	f.storeHandoff(handoff, pages)

	redirectURL, err := url.Parse(f.config.AppURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ตั้งค่า APP_URL ไม่ถูกต้อง"})
		return
	}
	redirectURL.Path = "/connect-facebook"
	query := redirectURL.Query()
	query.Set("facebook_connect", handoff)
	redirectURL.RawQuery = query.Encode()
	c.Redirect(http.StatusFound, redirectURL.String())
}

func (f *facebookOAuth) handoff(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่พบรหัสยืนยัน Facebook Login"})
		return
	}

	f.mu.Lock()
	entry, ok := f.handoffs[code]
	delete(f.handoffs, code)
	f.mu.Unlock()
	if !ok || time.Now().After(entry.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session Facebook Login หมดอายุแล้ว กรุณาเข้าสู่ระบบอีกครั้ง"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pages": entry.Pages})
}

func (f *facebookOAuth) exchangeCode(ctx context.Context, code string) (string, error) {
	tokenURL := url.URL{Scheme: "https", Host: "graph.facebook.com", Path: "/" + f.config.GraphVersion + "/oauth/access_token"}
	query := tokenURL.Query()
	query.Set("client_id", f.config.AppID)
	query.Set("client_secret", f.config.AppSecret)
	query.Set("redirect_uri", f.config.RedirectURI)
	query.Set("code", code)
	tokenURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, tokenURL.String(), nil)
	if err != nil {
		return "", err
	}
	response, err := f.http.Do(request)
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

func (f *facebookOAuth) fetchPages(ctx context.Context, accessToken string) ([]FacebookPage, error) {
	pagesURL := url.URL{Scheme: "https", Host: "graph.facebook.com", Path: "/" + f.config.GraphVersion + "/me/accounts"}
	query := pagesURL.Query()
	query.Set("fields", "id,name,category")
	query.Set("access_token", accessToken)
	pagesURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, pagesURL.String(), nil)
	if err != nil {
		return nil, err
	}
	response, err := f.http.Do(request)
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
	pages := make([]FacebookPage, 0, len(payload.Data))
	for _, page := range payload.Data {
		if page.ID == "" || page.Name == "" {
			continue
		}
		category := page.Category
		if category == "" {
			category = "Facebook Page"
		}
		pages = append(pages, FacebookPage{PageID: page.ID, PageName: page.Name, Category: category, IsActive: true})
	}
	return pages, nil
}

func (f *facebookOAuth) storeHandoff(code string, pages []FacebookPage) {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := time.Now()
	for key, entry := range f.handoffs {
		if now.After(entry.ExpiresAt) {
			delete(f.handoffs, key)
		}
	}
	f.handoffs[code] = facebookHandoff{ExpiresAt: now.Add(handoffLifetime), Pages: pages}
}

func (f *facebookOAuth) redirectWithError(c *gin.Context, message string) {
	redirectURL, err := url.Parse(f.config.AppURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": message})
		return
	}
	redirectURL.Path = "/connect-facebook"
	query := redirectURL.Query()
	query.Set("facebook_error", message)
	redirectURL.RawQuery = query.Encode()
	c.Redirect(http.StatusFound, redirectURL.String())
}

func secureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", bytes), nil
}

func isHTTPS(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	return err == nil && strings.EqualFold(parsed.Scheme, "https")
}
