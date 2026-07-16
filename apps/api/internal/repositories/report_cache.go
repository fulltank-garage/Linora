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

const oauthStateTTL = 5 * time.Minute

type ReportCache interface {
	Delete(context.Context, string, string) error
	Get(context.Context, string, string) (models.AnalysisReport, bool, error)
	Set(context.Context, string, string, models.AnalysisReport) error
}

type RedisReportCache struct {
	client *redis.Client
}

func NewRedisReportCache(ctx context.Context, redisURL string) (*RedisReportCache, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &RedisReportCache{client: client}, nil
}

func (c *RedisReportCache) Close() error {
	return c.client.Close()
}

func (c *RedisReportCache) Delete(ctx context.Context, ownerID string, pageID string) error {
	return c.client.Del(ctx, reportCacheKey(ownerID, pageID)).Err()
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

func oauthStateKey(state string) string {
	return "linora:facebook:oauth-state:" + state
}
