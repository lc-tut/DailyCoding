package domain

// ExecuteRequest はコード実行に必要な入力を表す
type ExecuteRequest struct {
	Code     string
	Language string
	Stdin    string
}

// ExecuteResult はコード実行の結果を表す
type ExecuteResult struct {
	Stdout string
	Stderr string
	TimeMs float64
}
