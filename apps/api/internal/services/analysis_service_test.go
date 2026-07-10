package services

import (
	"testing"

	"github.com/fulltank-garage/linora/apps/api/internal/models"
)

func TestAnalyzeManualInputReturnsReport(t *testing.T) {
	service := NewAnalysisService()

	report, err := service.AnalyzeManualInput(models.ManualAnalysisInput{
		PageName:          "Linora Cafe",
		PostContent:       "โปรโมชันเครื่องดื่มใหม่วันนี้",
		Likes:             120,
		Comments:          18,
		Shares:            9,
		ImportantComments: "ราคาเท่าไหร่\nจองได้ไหม",
		ExtraNotes:        "ลูกค้าสนใจช่วงเย็น",
	})

	if err != nil {
		t.Fatalf("AnalyzeManualInput returned error: %v", err)
	}
	if report.PageName != "Linora Cafe" {
		t.Fatalf("PageName = %q, want Linora Cafe", report.PageName)
	}
	if report.HealthScore < 0 || report.HealthScore > 100 {
		t.Fatalf("HealthScore = %d, want 0..100", report.HealthScore)
	}
	if len(report.ImportantComments) != 2 {
		t.Fatalf("ImportantComments length = %d, want 2", len(report.ImportantComments))
	}
	if report.LineSummaryMessage == "" {
		t.Fatal("LineSummaryMessage is empty")
	}
}

func TestAnalyzeManualInputRequiresPageName(t *testing.T) {
	service := NewAnalysisService()

	_, err := service.AnalyzeManualInput(models.ManualAnalysisInput{
		PostContent: "โพสต์ใหม่",
	})

	if err == nil {
		t.Fatal("AnalyzeManualInput error = nil, want validation error")
	}
}

func TestAnalyzeManualInputRequiresPostContent(t *testing.T) {
	service := NewAnalysisService()

	_, err := service.AnalyzeManualInput(models.ManualAnalysisInput{
		PageName: "Linora Cafe",
	})

	if err == nil {
		t.Fatal("AnalyzeManualInput error = nil, want validation error")
	}
}
