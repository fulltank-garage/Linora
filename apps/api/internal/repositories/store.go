package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/fulltank-garage/linora/apps/api/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("record not found")

type PageConnection struct {
	AccessToken string
	Category    string
	ConnectedAt time.Time
	OwnerID     string
	PageID      string
	PageName    string
	SyncedAt    *time.Time
}

type Store interface {
	CreateLinkCode(context.Context, string, string, string, time.Time) error
	DeletePage(context.Context, string, string) error
	DisconnectPage(context.Context, string, string) error
	EnsureLineUser(context.Context, string) error
	GetConnection(context.Context, string, string) (PageConnection, error)
	GetLatestReport(context.Context, string, string) (models.AnalysisReport, error)
	GetLinkedPage(context.Context, string) (string, error)
	ListMetrics(context.Context, string, string, time.Time, time.Time) ([]models.DailyPageMetrics, error)
	LinkPageToLineUser(context.Context, string, string) error
	ListConnections(context.Context, string) ([]PageConnection, error)
	Migrate(context.Context) error
	SaveMetrics(context.Context, string, string, models.PageMetrics) error
	SaveReport(context.Context, string, string, models.AnalysisReport) error
	UpsertConnection(context.Context, PageConnection) error
	UseLinkCode(context.Context, string, string) (string, error)
}

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Close() { s.pool.Close() }

func (s *PostgresStore) Migrate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS line_users (
			line_user_id TEXT PRIMARY KEY,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS page_connections (
			owner_id TEXT NOT NULL,
			page_id TEXT NOT NULL,
			page_name TEXT NOT NULL,
			category TEXT NOT NULL,
			encrypted_access_token TEXT NOT NULL,
			connected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			last_synced_at TIMESTAMPTZ,
			PRIMARY KEY (owner_id, page_id)
		);
		CREATE TABLE IF NOT EXISTS page_metrics (
			owner_id TEXT NOT NULL,
			page_id TEXT NOT NULL,
			recorded_on DATE NOT NULL,
			reach BIGINT NOT NULL DEFAULT 0,
			impressions BIGINT NOT NULL DEFAULT 0,
			engagements BIGINT NOT NULL DEFAULT 0,
			clicks BIGINT NOT NULL DEFAULT 0,
			PRIMARY KEY (owner_id, page_id, recorded_on)
		);
		CREATE TABLE IF NOT EXISTS analysis_reports (
			id TEXT PRIMARY KEY,
			owner_id TEXT NOT NULL,
			page_id TEXT NOT NULL,
			payload JSONB NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
		CREATE TABLE IF NOT EXISTS line_link_codes (
			code TEXT PRIMARY KEY,
			owner_id TEXT NOT NULL,
			page_id TEXT NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			used_at TIMESTAMPTZ
		);
		CREATE TABLE IF NOT EXISTS line_page_links (
			line_user_id TEXT PRIMARY KEY,
			page_id TEXT NOT NULL,
			linked_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);

		ALTER TABLE page_connections ADD COLUMN IF NOT EXISTS owner_id TEXT NOT NULL DEFAULT 'legacy';
		ALTER TABLE page_metrics ADD COLUMN IF NOT EXISTS owner_id TEXT NOT NULL DEFAULT 'legacy';
		ALTER TABLE analysis_reports ADD COLUMN IF NOT EXISTS owner_id TEXT NOT NULL DEFAULT 'legacy';
		ALTER TABLE line_link_codes ADD COLUMN IF NOT EXISTS owner_id TEXT NOT NULL DEFAULT 'legacy';
		ALTER TABLE page_metrics DROP CONSTRAINT IF EXISTS page_metrics_page_id_fkey;
		ALTER TABLE analysis_reports DROP CONSTRAINT IF EXISTS analysis_reports_page_id_fkey;
		ALTER TABLE line_link_codes DROP CONSTRAINT IF EXISTS line_link_codes_page_id_fkey;
		ALTER TABLE line_page_links DROP CONSTRAINT IF EXISTS line_page_links_page_id_fkey;
		ALTER TABLE page_connections DROP CONSTRAINT IF EXISTS page_connections_pkey;
		ALTER TABLE page_connections ADD PRIMARY KEY (owner_id, page_id);
		ALTER TABLE page_metrics DROP CONSTRAINT IF EXISTS page_metrics_pkey;
		ALTER TABLE page_metrics ADD PRIMARY KEY (owner_id, page_id, recorded_on);
		CREATE INDEX IF NOT EXISTS analysis_reports_owner_page_created_idx ON analysis_reports(owner_id, page_id, created_at DESC);
	`)
	return err
}

func (s *PostgresStore) EnsureLineUser(ctx context.Context, ownerID string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO line_users (line_user_id) VALUES ($1)
		ON CONFLICT (line_user_id) DO UPDATE SET last_seen_at = now()
	`, ownerID)
	return err
}

func (s *PostgresStore) UpsertConnection(ctx context.Context, connection PageConnection) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO page_connections (owner_id, page_id, page_name, category, encrypted_access_token)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (owner_id, page_id) DO UPDATE SET
			page_name = EXCLUDED.page_name,
			category = EXCLUDED.category,
			encrypted_access_token = EXCLUDED.encrypted_access_token,
			connected_at = now()
	`, connection.OwnerID, connection.PageID, connection.PageName, connection.Category, connection.AccessToken)
	return err
}

func (s *PostgresStore) GetConnection(ctx context.Context, ownerID string, pageID string) (PageConnection, error) {
	var connection PageConnection
	err := s.pool.QueryRow(ctx, `
		SELECT owner_id, page_id, page_name, category, encrypted_access_token, connected_at, last_synced_at
		FROM page_connections WHERE owner_id = $1 AND page_id = $2
	`, ownerID, pageID).Scan(&connection.OwnerID, &connection.PageID, &connection.PageName, &connection.Category, &connection.AccessToken, &connection.ConnectedAt, &connection.SyncedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return PageConnection{}, ErrNotFound
	}
	return connection, err
}

func (s *PostgresStore) ListConnections(ctx context.Context, ownerID string) ([]PageConnection, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT owner_id, page_id, page_name, category, encrypted_access_token, connected_at, last_synced_at
		FROM page_connections WHERE owner_id = $1 ORDER BY page_name ASC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	connections := make([]PageConnection, 0)
	for rows.Next() {
		var connection PageConnection
		if err := rows.Scan(&connection.OwnerID, &connection.PageID, &connection.PageName, &connection.Category, &connection.AccessToken, &connection.ConnectedAt, &connection.SyncedAt); err != nil {
			return nil, err
		}
		connections = append(connections, connection)
	}
	return connections, rows.Err()
}

func (s *PostgresStore) SaveMetrics(ctx context.Context, ownerID string, pageID string, metrics models.PageMetrics) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO page_metrics (owner_id, page_id, recorded_on, reach, impressions, engagements, clicks)
		VALUES ($1, $2, (now() AT TIME ZONE 'Asia/Bangkok')::date, $3, $4, $5, $6)
		ON CONFLICT (owner_id, page_id, recorded_on) DO UPDATE SET
			reach = EXCLUDED.reach, impressions = EXCLUDED.impressions,
			engagements = EXCLUDED.engagements, clicks = EXCLUDED.clicks
	`, ownerID, pageID, metrics.Reach, metrics.Impressions, metrics.Engagements, metrics.Clicks)
	return err
}

func (s *PostgresStore) ListMetrics(ctx context.Context, ownerID string, pageID string, startDate time.Time, endDate time.Time) ([]models.DailyPageMetrics, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT recorded_on, reach, impressions, engagements, clicks
		FROM page_metrics
		WHERE owner_id = $1 AND page_id = $2 AND recorded_on BETWEEN $3::date AND $4::date
		ORDER BY recorded_on ASC
	`, ownerID, pageID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := make([]models.DailyPageMetrics, 0, 7)
	for rows.Next() {
		var recordedOn time.Time
		var item models.DailyPageMetrics
		if err := rows.Scan(&recordedOn, &item.Metrics.Reach, &item.Metrics.Impressions, &item.Metrics.Engagements, &item.Metrics.Clicks); err != nil {
			return nil, err
		}
		item.RecordedOn = recordedOn.Format("2006-01-02")
		metrics = append(metrics, item)
	}
	return metrics, rows.Err()
}

func (s *PostgresStore) SaveReport(ctx context.Context, ownerID string, pageID string, report models.AnalysisReport) error {
	payload, err := json.Marshal(report)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO analysis_reports (id, owner_id, page_id, payload) VALUES ($1, $2, $3, $4)`, report.ID, ownerID, pageID, payload)
	return err
}

func (s *PostgresStore) GetLatestReport(ctx context.Context, ownerID string, pageID string) (models.AnalysisReport, error) {
	var payload []byte
	err := s.pool.QueryRow(ctx, `SELECT payload FROM analysis_reports WHERE owner_id = $1 AND page_id = $2 ORDER BY created_at DESC LIMIT 1`, ownerID, pageID).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.AnalysisReport{}, ErrNotFound
	}
	if err != nil {
		return models.AnalysisReport{}, err
	}
	var report models.AnalysisReport
	if err := json.Unmarshal(payload, &report); err != nil {
		return models.AnalysisReport{}, err
	}
	return report, nil
}

func (s *PostgresStore) CreateLinkCode(ctx context.Context, ownerID string, code string, pageID string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO line_link_codes (code, owner_id, page_id, expires_at) VALUES ($1, $2, $3, $4)`, code, ownerID, pageID, expiresAt)
	return err
}

func (s *PostgresStore) UseLinkCode(ctx context.Context, code string, lineUserID string) (string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	var pageID string
	err = tx.QueryRow(ctx, `
		UPDATE line_link_codes SET used_at = now()
		WHERE code = $1 AND owner_id = $2 AND used_at IS NULL AND expires_at > now()
		RETURNING page_id
	`, code, lineUserID).Scan(&pageID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO line_page_links (line_user_id, page_id) VALUES ($1, $2)
		ON CONFLICT (line_user_id) DO UPDATE SET page_id = EXCLUDED.page_id, linked_at = now()
	`, lineUserID, pageID)
	if err != nil {
		return "", err
	}
	return pageID, tx.Commit(ctx)
}

func (s *PostgresStore) GetLinkedPage(ctx context.Context, lineUserID string) (string, error) {
	var pageID string
	err := s.pool.QueryRow(ctx, `SELECT page_id FROM line_page_links WHERE line_user_id = $1`, lineUserID).Scan(&pageID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return pageID, err
}

func (s *PostgresStore) LinkPageToLineUser(ctx context.Context, lineUserID string, pageID string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO line_page_links (line_user_id, page_id) VALUES ($1, $2)
		ON CONFLICT (line_user_id) DO UPDATE SET page_id = EXCLUDED.page_id, linked_at = now()
	`, lineUserID, pageID)
	return err
}

func (s *PostgresStore) DeletePage(ctx context.Context, ownerID string, pageID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, `DELETE FROM line_page_links WHERE line_user_id = $1 AND page_id = $2`, ownerID, pageID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM line_link_codes WHERE owner_id = $1 AND page_id = $2`, ownerID, pageID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM page_metrics WHERE owner_id = $1 AND page_id = $2`, ownerID, pageID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM analysis_reports WHERE owner_id = $1 AND page_id = $2`, ownerID, pageID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM page_connections WHERE owner_id = $1 AND page_id = $2`, ownerID, pageID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *PostgresStore) DisconnectPage(ctx context.Context, ownerID string, pageID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, `DELETE FROM line_page_links WHERE line_user_id = $1 AND page_id = $2`, ownerID, pageID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM line_link_codes WHERE owner_id = $1 AND page_id = $2`, ownerID, pageID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM page_connections WHERE owner_id = $1 AND page_id = $2`, ownerID, pageID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
