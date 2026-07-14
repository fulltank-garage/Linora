package services

import (
	"testing"

	"github.com/fulltank-garage/linora/apps/api/internal/models"
)

func TestAnalyzePageSnapshotCalculatesPostingTimeFromPostEngagement(t *testing.T) {
	report := NewAnalysisService().AnalyzePageSnapshot(models.FacebookPage{PageName: "Linora"}, models.PageSnapshot{
		Posts: []models.FacebookPost{
			{ID: "monday", CreatedAt: "2026-07-06T10:00:00+07:00", Reactions: 10, Comments: 2, Shares: 1},
			{ID: "wednesday", CreatedAt: "2026-07-08T19:00:00+07:00", Reactions: 50, Comments: 10, Shares: 5},
		},
	})

	if report.PostingTimeInsight.BasedOnPosts != 2 {
		t.Fatalf("basedOnPosts = %d, want 2", report.PostingTimeInsight.BasedOnPosts)
	}
	if report.PostingTimeInsight.BestDay != "พ." {
		t.Fatalf("bestDay = %q, want พ.", report.PostingTimeInsight.BestDay)
	}
	if report.PostingTimeInsight.BestTime != "18:00 - 20:00" {
		t.Fatalf("bestTime = %q, want 18:00 - 20:00", report.PostingTimeInsight.BestTime)
	}
	if len(report.BestPostingTimes) != 1 || report.BestPostingTimes[0] != report.PostingTimeInsight.BestTime {
		t.Fatalf("bestPostingTimes = %#v, want derived best time", report.BestPostingTimes)
	}
}

func TestAnalyzePageSnapshotSupportsFacebookOffsetWithoutColon(t *testing.T) {
	report := NewAnalysisService().AnalyzePageSnapshot(models.FacebookPage{PageName: "Linora"}, models.PageSnapshot{
		Posts: []models.FacebookPost{
			{ID: "facebook-format", CreatedAt: "2026-07-10T19:00:00+0000", Reactions: 10},
		},
	})

	if report.PostingTimeInsight.BasedOnPosts != 1 {
		t.Fatalf("basedOnPosts = %d, want 1", report.PostingTimeInsight.BasedOnPosts)
	}
	if report.PostingTimeInsight.BestTime != "00:00 - 02:00" {
		t.Fatalf("bestTime = %q, want 00:00 - 02:00", report.PostingTimeInsight.BestTime)
	}
}
