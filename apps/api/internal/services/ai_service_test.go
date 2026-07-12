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

func TestCompleteGeminiUsesGenerateContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/models/gemini-test:generateContent" {
			t.Fatalf("unexpected request path: %s", request.URL.Path)
		}
		if got := request.Header.Get("x-goog-api-key"); got != "gemini-key" {
			t.Fatalf("unexpected API key header: %q", got)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"คำแนะนำจาก Gemini"}]}}]}`))
	}))
	defer server.Close()

	service := &AIService{
		config: config.AIConfig{APIKey: "gemini-key", BaseURL: server.URL, Model: "gemini-test", Provider: "gemini"},
		http:   server.Client(),
	}

	answer, err := service.complete(context.Background(), "วิเคราะห์เพจนี้")
	if err != nil {
		t.Fatalf("complete returned an error: %v", err)
	}
	if answer != "คำแนะนำจาก Gemini" {
		t.Fatalf("unexpected answer: %q", answer)
	}
}

func TestCompleteAnthropicUsesMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/messages" {
			t.Fatalf("unexpected request path: %s", request.URL.Path)
		}
		if got := request.Header.Get("x-api-key"); got != "anthropic-key" {
			t.Fatalf("unexpected API key header: %q", got)
		}
		if got := request.Header.Get("anthropic-version"); got != "2023-06-01" {
			t.Fatalf("unexpected Anthropic version: %q", got)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"content":[{"type":"text","text":"คำแนะนำจาก Claude"}]}`))
	}))
	defer server.Close()

	service := &AIService{
		config: config.AIConfig{APIKey: "anthropic-key", BaseURL: server.URL, Model: "claude-test", Provider: "anthropic"},
		http:   server.Client(),
	}

	answer, err := service.complete(context.Background(), "วิเคราะห์เพจนี้")
	if err != nil {
		t.Fatalf("complete returned an error: %v", err)
	}
	if answer != "คำแนะนำจาก Claude" {
		t.Fatalf("unexpected answer: %q", answer)
	}
}
