package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"lostandfound/internal/domain"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Gemini struct {
	key       string
	model     string
	uploadDir string
	client    *http.Client
}

func NewGemini(key, model, uploadDir string) *Gemini {
	return &Gemini{
		key:       key,
		model:     model,
		uploadDir: uploadDir,
		client:    &http.Client{Timeout: 60 * time.Second},
	}
}

func (g *Gemini) Enabled() bool {
	return g != nil && g.key != ""
}

func (g *Gemini) CompareLostToFound(ctx context.Context, lost domain.Report, found []domain.Report) ([]domain.ScoreSet, error) {
	if !g.Enabled() || len(found) == 0 {
		return nil, nil
	}
	return g.compare(ctx, lost, found)
}

func (g *Gemini) compare(ctx context.Context, anchor domain.Report, candidates []domain.Report) ([]domain.ScoreSet, error) {
	limited := candidates
	if len(limited) > 6 {
		limited = limited[:6]
	}
	parts := []map[string]any{
		{"text": systemPrompt + "\n\n" + buildUserPrompt(anchor, limited, true)},
	}
	parts = append(parts, g.imageParts("ANCHOR", anchor)...)
	for _, c := range limited {
		parts = append(parts, g.imageParts("CANDIDATE "+c.ID, c)...)
	}
	body := map[string]any{
		"contents": []map[string]any{
			{"parts": parts},
		},
		"generationConfig": map[string]any{
			"temperature":      0.2,
			"responseMimeType": "application/json",
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", g.model, g.key)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	payload, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("gemini: %s", truncate(payload, 400))
	}
	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini: empty response")
	}
	return parseScores(parsed.Candidates[0].Content.Parts[0].Text, anchor, limited), nil
}

func (g *Gemini) imageParts(label string, report domain.Report) []map[string]any {
	var parts []map[string]any
	limit := 2
	if len(report.Photos) < limit {
		limit = len(report.Photos)
	}
	for i := 0; i < limit; i++ {
		mime, data, err := g.readPhoto(report.Photos[i])
		if err != nil {
			continue
		}
		parts = append(parts, map[string]any{"text": fmt.Sprintf("Photo %d for %s (%s report %s)", i+1, label, report.Type, report.ID)})
		parts = append(parts, map[string]any{
			"inlineData": map[string]string{
				"mimeType": mime,
				"data":     data,
			},
		})
	}
	return parts
}

func (g *Gemini) readPhoto(name string) (string, string, error) {
	name = filepath.Base(name)
	path := filepath.Join(g.uploadDir, name)
	b, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	if len(b) > 4_000_000 {
		return "", "", fmt.Errorf("too large")
	}
	ext := strings.ToLower(filepath.Ext(name))
	mime := "image/jpeg"
	switch ext {
	case ".png":
		mime = "image/png"
	case ".webp":
		mime = "image/webp"
	case ".jpg", ".jpeg":
		mime = "image/jpeg"
	default:
		return "", "", fmt.Errorf("unsupported")
	}
	return mime, base64.StdEncoding.EncodeToString(b), nil
}
