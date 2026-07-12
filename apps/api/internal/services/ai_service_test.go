package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fulltank-garage/linora/apps/api/internal/config"
)

func TestCompleteDeepSeekUsesChatCompletions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected request path: %s", request.URL.Path)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer deepseek-key" {
			t.Fatalf("unexpected authorization header: %q", got)
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["model"] != "deepseek-v4-flash" {
			t.Fatalf("unexpected model: %v", body["model"])
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"choices":[{"message":{"content":"คำแนะนำจาก DeepSeek"}}]}`))
	}))
	defer server.Close()

	service := &AIService{
		config: config.AIConfig{
			APIKey:   "deepseek-key",
			BaseURL:  server.URL,
			Model:    "deepseek-v4-flash",
			Provider: "deepseek",
		},
		http: server.Client(),
	}

	answer, err := service.complete(context.Background(), "วิเคราะห์เพจนี้")
	if err != nil {
		t.Fatalf("complete returned an error: %v", err)
	}
	if answer != "คำแนะนำจาก DeepSeek" {
		t.Fatalf("unexpected answer: %q", answer)
	}
}
