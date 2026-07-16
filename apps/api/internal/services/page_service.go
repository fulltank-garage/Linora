package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fulltank-garage/linora/apps/api/internal/models"
	"github.com/fulltank-garage/linora/apps/api/internal/repositories"
)

var ErrFacebookReconnectionRequired = errors.New("Facebook access needs to be connected again")

var bangkokLocation = time.FixedZone("Asia/Bangkok", 7*60*60)

const analysisRefreshInterval = 15 * time.Minute
const analysisRetryDelay = time.Minute

type PageService struct {
	analysis *AnalysisService
	ai       *AIService
	cipher   *TokenCipher
	cache    repositories.ReportCache
	facebook *FacebookService
	jobs     map[string]struct{}
	jobsMu   sync.Mutex
	queue    repositories.AnalysisJobQueue
	store    repositories.Store
}

func NewPageService(store repositories.Store, cache repositories.ReportCache, cipher *TokenCipher, facebook *FacebookService, analysis *AnalysisService, ai *AIService) *PageService {
	service := &PageService{analysis: analysis, ai: ai, cache: cache, cipher: cipher, facebook: facebook, jobs: make(map[string]struct{}), store: store}
	if queue, ok := cache.(repositories.AnalysisJobQueue); ok {
		service.queue = queue
	}
	return service
}

func (s *PageService) Connect(ctx context.Context, ownerID string, handoffCode string, pageID string) (models.ConnectedPageResponse, error) {
	if err := s.store.EnsureLineUser(ctx, ownerID); err != nil {
		return models.ConnectedPageResponse{}, err
	}
	page, authorizedPages, facebookUserID, err := s.facebook.ConsumeConnection(handoffCode, ownerID, pageID)
	if err != nil {
		return models.ConnectedPageResponse{}, err
	}
	if err := s.store.LinkFacebookUser(ctx, ownerID, facebookUserID); err != nil {
		return models.ConnectedPageResponse{}, err
	}
	for _, authorizedPage := range authorizedPages {
		encryptedToken, err := s.cipher.Encrypt(authorizedPage.AccessToken)
		if err != nil {
			return models.ConnectedPageResponse{}, err
		}
		if err := s.store.UpsertConnection(ctx, repositories.PageConnection{
			AccessToken:    encryptedToken,
			Category:       authorizedPage.Category,
			FacebookUserID: facebookUserID,
			OwnerID:        ownerID,
			PageID:         authorizedPage.PageID,
			PageName:       authorizedPage.PageName,
		}); err != nil {
			return models.ConnectedPageResponse{}, err
		}
	}
	if err := s.store.LinkPageToLineUser(ctx, ownerID, page.PageID); err != nil {
		return models.ConnectedPageResponse{}, err
	}
	page.AccessToken = ""
	return s.pageResponse(ctx, ownerID, page)
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
	return s.pageResponse(ctx, ownerID, models.FacebookPage{
		Category: connection.Category,
		IsActive: true,
		PageID:   connection.PageID,
		PageName: connection.PageName,
	})
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
	if err := s.store.LinkPageToLineUser(ctx, ownerID, pageID); err != nil {
		return models.ConnectedPageResponse{}, err
	}
	return s.pageResponse(ctx, ownerID, models.FacebookPage{
		Category: connection.Category,
		IsActive: true,
		PageID:   connection.PageID,
		PageName: connection.PageName,
	})
}

// pageResponse returns immediately with the most recent report, then queues a
// refresh when the report is missing or older than the cache window.
func (s *PageService) pageResponse(ctx context.Context, ownerID string, page models.FacebookPage) (models.ConnectedPageResponse, error) {
	response := models.ConnectedPageResponse{Page: page, AnalysisStatus: s.analysisStatus(ctx, ownerID, page.PageID)}
	report, err := s.LatestReport(ctx, ownerID, page.PageID)
	if err == nil {
		response.Report = &report
		if s.isReportFresh(report) {
			response.AnalysisStatus = models.AnalysisStatus{State: "ready", UpdatedAt: report.CreatedAt}
			return response, nil
		}
	}
	if response.AnalysisStatus.State == "failed" && s.isRecentStatus(response.AnalysisStatus) {
		return response, nil
	}

	state := "queued"
	if response.Report != nil {
		state = "refreshing"
	}
	response.AnalysisStatus = models.AnalysisStatus{State: state, UpdatedAt: time.Now().UTC().Format(time.RFC3339)}
	s.setAnalysisStatus(context.Background(), ownerID, page.PageID, response.AnalysisStatus)
	s.scheduleSync(ownerID, page.PageID)
	return response, nil
}

func (s *PageService) scheduleSync(ownerID string, pageID string) {
	if s.queue != nil {
		if _, err := s.queue.EnqueueAnalysis(context.Background(), ownerID, pageID); err != nil {
			log.Printf("queue analysis for page %s: %v", pageID, err)
		}
		return
	}
	s.runInProcessSync(ownerID, pageID)
}

// StartBackgroundWorker consumes durable Redis jobs. It is intentionally kept
// inside the API service so Railway needs no separate worker deployment.
func (s *PageService) StartBackgroundWorker(ctx context.Context) {
	if s.queue == nil {
		log.Printf("analysis worker is using the in-process fallback because Redis is unavailable")
		return
	}
	go func() {
		log.Printf("Redis analysis worker started")
		if err := s.queue.RecoverAnalysisJobs(ctx); err != nil {
			log.Printf("recover queued analysis jobs: %v", err)
		}
		for {
			job, found, err := s.queue.DequeueAnalysis(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("dequeue analysis job: %v", err)
				continue
			}
			if !found {
				continue
			}
			s.runSyncJob(job.OwnerID, job.PageID)
			if err := s.queue.AcknowledgeAnalysis(context.Background(), job); err != nil {
				log.Printf("acknowledge analysis job for page %s: %v", job.PageID, err)
			}
		}
	}()
}

func (s *PageService) runInProcessSync(ownerID string, pageID string) {
	key := ownerID + ":" + pageID
	s.jobsMu.Lock()
	if _, running := s.jobs[key]; running {
		s.jobsMu.Unlock()
		return
	}
	s.jobs[key] = struct{}{}
	s.jobsMu.Unlock()

	go func() {
		defer func() {
			s.jobsMu.Lock()
			delete(s.jobs, key)
			s.jobsMu.Unlock()
		}()
		s.runSyncJob(ownerID, pageID)
	}()
}

func (s *PageService) runSyncJob(ownerID string, pageID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	s.setAnalysisStatus(ctx, ownerID, pageID, models.AnalysisStatus{State: "running", UpdatedAt: time.Now().UTC().Format(time.RFC3339)})
	report, err := s.Sync(ctx, ownerID, pageID)
	if err != nil {
		log.Printf("analyze page %s: %v", pageID, err)
		s.setAnalysisStatus(context.Background(), ownerID, pageID, models.AnalysisStatus{State: "failed", UpdatedAt: time.Now().UTC().Format(time.RFC3339)})
		return
	}
	s.setAnalysisStatus(context.Background(), ownerID, pageID, models.AnalysisStatus{State: "ready", UpdatedAt: report.CreatedAt})
}

func (s *PageService) analysisStatus(ctx context.Context, ownerID string, pageID string) models.AnalysisStatus {
	if store, ok := s.cache.(repositories.AnalysisStatusCache); ok {
		if status, found, err := store.GetAnalysisStatus(ctx, ownerID, pageID); err == nil && found {
			return status
		}
	}
	return models.AnalysisStatus{State: "queued", UpdatedAt: time.Now().UTC().Format(time.RFC3339)}
}

func (s *PageService) setAnalysisStatus(ctx context.Context, ownerID string, pageID string, status models.AnalysisStatus) {
	if store, ok := s.cache.(repositories.AnalysisStatusCache); ok {
		_ = store.SetAnalysisStatus(ctx, ownerID, pageID, status)
	}
}

func (s *PageService) isReportFresh(report models.AnalysisReport) bool {
	createdAt, err := time.Parse(time.RFC3339, report.CreatedAt)
	return err == nil && time.Since(createdAt) <= analysisRefreshInterval
}

func (s *PageService) isRecentStatus(status models.AnalysisStatus) bool {
	updatedAt, err := time.Parse(time.RFC3339, status.UpdatedAt)
	return err == nil && time.Since(updatedAt) < analysisRetryDelay
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
	if previous, err := s.LatestReport(ctx, ownerID, pageID); err == nil && previous.SourceFingerprint == snapshotFingerprint(snapshot) {
		previous.CreatedAt = time.Now().UTC().Format(time.RFC3339)
		if err := s.store.SaveMetrics(ctx, ownerID, pageID, snapshot.Metrics); err != nil {
			return models.AnalysisReport{}, err
		}
		if err := s.store.TouchReport(ctx, ownerID, pageID, previous); err != nil {
			return models.AnalysisReport{}, err
		}
		if s.cache != nil {
			_ = s.cache.Set(ctx, ownerID, pageID, previous)
		}
		return previous, nil
	}
	page := models.FacebookPage{Category: connection.Category, IsActive: true, PageID: connection.PageID, PageName: connection.PageName}
	report := s.analysis.AnalyzePageSnapshot(page, snapshot)
	report.SourceFingerprint = snapshotFingerprint(snapshot)
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

func snapshotFingerprint(snapshot models.PageSnapshot) string {
	payload, err := json.Marshal(struct {
		Metrics models.PageMetrics    `json:"metrics"`
		Posts   []models.FacebookPost `json:"posts"`
	}{Metrics: snapshot.Metrics, Posts: snapshot.Posts})
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
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
	if s.queue != nil {
		_ = s.queue.CancelAnalysis(ctx, ownerID, pageID)
	}
	if s.cache != nil {
		_ = s.cache.Delete(ctx, ownerID, pageID)
	}
	return s.store.DeletePage(ctx, ownerID, pageID)
}

func (s *PageService) DeleteFacebookUserData(ctx context.Context, facebookUserID string) error {
	connections, err := s.store.DeleteFacebookUserData(ctx, facebookUserID)
	if err != nil {
		return err
	}
	if s.cache != nil {
		for _, connection := range connections {
			_ = s.cache.Delete(ctx, connection.OwnerID, connection.PageID)
			if s.queue != nil {
				_ = s.queue.CancelAnalysis(ctx, connection.OwnerID, connection.PageID)
			}
		}
	}
	return nil
}

func (s *PageService) Disconnect(ctx context.Context, ownerID string, pageID string) error {
	if s.queue != nil {
		_ = s.queue.CancelAnalysis(ctx, ownerID, pageID)
	}
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
