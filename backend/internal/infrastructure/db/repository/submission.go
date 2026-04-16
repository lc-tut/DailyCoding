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
		INSERT INTO submissions (user_id, problem_id, code, language, overall_status, score)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, problem_id, code, language, overall_status, score, ai_review,
		          created_at, updated_at
	`,
		params.UserID, params.ProblemID, params.Code, params.Language,
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

// ListByProblemID は問題 ID に紐づく提出一覧をコードなしで返す（降順）
func (r *SubmissionRepository) ListByProblemID(ctx context.Context, problemID uuid.UUID) ([]domain.SubmissionSummary, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, problem_id, language, overall_status, score, created_at
		FROM   submissions
		WHERE  problem_id = $1
		ORDER  BY created_at DESC
	`, problemID)
	if err != nil {
		return nil, fmt.Errorf("query submissions: %w", err)
	}
	defer rows.Close()

	var out []domain.SubmissionSummary
	for rows.Next() {
		var s domain.SubmissionSummary
		var idStr, problemIDStr string
		if err := rows.Scan(&idStr, &problemIDStr, &s.Language, &s.OverallStatus, &s.Score, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan submission: %w", err)
		}
		parsed, parseErr := uuid.Parse(idStr)
		if parseErr != nil {
			return nil, fmt.Errorf("parse submission id: %w", parseErr)
		}
		s.ID = parsed
		parsed, parseErr = uuid.Parse(problemIDStr)
		if parseErr != nil {
			return nil, fmt.Errorf("parse problem id: %w", parseErr)
		}
		s.ProblemID = parsed
		out = append(out, s)
	}
	return out, rows.Err()
}

// GetByID は提出 ID に紐づく提出詳細（テストケース結果付き）を返す。
// 非サンプルテストケースの stdin / expected_stdout は空文字列で返す。
func (r *SubmissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.SubmissionDetail, error) {
	// 1. 提出本体を取得
	var sub domain.Submission
	var idStr, problemIDStr string
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, problem_id, code, language, overall_status, score, ai_review,
		       created_at, updated_at
		FROM   submissions
		WHERE  id = $1
	`, id).Scan(
		&idStr, &sub.UserID, &problemIDStr,
		&sub.Code, &sub.Language, &sub.OverallStatus, &sub.Score, &sub.AIReview,
		&sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("query submission: %w", err)
	}
	var parseErr error
	sub.ID, parseErr = uuid.Parse(idStr)
	if parseErr != nil {
		return nil, fmt.Errorf("parse submission id: %w", parseErr)
	}
	sub.ProblemID, parseErr = uuid.Parse(problemIDStr)
	if parseErr != nil {
		return nil, fmt.Errorf("parse problem id: %w", parseErr)
	}

	// 2. テストケース結果を取得（非サンプルの stdin/expected_stdout は空文字列）
	rows, err := r.db.Query(ctx, `
		SELECT sr.id, sr.test_case_id, tc.order_index, tc.is_sample,
		       CASE WHEN tc.is_sample THEN tc.stdin          ELSE '' END,
		       CASE WHEN tc.is_sample THEN tc.expected_stdout ELSE '' END,
		       sr.status, sr.stdout, sr.stderr, sr.execution_time_ms
		FROM   submission_results sr
		JOIN   test_cases tc ON tc.id = sr.test_case_id
		WHERE  sr.submission_id = $1
		ORDER  BY tc.order_index ASC
	`, id)
	if err != nil {
		return nil, fmt.Errorf("query submission results: %w", err)
	}
	defer rows.Close()

	var results []domain.SubmissionResultWithCase
	for rows.Next() {
		var res domain.SubmissionResultWithCase
		var resIDStr, tcIDStr string
		if err := rows.Scan(
			&resIDStr, &tcIDStr,
			&res.OrderIndex, &res.IsSample,
			&res.Stdin, &res.ExpectedStdout,
			&res.Status, &res.Stdout, &res.Stderr, &res.ExecutionTimeMs,
		); err != nil {
			return nil, fmt.Errorf("scan submission result: %w", err)
		}
		res.ID, parseErr = uuid.Parse(resIDStr)
		if parseErr != nil {
			return nil, fmt.Errorf("parse result id: %w", parseErr)
		}
		res.TestCaseID, parseErr = uuid.Parse(tcIDStr)
		if parseErr != nil {
			return nil, fmt.Errorf("parse test_case id: %w", parseErr)
		}
		results = append(results, res)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate results: %w", err)
	}

	return &domain.SubmissionDetail{Submission: sub, Results: results}, nil
}
