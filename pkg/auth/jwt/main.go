package jwt

import (
	"maps"
	"math/rand/v2"
	"time"

	ar "github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
)

var (
	config JWTConfig
	log logger.Logger
)

func randomString() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!§$%&/()=?.,;:-_#+*^°"
	n := rand.IntN(32) + 32

	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.IntN(len(letters))]
	}

	return string(b)
}

type SigningMethod = jwt.SigningMethod
type Claims = jwt.MapClaims

type JWTConfig struct {
	Issuer string

	SecretRefreshToken string
	SecretAccessToken string

	SigningMethod SigningMethod

	TTL time.Duration
	TTLRefreshToken time.Duration
	// MaxRefreshTokenAge is the maximum age of a refresh token.
	// If set to 0, RefreshToken will live forever.
	MaxRefreshTokenAge time.Duration

	AdditionalClaimsFunc func(ar.Ctx) Claims
}

func SetLogger(logger logger.Logger) {
	log = logger
}

func DefaultJWTConfig() *JWTConfig {
	config = JWTConfig{
		Issuer: "Apiright AuthServer",
		SecretRefreshToken: randomString(),
		SecretAccessToken: randomString(),
		SigningMethod: jwt.SigningMethodHS256,
		TTL: time.Hour,
		TTLRefreshToken: time.Duration(15 * time.Minute),
		MaxRefreshTokenAge: time.Duration(0),
		AdditionalClaimsFunc: nil,
	}

	return &config
}

type TokenPair struct {
	AccessToken string `json:"access_token" xml:"access_token" yaml:"access_token"`
	RefreshToken string `json:"refresh_token" xml:"refresh_token" yaml:"refresh_token"`
}

func JWTMiddleware(config JWTConfig) ar.Middleware {
	return ar.Middleware(func(next ar.Handler) ar.Handler {
		return func(c *ar.Ctx) error {
			return next(c)
		}
	})
}

func NewTokenPair(c *ar.Ctx) (TokenPair, error) {
	var (
		accessToken, refreshToken string
		err error
	)
	log.Debug("NewTokenPairFunc")

	claims := jwt.MapClaims{
		"iss": config.Issuer,
		"sub": c.Session["userID"],
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(config.TTL).Unix(),
	}
	log.Debugf("Claims: %v", claims)

	if config.AdditionalClaimsFunc != nil {
		log.Debug("AdditionalClaimsFunc")
		maps.Copy(claims, config.AdditionalClaimsFunc(*c))
		log.Debugf("Claims: %v", claims)
	}

	log.Debug(config.SecretAccessToken , " ", len([]byte(config.SecretAccessToken)))
	accessToken, err = jwt.NewWithClaims(config.SigningMethod, claims).SignedString([]byte(config.SecretAccessToken))
	if err != nil {
		return TokenPair{}, err
	}
	log.Debugf("AccessToken: %s", accessToken)

	log.Debug(config.SecretRefreshToken, " ", len([]byte(config.SecretRefreshToken)))
	refreshToken, err = jwt.NewWithClaims(config.SigningMethod, claims).SignedString([]byte(config.SecretRefreshToken))
	if err != nil {
		return TokenPair{}, err
	}
	log.Debugf("RefreshToken: %s", refreshToken)

	return TokenPair{
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func NewAccessToken(c ar.Ctx) (string, error) {
	return "", nil
}
