package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fulltank-garage/linora/apps/api/internal/models"
)

type AnalysisService struct{}

func NewAnalysisService() *AnalysisService {
	return &AnalysisService{}
}

func (s *AnalysisService) AnalyzeManualInput(input models.ManualAnalysisInput) (models.AnalysisReport, error) {
	pageName := strings.TrimSpace(input.PageName)
	postContent := strings.TrimSpace(input.PostContent)
	if pageName == "" {
		return models.AnalysisReport{}, errors.New("กรุณากรอกชื่อเพจ")
	}
	if postContent == "" {
		return models.AnalysisReport{}, errors.New("กรุณากรอกเนื้อหาโพสต์")
	}

	score := calculateHealthScore(input)
	comments := parseImportantComments(input.ImportantComments)
	summary := fmt.Sprintf("เพจ %s มี engagement จากข้อมูลที่กรอกอยู่ในระดับ %s โดยโพสต์นี้มี %d likes, %d comments และ %d shares", pageName, scoreLabel(score), input.Likes, input.Comments, input.Shares)

	return models.AnalysisReport{
		ID:          fmt.Sprintf("manual-%d", time.Now().Unix()),
		PageName:    pageName,
		Summary:     summary,
		HealthScore: score,
		TopPosts: []models.TopPost{
			{
				PostID:         "manual-post",
				Reason:         "โพสต์นี้ถูกใช้เป็นข้อมูลตั้งต้นสำหรับการวิเคราะห์แบบ manual",
				Recommendation: "เพิ่ม call-to-action ให้ชัด และต่อยอดหัวข้อที่มีคนคอมเมนต์มากที่สุด",
			},
		},
		ImportantComments: comments,
		ContentRecommendations: []string{
			"โพสต์รีวิวหรือผลลัพธ์จากลูกค้าจริงเพื่อเพิ่มความน่าเชื่อถือ",
			"ทำโพสต์โปรโมชันพร้อม call-to-action ที่ตอบได้ทันทีใน LINE",
			"นำคำถามจากคอมเมนต์มาทำเป็นคอนเทนต์ FAQ",
		},
		BestPostingTimes:   []string{"18:00 - 20:00", "11:00 - 13:00"},
		LineSummaryMessage: fmt.Sprintf("สรุปเพจวันนี้จาก Linora\n\nคะแนนภาพรวม: %d/100\n\nคำแนะนำ: เพิ่ม CTA ให้ชัดและตอบคอมเมนต์สำคัญให้เร็วขึ้น", score),
		CreatedAt:          time.Now().Format(time.RFC3339),
	}, nil
}

func calculateHealthScore(input models.ManualAnalysisInput) int {
	score := 45 + min(input.Likes/10, 25) + min(input.Comments*2, 20) + min(input.Shares*2, 10)
	if strings.TrimSpace(input.ImportantComments) != "" {
		score += 5
	}
	if score > 100 {
		return 100
	}
	if score < 0 {
		return 0
	}
	return score
}

func parseImportantComments(raw string) []models.ImportantComment {
	lines := strings.Split(raw, "\n")
	comments := make([]models.ImportantComment, 0, len(lines))
	for _, line := range lines {
		message := strings.TrimSpace(line)
		if message == "" {
			continue
		}
		comments = append(comments, models.ImportantComment{
			CommentID:      fmt.Sprintf("manual-comment-%d", len(comments)+1),
			Message:        message,
			Reason:         "เป็นคอมเมนต์ที่ควรตอบเร็ว เพราะอาจเกี่ยวข้องกับการตัดสินใจซื้อหรือจอง",
			SuggestedReply: "ขอบคุณที่สนใจค่ะ รบกวนแจ้งรายละเอียดเพิ่มเติมทางแชท LINE ได้เลยนะคะ",
		})
	}
	if len(comments) == 0 {
		return []models.ImportantComment{
			{
				CommentID:      "manual-comment-1",
				Message:        "ยังไม่มีคอมเมนต์สำคัญที่ระบุ",
				Reason:         "ควรติดตามคอมเมนต์ใหม่หลังโพสต์เพื่อหาโอกาสในการตอบกลับ",
				SuggestedReply: "ขอบคุณที่สนใจค่ะ สอบถามเพิ่มเติมได้ทาง LINE นี้เลยนะคะ",
			},
		}
	}
	return comments
}

func scoreLabel(score int) string {
	if score >= 80 {
		return "ดีมาก"
	}
	if score >= 60 {
		return "ดี"
	}
	return "ควรปรับปรุง"
}
