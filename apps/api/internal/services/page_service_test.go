package services

import (
	"testing"
	"time"

	"github.com/fulltank-garage/linora/apps/api/internal/models"
)

func TestSnapshotFingerprintIsStableAndTracksChanges(t *testing.T) {
	snapshot := models.PageSnapshot{
		Metrics: models.PageMetrics{Engagements: 12, Impressions: 120},
		Posts:   []models.FacebookPost{{ID: "post-1", Reactions: 10, Comments: 2}},
	}

	first := snapshotFingerprint(snapshot)
	second := snapshotFingerprint(snapshot)
	if first == "" || first != second {
		t.Fatalf("expected identical snapshots to have the same fingerprint, got %q and %q", first, second)
	}

	snapshot.Posts[0].Comments++
	if changed := snapshotFingerprint(snapshot); changed == first {
		t.Fatal("expected changed Facebook data to require a new analysis fingerprint")
	}
}

func TestPageServiceReportFreshness(t *testing.T) {
	service := NewPageService(nil, nil, nil, nil, nil, nil)
	if !service.isReportFresh(models.AnalysisReport{CreatedAt: time.Now().UTC().Format(time.RFC3339)}) {
		t.Fatal("expected a newly created report to be fresh")
	}
	if service.isReportFresh(models.AnalysisReport{CreatedAt: time.Now().Add(-analysisRefreshInterval - time.Second).UTC().Format(time.RFC3339)}) {
		t.Fatal("expected an expired report to be refreshed in the background")
	}
}

func TestPageServiceFailedStatusHasRetryDelay(t *testing.T) {
	service := NewPageService(nil, nil, nil, nil, nil, nil)
	if !service.isRecentStatus(models.AnalysisStatus{State: "failed", UpdatedAt: time.Now().UTC().Format(time.RFC3339)}) {
		t.Fatal("expected a recent failed job to wait before retrying")
	}
	if service.isRecentStatus(models.AnalysisStatus{State: "failed", UpdatedAt: time.Now().Add(-analysisRetryDelay - time.Second).UTC().Format(time.RFC3339)}) {
		t.Fatal("expected an older failed job to be eligible for retry")
	}
}
