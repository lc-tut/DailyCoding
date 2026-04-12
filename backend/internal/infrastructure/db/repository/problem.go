package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"daileycoding/backend/internal/domain"
)

// ProblemRepository は problems / test_cases テーブルへのアクセスを担う
type ProblemRepository struct {
	db *pgxpool.Pool
}

func NewProblemRepository(db *pgxpool.Pool) *ProblemRepository {
	return &ProblemRepository{db: db}
}

// ListActive はアクティブな問題を公開日の降順で返す
func (r *ProblemRepository) ListActive(ctx context.Context) ([]domain.Problem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, title, description, is_active, published_at, created_at, updated_at
		FROM   problems
		WHERE  is_active = true
		ORDER  BY published_at DESC NULLS LAST, created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query problems: %w", err)
	}
	defer rows.Close()

	var out []domain.Problem
	for rows.Next() {
		var p domain.Problem
		var idStr string
		if err := rows.Scan(
			&idStr, &p.Title, &p.Description,
			&p.IsActive, &p.PublishedAt,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan problem: %w", err)
		}
		p.ID, err = uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("parse problem id: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// Create は問題とテストケースをトランザクションで一括登録する
func (r *ProblemRepository) Create(ctx context.Context, params domain.CreateProblemParams) (*domain.Problem, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var p domain.Problem
	var idStr string
	err = tx.QueryRow(ctx, `
		INSERT INTO problems (title, description, is_active, published_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, description, is_active, published_at, created_at, updated_at
	`, params.Title, params.Description, params.IsActive, params.PublishedAt).Scan(
		&idStr, &p.Title, &p.Description,
		&p.IsActive, &p.PublishedAt,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert problem: %w", err)
	}
	p.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("parse problem id: %w", err)
	}

	for _, tc := range params.TestCases {
		sw := tc.ScoreWeight
		if sw == 0 {
			sw = 1
		}
		if _, err = tx.Exec(ctx, `
			INSERT INTO test_cases
			  (problem_id, stdin, expected_stdout, order_index, is_sample, hint, score_weight)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, p.ID, tc.Stdin, tc.ExpectedStdout, tc.OrderIndex, tc.IsSample, tc.Hint, sw); err != nil {
			return nil, fmt.Errorf("insert test_case %d: %w", tc.OrderIndex, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &p, nil
}

// GetTestCasesByProblemID は指定した問題のテストケースを order_index 順で返す
func (r *ProblemRepository) GetTestCasesByProblemID(ctx context.Context, problemID uuid.UUID) ([]domain.TestCase, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, problem_id, stdin, expected_stdout,
		       order_index, is_sample, hint, score_weight,
		       created_at, updated_at
		FROM   test_cases
		WHERE  problem_id = $1
		ORDER  BY order_index ASC
	`, problemID)
	if err != nil {
		return nil, fmt.Errorf("query test_cases: %w", err)
	}
	defer rows.Close()

	var out []domain.TestCase
	for rows.Next() {
		var tc domain.TestCase
		var idStr, problemIDStr string
		if err := rows.Scan(
			&idStr, &problemIDStr,
			&tc.Stdin, &tc.ExpectedStdout,
			&tc.OrderIndex, &tc.IsSample, &tc.Hint, &tc.ScoreWeight,
			&tc.CreatedAt, &tc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan test_case: %w", err)
		}
		tc.ID, err = uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("parse test_case id: %w", err)
		}
		tc.ProblemID, err = uuid.Parse(problemIDStr)
		if err != nil {
			return nil, fmt.Errorf("parse test_case problem_id: %w", err)
		}
		out = append(out, tc)
	}
	return out, rows.Err()
}
