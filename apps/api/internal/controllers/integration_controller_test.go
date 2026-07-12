package controllers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fulltank-garage/linora/apps/api/internal/config"
	"github.com/fulltank-garage/linora/apps/api/internal/services"
	"github.com/gin-gonic/gin"
)

func TestLineWebhookAcceptsVerifiedPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "line-webhook-secret"
	line := services.NewLineService(nil, nil, config.LineConfig{})
	controller := NewIntegrationController(nil, line, secret)
	router := gin.New()
	router.POST("/webhook", controller.LineWebhook)

	body := []byte(`{"events":[{"source":{"userId":"U123"},"message":{"type":"follow"}}]}`)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	request := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	request.Header.Set("X-Line-Signature", signature)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}
