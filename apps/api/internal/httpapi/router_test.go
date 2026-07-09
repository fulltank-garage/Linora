package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fulltank-garage/linora/apps/api/internal/analysis"
)

func TestHealthRoute(t *testing.T) {
	router := NewRouter(analysis.NewService())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestManualAnalysisRoute(t *testing.T) {
	router := NewRouter(analysis.NewService())
	body := map[string]any{
		"pageName":          "Linora Cafe",
		"postContent":       "โปรโมชันวันนี้",
		"likes":             80,
		"comments":          12,
		"shares":            6,
		"importantComments": "ราคาเท่าไหร่",
		"extraNotes":        "เน้นลูกค้าใหม่",
	}
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/analysis/manual", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte("report")) {
		t.Fatalf("body = %s, want report field", recorder.Body.String())
	}
}

func TestManualAnalysisRouteReturnsBadRequest(t *testing.T) {
	router := NewRouter(analysis.NewService())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/analysis/manual", bytes.NewReader([]byte(`{"pageName":""}`)))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestManualAnalysisRouteAllowsLocalDevCors(t *testing.T) {
	router := NewRouter(analysis.NewService())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/api/analysis/manual", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want http://localhost:5173", got)
	}
}
