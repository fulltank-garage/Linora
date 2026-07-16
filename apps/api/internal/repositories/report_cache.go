package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fulltank-garage/linora/apps/api/internal/models"
	"github.com/redis/go-redis/v9"
)

const reportCacheTTL = 15 * time.Minute

const analysisStatusTTL = time.Hour

const analysisJobLockTTL = 5 * time.Minute

const oauthStateTTL = 5 * time.Minute

type ReportCache interface {
	Delete(context.Context, string, string) error
	Get(context.Context, string, string) (models.AnalysisReport, bool, error)
	Set(context.Context, string, string, models.AnalysisReport) error
}

type AnalysisStatusCache interface {
	GetAnalysisStatus(context.Context, string, string) (models.AnalysisStatus, bool, error)
	SetAnalysisStatus(context.Context, string, string, models.AnalysisStatus) error
}

type AnalysisJob struct {
	OwnerID string `json:"ownerId"`
	PageID  string `json:"pageId"`
}

// AnalysisJobQueue persists background analysis work in Redis. Jobs are moved
// to a processing list before they run so they can be recovered after restart.
type AnalysisJobQueue interface {
	AcknowledgeAnalysis(context.Context, AnalysisJob) error
	CancelAnalysis(context.Context, string, string) error
	DequeueAnalysis(context.Context) (AnalysisJob, bool, error)
	EnqueueAnalysis(context.Context, string, string) (bool, error)
	RecoverAnalysisJobs(context.Context) error
}

type RedisReportCache struct {
	analysisQueuePrefix string
	client              *redis.Client
}

func NewRedisReportCache(ctx context.Context, redisURL string) (*RedisReportCache, error) {
	return newRedisReportCache(ctx, redisURL, "linora:analysis")
}

func newRedisReportCache(ctx context.Context, redisURL string, analysisQueuePrefix string) (*RedisReportCache, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &RedisReportCache{analysisQueuePrefix: analysisQueuePrefix, client: client}, nil
}

func (c *RedisReportCache) Close() error {
	return c.client.Close()
}

func (c *RedisReportCache) Delete(ctx context.Context, ownerID string, pageID string) error {
	return c.client.Del(ctx, reportCacheKey(ownerID, pageID), analysisStatusCacheKey(ownerID, pageID)).Err()
}

func (c *RedisReportCache) Get(ctx context.Context, ownerID string, pageID string) (models.AnalysisReport, bool, error) {
	payload, err := c.client.Get(ctx, reportCacheKey(ownerID, pageID)).Bytes()
	if err == redis.Nil {
		return models.AnalysisReport{}, false, nil
	}
	if err != nil {
		return models.AnalysisReport{}, false, err
	}
	var report models.AnalysisReport
	if err := json.Unmarshal(payload, &report); err != nil {
		return models.AnalysisReport{}, false, err
	}
	return report, true, nil
}

func (c *RedisReportCache) Set(ctx context.Context, ownerID string, pageID string, report models.AnalysisReport) error {
	payload, err := json.Marshal(report)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, reportCacheKey(ownerID, pageID), payload, reportCacheTTL).Err()
}

func (c *RedisReportCache) GetAnalysisStatus(ctx context.Context, ownerID string, pageID string) (models.AnalysisStatus, bool, error) {
	payload, err := c.client.Get(ctx, analysisStatusCacheKey(ownerID, pageID)).Bytes()
	if err == redis.Nil {
		return models.AnalysisStatus{}, false, nil
	}
	if err != nil {
		return models.AnalysisStatus{}, false, err
	}
	var status models.AnalysisStatus
	if err := json.Unmarshal(payload, &status); err != nil {
		return models.AnalysisStatus{}, false, err
	}
	return status, true, nil
}

func (c *RedisReportCache) SetAnalysisStatus(ctx context.Context, ownerID string, pageID string, status models.AnalysisStatus) error {
	payload, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, analysisStatusCacheKey(ownerID, pageID), payload, analysisStatusTTL).Err()
}

func (c *RedisReportCache) EnqueueAnalysis(ctx context.Context, ownerID string, pageID string) (bool, error) {
	job := AnalysisJob{OwnerID: ownerID, PageID: pageID}
	payload, err := json.Marshal(job)
	if err != nil {
		return false, err
	}
	queued, err := c.client.SetNX(ctx, c.analysisJobLockKey(ownerID, pageID), "1", analysisJobLockTTL).Result()
	if err != nil || !queued {
		return queued, err
	}
	if err := c.client.LPush(ctx, c.analysisJobQueueKey(), payload).Err(); err != nil {
		_ = c.client.Del(ctx, c.analysisJobLockKey(ownerID, pageID)).Err()
		return false, err
	}
	return true, nil
}

func (c *RedisReportCache) DequeueAnalysis(ctx context.Context) (AnalysisJob, bool, error) {
	payload, err := c.client.BRPopLPush(ctx, c.analysisJobQueueKey(), c.analysisProcessingQueueKey(), 5*time.Second).Result()
	if err == redis.Nil {
		return AnalysisJob{}, false, nil
	}
	if err != nil {
		return AnalysisJob{}, false, err
	}
	var job AnalysisJob
	if err := json.Unmarshal([]byte(payload), &job); err != nil {
		_ = c.client.LRem(ctx, c.analysisProcessingQueueKey(), 1, payload).Err()
		return AnalysisJob{}, false, err
	}
	return job, true, nil
}

func (c *RedisReportCache) AcknowledgeAnalysis(ctx context.Context, job AnalysisJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}
	_, err = c.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.LRem(ctx, c.analysisProcessingQueueKey(), 1, payload)
		pipe.Del(ctx, c.analysisJobLockKey(job.OwnerID, job.PageID))
		return nil
	})
	return err
}

func (c *RedisReportCache) CancelAnalysis(ctx context.Context, ownerID string, pageID string) error {
	payload, err := json.Marshal(AnalysisJob{OwnerID: ownerID, PageID: pageID})
	if err != nil {
		return err
	}
	_, err = c.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.LRem(ctx, c.analysisJobQueueKey(), 0, payload)
		pipe.LRem(ctx, c.analysisProcessingQueueKey(), 0, payload)
		pipe.Del(ctx, c.analysisJobLockKey(ownerID, pageID))
		return nil
	})
	return err
}

func (c *RedisReportCache) RecoverAnalysisJobs(ctx context.Context) error {
	jobs, err := c.client.LRange(ctx, c.analysisProcessingQueueKey(), 0, -1).Result()
	if err != nil || len(jobs) == 0 {
		return err
	}
	_, err = c.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.LPush(ctx, c.analysisJobQueueKey(), jobs)
		pipe.Del(ctx, c.analysisProcessingQueueKey())
		return nil
	})
	return err
}

// SaveOAuthState stores the one-time Facebook OAuth state outside the API
// process so an application restart does not invalidate an in-progress login.
func (c *RedisReportCache) SaveOAuthState(ctx context.Context, state string, ownerID string) error {
	return c.client.Set(ctx, oauthStateKey(state), ownerID, oauthStateTTL).Err()
}

// ConsumeOAuthState returns and removes a one-time Facebook OAuth state.
func (c *RedisReportCache) ConsumeOAuthState(ctx context.Context, state string) (string, bool, error) {
	ownerID, err := c.client.GetDel(ctx, oauthStateKey(state)).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return ownerID, true, nil
}

func reportCacheKey(ownerID string, pageID string) string {
	return fmt.Sprintf("linora:user:%s:page:%s:latest-report", ownerID, pageID)
}

func analysisStatusCacheKey(ownerID string, pageID string) string {
	return fmt.Sprintf("linora:user:%s:page:%s:analysis-status", ownerID, pageID)
}

func (c *RedisReportCache) analysisJobLockKey(ownerID string, pageID string) string {
	return fmt.Sprintf("%s:user:%s:page:%s:job", c.analysisQueuePrefix, ownerID, pageID)
}

func (c *RedisReportCache) analysisJobQueueKey() string {
	return c.analysisQueuePrefix + ":queue"
}

func (c *RedisReportCache) analysisProcessingQueueKey() string {
	return c.analysisQueuePrefix + ":processing"
}

func oauthStateKey(state string) string {
	return "linora:facebook:oauth-state:" + state
}
