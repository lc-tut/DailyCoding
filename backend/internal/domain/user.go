package domain

import (
	"time"

	"github.com/google/uuid"
)

// User は users テーブルのエンティティ
type User struct {
	ID        uuid.UUID
	Email     string
	Name      string
	Role      string // "student" | "admin"
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserIdentity は user_identities テーブルのエンティティ
type UserIdentity struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Provider    string // "google" | "keycloak"
	ProviderSub string
	CreatedAt   time.Time
}
