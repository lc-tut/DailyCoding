package oidcinfra

import (
	"context"
	"errors"

	"github.com/coreos/go-oidc/v3/oidc"

	"daileycoding/backend/internal/domain"
)

// Verifier は *oidc.IDTokenVerifier をラップし、
// ID Token の署名・クレーム検証と nonce チェックを行う。
type Verifier struct {
	inner *oidc.IDTokenVerifier
}

func NewVerifier(inner *oidc.IDTokenVerifier) *Verifier {
	return &Verifier{inner: inner}
}

// Verify は rawIDToken を検証し OIDCClaims を返す。
// 署名・iss・aud・exp の検証は go-oidc が行い、nonce はここで照合する。
func (v *Verifier) Verify(ctx context.Context, rawIDToken, expectedNonce string) (*domain.OIDCClaims, error) {
	idToken, err := v.inner.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}

	var c struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Nonce         string `json:"nonce"`
	}
	if err := idToken.Claims(&c); err != nil {
		return nil, err
	}

	if c.Nonce != expectedNonce {
		return nil, errors.New("nonce mismatch")
	}

	return &domain.OIDCClaims{
		Sub:           c.Sub,
		Email:         c.Email,
		EmailVerified: c.EmailVerified,
		Name:          c.Name,
	}, nil
}
