package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"daileycoding/backend/internal/domain"
)

// UserRepository は users / user_identities テーブルへのアクセスを担う
type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// GetByID は UUID でユーザーを取得する
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var u domain.User
	var idStr string
	err := r.db.QueryRow(ctx, `
		SELECT id, email, name, role, created_at, updated_at
		FROM   users
		WHERE  id = $1
	`, id).Scan(&idStr, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	u.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}
	return &u, nil
}

// UpsertWithIdentity は JIT プロビジョニングを行う。
//
// 処理順序:
//  1. user_identities を (provider, providerSub) で検索 → 見つかればそのユーザーを返す
//  2. users を email で検索 → 見つかれば identity を追加して返す（別プロバイダーでの同一メール = アカウント統合）
//  3. どちらも存在しない → users + user_identities を INSERT して返す
func (r *UserRepository) UpsertWithIdentity(
	ctx context.Context,
	email, name, provider, providerSub string,
) (*domain.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// 1. 既存の identity を検索
	var userIDStr string
	err = tx.QueryRow(ctx, `
		SELECT u.id
		FROM   user_identities ui
		JOIN   users u ON u.id = ui.user_id
		WHERE  ui.provider = $1 AND ui.provider_sub = $2
	`, provider, providerSub).Scan(&userIDStr)

	if err != nil && err.Error() != "no rows in result set" {
		return nil, fmt.Errorf("lookup identity: %w", err)
	}

	if err == nil {
		// identity が既存 → ユーザーを返す（name のみ更新）
		var u domain.User
		var idStr string
		if err := tx.QueryRow(ctx, `
			UPDATE users SET name = $2, updated_at = NOW()
			WHERE  id = $1
			RETURNING id, email, name, role, created_at, updated_at
		`, userIDStr, name).Scan(&idStr, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("update user name: %w", err)
		}
		u.ID, err = uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("parse user id: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("commit: %w", err)
		}
		return &u, nil
	}

	// 2. email で既存ユーザーを検索（別プロバイダーからのアカウント統合）
	var u domain.User
	var idStr string
	err = tx.QueryRow(ctx, `
		SELECT id, email, name, role, created_at, updated_at
		FROM   users WHERE email = $1
	`, email).Scan(&idStr, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)

	if err != nil && err.Error() != "no rows in result set" {
		return nil, fmt.Errorf("lookup user by email: %w", err)
	}

	if err != nil {
		// 3. 新規ユーザー作成
		if err := tx.QueryRow(ctx, `
			INSERT INTO users (email, name) VALUES ($1, $2)
			RETURNING id, email, name, role, created_at, updated_at
		`, email, name).Scan(&idStr, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("insert user: %w", err)
		}
	}

	u.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	// identity を追加
	if _, err := tx.Exec(ctx, `
		INSERT INTO user_identities (user_id, provider, provider_sub)
		VALUES ($1, $2, $3)
		ON CONFLICT (provider, provider_sub) DO NOTHING
	`, u.ID, provider, providerSub); err != nil {
		return nil, fmt.Errorf("insert identity: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &u, nil
}
