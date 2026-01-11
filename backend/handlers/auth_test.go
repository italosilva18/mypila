package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"m2m-backend/config"
	"m2m-backend/database"
	"m2m-backend/models"
)

// init initializes database for testing to prevent panic from global variables
func init() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try Docker port first (27018) with auth, then local (27017) without auth
	mongoURI := "mongodb://admin:admin123@localhost:27018"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		// Try local MongoDB without auth
		mongoURI = "mongodb://localhost:27017"
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
		if err != nil {
			// If MongoDB is not available, use a dummy database reference
			// Tests will skip gracefully
			return
		}
	}

	database.DB = client.Database("m2m_test_init")
	config.JWTSecret = "test-secret-for-init"
}

// setupTestDBForAuth initializes test database for auth tests
func setupTestDBForAuth(t *testing.T) func() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try Docker port first (27018) with auth, then local (27017) without auth
	mongoURI := "mongodb://admin:admin123@localhost:27018"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		// Try local MongoDB without auth
		mongoURI = "mongodb://localhost:27017"
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
		if err != nil {
			t.Skipf("Skipping test: MongoDB not available - %v", err)
			return nil
		}
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("Skipping test: MongoDB not responding - %v", err)
		return nil
	}

	database.DB = client.Database("m2m_test_auth")

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		database.DB.Drop(ctx)
		client.Disconnect(ctx)
	}

	return cleanup
}

// setupTestApp creates a test Fiber app with routes
func setupTestApp() *fiber.App {
	app := fiber.New()
	app.Post("/register", Register)
	app.Post("/login", Login)
	app.Get("/me", GetMe)
	return app
}

func TestRegister(t *testing.T) {
	cleanup := setupTestDBForAuth(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	// Initialize JWT secret for tests
	config.JWTSecret = "test-secret-key-for-testing-only"

	tests := []struct {
		name           string
		requestBody    models.RegisterRequest
		expectedStatus int
		expectToken    bool
		expectError    bool
	}{
		{
			name: "Valid registration",
			requestBody: models.RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			expectedStatus: 201,
			expectToken:    true,
			expectError:    false,
		},
		{
			name: "Duplicate email",
			requestBody: models.RegisterRequest{
				Name:     "Jane Doe",
				Email:    "duplicate@example.com",
				Password: "password123",
			},
			expectedStatus: 409,
			expectToken:    false,
			expectError:    true,
		},
		{
			name: "Missing name",
			requestBody: models.RegisterRequest{
				Name:     "",
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus: 400,
			expectToken:    false,
			expectError:    true,
		},
		{
			name: "Invalid email format",
			requestBody: models.RegisterRequest{
				Name:     "Test User",
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedStatus: 400,
			expectToken:    false,
			expectError:    true,
		},
		{
			name: "Password too short",
			requestBody: models.RegisterRequest{
				Name:     "Test User",
				Email:    "short@example.com",
				Password: "12345",
			},
			expectedStatus: 400,
			expectToken:    false,
			expectError:    true,
		},
		{
			name: "Empty email",
			requestBody: models.RegisterRequest{
				Name:     "Test User",
				Email:    "",
				Password: "password123",
			},
			expectedStatus: 400,
			expectToken:    false,
			expectError:    true,
		},
		{
			name: "XSS attempt in name",
			requestBody: models.RegisterRequest{
				Name:     "<script>alert('xss')</script>",
				Email:    "xss@example.com",
				Password: "password123",
			},
			expectedStatus: 400,
			expectToken:    false,
			expectError:    true,
		},
	}

	// Pre-create duplicate user
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	database.GetCollection("users").InsertOne(ctx, bson.M{
		"name":         "Duplicate User",
		"email":        "duplicate@example.com",
		"passwordHash": string(hash),
		"createdAt":    time.Now(),
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestApp()

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, resp.StatusCode, string(body))
			}

			body, _ := io.ReadAll(resp.Body)

			if tt.expectToken {
				var authResp models.AuthResponseWithTokens
				err := json.Unmarshal(body, &authResp)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if authResp.AccessToken == "" {
					t.Errorf("Expected access token in response, got empty string")
				}
				if authResp.RefreshToken == "" {
					t.Errorf("Expected refresh token in response, got empty string")
				}
				if authResp.User.Email != tt.requestBody.Email {
					t.Errorf("Expected user email %s, got %s", tt.requestBody.Email, authResp.User.Email)
				}
			}

			if tt.expectError {
				var errorResp map[string]interface{}
				err := json.Unmarshal(body, &errorResp)
				if err != nil {
					t.Errorf("Failed to unmarshal error response: %v", err)
				}
				// Check for either "error" (API errors) or "errors" (validation errors)
				_, hasError := errorResp["error"]
				_, hasErrors := errorResp["errors"]
				if !hasError && !hasErrors {
					t.Errorf("Expected error or errors field in response, got: %s", string(body))
				}
			}
		})
	}
}

func TestLogin(t *testing.T) {
	cleanup := setupTestDBForAuth(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	config.JWTSecret = "test-secret-key-for-testing-only"

	// Create test user
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	database.GetCollection("users").InsertOne(ctx, bson.M{
		"name":         "Test User",
		"email":        "test@example.com",
		"passwordHash": string(hash),
		"createdAt":    time.Now(),
	})

	tests := []struct {
		name           string
		requestBody    models.LoginRequest
		expectedStatus int
		expectToken    bool
		expectError    bool
	}{
		{
			name: "Valid login credentials",
			requestBody: models.LoginRequest{
				Email:    "test@example.com",
				Password: "correctpassword",
			},
			expectedStatus: 200,
			expectToken:    true,
			expectError:    false,
		},
		{
			name: "Wrong password",
			requestBody: models.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			expectedStatus: 401,
			expectToken:    false,
			expectError:    true,
		},
		{
			name: "Non-existent user",
			requestBody: models.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "anypassword",
			},
			expectedStatus: 401,
			expectToken:    false,
			expectError:    true,
		},
		{
			name: "Empty email",
			requestBody: models.LoginRequest{
				Email:    "",
				Password: "password123",
			},
			expectedStatus: 400,
			expectToken:    false,
			expectError:    true,
		},
		{
			name: "Empty password",
			requestBody: models.LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			expectedStatus: 400,
			expectToken:    false,
			expectError:    true,
		},
		{
			name: "Invalid email format",
			requestBody: models.LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedStatus: 400,
			expectToken:    false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestApp()

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, resp.StatusCode, string(body))
			}

			body, _ := io.ReadAll(resp.Body)

			if tt.expectToken {
				var authResp models.AuthResponseWithTokens
				err := json.Unmarshal(body, &authResp)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if authResp.AccessToken == "" {
					t.Errorf("Expected access token in response, got empty string")
				}
				if authResp.RefreshToken == "" {
					t.Errorf("Expected refresh token in response, got empty string")
				}
				if authResp.User.Email != tt.requestBody.Email {
					t.Errorf("Expected user email %s, got %s", tt.requestBody.Email, authResp.User.Email)
				}

				// Verify access token is valid
				token, err := jwt.Parse(authResp.AccessToken, func(token *jwt.Token) (interface{}, error) {
					return []byte(config.JWTSecret), nil
				})
				if err != nil || !token.Valid {
					t.Errorf("Generated access token is invalid: %v", err)
				}
			}

			if tt.expectError {
				var errorResp map[string]interface{}
				err := json.Unmarshal(body, &errorResp)
				if err != nil {
					t.Errorf("Failed to unmarshal error response: %v", err)
				}
				// Check for either "error" (API errors) or "errors" (validation errors)
				_, hasError := errorResp["error"]
				_, hasErrors := errorResp["errors"]
				if !hasError && !hasErrors {
					t.Errorf("Expected error or errors field in response, got: %s", string(body))
				}
			}
		})
	}
}

func TestGetMe(t *testing.T) {
	cleanup := setupTestDBForAuth(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	config.JWTSecret = "test-secret-key-for-testing-only"

	// Create test user
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	result, _ := database.GetCollection("users").InsertOne(ctx, bson.M{
		"name":         "Test User",
		"email":        "me@example.com",
		"passwordHash": string(hash),
		"createdAt":    time.Now(),
	})
	userID := result.InsertedID.(primitive.ObjectID)

	tests := []struct {
		name           string
		userIDInCtx    interface{}
		expectedStatus int
		expectUser     bool
	}{
		{
			name:           "Valid user ID in context",
			userIDInCtx:    userID.Hex(), // Pass as string, as expected by the handler
			expectedStatus: 200,
			expectUser:     true,
		},
		{
			name:           "Missing user ID in context",
			userIDInCtx:    nil,
			expectedStatus: 401,
			expectUser:     false,
		},
		{
			name:           "Invalid user ID type",
			userIDInCtx:    12345,
			expectedStatus: 401,
			expectUser:     false,
		},
		{
			name:           "Invalid user ID format",
			userIDInCtx:    "invalid-id",
			expectedStatus: 400,
			expectUser:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/me", func(c *fiber.Ctx) error {
				if tt.userIDInCtx != nil {
					c.Locals("userId", tt.userIDInCtx)
				}
				return GetMe(c)
			})

			req := httptest.NewRequest("GET", "/me", nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, resp.StatusCode, string(body))
			}

			body, _ := io.ReadAll(resp.Body)

			if tt.expectUser {
				var userResp map[string]models.User
				err := json.Unmarshal(body, &userResp)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if user, exists := userResp["user"]; !exists {
					t.Errorf("Expected user field in response")
				} else if user.Email != "me@example.com" {
					t.Errorf("Expected email me@example.com, got %s", user.Email)
				}
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	config.JWTSecret = "test-secret-key-for-testing-only"

	user := models.User{
		Email: "test@example.com",
	}

	token, err := generateAccessToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Errorf("Expected non-empty token")
	}

	// Parse and validate token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(config.JWTSecret), nil
	})

	if err != nil {
		t.Errorf("Failed to parse token: %v", err)
	}

	if !parsedToken.Valid {
		t.Errorf("Generated token is not valid")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("Failed to extract claims from token")
	}

	if claims["email"] != user.Email {
		t.Errorf("Expected email %s in claims, got %v", user.Email, claims["email"])
	}

	if _, exists := claims["userId"]; !exists {
		t.Errorf("Expected userId in claims")
	}

	if _, exists := claims["exp"]; !exists {
		t.Errorf("Expected exp (expiration) in claims")
	}
}

func TestRegisterInvalidJSON(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("Expected status 400 for invalid JSON, got %d", resp.StatusCode)
	}
}

func TestLoginInvalidJSON(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("Expected status 400 for invalid JSON, got %d", resp.StatusCode)
	}
}

// Edge case tests
func TestRegisterEdgeCases(t *testing.T) {
	cleanup := setupTestDBForAuth(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	config.JWTSecret = "test-secret-key-for-testing-only"

	tests := []struct {
		name           string
		requestBody    models.RegisterRequest
		expectedStatus int
	}{
		{
			name: "Very long name",
			requestBody: models.RegisterRequest{
				Name:     string(make([]byte, 1000)),
				Email:    "longname@example.com",
				Password: "password123",
			},
			expectedStatus: 400,
		},
		{
			name: "SQL injection in name",
			requestBody: models.RegisterRequest{
				Name:     "'; DROP TABLE users; --",
				Email:    "sqlinjection@example.com",
				Password: "password123",
			},
			expectedStatus: 400,
		},
		{
			name: "MongoDB injection in name",
			requestBody: models.RegisterRequest{
				Name:     "$ne",
				Email:    "mongoinjection@example.com",
				Password: "password123",
			},
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestApp()

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

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
func BenchmarkGenerateToken(b *testing.B) {
	config.JWTSecret = "test-secret-key-for-testing-only"
	user := models.User{
		Email: "bench@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateAccessToken(user)
	}
}
