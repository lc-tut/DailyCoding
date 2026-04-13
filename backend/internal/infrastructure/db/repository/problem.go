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

// ListActive はアクティブな問題を公開日の降順で返す。
// is_active=true のものに加え、published_at が当日以前の問題も自動的に含む。
func (r *ProblemRepository) ListActive(ctx context.Context) ([]domain.Problem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, title, description, is_active, published_at, created_at, updated_at
		FROM   problems
		WHERE  is_active = true
		   OR  (published_at IS NOT NULL AND published_at <= CURRENT_DATE)
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

// Update は問題のタイトル・説明・公開設定を更新する（テストケースは変更しない）
func (r *ProblemRepository) Update(ctx context.Context, id uuid.UUID, params domain.UpdateProblemParams) (*domain.Problem, error) {
	var p domain.Problem
	var idStr string
	err := r.db.QueryRow(ctx, `
		UPDATE problems
		SET    title        = $2,
		       description  = $3,
		       is_active    = $4,
		       published_at = $5,
		       updated_at   = NOW()
		WHERE  id = $1
		RETURNING id, title, description, is_active, published_at, created_at, updated_at
	`, id, params.Title, params.Description, params.IsActive, params.PublishedAt).Scan(
		&idStr, &p.Title, &p.Description,
		&p.IsActive, &p.PublishedAt,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("update problem: %w", err)
	}
	p.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("parse problem id: %w", err)
	}
	return &p, nil
}

// GetByID は指定した ID の問題を1件返す。存在しない場合は nil, nil を返す
func (r *ProblemRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Problem, error) {
	var p domain.Problem
	var idStr string
	err := r.db.QueryRow(ctx, `
		SELECT id, title, description, is_active, published_at, created_at, updated_at
		FROM   problems
		WHERE  id = $1
	`, id).Scan(
		&idStr, &p.Title, &p.Description,
		&p.IsActive, &p.PublishedAt,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("query problem by id: %w", err)
	}
	p.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("parse problem id: %w", err)
	}
	return &p, nil
}

// ListAll は管理画面用に全問題を返す。params で絞込・並び替えが可能。
func (r *ProblemRepository) ListAll(ctx context.Context, params domain.ListAllParams) ([]domain.Problem, error) {
	// 絞込条件（値はコード側で決定するため SQL インジェクションリスクなし）
	where := ""
	switch params.Filter {
	case "active":
		where = "WHERE is_active = true OR (published_at IS NOT NULL AND published_at <= CURRENT_DATE)"
	case "inactive":
		where = "WHERE NOT (is_active = true OR (published_at IS NOT NULL AND published_at <= CURRENT_DATE))"
	case "undated":
		where = "WHERE published_at IS NULL"
	case "inactive_undated":
		where = "WHERE is_active = false AND published_at IS NULL"
	}

	// 並び替え
	orderBy := "ORDER BY created_at DESC"
	switch params.Sort {
	case "published":
		orderBy = "ORDER BY published_at DESC NULLS LAST, created_at DESC"
	case "title":
		orderBy = "ORDER BY title ASC"
	}

	//nolint:gosec // where/orderBy はコード側で決定済みの固定文字列
	rows, err := r.db.Query(ctx, `
		SELECT id, title, description, is_active, published_at, created_at, updated_at
		FROM   problems
		`+where+`
		`+orderBy)
	if err != nil {
		return nil, fmt.Errorf("query all problems: %w", err)
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

// GetToday は JST 当日の published_at を持つアクティブな問題を1件返す。
// 該当なしの場合は nil, nil を返す
func (r *ProblemRepository) GetToday(ctx context.Context, today string) (*domain.Problem, error) {
	var p domain.Problem
	var idStr string
	err := r.db.QueryRow(ctx, `
		SELECT id, title, description, is_active, published_at, created_at, updated_at
		FROM   problems
		WHERE  published_at = $1
		LIMIT  1
	`, today).Scan(
		&idStr, &p.Title, &p.Description,
		&p.IsActive, &p.PublishedAt,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("query today problem: %w", err)
	}
	p.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("parse problem id: %w", err)
	}
	return &p, nil
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
