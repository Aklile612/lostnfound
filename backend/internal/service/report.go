package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"lostandfound/internal/ai"
	"lostandfound/internal/domain"
	"lostandfound/internal/port"
	"lostandfound/internal/repository"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	ErrValidation = errors.New("validation")
	ErrNotFound   = repository.ErrNotFound
)

type ReportInput struct {
	Type            domain.ReportType
	Category        string
	Title           string
	Description     string
	UniqueFeatures  string
	Location        string
	LocationDetails string
	IncidentDate    time.Time
	Photos          []string
	Phone           string
	Telegram        string
}

type Service struct {
	reports   port.ReportRepository
	matches   port.MatchRepository
	groq      port.TextMatcher
	gemini    port.VisionMatcher
	window    int
	threshold float64
}

func New(reports port.ReportRepository, matches port.MatchRepository, groq port.TextMatcher, gemini port.VisionMatcher, window int, threshold float64) *Service {
	if window <= 0 {
		window = 7
	}
	if threshold <= 0 {
		threshold = 40
	}
	return &Service{
		reports:   reports,
		matches:   matches,
		groq:      groq,
		gemini:    gemini,
		window:    window,
		threshold: threshold,
	}
}

func (s *Service) Create(ctx context.Context, in ReportInput) (domain.Report, []domain.Match, error) {
	report, err := buildReport(in)
	if err != nil {
		return domain.Report{}, nil, err
	}
	if err := s.reports.Create(ctx, report); err != nil {
		return domain.Report{}, nil, err
	}
	matched, err := s.Match(ctx, report)
	if err != nil {
		return report, nil, err
	}
	return report, matched, nil
}

func (s *Service) Get(ctx context.Context, id string) (domain.Report, error) {
	if _, err := uuid.Parse(id); err != nil {
		return domain.Report{}, fmt.Errorf("%w: invalid id", ErrValidation)
	}
	return s.reports.GetByID(ctx, id)
}

func (s *Service) ListMatches(ctx context.Context, id string) (domain.Report, []domain.Match, error) {
	report, err := s.Get(ctx, id)
	if err != nil {
		return domain.Report{}, nil, err
	}
	var list []domain.Match
	if report.Type == domain.TypeLost {
		list, err = s.matches.ListByLostID(ctx, report.ID)
	} else {
		list, err = s.matches.ListByFoundID(ctx, report.ID)
	}
	return report, list, err
}

func (s *Service) RefreshMatches(ctx context.Context, id string) (domain.Report, []domain.Match, error) {
	report, err := s.Get(ctx, id)
	if err != nil {
		return domain.Report{}, nil, err
	}
	list, err := s.Match(ctx, report)
	return report, list, err
}

func (s *Service) Claim(ctx context.Context, foundID string) (domain.Report, error) {
	report, err := s.Get(ctx, foundID)
	if err != nil {
		return domain.Report{}, err
	}
	if report.Type != domain.TypeFound {
		return domain.Report{}, fmt.Errorf("%w: only found items can be claimed", ErrValidation)
	}
	if report.Status == domain.StatusClaimed {
		return report, nil
	}
	if err := s.reports.UpdateStatus(ctx, report.ID, domain.StatusClaimed); err != nil {
		return domain.Report{}, err
	}
	report.Status = domain.StatusClaimed
	return report, nil
}

func (s *Service) Match(ctx context.Context, report domain.Report) ([]domain.Match, error) {
	var (
		candidates []domain.Report
		err        error
	)
	if report.Type == domain.TypeLost {
		candidates, err = s.reports.FindFoundCandidates(ctx, report.Category, report.IncidentDate, s.window)
	} else {
		candidates, err = s.reports.FindLostCandidates(ctx, report.Category, report.IncidentDate, s.window)
	}
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return []domain.Match{}, nil
	}

	var groqScores []domain.ScoreSet
	if s.groq != nil {
		groqScores, err = s.groq.CompareLostToFound(ctx, report, candidates)
		if err != nil {
			slog.Warn("groq matching failed", "err", err)
			groqScores = nil
		}
	}
	var geminiScores []domain.ScoreSet
	if s.gemini != nil {
		geminiScores, err = s.gemini.CompareLostToFound(ctx, report, candidates)
		if err != nil {
			slog.Warn("gemini matching failed", "err", err)
			geminiScores = nil
		}
	}
	var combined []domain.ScoreSet
	if len(groqScores) == 0 && len(geminiScores) == 0 {
		for _, row := range ai.HeuristicScores(report, candidates) {
			if row.Score >= s.threshold {
				combined = append(combined, row)
			}
		}
	} else {
		combined = ai.Combine(groqScores, geminiScores, s.threshold)
	}
	now := time.Now().UTC()
	stored := make([]domain.Match, 0, len(combined))
	for _, row := range combined {
		stored = append(stored, domain.Match{
			ID:            uuid.NewString(),
			LostReportID:  row.LostID,
			FoundReportID: row.FoundID,
			GroqScore:     row.GroqScore,
			GeminiScore:   row.GeminiScore,
			CombinedScore: row.Score,
			Reasoning:     row.Reasoning,
			CreatedAt:     now,
		})
	}
	if err := s.matches.UpsertMany(ctx, stored); err != nil {
		return nil, err
	}
	if report.Type == domain.TypeLost {
		return s.matches.ListByLostID(ctx, report.ID)
	}
	return s.matches.ListByFoundID(ctx, report.ID)
}

func buildReport(in ReportInput) (domain.Report, error) {
	in.Category = strings.TrimSpace(in.Category)
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	in.UniqueFeatures = strings.TrimSpace(in.UniqueFeatures)
	in.Location = strings.TrimSpace(in.Location)
	in.LocationDetails = strings.TrimSpace(in.LocationDetails)
	in.Phone = strings.TrimSpace(in.Phone)
	in.Telegram = strings.TrimPrefix(strings.TrimSpace(in.Telegram), "@")

	if !domain.ValidCategory(in.Category) {
		return domain.Report{}, fmt.Errorf("%w: unknown category", ErrValidation)
	}
	if !domain.ValidLocation(in.Location) {
		return domain.Report{}, fmt.Errorf("%w: unknown location", ErrValidation)
	}
	if in.IncidentDate.IsZero() {
		return domain.Report{}, fmt.Errorf("%w: date is required", ErrValidation)
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if in.IncidentDate.After(today.Add(24 * time.Hour)) {
		return domain.Report{}, fmt.Errorf("%w: date cannot be in the future", ErrValidation)
	}

	switch in.Type {
	case domain.TypeLost:
		if utf8.RuneCountInString(in.Description) < 8 {
			return domain.Report{}, fmt.Errorf("%w: describe the item in more detail", ErrValidation)
		}
	case domain.TypeFound:
		if len(in.Photos) == 0 && utf8.RuneCountInString(in.Description) < 8 {
			return domain.Report{}, fmt.Errorf("%w: add a photo or a description", ErrValidation)
		}
		if in.Phone == "" {
			return domain.Report{}, fmt.Errorf("%w: phone number is required", ErrValidation)
		}
		if in.Telegram == "" {
			return domain.Report{}, fmt.Errorf("%w: telegram username is required", ErrValidation)
		}
	default:
		return domain.Report{}, fmt.Errorf("%w: unknown report type", ErrValidation)
	}

	if in.Title == "" {
		if in.Description != "" {
			in.Title = clip(in.Description, 80)
		} else {
			in.Title = domain.Categories[in.Category]
		}
	}

	status := domain.StatusOpen
	if in.Type == domain.TypeFound {
		status = domain.StatusUnclaimed
	}

	if in.Photos == nil {
		in.Photos = []string{}
	}

	now := time.Now().UTC()
	return domain.Report{
		ID:              uuid.NewString(),
		Type:            in.Type,
		Category:        in.Category,
		Title:           in.Title,
		Description:     in.Description,
		UniqueFeatures:  in.UniqueFeatures,
		Location:        in.Location,
		LocationDetails: in.LocationDetails,
		IncidentDate:    in.IncidentDate,
		Photos:          in.Photos,
		Phone:           in.Phone,
		Telegram:        in.Telegram,
		Status:          status,
		CreatedAt:       now,
	}, nil
}

func clip(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}
