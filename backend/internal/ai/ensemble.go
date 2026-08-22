package ai

import "lostandfound/internal/domain"

func Combine(groq, gemini []domain.ScoreSet, threshold float64) []domain.ScoreSet {
	type acc struct {
		lostID  string
		foundID string
		groq    *float64
		gemini  *float64
		gReason string
		vReason string
	}
	index := map[string]*acc{}
	key := func(lostID, foundID string) string { return lostID + "|" + foundID }

	put := func(list []domain.ScoreSet, vision bool) {
		for _, s := range list {
			k := key(s.LostID, s.FoundID)
			row, ok := index[k]
			if !ok {
				row = &acc{lostID: s.LostID, foundID: s.FoundID}
				index[k] = row
			}
			score := s.Score
			if vision {
				row.gemini = &score
				row.vReason = s.Reasoning
			} else {
				row.groq = &score
				row.gReason = s.Reasoning
			}
		}
	}
	put(groq, false)
	put(gemini, true)

	var out []domain.ScoreSet
	for _, row := range index {
		combined := 0.0
		switch {
		case row.groq != nil && row.gemini != nil:
			combined = 0.45*(*row.groq) + 0.55*(*row.gemini)
		case row.gemini != nil:
			combined = *row.gemini
		case row.groq != nil:
			combined = *row.groq
		}
		if combined < threshold {
			continue
		}
		reason := row.vReason
		if reason == "" {
			reason = row.gReason
		} else if row.gReason != "" && row.gReason != row.vReason {
			reason = reason + " Text model: " + row.gReason
		}
		out = append(out, domain.ScoreSet{
			LostID:      row.lostID,
			FoundID:     row.foundID,
			GroqScore:   row.groq,
			GeminiScore: row.gemini,
			Score:       combined,
			Reasoning:   reason,
		})
	}
	return out
}
