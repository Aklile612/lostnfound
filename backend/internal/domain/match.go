package domain

import "time"

type ScoreSet struct {
	FoundID     string
	LostID      string
	GroqScore   *float64
	GeminiScore *float64
	Score       float64
	Reasoning   string
}

type Match struct {
	ID            string
	LostReportID  string
	FoundReportID string
	GroqScore     *float64
	GeminiScore   *float64
	CombinedScore float64
	Reasoning     string
	CreatedAt     time.Time
	LostReport    *Report
	FoundReport   *Report
}

type CandidateScores struct {
	Scores []ScoreSet
}
