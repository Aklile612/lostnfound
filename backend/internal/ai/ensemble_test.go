package ai

import (
	"lostandfound/internal/domain"
	"testing"
)

func TestCombinePrefersVisionWhenBothPresent(t *testing.T) {
	groq := 80.0
	gemini := 60.0
	out := Combine(
		[]domain.ScoreSet{{LostID: "l1", FoundID: "f1", Score: groq, Reasoning: "text"}},
		[]domain.ScoreSet{{LostID: "l1", FoundID: "f1", Score: gemini, Reasoning: "vision"}},
		40,
	)
	if len(out) != 1 {
		t.Fatalf("got %d matches", len(out))
	}
	want := 0.45*groq + 0.55*gemini
	if out[0].Score != want {
		t.Fatalf("score %v want %v", out[0].Score, want)
	}
	if out[0].GroqScore == nil || *out[0].GroqScore != groq {
		t.Fatal("missing groq score")
	}
	if out[0].GeminiScore == nil || *out[0].GeminiScore != gemini {
		t.Fatal("missing gemini score")
	}
}

func TestCombineDropsBelowThreshold(t *testing.T) {
	out := Combine(
		[]domain.ScoreSet{{LostID: "l1", FoundID: "f1", Score: 20, Reasoning: "weak"}},
		nil,
		40,
	)
	if len(out) != 0 {
		t.Fatalf("expected empty, got %#v", out)
	}
}

func TestHeuristicKeepsNearbyCampusPlaces(t *testing.T) {
	lost := domain.Report{
		ID: "lost-1", Type: domain.TypeLost, Category: "bags",
		Title: "black backpack", Description: "black jansport backpack with a cracked zipper",
		UniqueFeatures: "cracked left zipper", Location: "cafeteria",
	}
	found := domain.Report{
		ID: "found-1", Type: domain.TypeFound, Category: "bags",
		Title: "dark backpack", Description: "dark backpack cracked zipper",
		Location: "coffee_shop",
	}
	out := HeuristicScores(lost, []domain.Report{found})
	if len(out) == 0 {
		t.Fatal("expected a heuristic match")
	}
}
