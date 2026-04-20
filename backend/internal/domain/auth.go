package domain

import "github.com/golang-jwt/jwt/v5"

// OIDCClaims は OIDC ID Token から取り出した検証済みクレーム
type OIDCClaims struct {
	Sub           string
	Email         string
	EmailVerified bool
	Name          string
}

// AppClaims はアプリ独自 JWT のペイロード
type AppClaims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
