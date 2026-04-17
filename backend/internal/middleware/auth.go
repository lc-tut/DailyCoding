package middleware

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"

	"daileycoding/backend/internal/domain"
)

const contextUserKey = "auth_user"

// Required は Cookie の JWT を検証し、ユーザーを context にセットする。
// bypass=true のとき検証をスキップしてモックユーザーをセットする。
func Required(jwtSecret []byte, bypass bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if bypass {
				(*c).Set(contextUserKey, &domain.User{
					ID:    uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Email: "admin@linuxclub.example",
					Name:  "テスト管理者",
					Role:  "admin",
				})
				return next(c)
			}

			cookie, err := (*c).Cookie("dc_session")
			if err != nil {
				return (*c).JSON(http.StatusUnauthorized, map[string]any{"error": "not authenticated"})
			}

			claims := &domain.AppClaims{}
			token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return jwtSecret, nil
			})
			if err != nil || !token.Valid {
				return (*c).JSON(http.StatusUnauthorized, map[string]any{"error": "invalid or expired token"})
			}

			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				return (*c).JSON(http.StatusUnauthorized, map[string]any{"error": "invalid token claims"})
			}

			(*c).Set(contextUserKey, &domain.User{
				ID:    userID,
				Email: claims.Email,
				Name:  claims.Name,
				Role:  claims.Role,
			})
			return next(c)
		}
	}
}

// AdminRequired は Required の後に使う。role != "admin" なら 403 を返す。
func AdminRequired() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			user := UserFromContext(c)
			if user == nil || user.Role != "admin" {
				return (*c).JSON(http.StatusForbidden, map[string]any{"error": "admin only"})
			}
			return next(c)
		}
	}
}

// UserFromContext は echo.Context からユーザーを取り出す
func UserFromContext(c *echo.Context) *domain.User {
	u, _ := (*c).Get(contextUserKey).(*domain.User)
	return u
}
