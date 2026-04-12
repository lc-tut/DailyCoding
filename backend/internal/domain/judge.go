package domain

// TestCase はひとつのテストケースを表す
type TestCase struct {
	Input    string
	Expected string
}

// TestResult はひとつのテストケースの採点結果
type TestResult struct {
	CaseIndex int     `json:"case_index"`
	Status    string  `json:"status"` // "Pass" or "Fail"
	Stdout    string  `json:"stdout"`
	Expected  string  `json:"expected"`
	Stderr    string  `json:"stderr"`
	TimeMs    float64 `json:"time_ms"`
}

// JudgeResponse はフロントエンドへ返すレスポンス全体
type JudgeResponse struct {
	Overall string       `json:"overall"` // "Pass" or "Fail"
	Results []TestResult `json:"results"`
}

// SubmitRequest はフロントエンドから受け取るリクエスト
type SubmitRequest struct {
	Code string `json:"code"`
}
