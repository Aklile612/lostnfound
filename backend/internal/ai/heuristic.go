package ai

import (
	"regexp"
	"strings"
	"unicode"

	"lostandfound/internal/domain"
)

var stop = map[string]struct{}{
	"a": {}, "an": {}, "the": {}, "and": {}, "or": {}, "of": {}, "in": {}, "on": {}, "at": {},
	"to": {}, "for": {}, "with": {}, "my": {}, "it": {}, "this": {}, "that": {}, "near": {},
	"around": {}, "item": {}, "lost": {}, "found": {}, "please": {}, "has": {}, "have": {},
}

var wordRe = regexp.MustCompile(`[a-z0-9]+`)

func HeuristicScores(anchor domain.Report, candidates []domain.Report) []domain.ScoreSet {
	var out []domain.ScoreSet
	for _, c := range candidates {
		score, reason := heuristicPair(anchor, c)
		if score < 40 {
			continue
		}
		item := domain.ScoreSet{Score: score, Reasoning: reason}
		if anchor.Type == domain.TypeLost {
			item.LostID = anchor.ID
			item.FoundID = c.ID
		} else {
			item.FoundID = anchor.ID
			item.LostID = c.ID
		}
		out = append(out, item)
	}
	return out
}

func heuristicPair(a, b domain.Report) (float64, string) {
	lost, found := a, b
	if a.Type == domain.TypeFound {
		lost, found = b, a
	}
	score := 38.0
	reasons := []string{}

	if lost.Location == found.Location {
		score += 22
		reasons = append(reasons, "same campus place")
	} else if relatedPlace(lost.Location, found.Location) {
		score += 12
		reasons = append(reasons, "nearby campus places")
	}

	overlap := tokenOverlap(
		lost.Title+" "+lost.Description+" "+lost.UniqueFeatures,
		found.Title+" "+found.Description+" "+found.UniqueFeatures,
	)
	score += overlap * 28
	if overlap >= 0.35 {
		reasons = append(reasons, "descriptions share specific words")
	}

	days := found.IncidentDate.Sub(lost.IncidentDate).Hours() / 24
	if days < 0 {
		days = -days
	}
	if days == 0 {
		score += 10
		reasons = append(reasons, "same day")
	} else if days <= 2 {
		score += 6
		reasons = append(reasons, "close dates")
	}

	if lost.UniqueFeatures != "" && tokenOverlap(lost.UniqueFeatures, found.Description+" "+found.UniqueFeatures) >= 0.25 {
		score += 12
		reasons = append(reasons, "unique marks overlap")
	}

	if score > 100 {
		score = 100
	}
	reason := "Heuristic fallback (no AI keys): " + strings.Join(reasons, "; ") + "."
	if len(reasons) == 0 {
		reason = "Heuristic fallback (no AI keys): same category and a plausible date window."
	}
	return score, reason
}

func relatedPlace(a, b string) bool {
	for _, n := range domain.NearbyLocations(a) {
		if n == b {
			return true
		}
	}
	return false
}

func tokenOverlap(a, b string) float64 {
	left := tokens(a)
	right := tokens(b)
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	set := map[string]struct{}{}
	for t := range right {
		set[t] = struct{}{}
	}
	var hit int
	for t := range left {
		if _, ok := set[t]; ok {
			hit++
		}
	}
	smaller := len(left)
	if len(right) < smaller {
		smaller = len(right)
	}
	return float64(hit) / float64(smaller)
}

func tokens(s string) map[string]struct{} {
	s = strings.ToLower(s)
	out := map[string]struct{}{}
	for _, w := range wordRe.FindAllString(s, -1) {
		if len(w) < 3 {
			continue
		}
		if _, skip := stop[w]; skip {
			continue
		}
		if !hasLetter(w) {
			continue
		}
		out[w] = struct{}{}
	}
	return out
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
