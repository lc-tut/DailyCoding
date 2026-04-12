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
