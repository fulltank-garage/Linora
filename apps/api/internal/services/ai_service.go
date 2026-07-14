package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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
	return s.config.Validate() == nil
}

func (s *AIService) EnhanceReport(ctx context.Context, report models.AnalysisReport) models.AnalysisReport {
	if !s.Enabled() {
		return report
	}
	prompt := fmt.Sprintf(`คุณคือ Linora AI ผู้ช่วยวิเคราะห์คอนเทนต์ Facebook ตอบภาษาไทยแบบกระชับ
วิเคราะห์เฉพาะข้อมูลโพสต์, engagement, comments และ shares ในรายงานนี้: %s
ให้คำแนะนำที่ผู้ดูแลเพจอ่านแล้วนำไปทำต่อได้ทันที โดยอ้างอิงรูปแบบเนื้อหาหรือการมีส่วนร่วมที่พบจริง
ห้ามพูดถึงวันหรือเวลาโพสต์ ห้ามแต่งตัวเลข อ้างข้อมูลนอกเหนือจากนี้ หรือรับประกันผลลัพธ์
ใช้ภาษาคนทั่วไป หลีกเลี่ยงคำเทคนิค เช่น engagement, CTA, conversion หรือ content format
ตอบเป็น 3 บรรทัดตามนี้เท่านั้น:
แนะนำ: [บอกสิ่งที่ควรทำด้วยคำกริยาชัดเจน]
วิธีทำ: [อธิบายแนวโพสต์หรือข้อความที่ควรลองแบบทำตามได้]
เหตุผล: [อธิบายสั้น ๆ ว่าแนวทางนี้สัมพันธ์กับข้อมูลเพจอย่างไร]
แต่ละบรรทัดยาวไม่เกิน 1 ประโยค ไม่ต้องใส่ Markdown`, mustJSON(report))
	if answer, err := s.complete(ctx, prompt); err == nil && answer != "" {
		report.AIContentRecommendation = answer
	}
	if report.PostingTimeInsight.BasedOnPosts >= 3 && report.PostingTimeInsight.BestTime != "" {
		postingTimePrompt := fmt.Sprintf(`คุณคือ Linora AI ผู้ช่วยวิเคราะห์เพจ Facebook ตอบภาษาไทยแบบกระชับ
วิเคราะห์เฉพาะข้อมูลผลการโพสต์นี้: %s
เขียนคำแนะนำ 2 ข้อที่เป็นประโยชน์ต่อผู้ดูแลเพจ โดยขึ้นบรรทัดใหม่ทุกข้อและใช้รูปแบบนี้เท่านั้น:
1. [ตีความแนวโน้มที่พบจากวันและช่วงเวลาที่ทำผลงานดี โดยไม่คัดลอกข้อมูลดิบมาเรียงใหม่]
2. [แผนทดลองโพสต์ครั้งถัดไปที่ทำได้จริง เช่น เปรียบเทียบรูปแบบเนื้อหาหรือคำกระตุ้นให้มีส่วนร่วม]
แต่ละข้อมีเพียง 1 ประโยค ไม่ต้องใส่หัวข้อเพิ่มหรือ Markdown
อาจกล่าวถึงวันหรือเวลาได้เพียงครั้งเดียวเมื่อช่วยให้ลงมือทำได้ ห้ามแต่งตัวเลข อ้างข้อมูลนอกเหนือจากนี้ หรือรับประกันผลลัพธ์`, mustJSON(report.PostingTimeInsight))
		if answer, err := s.complete(ctx, postingTimePrompt); err == nil && answer != "" {
			report.PostingTimeRecommendation = answer
		}
	}
	return report
}

func (s *AIService) Answer(ctx context.Context, report *models.AnalysisReport, question string) string {
	question = strings.TrimSpace(question)
	if question == "" {
		return "พิมพ์คำถามเกี่ยวกับเพจที่เลือก หรือขอไอเดียคอนเทนต์ได้เลยครับ"
	}
	return s.answerFromReport(ctx, report, question)
}

func (s *AIService) answerFromReport(ctx context.Context, report *models.AnalysisReport, question string) string {
	if report == nil {
		return "กรุณาเลือกเพจใน Linora ก่อนครับ"
	}
	if !s.Enabled() {
		return fmt.Sprintf("สรุปเพจ %s: %s", report.PageName, report.Summary)
	}
	prompt := fmt.Sprintf(`คุณคือ Linora ผู้ช่วยวางกลยุทธ์คอนเทนต์ Facebook ของเพจที่ผู้ใช้เลือกอยู่ ตอบภาษาไทยที่เข้าใจง่าย สุภาพ และนำไปทำต่อได้จริง
คำตอบนี้อ้างอิงได้เฉพาะรายงานของเพจที่เลือกอยู่ด้านล่างเท่านั้น ห้ามอ้างว่าเห็นข้อมูลของเพจอื่น ห้ามแต่งตัวเลข ห้ามบอกว่าคุณโพสต์ ตอบคอมเมนต์ หรือแก้ไขเพจให้ผู้ใช้แล้ว

เมื่อคำถามต้องการไอเดียคอนเทนต์ กลยุทธ์ หรือแนวทางทำโพสต์ ให้ตอบละเอียดตามหัวข้อต่อไปนี้ โดยใช้เฉพาะหัวข้อที่เกี่ยวข้อง:
🎯 แนวทางที่ควรทำ
💡 หัวข้อโพสต์ที่น่าลอง
🪝 จุดเปิดโพสต์
📝 วิธีทำคอนเทนต์
✨ จุดขายที่ควรสื่อ
🖼️ ภาพหรือวิดีโอที่เหมาะ
📣 ชวนผู้อ่านทำอะไรต่อ
📌 เหตุผลจากข้อมูลเพจ

ใต้แต่ละหัวข้อให้เขียนเป็นข้อความต่อเนื่องหรือรายการสั้น ๆ ที่ทำตามได้จริง เว้นบรรทัดว่าง 1 บรรทัดระหว่างทุกหัวข้อ ใช้อีโมจิเฉพาะต้นหัวข้อ และอย่าใช้อีโมจิในเนื้อหาจนรก
ถ้าผู้ใช้ถามข้อมูลเชิงข้อเท็จจริง ให้ตอบตรงคำถามจากรายงานก่อน แล้วค่อยให้ข้อเสนอแนะเมื่อมีประโยชน์ หากรายงานไม่มีข้อมูลที่ถาม ให้บอกอย่างตรงไปตรงมา

รายงานเพจที่เลือก: %s

คำถามผู้ใช้: %s`, mustJSON(report), question)
	answer, err := s.complete(ctx, prompt)
	if err != nil || answer == "" {
		return fmt.Sprintf("ยังตอบเชิงลึกไม่ได้ในขณะนี้ สรุปที่มีคือ: %s", report.Summary)
	}
	return answer
}

func (s *AIService) complete(ctx context.Context, prompt string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s.config.Provider)) {
	case "deepseek", "openai-compatible":
		return s.completeChatCompletions(ctx, prompt)
	case "openai":
		return s.completeResponses(ctx, prompt)
	case "gemini":
		return s.completeGemini(ctx, prompt)
	case "anthropic":
		return s.completeAnthropic(ctx, prompt)
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

func (s *AIService) completeGemini(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(map[string]any{
		"contents":         []map[string]any{{"parts": []map[string]string{{"text": prompt}}}},
		"generationConfig": map[string]float64{"temperature": 0.2},
	})
	if err != nil {
		return "", err
	}
	endpoint := strings.TrimRight(s.config.BaseURL, "/") + "/models/" + url.PathEscape(s.config.Model) + ":generateContent"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("x-goog-api-key", s.config.APIKey)
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
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	if len(payload.Candidates) == 0 || len(payload.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("AI response contains no content")
	}
	return strings.TrimSpace(payload.Candidates[0].Content.Parts[0].Text), nil
}

func (s *AIService) completeAnthropic(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(map[string]any{
		"max_tokens":  1024,
		"messages":    []map[string]string{{"role": "user", "content": prompt}},
		"model":       s.config.Model,
		"temperature": 0.2,
	})
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(s.config.BaseURL, "/")+"/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("anthropic-version", "2023-06-01")
	request.Header.Set("x-api-key", s.config.APIKey)
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
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	if len(payload.Content) == 0 {
		return "", errors.New("AI response contains no content")
	}
	return strings.TrimSpace(payload.Content[0].Text), nil
}

func mustJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}
