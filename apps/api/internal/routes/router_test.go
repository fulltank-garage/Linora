package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fulltank-garage/linora/apps/api/internal/config"
	"github.com/fulltank-garage/linora/apps/api/internal/services"
)

func newTestRouter() http.Handler {
	cfg := config.Config{Environment: "development", Port: "8080"}
	return NewRouter(cfg, services.NewAnalysisService(), services.NewFacebookService(cfg.Facebook), nil, nil, services.NewLineIdentityService(cfg.Line))
}

func TestHealthRoute(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestRouter().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestFacebookLoginRequiresConfiguration(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/facebook/login", nil)
	request.Header.Set("X-Linora-Dev-User", "test-line-user")
	newTestRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
}

func TestFacebookDeauthorizeRouteIsPublic(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestRouter().ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/facebook/deauthorize", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
}
