package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AllowedEmailRepository は allowed_emails テーブルへのアクセスを担う
type AllowedEmailRepository struct {
	db *pgxpool.Pool
}

func NewAllowedEmailRepository(db *pgxpool.Pool) *AllowedEmailRepository {
	return &AllowedEmailRepository{db: db}
}

// IsAllowed は指定したメールアドレスが allowed_emails テーブルに登録されているか返す
func (r *AllowedEmailRepository) IsAllowed(ctx context.Context, email string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM allowed_emails WHERE email = $1`, email,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check allowed_email: %w", err)
	}
	return count > 0, nil
}
