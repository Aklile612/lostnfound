package port

import (
	"context"
	"lostandfound/internal/domain"
	"time"
)

type ReportRepository interface {
	Create(ctx context.Context, report domain.Report) error
	GetByID(ctx context.Context, id string) (domain.Report, error)
	UpdateStatus(ctx context.Context, id string, status domain.Status) error
	FindFoundCandidates(ctx context.Context, category string, lostDate time.Time, windowDays int) ([]domain.Report, error)
	FindLostCandidates(ctx context.Context, category string, foundDate time.Time, windowDays int) ([]domain.Report, error)
}

type MatchRepository interface {
	UpsertMany(ctx context.Context, matches []domain.Match) error
	ListByLostID(ctx context.Context, lostID string) ([]domain.Match, error)
	ListByFoundID(ctx context.Context, foundID string) ([]domain.Match, error)
}

type TextMatcher interface {
	CompareLostToFound(ctx context.Context, lost domain.Report, found []domain.Report) ([]domain.ScoreSet, error)
}

type VisionMatcher interface {
	CompareLostToFound(ctx context.Context, lost domain.Report, found []domain.Report) ([]domain.ScoreSet, error)
}
