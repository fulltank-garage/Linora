package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fulltank-garage/linora/apps/api/internal/models"
	"github.com/fulltank-garage/linora/apps/api/internal/repositories"
)

var ErrFacebookReconnectionRequired = errors.New("Facebook access needs to be connected again")

var bangkokLocation = time.FixedZone("Asia/Bangkok", 7*60*60)

type PageService struct {
	analysis *AnalysisService
	ai       *AIService
	cipher   *TokenCipher
	cache    repositories.ReportCache
	facebook *FacebookService
	store    repositories.Store
}

func NewPageService(store repositories.Store, cache repositories.ReportCache, cipher *TokenCipher, facebook *FacebookService, analysis *AnalysisService, ai *AIService) *PageService {
	return &PageService{analysis: analysis, ai: ai, cache: cache, cipher: cipher, facebook: facebook, store: store}
}

func (s *PageService) Connect(ctx context.Context, ownerID string, handoffCode string, pageID string) (models.ConnectedPageResponse, error) {
	if err := s.store.EnsureLineUser(ctx, ownerID); err != nil {
		return models.ConnectedPageResponse{}, err
	}
	page, authorizedPages, err := s.facebook.ConsumePages(handoffCode, ownerID, pageID)
	if err != nil {
		return models.ConnectedPageResponse{}, err
	}
	for _, authorizedPage := range authorizedPages {
		encryptedToken, err := s.cipher.Encrypt(authorizedPage.AccessToken)
		if err != nil {
			return models.ConnectedPageResponse{}, err
		}
		if err := s.store.UpsertConnection(ctx, repositories.PageConnection{
			AccessToken: encryptedToken,
			Category:    authorizedPage.Category,
			OwnerID:     ownerID,
			PageID:      authorizedPage.PageID,
			PageName:    authorizedPage.PageName,
		}); err != nil {
			return models.ConnectedPageResponse{}, err
		}
	}
	report, err := s.Sync(ctx, ownerID, page.PageID)
	if err != nil {
		return models.ConnectedPageResponse{}, err
	}
	if err := s.store.LinkPageToLineUser(ctx, ownerID, page.PageID); err != nil {
		return models.ConnectedPageResponse{}, err
	}
	page.AccessToken = ""
	return models.ConnectedPageResponse{Page: page, Report: report}, nil
}

func (s *PageService) List(ctx context.Context, ownerID string) ([]models.FacebookPage, error) {
	connections, err := s.store.ListConnections(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	pages := make([]models.FacebookPage, 0, len(connections))
	for _, connection := range connections {
		pages = append(pages, models.FacebookPage{
			Category: connection.Category,
			IsActive: true,
			PageID:   connection.PageID,
			PageName: connection.PageName,
		})
	}
	return pages, nil
}

func (s *PageService) Dashboard(ctx context.Context, ownerID string) (models.ConnectedPageResponse, error) {
	pageID, err := s.store.GetLinkedPage(ctx, ownerID)
	if err != nil {
		return models.ConnectedPageResponse{}, err
	}
	connection, err := s.store.GetConnection(ctx, ownerID, pageID)
	if err != nil {
		return models.ConnectedPageResponse{}, err
	}
	report, err := s.LatestReport(ctx, ownerID, pageID)
	if err != nil {
		return models.ConnectedPageResponse{}, err
	}
	return models.ConnectedPageResponse{
		Page: models.FacebookPage{
			Category: connection.Category,
			IsActive: true,
			PageID:   connection.PageID,
			PageName: connection.PageName,
		},
		Report: report,
	}, nil
}

func (s *PageService) WeeklyReport(ctx context.Context, ownerID string, pageID string) (models.WeeklyReport, error) {
	if _, err := s.store.GetConnection(ctx, ownerID, pageID); err != nil {
		return models.WeeklyReport{}, err
	}
	endDate := time.Now().In(bangkokLocation)
	startDate := endDate.AddDate(0, 0, -6)
	dailyMetrics, err := s.store.ListMetrics(ctx, ownerID, pageID, startDate, endDate)
	if err != nil {
		return models.WeeklyReport{}, err
	}
	weekly := models.WeeklyReport{
		DaysWithData: len(dailyMetrics),
		EndDate:      endDate.Format("2006-01-02"),
		StartDate:    startDate.Format("2006-01-02"),
	}
	for _, daily := range dailyMetrics {
		weekly.Metrics.Reach += daily.Metrics.Reach
		weekly.Metrics.Impressions += daily.Metrics.Impressions
		weekly.Metrics.Engagements += daily.Metrics.Engagements
		weekly.Metrics.Clicks += daily.Metrics.Clicks
	}
	return weekly, nil
}

func (s *PageService) Select(ctx context.Context, ownerID string, pageID string) (models.ConnectedPageResponse, error) {
	connection, err := s.store.GetConnection(ctx, ownerID, pageID)
	if err != nil {
		return models.ConnectedPageResponse{}, err
	}
	report, err := s.latestFreshReport(ctx, ownerID, pageID)
	if err != nil {
		report, err = s.Sync(ctx, ownerID, pageID)
	}
	if err != nil {
		return models.ConnectedPageResponse{}, err
	}
	if err := s.store.LinkPageToLineUser(ctx, ownerID, pageID); err != nil {
		return models.ConnectedPageResponse{}, err
	}
	return models.ConnectedPageResponse{
		Page: models.FacebookPage{
			Category: connection.Category,
			IsActive: true,
			PageID:   connection.PageID,
			PageName: connection.PageName,
		},
		Report: report,
	}, nil
}

func (s *PageService) Sync(ctx context.Context, ownerID string, pageID string) (models.AnalysisReport, error) {
	connection, err := s.store.GetConnection(ctx, ownerID, pageID)
	if err != nil {
		return models.AnalysisReport{}, err
	}
	accessToken, err := s.cipher.Decrypt(connection.AccessToken)
	if err != nil {
		return models.AnalysisReport{}, err
	}
	snapshot, err := s.facebook.FetchPageSnapshot(ctx, pageID, accessToken)
	if err != nil {
		if isFacebookTokenError(err) {
			_ = s.Delete(ctx, ownerID, pageID)
			return models.AnalysisReport{}, ErrFacebookReconnectionRequired
		}
		return models.AnalysisReport{}, err
	}
	page := models.FacebookPage{Category: connection.Category, IsActive: true, PageID: connection.PageID, PageName: connection.PageName}
	report := s.analysis.AnalyzePageSnapshot(page, snapshot)
	report = s.ai.EnhanceReport(ctx, report)
	if err := s.store.SaveMetrics(ctx, ownerID, pageID, snapshot.Metrics); err != nil {
		return models.AnalysisReport{}, err
	}
	if err := s.store.SaveReport(ctx, ownerID, pageID, report); err != nil {
		return models.AnalysisReport{}, err
	}
	if s.cache != nil {
		_ = s.cache.Set(ctx, ownerID, pageID, report)
	}
	return report, nil
}

func (s *PageService) LatestReport(ctx context.Context, ownerID string, pageID string) (models.AnalysisReport, error) {
	if s.cache != nil {
		if report, found, err := s.cache.Get(ctx, ownerID, pageID); err == nil && found {
			return report, nil
		}
	}
	return s.store.GetLatestReport(ctx, ownerID, pageID)
}

func (s *PageService) Delete(ctx context.Context, ownerID string, pageID string) error {
	if s.cache != nil {
		_ = s.cache.Delete(ctx, ownerID, pageID)
	}
	return s.store.DeletePage(ctx, ownerID, pageID)
}

func (s *PageService) Disconnect(ctx context.Context, ownerID string, pageID string) error {
	if s.cache != nil {
		_ = s.cache.Delete(ctx, ownerID, pageID)
	}
	return s.store.DisconnectPage(ctx, ownerID, pageID)
}

func (s *PageService) CreateLineLinkCode(ctx context.Context, ownerID string, pageID string) (string, error) {
	if _, err := s.store.GetConnection(ctx, ownerID, pageID); err != nil {
		return "", err
	}
	code, err := SecureToken()
	if err != nil {
		return "", err
	}
	shortCode := fmt.Sprintf("LIN-%s", code[:8])
	return shortCode, s.store.CreateLinkCode(ctx, ownerID, shortCode, pageID, time.Now().Add(10*time.Minute))
}

func (s *PageService) latestFreshReport(ctx context.Context, ownerID string, pageID string) (models.AnalysisReport, error) {
	report, err := s.LatestReport(ctx, ownerID, pageID)
	if err != nil {
		return models.AnalysisReport{}, err
	}
	createdAt, err := time.Parse(time.RFC3339, report.CreatedAt)
	if err != nil || time.Since(createdAt) > 15*time.Minute {
		return models.AnalysisReport{}, repositories.ErrNotFound
	}
	return report, nil
}

func isFacebookTokenError(err error) bool {
	return IsFacebookAccessTokenError(err)
}
