package domain

import (
	"time"

	"github.com/google/uuid"
)

// Problem は problems テーブルのエンティティ
type Problem struct {
	ID          uuid.UUID
	Title       string
	Description string
	IsActive    bool
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TestCase は test_cases テーブルのエンティティ
type TestCase struct {
	ID             uuid.UUID
	ProblemID      uuid.UUID
	Stdin          string
	ExpectedStdout string
	OrderIndex     int16
	IsSample       bool
	Hint           *string
	ScoreWeight    int16
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// CreateProblemParams は問題作成時の入力パラメータ
type CreateProblemParams struct {
	Title       string
	Description string
	IsActive    bool
	PublishedAt *time.Time
	TestCases   []CreateTestCaseParams
}

// CreateTestCaseParams はテストケース作成時の入力パラメータ
type CreateTestCaseParams struct {
	Stdin          string
	ExpectedStdout string
	OrderIndex     int16
	IsSample       bool
	Hint           *string
	ScoreWeight    int16
}
