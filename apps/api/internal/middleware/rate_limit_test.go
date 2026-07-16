package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimitRejectsRequestsAfterLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimit(2, time.Minute))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	for requestNumber := 0; requestNumber < 2; requestNumber++ {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
		if response.Code != http.StatusNoContent {
			t.Fatalf("request %d status = %d, want %d", requestNumber+1, response.Code, http.StatusNoContent)
		}
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusTooManyRequests)
	}
	if response.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
}
