package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fulltank-garage/linora/apps/api/internal/config"
	"github.com/fulltank-garage/linora/apps/api/internal/models"
)

func TestAnswerRequiresSelectedPageReport(t *testing.T) {
	service := &AIService{}
	answer := service.Answer(context.Background(), nil, "ช่วยคิดหัวข้อโพสต์ร้านกาแฟ")
	if !strings.Contains(answer, "เชื่อมต่อ Facebook") || !strings.Contains(answer, "เลือกเพจ") {
		t.Fatalf("unexpected answer without selected page: %q", answer)
	}
}

func TestAnswerBuildsDetailedPageStrategyPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		messages, ok := body["messages"].([]any)
		if !ok || len(messages) != 1 {
			t.Fatalf("expected one chat message, got %#v", body["messages"])
		}
		content := messages[0].(map[string]any)["content"].(string)
		if !strings.Contains(content, baseAIBehavior) || !strings.Contains(content, "🎯 แนวทางที่ควรทำ") || !strings.Contains(content, "✅ สรุปพร้อมนำไปใช้") || !strings.Contains(content, "Linora Cafe") {
			t.Fatalf("expected detailed strategy prompt, got %q", content)
		}
		_, _ = writer.Write([]byte(`{"choices":[{"message":{"content":"🎯 แนวทางที่ควรทำ\nโพสต์เรื่องเมนูเด่น\n\n💡 หัวข้อโพสต์ที่น่าลอง\nเมนูที่ลูกค้ากลับมาซ้ำ"}}]}`))
	}))
	defer server.Close()

	service := &AIService{config: config.AIConfig{APIKey: "key", BaseURL: server.URL, Model: "test", Provider: "deepseek"}, http: server.Client()}
	report := &models.AnalysisReport{PageName: "Linora Cafe", Summary: "มีข้อมูลล่าสุด"}
	answer := service.Answer(context.Background(), report, "ช่วยคิดหัวข้อโพสต์ร้านกาแฟ")
	if !strings.Contains(answer, "🎯 แนวทางที่ควรทำ") || !strings.Contains(answer, "หัวข้อโพสต์") {
		t.Fatalf("unexpected strategy answer: %q", answer)
	}
}

func TestAnswerKeepsPageQuestionsGrounded(t *testing.T) {
	service := &AIService{}
	answer := service.Answer(context.Background(), nil, "สรุปยอดเข้าถึงของเพจฉันให้หน่อย")
	if !strings.Contains(answer, "เชื่อมต่อ Facebook") || !strings.Contains(answer, "เลือกเพจ") {
		t.Fatalf("expected request to select a page, got %q", answer)
	}

	report := &models.AnalysisReport{PageName: "Linora Demo", Summary: "มีข้อมูลล่าสุด"}
	if got := service.Answer(context.Background(), report, ""); got == "" {
		t.Fatal("expected empty questions to return guidance")
	}
}

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
