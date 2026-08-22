package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"lostandfound/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Postgres struct {
	pool *pgxpool.Pool
}

func Connect(ctx context.Context, url string) (*Postgres, error) {
	var pool *pgxpool.Pool
	var err error
	for i := 0; i < 20; i++ {
		pool, err = pgxpool.New(ctx, url)
		if err == nil {
			err = pool.Ping(ctx)
		}
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	db := &Postgres{pool: pool}
	if err := db.migrate(ctx); err != nil {
		return nil, err
	}
	return db, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) migrate(ctx context.Context) error {
	_, err := p.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS reports (
			id UUID PRIMARY KEY,
			type VARCHAR(16) NOT NULL,
			category VARCHAR(64) NOT NULL,
			title VARCHAR(255) NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			unique_features TEXT NOT NULL DEFAULT '',
			location VARCHAR(255) NOT NULL,
			location_details TEXT NOT NULL DEFAULT '',
			incident_date DATE NOT NULL,
			photos JSONB NOT NULL DEFAULT '[]',
			phone VARCHAR(64) NOT NULL DEFAULT '',
			telegram VARCHAR(128) NOT NULL DEFAULT '',
			status VARCHAR(32) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_reports_match
			ON reports (type, category, status, incident_date);
		CREATE TABLE IF NOT EXISTS matches (
			id UUID PRIMARY KEY,
			lost_report_id UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
			found_report_id UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
			groq_score DOUBLE PRECISION,
			gemini_score DOUBLE PRECISION,
			combined_score DOUBLE PRECISION NOT NULL,
			reasoning TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (lost_report_id, found_report_id)
		);
	`)
	return err
}

func (p *Postgres) Create(ctx context.Context, report domain.Report) error {
	photos, err := json.Marshal(report.Photos)
	if err != nil {
		return err
	}
	if photos == nil {
		photos = []byte("[]")
	}
	_, err = p.pool.Exec(ctx, `
		INSERT INTO reports (
			id, type, category, title, description, unique_features,
			location, location_details, incident_date, photos, phone, telegram, status, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`, report.ID, report.Type, report.Category, report.Title, report.Description, report.UniqueFeatures,
		report.Location, report.LocationDetails, report.IncidentDate, photos, report.Phone, report.Telegram,
		report.Status, report.CreatedAt)
	return err
}

func (p *Postgres) GetByID(ctx context.Context, id string) (domain.Report, error) {
	row := p.pool.QueryRow(ctx, `
		SELECT id, type, category, title, description, unique_features,
			location, location_details, incident_date, photos, phone, telegram, status, created_at
		FROM reports WHERE id = $1
	`, id)
	report, err := scanReport(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Report{}, ErrNotFound
	}
	return report, err
}

func (p *Postgres) UpdateStatus(ctx context.Context, id string, status domain.Status) error {
	tag, err := p.pool.Exec(ctx, `UPDATE reports SET status = $2 WHERE id = $1`, id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) FindFoundCandidates(ctx context.Context, category string, lostDate time.Time, windowDays int) ([]domain.Report, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT id, type, category, title, description, unique_features,
			location, location_details, incident_date, photos, phone, telegram, status, created_at
		FROM reports
		WHERE type = 'found'
			AND category = $1
			AND status = 'unclaimed'
			AND incident_date >= $2::date
			AND incident_date <= $2::date + ($3 * INTERVAL '1 day')
		ORDER BY incident_date ASC
		LIMIT 15
	`, category, lostDate, windowDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectReports(rows)
}

func (p *Postgres) FindLostCandidates(ctx context.Context, category string, foundDate time.Time, windowDays int) ([]domain.Report, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT id, type, category, title, description, unique_features,
			location, location_details, incident_date, photos, phone, telegram, status, created_at
		FROM reports
		WHERE type = 'lost'
			AND category = $1
			AND status = 'open'
			AND incident_date <= $2::date
			AND incident_date >= $2::date - ($3 * INTERVAL '1 day')
		ORDER BY incident_date DESC
		LIMIT 15
	`, category, foundDate, windowDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectReports(rows)
}

func (p *Postgres) UpsertMany(ctx context.Context, matches []domain.Match) error {
	if len(matches) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, m := range matches {
		batch.Queue(`
			INSERT INTO matches (id, lost_report_id, found_report_id, groq_score, gemini_score, combined_score, reasoning, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			ON CONFLICT (lost_report_id, found_report_id) DO UPDATE SET
				groq_score = EXCLUDED.groq_score,
				gemini_score = EXCLUDED.gemini_score,
				combined_score = EXCLUDED.combined_score,
				reasoning = EXCLUDED.reasoning,
				created_at = EXCLUDED.created_at
		`, m.ID, m.LostReportID, m.FoundReportID, m.GroqScore, m.GeminiScore, m.CombinedScore, m.Reasoning, m.CreatedAt)
	}
	br := p.pool.SendBatch(ctx, batch)
	defer br.Close()
	for range matches {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (p *Postgres) ListByLostID(ctx context.Context, lostID string) ([]domain.Match, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT m.id, m.lost_report_id, m.found_report_id, m.groq_score, m.gemini_score,
			m.combined_score, m.reasoning, m.created_at,
			f.id, f.type, f.category, f.title, f.description, f.unique_features,
			f.location, f.location_details, f.incident_date, f.photos, f.phone, f.telegram, f.status, f.created_at
		FROM matches m
		JOIN reports f ON f.id = m.found_report_id
		WHERE m.lost_report_id = $1
		ORDER BY m.combined_score DESC
	`, lostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectMatches(rows, true)
}

func (p *Postgres) ListByFoundID(ctx context.Context, foundID string) ([]domain.Match, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT m.id, m.lost_report_id, m.found_report_id, m.groq_score, m.gemini_score,
			m.combined_score, m.reasoning, m.created_at,
			l.id, l.type, l.category, l.title, l.description, l.unique_features,
			l.location, l.location_details, l.incident_date, l.photos, l.phone, l.telegram, l.status, l.created_at
		FROM matches m
		JOIN reports l ON l.id = m.lost_report_id
		WHERE m.found_report_id = $1
		ORDER BY m.combined_score DESC
	`, foundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectMatches(rows, false)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanReport(row rowScanner) (domain.Report, error) {
	var r domain.Report
	var photos []byte
	err := row.Scan(
		&r.ID, &r.Type, &r.Category, &r.Title, &r.Description, &r.UniqueFeatures,
		&r.Location, &r.LocationDetails, &r.IncidentDate, &photos, &r.Phone, &r.Telegram, &r.Status, &r.CreatedAt,
	)
	if err != nil {
		return domain.Report{}, err
	}
	if len(photos) > 0 {
		_ = json.Unmarshal(photos, &r.Photos)
	}
	if r.Photos == nil {
		r.Photos = []string{}
	}
	return r, nil
}

func collectReports(rows pgx.Rows) ([]domain.Report, error) {
	var out []domain.Report
	for rows.Next() {
		r, err := scanReport(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func collectMatches(rows pgx.Rows, hydrateFound bool) ([]domain.Match, error) {
	var out []domain.Match
	for rows.Next() {
		var m domain.Match
		var photos []byte
		var side domain.Report
		err := rows.Scan(
			&m.ID, &m.LostReportID, &m.FoundReportID, &m.GroqScore, &m.GeminiScore,
			&m.CombinedScore, &m.Reasoning, &m.CreatedAt,
			&side.ID, &side.Type, &side.Category, &side.Title, &side.Description, &side.UniqueFeatures,
			&side.Location, &side.LocationDetails, &side.IncidentDate, &photos, &side.Phone, &side.Telegram, &side.Status, &side.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if len(photos) > 0 {
			_ = json.Unmarshal(photos, &side.Photos)
		}
		if side.Photos == nil {
			side.Photos = []string{}
		}
		if hydrateFound {
			m.FoundReport = &side
		} else {
			m.LostReport = &side
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
