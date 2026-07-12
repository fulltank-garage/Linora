package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fulltank-garage/linora/apps/api/internal/config"
	"github.com/fulltank-garage/linora/apps/api/internal/models"
)

type AIService struct {
	config config.AIConfig
	http   *http.Client
}

func NewAIService(cfg config.AIConfig) *AIService {
	return &AIService{config: cfg, http: &http.Client{Timeout: 20 * time.Second}}
}

func (s *AIService) Enabled() bool {
	provider := strings.ToLower(strings.TrimSpace(s.config.Provider))
	return s.config.APIKey != "" && s.config.BaseURL != "" && s.config.Model != "" && (provider == "openai" || provider == "deepseek")
}

func (s *AIService) EnhanceReport(ctx context.Context, report models.AnalysisReport) models.AnalysisReport {
	if !s.Enabled() {
		return report
	}
	prompt := fmt.Sprintf(`คุณคือ Linora ผู้ช่วยวิเคราะห์เพจ Facebook ตอบภาษาไทยแบบกระชับ
ข้อมูลที่อนุญาตให้ใช้: %s
สร้างคำแนะนำเชิงปฏิบัติ 1 ข้อ ห้ามแต่งตัวเลขหรืออ้างถึงข้อมูลส่วนบุคคล`, mustJSON(report))
	if answer, err := s.complete(ctx, prompt); err == nil && answer != "" {
		report.ContentRecommendations = append([]string{answer}, report.ContentRecommendations...)
	}
	return report
}

func (s *AIService) Answer(ctx context.Context, report models.AnalysisReport, question string) string {
	question = strings.TrimSpace(question)
	if question == "" {
		return "พิมพ์คำถามเกี่ยวกับผลวิเคราะห์เพจได้เลยครับ"
	}
	if !s.Enabled() {
		return fmt.Sprintf("สรุปเพจ %s: %s", report.PageName, report.Summary)
	}
	prompt := fmt.Sprintf(`คุณคือ Linora ตอบคำถามผู้ดูแลเพจเป็นภาษาไทยอย่างสุภาพและกระชับ
ตอบจากรายงานนี้เท่านั้น หากไม่มีข้อมูลให้บอกตรง ๆ ว่ายังไม่มีข้อมูลเพียงพอ
รายงาน: %s
คำถาม: %s`, mustJSON(report), question)
	answer, err := s.complete(ctx, prompt)
	if err != nil || answer == "" {
		return fmt.Sprintf("ยังตอบเชิงลึกไม่ได้ในขณะนี้ สรุปที่มีคือ: %s", report.Summary)
	}
	return answer
}

func (s *AIService) complete(ctx context.Context, prompt string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s.config.Provider)) {
	case "deepseek":
		return s.completeChatCompletions(ctx, prompt)
	case "openai":
		return s.completeResponses(ctx, prompt)
	default:
		return "", fmt.Errorf("unsupported AI provider: %s", s.config.Provider)
	}
}

func (s *AIService) completeResponses(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(map[string]any{"model": s.config.Model, "input": prompt, "temperature": 0.2})
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(s.config.BaseURL, "/")+"/responses", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	request.Header.Set("Content-Type", "application/json")
	response, err := s.http.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("AI request returned %s", response.Status)
	}
	var payload struct {
		OutputText string `json:"output_text"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	return strings.TrimSpace(payload.OutputText), nil
}

func (s *AIService) completeChatCompletions(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(map[string]any{
		"model":       s.config.Model,
		"messages":    []map[string]string{{"role": "user", "content": prompt}},
		"temperature": 0.2,
	})
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(s.config.BaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	request.Header.Set("Content-Type", "application/json")
	response, err := s.http.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("AI request returned %s", response.Status)
	}
	var payload struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	if len(payload.Choices) == 0 {
		return "", errors.New("AI response contains no choices")
	}
	return strings.TrimSpace(payload.Choices[0].Message.Content), nil
}

func mustJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}
