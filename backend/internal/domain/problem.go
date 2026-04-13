package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrDateConflict は published_at が他の問題と重複するときに返す
var ErrDateConflict = errors.New("another problem already has this published_at date")

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

// ListAllParams は管理用一覧の絞込・並び替えパラメータ
// Filter: "" | "active" | "inactive" | "undated" | "inactive_undated"
// Sort:   "" | "created" | "published" | "title"
type ListAllParams struct {
	Filter string
	Sort   string
}

// UpdateProblemParams は問題更新時の入力パラメータ（テストケースは変更しない）
type UpdateProblemParams struct {
	Title       string
	Description string
	IsActive    bool
	PublishedAt *time.Time // nil = 常時公開（日付なし）
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
