package handler

import (
	"encoding/json"
	"errors"
	"io"
	"lostandfound/internal/domain"
	"lostandfound/internal/repository"
	"lostandfound/internal/service"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Handler struct {
	svc       *service.Service
	uploadDir string
}

func New(svc *service.Service, uploadDir string) *Handler {
	return &Handler{svc: svc, uploadDir: uploadDir}
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Meta(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"categories": domain.Categories,
		"locations":  domain.Locations,
	})
}

func (h *Handler) CreateLost(w http.ResponseWriter, r *http.Request) {
	h.create(w, r, domain.TypeLost)
}

func (h *Handler) CreateFound(w http.ResponseWriter, r *http.Request) {
	h.create(w, r, domain.TypeFound)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request, kind domain.ReportType) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "could not read form")
		return
	}
	photos, err := h.savePhotos(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	date, err := time.Parse("2006-01-02", strings.TrimSpace(r.FormValue("incident_date")))
	if err != nil {
		writeError(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
		return
	}
	report, matches, err := h.svc.Create(r.Context(), service.ReportInput{
		Type:            kind,
		Category:        r.FormValue("category"),
		Title:           r.FormValue("title"),
		Description:     r.FormValue("description"),
		UniqueFeatures:  r.FormValue("unique_features"),
		Location:        r.FormValue("location"),
		LocationDetails: r.FormValue("location_details"),
		IncidentDate:    date,
		Photos:          photos,
		Phone:           r.FormValue("phone"),
		Telegram:        r.FormValue("telegram"),
	})
	if err != nil {
		h.handleErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"report":  toJSON(report),
		"matches": matchesJSON(matches),
	})
}

func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	report, matches, err := h.svc.ListMatches(r.Context(), r.PathValue("id"))
	if err != nil {
		h.handleErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"report":  toJSON(report),
		"matches": matchesJSON(matches),
	})
}

func (h *Handler) RefreshMatches(w http.ResponseWriter, r *http.Request) {
	report, matches, err := h.svc.RefreshMatches(r.Context(), r.PathValue("id"))
	if err != nil {
		h.handleErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"report":  toJSON(report),
		"matches": matchesJSON(matches),
	})
}

func (h *Handler) Claim(w http.ResponseWriter, r *http.Request) {
	report, err := h.svc.Claim(r.Context(), r.PathValue("id"))
	if err != nil {
		h.handleErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"report": toJSON(report)})
}

func (h *Handler) savePhotos(r *http.Request) ([]string, error) {
	if r.MultipartForm == nil {
		return nil, nil
	}
	files := r.MultipartForm.File["photos"]
	if len(files) > 3 {
		return nil, errors.New("at most 3 photos")
	}
	var names []string
	for _, fh := range files {
		if fh.Size > 5<<20 {
			return nil, errors.New("each photo must be under 5MB")
		}
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".webp":
		default:
			return nil, errors.New("photos must be jpg, png, or webp")
		}
		src, err := fh.Open()
		if err != nil {
			return nil, err
		}
		name := uuid.NewString() + ext
		dstPath := filepath.Join(h.uploadDir, name)
		dst, err := os.Create(dstPath)
		if err != nil {
			src.Close()
			return nil, err
		}
		_, err = io.Copy(dst, src)
		src.Close()
		dst.Close()
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func (h *Handler) handleErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		msg := strings.TrimPrefix(err.Error(), "validation: ")
		writeError(w, http.StatusBadRequest, msg)
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "report not found")
	default:
		writeError(w, http.StatusInternalServerError, "something went wrong")
	}
}

type reportJSON struct {
	ID              string   `json:"id"`
	Type            string   `json:"type"`
	Category        string   `json:"category"`
	CategoryLabel   string   `json:"category_label"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	UniqueFeatures  string   `json:"unique_features"`
	Location        string   `json:"location"`
	LocationLabel   string   `json:"location_label"`
	LocationDetails string   `json:"location_details"`
	IncidentDate    string   `json:"incident_date"`
	Photos          []string `json:"photos"`
	Phone           string   `json:"phone,omitempty"`
	Telegram        string   `json:"telegram,omitempty"`
	Status          string   `json:"status"`
	CreatedAt       string   `json:"created_at"`
}

type matchJSON struct {
	ID            string      `json:"id"`
	Score         float64     `json:"score"`
	GroqScore     *float64    `json:"groq_score"`
	GeminiScore   *float64    `json:"gemini_score"`
	Reasoning     string      `json:"reasoning"`
	LostReport    *reportJSON `json:"lost_report,omitempty"`
	FoundReport   *reportJSON `json:"found_report,omitempty"`
}

func toJSON(r domain.Report) reportJSON {
	photos := make([]string, 0, len(r.Photos))
	for _, p := range r.Photos {
		photos = append(photos, "/uploads/"+p)
	}
	out := reportJSON{
		ID:              r.ID,
		Type:            string(r.Type),
		Category:        r.Category,
		CategoryLabel:   domain.Categories[r.Category],
		Title:           r.Title,
		Description:     r.Description,
		UniqueFeatures:  r.UniqueFeatures,
		Location:        r.Location,
		LocationLabel:   domain.Locations[r.Location],
		LocationDetails: r.LocationDetails,
		IncidentDate:    r.IncidentDate.Format("2006-01-02"),
		Photos:          photos,
		Status:          string(r.Status),
		CreatedAt:       r.CreatedAt.Format(time.RFC3339),
	}
	if r.Type == domain.TypeFound {
		out.Phone = r.Phone
		out.Telegram = r.Telegram
	}
	return out
}

func matchesJSON(matches []domain.Match) []matchJSON {
	out := make([]matchJSON, 0, len(matches))
	for _, m := range matches {
		item := matchJSON{
			ID:          m.ID,
			Score:       m.CombinedScore,
			GroqScore:   m.GroqScore,
			GeminiScore: m.GeminiScore,
			Reasoning:   m.Reasoning,
		}
		if m.FoundReport != nil {
			rep := toJSON(*m.FoundReport)
			item.FoundReport = &rep
		}
		if m.LostReport != nil {
			rep := toJSON(*m.LostReport)
			item.LostReport = &rep
		}
		out = append(out, item)
	}
	return out
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
