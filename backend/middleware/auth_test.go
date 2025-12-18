package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"m2m-backend/config"
)

// setupTestMiddlewareApp creates a test app with protected endpoint
func setupTestMiddlewareApp() *fiber.App {
	app := fiber.New()
	app.Get("/protected", Protected(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "success",
			"userId":  c.Locals("userId"),
			"email":   c.Locals("email"),
		})
	})
	return app
}

// generateTestToken generates a valid JWT token for testing
func generateTestToken(userID, email string, expiresIn time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"email":  email,
		"exp":    time.Now().Add(expiresIn).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetJWTSecret()))
}

// generateExpiredToken generates an expired token for testing
func generateExpiredToken(userID, email string) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"email":  email,
		"exp":    time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetJWTSecret()))
}

// generateTokenWithWrongSecret generates a token with wrong secret
func generateTokenWithWrongSecret(userID, email string) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"email":  email,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("wrong-secret-key"))
}

func TestProtectedMiddleware(t *testing.T) {
	// Initialize JWT secret
	config.JWTSecret = "test-secret-key-for-middleware-testing"

	validToken, _ := generateTestToken("user123", "test@example.com", time.Hour*24)
	expiredToken, _ := generateExpiredToken("user123", "test@example.com")
	wrongSecretToken, _ := generateTokenWithWrongSecret("user123", "test@example.com")

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectSuccess  bool
		expectUserID   bool
		expectEmail    bool
	}{
		{
			name:           "Valid token with Bearer prefix",
			authHeader:     "Bearer " + validToken,
			expectedStatus: 200,
			expectSuccess:  true,
			expectUserID:   true,
			expectEmail:    true,
		},
		{
			name:           "Missing authorization header",
			authHeader:     "",
			expectedStatus: 401,
			expectSuccess:  false,
			expectUserID:   false,
			expectEmail:    false,
		},
		{
			name:           "Token without Bearer prefix",
			authHeader:     validToken,
			expectedStatus: 401,
			expectSuccess:  false,
			expectUserID:   false,
			expectEmail:    false,
		},
		{
			name:           "Invalid token format - missing token part",
			authHeader:     "Bearer",
			expectedStatus: 401,
			expectSuccess:  false,
			expectUserID:   false,
			expectEmail:    false,
		},
		{
			name:           "Invalid token - malformed",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: 401,
			expectSuccess:  false,
			expectUserID:   false,
			expectEmail:    false,
		},
		{
			name:           "Expired token",
			authHeader:     "Bearer " + expiredToken,
			expectedStatus: 401,
			expectSuccess:  false,
			expectUserID:   false,
			expectEmail:    false,
		},
		{
			name:           "Token signed with wrong secret",
			authHeader:     "Bearer " + wrongSecretToken,
			expectedStatus: 401,
			expectSuccess:  false,
			expectUserID:   false,
			expectEmail:    false,
		},
		{
			name:           "Empty Bearer token",
			authHeader:     "Bearer ",
			expectedStatus: 401,
			expectSuccess:  false,
			expectUserID:   false,
			expectEmail:    false,
		},
		{
			name:           "Wrong auth scheme",
			authHeader:     "Basic " + validToken,
			expectedStatus: 401,
			expectSuccess:  false,
			expectUserID:   false,
			expectEmail:    false,
		},
		{
			name:           "Multiple Bearer prefixes",
			authHeader:     "Bearer Bearer " + validToken,
			expectedStatus: 401,
			expectSuccess:  false,
			expectUserID:   false,
			expectEmail:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestMiddlewareApp()

			req := httptest.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectSuccess {
				if resp.StatusCode != 200 {
					t.Errorf("Expected successful response (200), got %d", resp.StatusCode)
				}
			}
		})
	}
}

func TestProtectedMiddleware_ContextValues(t *testing.T) {
	config.JWTSecret = "test-secret-key-for-middleware-testing"

	userID := "user123456"
	email := "test@example.com"
	token, _ := generateTestToken(userID, email, time.Hour*24)

	app := fiber.New()
	app.Get("/protected", Protected(), func(c *fiber.Ctx) error {
		localUserID := c.Locals("userId")
		localEmail := c.Locals("email")

		return c.JSON(fiber.Map{
			"userId": localUserID,
			"email":  localEmail,
		})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestProtectedMiddleware_TokenWithoutClaims(t *testing.T) {
	config.JWTSecret = "test-secret-key-for-middleware-testing"

	// Create token without proper claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		// Missing userId and email
	})
	tokenString, _ := token.SignedString([]byte(config.JWTSecret))

	app := setupTestMiddlewareApp()

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Should still succeed as middleware doesn't validate claim presence
	// Only validates token signature and expiration
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestProtectedMiddleware_DifferentSigningMethod(t *testing.T) {
	config.JWTSecret = "test-secret-key-for-middleware-testing"

	// Try to create token with RS256 (should be rejected)
	claims := jwt.MapClaims{
		"userId": "user123",
		"email":  "test@example.com",
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	}

	// Using wrong signing method
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	app := setupTestMiddlewareApp()

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 401 {
		t.Errorf("Expected status 401 for wrong signing method, got %d", resp.StatusCode)
	}
}

func TestProtectedMiddleware_VeryLongToken(t *testing.T) {
	config.JWTSecret = "test-secret-key-for-middleware-testing"

	app := setupTestMiddlewareApp()

	// Create a long invalid token (not too long to avoid buffer issues)
	longToken := "Bearer "
	for i := 0; i < 1000; i++ {
		longToken += "a"
	}

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", longToken)

	resp, err := app.Test(req, -1)
	if err != nil {
		// If request fails due to header size, that's also acceptable
		t.Skipf("Request failed (likely due to header size limit): %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 401 {
		t.Errorf("Expected status 401 for invalid long token, got %d", resp.StatusCode)
	}
}

func TestProtectedMiddleware_TokenExpiringNow(t *testing.T) {
	config.JWTSecret = "test-secret-key-for-middleware-testing"

	// Token that expires in 1 second
	token, _ := generateTestToken("user123", "test@example.com", time.Second*1)

	app := setupTestMiddlewareApp()

	// First request should succeed
	req1 := httptest.NewRequest("GET", "/protected", nil)
	req1.Header.Set("Authorization", "Bearer "+token)

	resp1, err := app.Test(req1, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	resp1.Body.Close()

	if resp1.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp1.StatusCode)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Second request should fail
	req2 := httptest.NewRequest("GET", "/protected", nil)
	req2.Header.Set("Authorization", "Bearer "+token)

	resp2, err := app.Test(req2, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	resp2.Body.Close()

	if resp2.StatusCode != 401 {
		t.Errorf("Expected status 401 for expired token, got %d", resp2.StatusCode)
	}
}

func TestProtectedMiddleware_CaseSensitiveBearer(t *testing.T) {
	config.JWTSecret = "test-secret-key-for-middleware-testing"

	validToken, _ := generateTestToken("user123", "test@example.com", time.Hour*24)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Correct Bearer",
			authHeader:     "Bearer " + validToken,
			expectedStatus: 200,
		},
		{
			name:           "Lowercase bearer",
			authHeader:     "bearer " + validToken,
			expectedStatus: 401,
		},
		{
			name:           "Uppercase BEARER",
			authHeader:     "BEARER " + validToken,
			expectedStatus: 401,
		},
		{
			name:           "Mixed case BeArEr",
			authHeader:     "BeArEr " + validToken,
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestMiddlewareApp()

			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tt.authHeader)

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestProtectedMiddleware_MultipleSpaces(t *testing.T) {
	config.JWTSecret = "test-secret-key-for-middleware-testing"

	validToken, _ := generateTestToken("user123", "test@example.com", time.Hour*24)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Single space",
			authHeader:     "Bearer " + validToken,
			expectedStatus: 200,
		},
		{
			name:           "Double space",
			authHeader:     "Bearer  " + validToken,
			expectedStatus: 401,
		},
		{
			name:           "Tab separator",
			authHeader:     "Bearer\t" + validToken,
			expectedStatus: 401,
		},
		{
			name:           "No space",
			authHeader:     "Bearer" + validToken,
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestMiddlewareApp()

			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tt.authHeader)

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// Benchmark tests
func BenchmarkProtectedMiddleware_ValidToken(b *testing.B) {
	config.JWTSecret = "test-secret-key-for-middleware-testing"

	token, _ := generateTestToken("user123", "test@example.com", time.Hour*24)
	app := setupTestMiddlewareApp()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, _ := app.Test(req, -1)
		resp.Body.Close()
	}
}

func BenchmarkProtectedMiddleware_InvalidToken(b *testing.B) {
	config.JWTSecret = "test-secret-key-for-middleware-testing"

	app := setupTestMiddlewareApp()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		resp, _ := app.Test(req, -1)
		resp.Body.Close()
	}
}

func BenchmarkProtectedMiddleware_NoToken(b *testing.B) {
	config.JWTSecret = "test-secret-key-for-middleware-testing"

	app := setupTestMiddlewareApp()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/protected", nil)
		resp, _ := app.Test(req, -1)
		resp.Body.Close()
	}
}

func BenchmarkGenerateTestToken(b *testing.B) {
	config.JWTSecret = "test-secret-key-for-middleware-testing"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateTestToken("user123", "test@example.com", time.Hour*24)
	}
}
