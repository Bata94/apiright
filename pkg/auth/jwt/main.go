package jwt

import (
	"errors"
	"maps"
	"math/rand/v2"
	"strings"
	"time"

	ar "github.com/bata94/apiright/pkg/core"
	"github.com/bata94/apiright/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
)

var (
	config JWTConfig
	log    logger.Logger
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
	SecretAccessToken  string

	SigningMethod SigningMethod

	TTL             time.Duration
	TTLRefreshToken time.Duration
	// MaxRefreshTokenAge is the maximum age of a refresh token.
	// If set to 0, RefreshToken will live forever.
	MaxRefreshTokenAge time.Duration
	Leeway             time.Duration

	AdditionalClaimsFunc func(*ar.Ctx) Claims
}

func SetLogger(logger logger.Logger) {
	log = logger
}

func DefaultJWTConfig() *JWTConfig {
	config = JWTConfig{
		Issuer:               "Apiright AuthServer",
		SecretRefreshToken:   randomString(),
		SecretAccessToken:    randomString(),
		SigningMethod:        jwt.SigningMethodHS256,
		TTL:                  time.Hour,
		Leeway:               time.Duration(15 * time.Second),
		TTLRefreshToken:      time.Duration(15 * time.Minute),
		MaxRefreshTokenAge:   time.Duration(0),
		AdditionalClaimsFunc: nil,
	}

	return &config
}

type TokenPair struct {
	AccessToken  string `json:"access_token" xml:"access_token" yaml:"access_token"`
	RefreshToken string `json:"refresh_token" xml:"refresh_token" yaml:"refresh_token"`
}

func JWTMiddleware(config JWTConfig) ar.Middleware {
	return ar.Middleware(func(next ar.Handler) ar.Handler {
		return func(c *ar.Ctx) error {
			if c.Request.Header.Get("Authorization") == "" {
				return errors.New("authorization header is missing")
			}

			accessToken := strings.Replace(c.Request.Header.Get("Authorization"), "Bearer ", "", 1)
			err := ValidateAccessToken(c, accessToken)
			if err != nil {
				return err
			}

			return next(c)
		}
	})
}

func NewTokenPair(c *ar.Ctx, userID any) (TokenPair, error) {
	var (
		accessToken, refreshToken string
		err                       error
	)

	claims := jwt.MapClaims{
		"iss": config.Issuer,
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(config.TTL).Unix(),
	}

	if config.AdditionalClaimsFunc != nil {
		log.Debug("AdditionalClaimsFunc")
		maps.Copy(claims, config.AdditionalClaimsFunc(c))
		log.Debugf("Claims: %v", claims)
	}

	accessToken, err = jwt.NewWithClaims(config.SigningMethod, claims).SignedString([]byte(config.SecretAccessToken))
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, err = jwt.NewWithClaims(config.SigningMethod, claims).SignedString([]byte(config.SecretRefreshToken))
	if err != nil {
		return TokenPair{}, err
	}

	c.Session["userID"] = userID

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func NewAccessToken(c *ar.Ctx, userID any) (string, error) {
	var (
		accessToken string
		err         error
	)

	claims := jwt.MapClaims{
		"iss": config.Issuer,
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(config.TTL).Unix(),
	}

	accessToken, err = jwt.NewWithClaims(config.SigningMethod, claims).SignedString([]byte(config.SecretAccessToken))
	if err != nil {
		return "", err
	}

	c.Session["userID"] = userID

	return accessToken, nil
}

func NewAccessTokenWithRefreshToken(c *ar.Ctx, userID any, refreshToken string) (string, error) {
	userID_RT, err := ValidateRefreshToken(c, refreshToken)
	if err != nil {
		return "", err
	} else if userID_RT == nil {
		return "", errors.New("userID is nil")
	} else if userID_RT != userID {
		return "", errors.New("userID is not the same")
	}

	accessToken, err := NewAccessToken(c, userID)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func ValidateAccessToken(c *ar.Ctx, accessToken string) error {
	var (
		claims Claims
		err    error
	)

	if accessToken == "" {
		return errors.New("token is empty")
	}

	accessToken = strings.Replace(accessToken, "Bearer ", "", 1)
	_, err = jwt.ParseWithClaims(accessToken, &claims, func(token *jwt.Token) (any, error) {
		return []byte(config.SecretAccessToken), nil
	},
		// BUG: Leeway is not working
		jwt.WithLeeway(config.Leeway),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(config.Issuer),
		jwt.WithStrictDecoding(),
	)
	if err != nil {
		return err
	}

	c.Session["userID"] = claims["sub"]

	return nil
}

func ValidateRefreshToken(c *ar.Ctx, refreshToken string) (any, error) {
	var (
		claims Claims
		err    error
	)

	if refreshToken == "" {
		return nil, errors.New("token is empty")
	}

	_, err = jwt.ParseWithClaims(refreshToken, &claims, func(token *jwt.Token) (any, error) {
		return []byte(config.SecretRefreshToken), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(config.Issuer),
		jwt.WithStrictDecoding(),
	)
	if err != nil {
		return nil, err
	}

	return claims["sub"], nil
}

func GetSubFromToken(c *ar.Ctx) (any, error) {
	var (
		claims Claims
		err    error
	)

	authHeaderToken := c.Request.Header.Get("Authorization")
	if authHeaderToken == "" {
		return nil, errors.New("authorization header is missing")
	}

	accessToken := strings.Replace(authHeaderToken, "Bearer ", "", 1)
	token, err := jwt.ParseWithClaims(accessToken, &claims, func(token *jwt.Token) (any, error) {
		return []byte(config.SecretAccessToken), nil
	})
	if err != nil {
		return nil, err
	}

	sub, ok := token.Claims.(jwt.MapClaims)["sub"]
	if !ok {
		return nil, errors.New("sub is not in claims")
	}

	return sub, nil
}
