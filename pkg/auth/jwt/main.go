
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/bata94/apiright/pkg/core"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrJWTMissing = errors.New("missing or malformed JWT")
	ErrJWTInvalid = errors.New("invalid or expired JWT")
)

type JWTConfig struct {
	Secret        []byte
	SigningMethod jwt.SigningMethod
	Expiration    time.Duration
	ContextKey    string
}

func NewJWTConfig(secret []byte) JWTConfig {
	return JWTConfig{
		Secret:        secret,
		SigningMethod: jwt.SigningMethodHS256,
		Expiration:    time.Hour * 24,
		ContextKey:    "user",
	}
}

func (c *JWTConfig) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(c.SigningMethod, claims)
	return token.SignedString(c.Secret)
}

func (c *JWTConfig) Middleware() core.Middleware {
	return func(next core.Handler) core.Handler {
		return func(ctx *core.Ctx) error {
			authHeader := ctx.Request.Header.Get("Authorization")
			if authHeader == "" {
				return ErrJWTMissing
			}

			tokenString := ""
			count, err := fmt.Sscanf(authHeader, "Bearer %s", &tokenString)
			if tokenString == "" {
				return ErrJWTMissing
			} else if count != 1 {
				return ErrJWTInvalid
			} else if err != nil {
				return ErrJWTInvalid
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return c.Secret, nil
			})

			if err != nil {
				return ErrJWTInvalid
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				ctx.Session[c.ContextKey] = claims
			} else {
				return ErrJWTInvalid
			}

			return next(ctx)
		}
	}
}
