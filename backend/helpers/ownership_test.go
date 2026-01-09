package helpers

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"m2m-backend/database"
	"m2m-backend/models"
)

// setupTestDB initializes a test database connection
func setupTestDB(t *testing.T) func() {
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

	// Set the test database
	database.DB = client.Database("m2m_test_ownership")

	// Cleanup function
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		database.DB.Drop(ctx)
		client.Disconnect(ctx)
	}

	return cleanup
}

// createTestUser creates a test user and returns the ID
func createTestUser(t *testing.T, email string) primitive.ObjectID {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userID := primitive.NewObjectID()
	user := models.User{
		ID:           userID,
		Name:         "Test User",
		Email:        email,
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
	}

	collection := database.GetCollection("users")
	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return userID
}

// createTestCompany creates a test company and returns the ID
func createTestCompany(t *testing.T, userID primitive.ObjectID, name string) primitive.ObjectID {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	companyID := primitive.NewObjectID()
	company := models.Company{
		ID:        companyID,
		UserID:    userID,
		Name:      name,
		CreatedAt: time.Now(),
	}

	collection := database.GetCollection("companies")
	_, err := collection.InsertOne(ctx, company)
	if err != nil {
		t.Fatalf("Failed to create test company: %v", err)
	}

	return companyID
}

// createTestTransaction creates a test transaction and returns the ID
func createTestTransaction(t *testing.T, companyID primitive.ObjectID) primitive.ObjectID {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transactionID := primitive.NewObjectID()
	transaction := models.Transaction{
		ID:          transactionID,
		CompanyID:   companyID,
		Month:       "Janeiro",
		Year:        2024,
		Amount:      1000.0,
		Category:    "Salário",
		Status:      models.StatusPaid,
		Description: "Test transaction",
	}

	collection := database.GetCollection("transactions")
	_, err := collection.InsertOne(ctx, transaction)
	if err != nil {
		t.Fatalf("Failed to create test transaction: %v", err)
	}

	return transactionID
}

func TestValidateCompanyOwnership(t *testing.T) {
	cleanup := setupTestDB(t)
	if cleanup == nil {
		return // Test skipped
	}
	defer cleanup()

	// Setup test data
	ownerUserID := createTestUser(t, "owner@example.com")
	otherUserID := createTestUser(t, "other@example.com")
	companyID := createTestCompany(t, ownerUserID, "Test Company")

	tests := []struct {
		name           string
		userID         primitive.ObjectID
		companyID      primitive.ObjectID
		expectError    bool
		expectedStatus int
	}{
		{
			name:           "Valid owner access",
			userID:         ownerUserID,
			companyID:      companyID,
			expectError:    false,
			expectedStatus: 0,
		},
		{
			name:           "Access denied - different user",
			userID:         otherUserID,
			companyID:      companyID,
			expectError:    true,
			expectedStatus: 403,
		},
		{
			name:           "Company not found",
			userID:         ownerUserID,
			companyID:      primitive.NewObjectID(),
			expectError:    true,
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			var company *models.Company
			var err error
			var statusCode int

			app.Get("/test", func(c *fiber.Ctx) error {
				c.Locals("userId", tt.userID.Hex())
				company, err = ValidateCompanyOwnership(c, tt.companyID)
				statusCode = c.Response().StatusCode()
				return nil
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, _ := app.Test(req, -1)
			defer resp.Body.Close()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if statusCode != tt.expectedStatus && statusCode != 0 {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, statusCode)
				}
				if company != nil {
					t.Errorf("Expected nil company on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if company == nil {
					t.Errorf("Expected company, got nil")
				}
				if company != nil && company.ID != tt.companyID {
					t.Errorf("Expected company ID %v, got %v", tt.companyID, company.ID)
				}
			}
		})
	}
}

func TestValidateCompanyOwnershipByString(t *testing.T) {
	cleanup := setupTestDB(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	ownerUserID := createTestUser(t, "owner@example.com")
	companyID := createTestCompany(t, ownerUserID, "Test Company")

	tests := []struct {
		name           string
		companyIDStr   string
		userID         primitive.ObjectID
		expectError    bool
		expectedStatus int
	}{
		{
			name:           "Valid company ID string",
			companyIDStr:   companyID.Hex(),
			userID:         ownerUserID,
			expectError:    false,
			expectedStatus: 0,
		},
		{
			name:           "Invalid company ID format",
			companyIDStr:   "invalid-id",
			userID:         ownerUserID,
			expectError:    true,
			expectedStatus: 400,
		},
		{
			name:           "Empty company ID",
			companyIDStr:   "",
			userID:         ownerUserID,
			expectError:    true,
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			var company *models.Company
			var err error

			app.Get("/test", func(c *fiber.Ctx) error {
				c.Locals("userId", tt.userID.Hex())
				company, err = ValidateCompanyOwnershipByString(c, tt.companyIDStr)
				return nil
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, _ := app.Test(req, -1)
			defer resp.Body.Close()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if company == nil {
					t.Errorf("Expected company, got nil")
				}
			}
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name        string
		userIDValue interface{}
		expectError bool
	}{
		{
			name:        "Valid user ID",
			userIDValue: primitive.NewObjectID().Hex(),
			expectError: false,
		},
		{
			name:        "Missing user ID in context",
			userIDValue: nil,
			expectError: true,
		},
		{
			name:        "Invalid user ID type",
			userIDValue: 12345,
			expectError: true,
		},
		{
			name:        "Invalid user ID format",
			userIDValue: "invalid-id",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			var userID primitive.ObjectID
			var err error

			app.Get("/test", func(c *fiber.Ctx) error {
				if tt.userIDValue != nil {
					c.Locals("userId", tt.userIDValue)
				}
				userID, err = GetUserIDFromContext(c)
				return nil
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, _ := app.Test(req, -1)
			defer resp.Body.Close()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if userID != primitive.NilObjectID {
					t.Errorf("Expected NilObjectID on error, got %v", userID)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if userID == primitive.NilObjectID {
					t.Errorf("Expected valid ObjectID, got NilObjectID")
				}
			}
		})
	}
}

func TestValidateTransactionOwnership(t *testing.T) {
	cleanup := setupTestDB(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	ownerUserID := createTestUser(t, "owner@example.com")
	otherUserID := createTestUser(t, "other@example.com")
	companyID := createTestCompany(t, ownerUserID, "Test Company")
	transactionID := createTestTransaction(t, companyID)

	tests := []struct {
		name          string
		userID        primitive.ObjectID
		transactionID primitive.ObjectID
		expectError   bool
	}{
		{
			name:          "Valid transaction access by owner",
			userID:        ownerUserID,
			transactionID: transactionID,
			expectError:   false,
		},
		{
			name:          "Access denied - different user",
			userID:        otherUserID,
			transactionID: transactionID,
			expectError:   true,
		},
		{
			name:          "Transaction not found",
			userID:        ownerUserID,
			transactionID: primitive.NewObjectID(),
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			var transaction *models.Transaction
			var err error

			app.Get("/test", func(c *fiber.Ctx) error {
				c.Locals("userId", tt.userID.Hex())
				transaction, err = ValidateTransactionOwnership(c, tt.transactionID)
				return nil
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, _ := app.Test(req, -1)
			defer resp.Body.Close()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if transaction != nil {
					t.Errorf("Expected nil transaction on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if transaction == nil {
					t.Errorf("Expected transaction, got nil")
				}
				if transaction != nil && transaction.ID != tt.transactionID {
					t.Errorf("Expected transaction ID %v, got %v", tt.transactionID, transaction.ID)
				}
			}
		})
	}
}

// Edge case tests
func TestValidateCompanyOwnershipByString_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		companyIDStr string
		expectError  bool
	}{
		{
			name:         "Very long invalid ID",
			companyIDStr: "abcdefghijklmnopqrstuvwxyz1234567890",
			expectError:  true,
		},
		{
			name:         "Special characters in ID",
			companyIDStr: "abc@#$%^&*()",
			expectError:  true,
		},
		{
			name:         "Spaces in ID",
			companyIDStr: "abc def 123",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			var err error

			app.Get("/test", func(c *fiber.Ctx) error {
				userID := primitive.NewObjectID()
				c.Locals("userId", userID.Hex())
				_, err = ValidateCompanyOwnershipByString(c, tt.companyIDStr)
				return nil
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, _ := app.Test(req, -1)
			defer resp.Body.Close()

			if tt.expectError && err == nil {
				t.Errorf("Expected error for invalid company ID format")
			}
		})
	}
}

func TestGetUserIDFromContext_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		userIDValue interface{}
		expectError bool
	}{
		{
			name:        "Empty string",
			userIDValue: "",
			expectError: true,
		},
		{
			name:        "Boolean value",
			userIDValue: true,
			expectError: true,
		},
		{
			name:        "Float value",
			userIDValue: 123.45,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			var userID primitive.ObjectID
			var err error

			app.Get("/test", func(c *fiber.Ctx) error {
				c.Locals("userId", tt.userIDValue)
				userID, err = GetUserIDFromContext(c)
				return nil
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, _ := app.Test(req, -1)
			defer resp.Body.Close()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if userID != primitive.NilObjectID {
					t.Errorf("Expected NilObjectID on error")
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkGetUserIDFromContext(b *testing.B) {
	app := fiber.New()
	userID := primitive.NewObjectID().Hex()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Get("/test", func(c *fiber.Ctx) error {
			c.Locals("userId", userID)
			GetUserIDFromContext(c)
			return nil
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, _ := app.Test(req, -1)
		resp.Body.Close()
	}
}
