package domain

import "github.com/google/uuid"

// SubmitRequest はフロントエンドから受け取るリクエスト
type SubmitRequest struct {
	ProblemID uuid.UUID `json:"problem_id"`
	Code      string    `json:"code"`
}

// TestResult はひとつのテストケースの採点結果（API レスポンス用）
type TestResult struct {
	CaseIndex int     `json:"case_index"`
	Status    string  `json:"status"` // "Pass" | "Fail"
	Stdout    string  `json:"stdout"`
	Expected  string  `json:"expected"`
	Stderr    string  `json:"stderr"`
	TimeMs    float64 `json:"time_ms"`
}

// JudgeResponse はフロントエンドへ返すレスポンス全体
type JudgeResponse struct {
	Overall string       `json:"overall"` // "Pass" | "Fail"
	Results []TestResult `json:"results"`
}
