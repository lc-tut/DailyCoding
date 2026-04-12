package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"daileycoding/backend/internal/domain"
)

// SubmissionRepository は submissions / submission_results テーブルへのアクセスを担う
type SubmissionRepository struct {
	db *pgxpool.Pool
}

func NewSubmissionRepository(db *pgxpool.Pool) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

// Save は submission と全 results を1トランザクションで保存する
func (r *SubmissionRepository) Save(
	ctx context.Context,
	params domain.CreateSubmissionParams,
	results []domain.CreateSubmissionResultParams,
) (*domain.Submission, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// submissions INSERT
	var s domain.Submission
	var idStr string
	err = tx.QueryRow(ctx, `
		INSERT INTO submissions (problem_id, code, language, overall_status, score)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, problem_id, code, language, overall_status, score, ai_review,
		          created_at, updated_at
	`,
		params.ProblemID, params.Code, params.Language,
		params.OverallStatus, params.Score,
	).Scan(
		&idStr,
		&s.UserID, // NULL 許容なので *uuid.UUID として受け取る
		&params.ProblemID, // 使い捨て
		&s.Code, &s.Language, &s.OverallStatus, &s.Score,
		&s.AIReview, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert submission: %w", err)
	}
	s.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("parse submission id: %w", err)
	}
	s.ProblemID = params.ProblemID

	// submission_results INSERT (ループ)
	for i := range results {
		results[i].SubmissionID = s.ID
		if _, err := tx.Exec(ctx, `
			INSERT INTO submission_results
			       (submission_id, test_case_id, status, stdout, stderr, execution_time_ms)
			VALUES ($1, $2, $3, $4, $5, $6)
		`,
			results[i].SubmissionID,
			results[i].TestCaseID,
			results[i].Status,
			results[i].Stdout,
			results[i].Stderr,
			results[i].ExecutionTimeMs,
		); err != nil {
			return nil, fmt.Errorf("insert submission_result[%d]: %w", i, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return &s, nil
}
