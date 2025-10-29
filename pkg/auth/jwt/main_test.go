package jwt

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bata94/apiright/pkg/core"
	"github.com/golang-jwt/jwt/v5"
)

func TestDefaultJWTConfig(t *testing.T) {
	config := DefaultJWTConfig()

	if config.Issuer != "Apiright AuthServer" {
		t.Errorf("Expected Issuer 'Apiright AuthServer', got '%s'", config.Issuer)
	}

	if config.SecretAccessToken == "" {
		t.Error("Expected non-empty SecretAccessToken")
	}

	if config.SecretRefreshToken == "" {
		t.Error("Expected non-empty SecretRefreshToken")
	}

	if config.SigningMethod != jwt.SigningMethodHS256 {
		t.Errorf("Expected SigningMethod HS256, got %v", config.SigningMethod)
	}

	if config.TTL != time.Hour {
		t.Errorf("Expected TTL 1h, got %v", config.TTL)
	}

	if config.TTLRefreshToken != 15*time.Minute {
		t.Errorf("Expected TTLRefreshToken 15m, got %v", config.TTLRefreshToken)
	}

	if config.Leeway != 15*time.Second {
		t.Errorf("Expected Leeway 15s, got %v", config.Leeway)
	}

	if config.MaxRefreshTokenAge != 0 {
		t.Errorf("Expected MaxRefreshTokenAge 0, got %v", config.MaxRefreshTokenAge)
	}
}

func TestRandomString(t *testing.T) {
	s1 := randomString()
	s2 := randomString()

	if len(s1) < 32 {
		t.Errorf("Expected random string length >= 32, got %d", len(s1))
	}

	if s1 == s2 {
		t.Error("Expected different random strings, got identical")
	}
}

func TestNewTokenPair(t *testing.T) {
	config := DefaultJWTConfig()
	ctx := core.NewCtx(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), core.Route{}, core.Endpoint{})

	userID := "user123"
	tokenPair, err := NewTokenPair(ctx, userID)
	if err != nil {
		t.Errorf("NewTokenPair returned error: %v", err)
	}

	if tokenPair.AccessToken == "" {
		t.Error("Expected non-empty AccessToken")
	}

	if tokenPair.RefreshToken == "" {
		t.Error("Expected non-empty RefreshToken")
	}

	// Verify tokens can be parsed
	_, err = jwt.Parse(tokenPair.AccessToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.SecretAccessToken), nil
	})
	if err != nil {
		t.Errorf("Failed to parse access token: %v", err)
	}

	_, err = jwt.Parse(tokenPair.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.SecretRefreshToken), nil
	})
	if err != nil {
		t.Errorf("Failed to parse refresh token: %v", err)
	}

	// Check session
	if ctx.Session["userID"] != userID {
		t.Errorf("Expected session userID '%s', got '%v'", userID, ctx.Session["userID"])
	}
}

func TestNewAccessToken(t *testing.T) {
	config := DefaultJWTConfig()
	ctx := core.NewCtx(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), core.Route{}, core.Endpoint{})

	userID := "user456"
	token, err := NewAccessToken(ctx, userID)
	if err != nil {
		t.Errorf("NewAccessToken returned error: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Verify token can be parsed
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.SecretAccessToken), nil
	})
	if err != nil {
		t.Errorf("Failed to parse token: %v", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Error("Expected MapClaims")
	}

	if claims["sub"] != userID {
		t.Errorf("Expected subject '%s', got '%v'", userID, claims["sub"])
	}

	if claims["iss"] != config.Issuer {
		t.Errorf("Expected issuer '%s', got '%v'", config.Issuer, claims["iss"])
	}

	// Check session
	if ctx.Session["userID"] != userID {
		t.Errorf("Expected session userID '%s', got '%v'", userID, ctx.Session["userID"])
	}
}

func TestNewAccessTokenWithRefreshToken(t *testing.T) {
	ctx := core.NewCtx(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), core.Route{}, core.Endpoint{})

	userID := "user789"

	// First create a token pair
	tokenPair, err := NewTokenPair(ctx, userID)
	if err != nil {
		t.Fatalf("NewTokenPair returned error: %v", err)
	}

	// Now create access token with refresh token
	newAccessToken, err := NewAccessTokenWithRefreshToken(ctx, userID, tokenPair.RefreshToken)
	if err != nil {
		t.Errorf("NewAccessTokenWithRefreshToken returned error: %v", err)
	}

	if newAccessToken == "" {
		t.Error("Expected non-empty token")
	}

	// Test with wrong userID
	_, err = NewAccessTokenWithRefreshToken(ctx, "wrong-user", tokenPair.RefreshToken)
	if err == nil {
		t.Error("Expected error for wrong userID")
	}

	// Test with invalid refresh token
	_, err = NewAccessTokenWithRefreshToken(ctx, userID, "invalid-token")
	if err == nil {
		t.Error("Expected error for invalid refresh token")
	}
}

func TestValidateAccessToken(t *testing.T) {
	ctx := core.NewCtx(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), core.Route{}, core.Endpoint{})

	userID := "user999"

	// Create a valid token
	token, err := NewAccessToken(ctx, userID)
	if err != nil {
		t.Fatalf("NewAccessToken returned error: %v", err)
	}

	// Test validation
	err = ValidateAccessToken(ctx, token)
	if err != nil {
		t.Errorf("ValidateAccessToken returned error for valid token: %v", err)
	}

	if ctx.Session["userID"] != userID {
		t.Errorf("Expected session userID '%s', got '%v'", userID, ctx.Session["userID"])
	}

	// Test with Bearer prefix
	ctx2 := core.NewCtx(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), core.Route{}, core.Endpoint{})
	err = ValidateAccessToken(ctx2, "Bearer "+token)
	if err != nil {
		t.Errorf("ValidateAccessToken returned error for token with Bearer prefix: %v", err)
	}

	// Test with empty token
	err = ValidateAccessToken(ctx, "")
	if err == nil {
		t.Error("Expected error for empty token")
	}

	// Test with invalid token
	err = ValidateAccessToken(ctx, "invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestValidateRefreshToken(t *testing.T) {
	ctx := core.NewCtx(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), core.Route{}, core.Endpoint{})

	userID := "user111"

	// Create a token pair
	tokenPair, err := NewTokenPair(ctx, userID)
	if err != nil {
		t.Fatalf("NewTokenPair returned error: %v", err)
	}

	// Test validation
	returnedUserID, err := ValidateRefreshToken(ctx, tokenPair.RefreshToken)
	if err != nil {
		t.Errorf("ValidateRefreshToken returned error for valid token: %v", err)
	}

	if returnedUserID != userID {
		t.Errorf("Expected userID '%s', got '%v'", userID, returnedUserID)
	}

	// Test with empty token
	_, err = ValidateRefreshToken(ctx, "")
	if err == nil {
		t.Error("Expected error for empty token")
	}

	// Test with invalid token
	_, err = ValidateRefreshToken(ctx, "invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestGetSubFromToken(t *testing.T) {
	ctx := core.NewCtx(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), core.Route{}, core.Endpoint{})

	userID := "user222"

	// Create a token
	token, err := NewAccessToken(ctx, userID)
	if err != nil {
		t.Fatalf("NewAccessToken returned error: %v", err)
	}

	// Set authorization header
	ctx.Request.Header.Set("Authorization", "Bearer "+token)

	// Test getting sub
	sub, err := GetSubFromToken(ctx)
	if err != nil {
		t.Errorf("GetSubFromToken returned error: %v", err)
	}

	if sub != userID {
		t.Errorf("Expected sub '%s', got '%v'", userID, sub)
	}

	// Test without Bearer prefix
	ctx2 := core.NewCtx(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), core.Route{}, core.Endpoint{})
	ctx2.Request.Header.Set("Authorization", token)
	sub2, err := GetSubFromToken(ctx2)
	if err != nil {
		t.Errorf("GetSubFromToken returned error without Bearer prefix: %v", err)
	}

	if sub2 != userID {
		t.Errorf("Expected sub '%s', got '%v'", userID, sub2)
	}

	// Test without authorization header
	ctx3 := core.NewCtx(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), core.Route{}, core.Endpoint{})
	_, err = GetSubFromToken(ctx3)
	if err == nil {
		t.Error("Expected error without authorization header")
	}

	// Test with invalid token
	ctx4 := core.NewCtx(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), core.Route{}, core.Endpoint{})
	ctx4.Request.Header.Set("Authorization", "Bearer invalid-token")
	_, err = GetSubFromToken(ctx4)
	if err == nil {
		t.Error("Expected error with invalid token")
	}
}

func TestJWTMiddleware(t *testing.T) {
	mockHandler := func(c *core.Ctx) error {
		c.Response.SetMessage("OK")
		return nil
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedError  bool
		expectedStatus int
	}{
		{
			name:           "valid token",
			authHeader:     "", // Will be set in test
			expectedError:  false,
			expectedStatus: 200,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedError:  true,
			expectedStatus: 0,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			expectedError:  true,
			expectedStatus: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			} else if tt.name == "valid token" {
				// Create a valid token for this test
				ctx := core.NewCtx(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil), core.Route{}, core.Endpoint{})
				token, err := NewAccessToken(ctx, "test-user")
				if err != nil {
					t.Fatalf("Failed to create token: %v", err)
				}
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			ctx := core.NewCtx(w, req, core.Route{}, core.Endpoint{})

			middleware := JWTMiddleware(JWTConfig{})
			handler := middleware(mockHandler)

			err := handler(ctx)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tt.expectedStatus != 0 {
				ctx.SendingReturn(w, nil)
				if w.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
				}
			}
		})
	}
}
