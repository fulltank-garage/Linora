package repositories

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestRedisAnalysisQueue(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL is required for Redis integration testing")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cache, err := newRedisReportCache(ctx, redisURL, fmt.Sprintf("linora:test:analysis:%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()

	job := AnalysisJob{OwnerID: "test-analysis-owner", PageID: "test-analysis-page"}
	defer cache.CancelAnalysis(context.Background(), job.OwnerID, job.PageID)
	if err := cache.CancelAnalysis(ctx, job.OwnerID, job.PageID); err != nil {
		t.Fatal(err)
	}
	queued, err := cache.EnqueueAnalysis(ctx, job.OwnerID, job.PageID)
	if err != nil || !queued {
		t.Fatalf("expected job to be queued, queued=%v err=%v", queued, err)
	}
	duplicate, err := cache.EnqueueAnalysis(ctx, job.OwnerID, job.PageID)
	if err != nil || duplicate {
		t.Fatalf("expected duplicate job to be ignored, queued=%v err=%v", duplicate, err)
	}

	dequeued, found, err := cache.DequeueAnalysis(ctx)
	if err != nil || !found {
		t.Fatalf("expected a queued job, found=%v err=%v", found, err)
	}
	if dequeued != job {
		t.Fatalf("unexpected job: %#v", dequeued)
	}
	if err := cache.AcknowledgeAnalysis(ctx, dequeued); err != nil {
		t.Fatal(err)
	}
}
