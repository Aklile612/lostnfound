package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"lostandfound/internal/domain"
	"net/http"
	"time"
)

type Groq struct {
	key    string
	model  string
	client *http.Client
}

func NewGroq(key, model string) *Groq {
	return &Groq{
		key:    key,
		model:  model,
		client: &http.Client{Timeout: 45 * time.Second},
	}
}

func (g *Groq) Enabled() bool {
	return g != nil && g.key != ""
}

func (g *Groq) CompareLostToFound(ctx context.Context, lost domain.Report, found []domain.Report) ([]domain.ScoreSet, error) {
	if !g.Enabled() || len(found) == 0 {
		return nil, nil
	}
	return g.compare(ctx, lost, found)
}

func (g *Groq) compare(ctx context.Context, anchor domain.Report, candidates []domain.Report) ([]domain.ScoreSet, error) {
	body := map[string]any{
		"model": g.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": buildUserPrompt(anchor, candidates, false)},
		},
		"temperature":     0.2,
		"response_format": map[string]string{"type": "json_object"},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.key)
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
		return nil, fmt.Errorf("groq: %s", truncate(payload, 400))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("groq: empty response")
	}
	return parseScores(parsed.Choices[0].Message.Content, anchor, candidates), nil
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n])
}
