package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSAllowsDeleteForConfiguredOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS("https://linora.up.railway.app"))
	router.OPTIONS("/api/facebook/pages/:pageID", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodOptions, "/api/facebook/pages/page-123", nil)
	request.Header.Set("Origin", "https://linora.up.railway.app")
	request.Header.Set("Access-Control-Request-Method", http.MethodDelete)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	if !strings.Contains(recorder.Header().Get("Access-Control-Allow-Methods"), http.MethodDelete) {
		t.Fatalf("DELETE is missing from Access-Control-Allow-Methods: %q", recorder.Header().Get("Access-Control-Allow-Methods"))
	}
}
