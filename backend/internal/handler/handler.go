package handler

import (
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

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc       *service.Service
	uploadDir string
}

func New(svc *service.Service, uploadDir string) *Handler {
	return &Handler{svc: svc, uploadDir: uploadDir}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Meta(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"categories": domain.Categories,
		"locations":  domain.Locations,
	})
}

func (h *Handler) CreateLost(c *gin.Context) {
	h.create(c, domain.TypeLost)
}

func (h *Handler) CreateFound(c *gin.Context) {
	h.create(c, domain.TypeFound)
}

func (h *Handler) create(c *gin.Context, kind domain.ReportType) {
	photos, err := h.savePhotos(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	date, err := time.Parse("2006-01-02", strings.TrimSpace(c.PostForm("incident_date")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date must be YYYY-MM-DD"})
		return
	}
	report, matches, err := h.svc.Create(c.Request.Context(), service.ReportInput{
		Type:            kind,
		Category:        c.PostForm("category"),
		Title:           c.PostForm("title"),
		Description:     c.PostForm("description"),
		UniqueFeatures:  c.PostForm("unique_features"),
		Location:        c.PostForm("location"),
		LocationDetails: c.PostForm("location_details"),
		IncidentDate:    date,
		Photos:          photos,
		Phone:           c.PostForm("phone"),
		Telegram:        c.PostForm("telegram"),
	})
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"report":  toJSON(report),
		"matches": matchesJSON(matches),
	})
}

func (h *Handler) GetReport(c *gin.Context) {
	report, matches, err := h.svc.ListMatches(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"report":  toJSON(report),
		"matches": matchesJSON(matches),
	})
}

func (h *Handler) RefreshMatches(c *gin.Context) {
	report, matches, err := h.svc.RefreshMatches(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"report":  toJSON(report),
		"matches": matchesJSON(matches),
	})
}

func (h *Handler) Claim(c *gin.Context) {
	report, err := h.svc.Claim(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"report": toJSON(report)})
}

func (h *Handler) savePhotos(c *gin.Context) ([]string, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, errors.New("could not read form")
	}
	files := form.File["photos"]
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
		dst, err := os.Create(filepath.Join(h.uploadDir, name))
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

func (h *Handler) handleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": strings.TrimPrefix(err.Error(), "validation: ")})
	case errors.Is(err, repository.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
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
	ID          string      `json:"id"`
	Score       float64     `json:"score"`
	GroqScore   *float64    `json:"groq_score"`
	GeminiScore *float64    `json:"gemini_score"`
	Reasoning   string      `json:"reasoning"`
	LostReport  *reportJSON `json:"lost_report,omitempty"`
	FoundReport *reportJSON `json:"found_report,omitempty"`
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
