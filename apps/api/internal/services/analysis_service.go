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

func (s *AnalysisService) AnalyzePageSnapshot(page models.FacebookPage, snapshot models.PageSnapshot) models.AnalysisReport {
	var best models.FacebookPost
	for _, post := range snapshot.Posts {
		if post.Reactions+post.Comments+post.Shares > best.Reactions+best.Comments+best.Shares {
			best = post
		}
	}
	score := 45 + min(int(snapshot.Metrics.Engagements/4), 35) + min(len(snapshot.Posts)*2, 15) + min(len(snapshot.Comments)*2, 5)
	if score > 100 {
		score = 100
	}
	important := snapshot.Comments
	if len(important) == 0 {
		important = []models.ImportantComment{{
			CommentID:      "none",
			Message:        "ยังไม่พบคอมเมนต์ที่ต้องติดตามเป็นพิเศษ",
			Reason:         "Linora จะตรวจสอบอีกครั้งเมื่อมีข้อมูลใหม่",
			SuggestedReply: "",
		}}
	}
	postReason := "ยังมีข้อมูลโพสต์ไม่เพียงพอสำหรับจัดอันดับ"
	postRecommendation := "เผยแพร่โพสต์เพิ่มเติมและกลับมาซิงก์ข้อมูลอีกครั้งเพื่อดูแนวโน้ม"
	if best.ID != "" {
		postReason = fmt.Sprintf("โพสต์นี้มีปฏิสัมพันธ์รวม %d ครั้ง จึงเป็นโพสต์ที่ทำผลงานเด่นในรอบล่าสุด", best.Reactions+best.Comments+best.Shares)
		postRecommendation = "ต่อยอดหัวข้อของโพสต์นี้ และเพิ่ม CTA ที่ชัดเจนเพื่อให้ลูกค้าทัก LINE ได้ทันที"
	}
	postingTimeInsight := calculatePostingTimeInsight(snapshot.Posts)
	bestPostingTimes := []string(nil)
	if postingTimeInsight.BestTime != "" {
		bestPostingTimes = []string{postingTimeInsight.BestTime}
	}
	return models.AnalysisReport{
		ID:                fmt.Sprintf("facebook-%d", time.Now().UnixNano()),
		PageName:          page.PageName,
		Summary:           fmt.Sprintf("วิเคราะห์จาก %d โพสต์ล่าสุด: มีปฏิสัมพันธ์รวม %d ครั้ง", len(snapshot.Posts), snapshot.Metrics.Engagements),
		HealthScore:       score,
		TopPosts:          []models.TopPost{{PostID: best.ID, Reason: postReason, Recommendation: postRecommendation}},
		ImportantComments: important,
		ContentRecommendations: []string{
			"ต่อยอดเนื้อหาที่ผู้ติดตามมีส่วนร่วมสูง พร้อม CTA ที่ตอบได้ทันที",
			"นำคำถามจากลูกค้ามาทำเป็นโพสต์ FAQ เพื่อลดเวลาตอบซ้ำ",
			"ติดตามผลหลังโพสต์และเปรียบเทียบแนวโน้มทุกสัปดาห์",
		},
		BestPostingTimes:   bestPostingTimes,
		PostingTimeInsight: postingTimeInsight,
		LineSummaryMessage: fmt.Sprintf("สรุปเพจ %s\nคะแนนภาพรวม %d/100\nปฏิสัมพันธ์ล่าสุด %d ครั้ง", page.PageName, score, snapshot.Metrics.Engagements),
		Metrics:            snapshot.Metrics,
		CreatedAt:          time.Now().Format(time.RFC3339),
	}
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
		BestPostingTimes:   nil,
		LineSummaryMessage: fmt.Sprintf("สรุปเพจวันนี้จาก Linora\n\nคะแนนภาพรวม: %d/100\n\nคำแนะนำ: เพิ่ม CTA ให้ชัดและตอบคอมเมนต์สำคัญให้เร็วขึ้น", score),
		Metrics: models.PageMetrics{
			Engagements: int64(input.Likes + input.Comments + input.Shares),
		},
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func calculatePostingTimeInsight(posts []models.FacebookPost) models.PostingTimeInsight {
	dayLabels := []string{"จ.", "อ.", "พ.", "พฤ.", "ศ.", "ส.", "อา."}
	insight := models.PostingTimeInsight{Days: make([]models.PostingDayInsight, len(dayLabels))}
	type aggregate struct {
		engagements int64
		postCount   int
	}
	days := make([]aggregate, len(dayLabels))
	timeSlots := make([]aggregate, 8)
	location := time.FixedZone("Asia/Bangkok", 7*60*60)

	for _, post := range posts {
		postedAt, err := time.Parse(time.RFC3339, post.CreatedAt)
		if err != nil {
			continue
		}
		engagements := post.Reactions + post.Comments + post.Shares
		dayIndex := (int(postedAt.In(location).Weekday()) + 6) % len(dayLabels)
		days[dayIndex].engagements += engagements
		days[dayIndex].postCount++
		timeSlot := postedAt.In(location).Hour() / 3
		timeSlots[timeSlot].engagements += engagements
		timeSlots[timeSlot].postCount++
		insight.BasedOnPosts++
	}

	bestDayAverage := int64(-1)
	for index, day := range days {
		average := int64(0)
		if day.postCount > 0 {
			average = day.engagements / int64(day.postCount)
			if average > bestDayAverage {
				bestDayAverage = average
				insight.BestDay = dayLabels[index]
			}
		}
		insight.Days[index] = models.PostingDayInsight{
			AverageEngagement: average,
			Day:               dayLabels[index],
			PostCount:         day.postCount,
		}
	}

	bestTimeAverage := int64(-1)
	for slot, item := range timeSlots {
		if item.postCount == 0 {
			continue
		}
		average := item.engagements / int64(item.postCount)
		if average <= bestTimeAverage {
			continue
		}
		bestTimeAverage = average
		startHour := slot * 3
		insight.BestTime = fmt.Sprintf("%02d:00 - %02d:00", startHour, startHour+2)
	}

	return insight
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
