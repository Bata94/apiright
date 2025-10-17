package jwt

import (
	"time"

	ar "github.com/bata94/apiright/pkg/core"
	"github.com/golang-jwt/jwt/v5"
)

type SigningMethod = jwt.SigningMethod
type Claims = jwt.Claims

type JWTConfig struct {
	Issuer string
	Secret string
	SigningMethod SigningMethod
	TTL time.Duration
	TTLRefreshToken time.Duration
	// MaxRefreshTokenAge is the maximum age of a refresh token.
	// If set to 0, RefreshToken will live forever.
	MaxRefreshTokenAge time.Duration

	AdditionalClaimsFunc func(ar.Ctx) Claims
}

var (
  key []byte
  t   *jwt.Token
  s   string
)

func init() {
  key = []byte("secret")

	t = jwt.NewWithClaims(jwt.SigningMethodES256,
		jwt.MapClaims{
			"iss": "my-auth-server",
			"sub": "john",
			"foo": 2,
		})

	s, err := t.SignedString(key)
	if err != nil {
		panic(err)
	}
}

func JWTMiddleware(config JWTConfig) ar.Middleware {
	return ar.Middleware(func(next ar.Handler) ar.Handler {
		return func(c *ar.Ctx) error {
			return next(c)
		}
	})
}
