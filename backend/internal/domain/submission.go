package domain

import (
	"time"

	"github.com/google/uuid"
)

// Submission は submissions テーブルのエンティティ
type Submission struct {
	ID            uuid.UUID
	UserID        *uuid.UUID // Phase1: NULL 許容
	ProblemID     uuid.UUID
	Code          string
	Language      string
	OverallStatus string
	Score         int16
	AIReview      *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// SubmissionResult は submission_results テーブルのエンティティ
type SubmissionResult struct {
	ID              uuid.UUID
	SubmissionID    uuid.UUID
	TestCaseID      uuid.UUID
	Status          string
	Stdout          string
	Stderr          string
	ExecutionTimeMs float32
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// CreateSubmissionParams は Submission 作成時の入力
type CreateSubmissionParams struct {
	ProblemID     uuid.UUID
	Code          string
	Language      string
	OverallStatus string
	Score         int16
}

// CreateSubmissionResultParams は SubmissionResult 作成時の入力
type CreateSubmissionResultParams struct {
	SubmissionID    uuid.UUID
	TestCaseID      uuid.UUID
	Status          string
	Stdout          string
	Stderr          string
	ExecutionTimeMs float32
}

// SubmissionSummary は一覧表示用の提出サマリ（コードを含まない）
type SubmissionSummary struct {
	ID            uuid.UUID
	ProblemID     uuid.UUID
	Language      string
	OverallStatus string
	Score         int16
	CreatedAt     time.Time
}

// SubmissionResultWithCase は提出結果とテストケース情報をまとめた読み取り専用型
// is_sample=false のとき Stdin / ExpectedStdout は空文字列
type SubmissionResultWithCase struct {
	ID              uuid.UUID
	TestCaseID      uuid.UUID
	OrderIndex      int16
	IsSample        bool
	Stdin           string
	ExpectedStdout  string
	Status          string
	Stdout          string
	Stderr          string
	ExecutionTimeMs float32
}

// SubmissionDetail は提出詳細（コード＋テストケース結果付き）
type SubmissionDetail struct {
	Submission
	Results []SubmissionResultWithCase
}
