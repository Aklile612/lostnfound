package ai

import (
	"encoding/json"
	"fmt"
	"lostandfound/internal/domain"
	"strings"
)

const systemPrompt = `You are a lost-and-found matching assistant for a university campus.

A hard filter already guaranteed:
- the items share the same category
- the found date is on or after the lost date
- the found date is within 7 days of the lost date
- the found item is still unclaimed

Your job is to decide how likely two reports describe the SAME physical object.

Score from 0 to 100.

Use all of the following:

1. Description identity
   Color, brand, model, size, material, and contents.
   Generic wording like "black bag" is weak evidence.
   Specific overlap (brand, model, a listed accessory) is strong.

2. Unique marks
   Cracks, stickers, engraving, scratches, missing parts, charms, handwriting, a broken zipper, a stained corner.
   The same distinctive mark on both sides is very strong evidence.
   A clear contradiction of a distinctive mark should drop the score below 30.

3. Location
   Same named place is strong.
   Nearby campus places (cafeteria vs coffee shop, library vs library entrance / student center) are moderate-strong.
   Distant unrelated places (library vs football field) are weak unless everything else is overwhelming.
   Extra details such as "near the entrance" or a room number can strengthen or weaken a match.

4. Timing
   Same day is strongest.
   The next day is still plausible.
   Later in the 7-day window is weaker, especially for small high-value items.

5. Photos
   When images are provided, treat appearance as primary evidence.
   A written description that matches a photo counts.
   A photo that shows a clearly different object should dominate and reject the match.
   Do not require both sides to have photos.
   A finder may upload only a photo and a place. Do not penalize a short or missing finder description.
   A loser may have no photo and only a description, including unique marks. Compare that text to the finder's photo and text.

6. Contradictions
   Different brands, colors, or object types inside the same category (AirPods vs a power bank) should score below 30.

Return JSON only in this shape:
{"matches":[{"id":"<candidate uuid>","score":72,"reasoning":"One or two sentences covering the main evidence and any doubt."}]}

Omit candidates you would score under 40.
If none are plausible, return {"matches":[]}.`

type modelMatch struct {
	ID        string  `json:"id"`
	FoundID   string  `json:"found_id"`
	LostID    string  `json:"lost_id"`
	Score     float64 `json:"score"`
	Reasoning string  `json:"reasoning"`
}

type modelResponse struct {
	Matches []modelMatch `json:"matches"`
}

func buildUserPrompt(anchor domain.Report, candidates []domain.Report, vision bool) string {
	var b strings.Builder
	if vision {
		b.WriteString("You can see attached photos. Use them as primary visual evidence.\n\n")
	} else {
		b.WriteString("You cannot see photos. A flag tells you whether photos exist. Score from text, location, dates, and unique marks only.\n\n")
	}

	if anchor.Type == domain.TypeLost {
		b.WriteString("LOST ITEM (the student is looking for this):\n")
	} else {
		b.WriteString("FOUND ITEM (already in hand):\n")
	}
	b.WriteString(formatReport(anchor, vision))
	b.WriteString("\nCANDIDATES:\n")
	for i, c := range candidates {
		fmt.Fprintf(&b, "\n--- candidate %d ---\nid: %s\n", i+1, c.ID)
		b.WriteString(formatReport(c, vision))
	}
	b.WriteString("\nReturn JSON only.")
	return b.String()
}

func formatReport(r domain.Report, vision bool) string {
	label := domain.Categories[r.Category]
	place := domain.Locations[r.Location]
	photoNote := "no"
	if r.HasPhotos() {
		if vision {
			photoNote = fmt.Sprintf("yes (%d image(s) attached)", len(r.Photos))
		} else {
			photoNote = fmt.Sprintf("yes (%d image(s), not visible to you)", len(r.Photos))
		}
	}
	title := r.Title
	if title == "" {
		title = "(none)"
	}
	desc := r.Description
	if desc == "" {
		desc = "(none — rely on photos and location if present)"
	}
	marks := r.UniqueFeatures
	if marks == "" {
		marks = "(none given)"
	}
	details := r.LocationDetails
	if details == "" {
		details = "(none)"
	}
	return fmt.Sprintf(
		"type: %s\ncategory: %s\ntitle: %s\ndescription: %s\nunique_marks: %s\nlocation: %s\nlocation_details: %s\ndate: %s\nhas_photos: %s\n",
		r.Type, label, title, desc, marks, place, details, r.IncidentDate.Format("2006-01-02"), photoNote,
	)
}

func parseScores(raw string, anchor domain.Report, candidates []domain.Report) []domain.ScoreSet {
	payload := extractJSON(raw)
	var parsed modelResponse
	if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
		return nil
	}
	known := map[string]domain.Report{}
	for _, c := range candidates {
		known[c.ID] = c
	}
	var out []domain.ScoreSet
	for _, m := range parsed.Matches {
		id := m.ID
		if id == "" {
			if anchor.Type == domain.TypeLost {
				id = m.FoundID
			} else {
				id = m.LostID
			}
		}
		c, ok := known[id]
		if !ok {
			continue
		}
		score := m.Score
		if score < 0 {
			score = 0
		}
		if score > 100 {
			score = 100
		}
		item := domain.ScoreSet{Score: score, Reasoning: strings.TrimSpace(m.Reasoning)}
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

func extractJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		return raw[start : end+1]
	}
	return raw
}
