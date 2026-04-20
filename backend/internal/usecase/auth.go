package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"daileycoding/backend/internal/domain"
)

// AuthUserRepository はユーザーの永続化操作を定義する
type AuthUserRepository interface {
	UpsertWithIdentity(ctx context.Context, email, name, provider, providerSub string) (*domain.User, error)
}

// AuthAllowedEmailRepository は個別許可メールアドレスの参照を定義する
type AuthAllowedEmailRepository interface {
	IsAllowed(ctx context.Context, email string) (bool, error)
}

// AuthUsecase は認証に関わるビジネスロジックを担う
type AuthUsecase struct {
	userRepo         AuthUserRepository
	allowedEmailRepo AuthAllowedEmailRepository
	allowDomains     []string
	jwtSecret        []byte
}

func NewAuthUsecase(
	userRepo AuthUserRepository,
	allowedEmailRepo AuthAllowedEmailRepository,
	allowDomains []string,
	jwtSecret []byte,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:         userRepo,
		allowedEmailRepo: allowedEmailRepo,
		allowDomains:     allowDomains,
		jwtSecret:        jwtSecret,
	}
}

// ProcessLogin はメール許可チェック → ユーザー JIT プロビジョニング → アプリ JWT 発行を行う。
// OIDC による検証済みクレームを受け取るため、このレイヤーで外部プロバイダーには依存しない。
func (u *AuthUsecase) ProcessLogin(ctx context.Context, provider string, claims *domain.OIDCClaims) (string, error) {
	if !claims.EmailVerified {
		return "", &AuthError{Code: "email_not_verified"}
	}

	allowed, err := u.isEmailAllowed(ctx, claims.Email)
	if err != nil {
		return "", fmt.Errorf("check allowlist: %w", err)
	}
	if !allowed {
		return "", &AuthError{Code: "access_denied"}
	}

	user, err := u.userRepo.UpsertWithIdentity(
		ctx,
		strings.ToLower(claims.Email), claims.Name, provider, claims.Sub,
	)
	if err != nil {
		return "", fmt.Errorf("provision user: %w", err)
	}

	token, err := u.issueAppToken(user)
	if err != nil {
		return "", fmt.Errorf("issue token: %w", err)
	}
	return token, nil
}

// isEmailAllowed はドメイン一致または個別許可メールアドレス登録を確認する。
// allowDomains と allowedEmailRepo の両方が未設定のときは全員許可（開発用）。
func (u *AuthUsecase) isEmailAllowed(ctx context.Context, email string) (bool, error) {
	if len(u.allowDomains) == 0 && u.allowedEmailRepo == nil {
		return true, nil
	}

	if idx := strings.LastIndex(email, "@"); idx >= 0 {
		domain := strings.ToLower(email[idx+1:])
		for _, d := range u.allowDomains {
			if domain == strings.ToLower(d) {
				return true, nil
			}
		}
	}

	if u.allowedEmailRepo != nil {
		ok, err := u.allowedEmailRepo.IsAllowed(ctx, strings.ToLower(email))
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	return false, nil
}

func (u *AuthUsecase) issueAppToken(user *domain.User) (string, error) {
	claims := domain.AppClaims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Name:   user.Name,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(u.jwtSecret)
}

// AuthError は認証ビジネスロジックのエラー型
type AuthError struct {
	Code string // "email_not_verified" | "access_denied"
}

func (e *AuthError) Error() string { return e.Code }
