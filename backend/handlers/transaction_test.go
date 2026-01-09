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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"m2m-backend/database"
	"m2m-backend/models"
)

// setupTestDBForTransactions initializes test database for transaction tests
func setupTestDBForTransactions(t *testing.T) func() {
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

	database.DB = client.Database("m2m_test_transactions")

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		database.DB.Drop(ctx)
		client.Disconnect(ctx)
	}

	return cleanup
}

// createTestUserAndCompany creates a test user and company, returning their IDs
func createTestUserAndCompany(t *testing.T) (primitive.ObjectID, primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create test user
	userResult, err := database.GetCollection("users").InsertOne(ctx, bson.M{
		"name":         "Test User",
		"email":        "testuser@example.com",
		"passwordHash": "hashedpassword",
		"createdAt":    time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	userID := userResult.InsertedID.(primitive.ObjectID)

	// Create test company
	companyResult, err := database.GetCollection("companies").InsertOne(ctx, bson.M{
		"userId":    userID,
		"name":      "Test Company",
		"createdAt": time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test company: %v", err)
	}
	companyID := companyResult.InsertedID.(primitive.ObjectID)

	return userID, companyID
}

// setupTransactionTestApp creates a Fiber app with transaction routes and user context
func setupTransactionTestApp(userID primitive.ObjectID) *fiber.App {
	app := fiber.New()

	// Middleware to set user context
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", userID.Hex())
		return c.Next()
	})

	app.Get("/transactions", GetAllTransactions)
	app.Get("/transactions/:id", GetTransaction)
	app.Post("/transactions", CreateTransaction)
	app.Put("/transactions/:id", UpdateTransaction)
	app.Delete("/transactions/:id", DeleteTransaction)
	app.Patch("/transactions/:id/toggle-status", ToggleStatus)
	app.Get("/stats", GetStats)

	return app
}

func TestGetAllTransactions(t *testing.T) {
	cleanup := setupTestDBForTransactions(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompany(t)

	// Create test transactions
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testTransactions := []interface{}{
		models.Transaction{
			ID:          primitive.NewObjectID(),
			CompanyID:   companyID,
			Month:       "Janeiro",
			Year:        2024,
			Amount:      1000.00,
			Category:    "Salario",
			Status:      models.StatusPaid,
			Description: "Salario Janeiro",
		},
		models.Transaction{
			ID:          primitive.NewObjectID(),
			CompanyID:   companyID,
			Month:       "Fevereiro",
			Year:        2024,
			Amount:      1500.00,
			Category:    "Bonus",
			Status:      models.StatusOpen,
			Description: "Bonus trimestral",
		},
	}
	database.GetCollection(collectionName).InsertMany(ctx, testTransactions)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Get all transactions for company",
			queryParams:    "?companyId=" + companyID.Hex(),
			expectedStatus: 200,
			expectedCount:  2,
		},
		{
			name:           "Get transactions with pagination",
			queryParams:    "?companyId=" + companyID.Hex() + "&page=1&limit=1",
			expectedStatus: 200,
			expectedCount:  1,
		},
		{
			name:           "Get transactions without companyId returns user transactions",
			queryParams:    "",
			expectedStatus: 200,
			expectedCount:  2,
		},
		{
			name:           "Invalid company ID format",
			queryParams:    "?companyId=invalid-id",
			expectedStatus: 400,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTransactionTestApp(userID)

			req := httptest.NewRequest("GET", "/transactions"+tt.queryParams, nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, resp.StatusCode, string(body))
			}

			if tt.expectedStatus == 200 && tt.expectedCount > 0 {
				body, _ := io.ReadAll(resp.Body)
				var response models.PaginatedTransactions
				if err := json.Unmarshal(body, &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if len(response.Data) != tt.expectedCount {
					t.Errorf("Expected %d transactions, got %d", tt.expectedCount, len(response.Data))
				}
			}
		})
	}
}

func TestCreateTransaction(t *testing.T) {
	cleanup := setupTestDBForTransactions(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompany(t)

	tests := []struct {
		name           string
		requestBody    models.CreateTransactionRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid transaction creation",
			requestBody: models.CreateTransactionRequest{
				CompanyID:   companyID.Hex(),
				Month:       "Janeiro",
				Year:        2024,
				Amount:      2500.00,
				Category:    "Salario",
				Status:      models.StatusPaid,
				Description: "Salario mensal",
			},
			expectedStatus: 201,
			expectError:    false,
		},
		{
			name: "Invalid month",
			requestBody: models.CreateTransactionRequest{
				CompanyID:   companyID.Hex(),
				Month:       "InvalidMonth",
				Year:        2024,
				Amount:      1000.00,
				Category:    "Salario",
				Status:      models.StatusPaid,
				Description: "Test",
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "Invalid year - too low",
			requestBody: models.CreateTransactionRequest{
				CompanyID:   companyID.Hex(),
				Month:       "Janeiro",
				Year:        1999,
				Amount:      1000.00,
				Category:    "Salario",
				Status:      models.StatusPaid,
				Description: "Test",
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "Invalid amount - negative",
			requestBody: models.CreateTransactionRequest{
				CompanyID:   companyID.Hex(),
				Month:       "Janeiro",
				Year:        2024,
				Amount:      -100.00,
				Category:    "Salario",
				Status:      models.StatusPaid,
				Description: "Test",
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "Missing category",
			requestBody: models.CreateTransactionRequest{
				CompanyID:   companyID.Hex(),
				Month:       "Janeiro",
				Year:        2024,
				Amount:      1000.00,
				Category:    "",
				Status:      models.StatusPaid,
				Description: "Test",
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "Invalid status",
			requestBody: models.CreateTransactionRequest{
				CompanyID:   companyID.Hex(),
				Month:       "Janeiro",
				Year:        2024,
				Amount:      1000.00,
				Category:    "Salario",
				Status:      "INVALID",
				Description: "Test",
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "Invalid company ID format",
			requestBody: models.CreateTransactionRequest{
				CompanyID:   "invalid-id",
				Month:       "Janeiro",
				Year:        2024,
				Amount:      1000.00,
				Category:    "Salario",
				Status:      models.StatusPaid,
				Description: "Test",
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "XSS attempt in description",
			requestBody: models.CreateTransactionRequest{
				CompanyID:   companyID.Hex(),
				Month:       "Janeiro",
				Year:        2024,
				Amount:      1000.00,
				Category:    "Salario",
				Status:      models.StatusPaid,
				Description: "<script>alert('xss')</script>",
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "SQL injection attempt",
			requestBody: models.CreateTransactionRequest{
				CompanyID:   companyID.Hex(),
				Month:       "Janeiro",
				Year:        2024,
				Amount:      1000.00,
				Category:    "'; DROP TABLE transactions; --",
				Status:      models.StatusPaid,
				Description: "Test",
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "Description too long",
			requestBody: models.CreateTransactionRequest{
				CompanyID:   companyID.Hex(),
				Month:       "Janeiro",
				Year:        2024,
				Amount:      1000.00,
				Category:    "Salario",
				Status:      models.StatusPaid,
				Description: string(make([]byte, 300)),
			},
			expectedStatus: 400,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTransactionTestApp(userID)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(jsonBody))
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

			if !tt.expectError && tt.expectedStatus == 201 {
				body, _ := io.ReadAll(resp.Body)
				var transaction models.Transaction
				if err := json.Unmarshal(body, &transaction); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if transaction.Amount != tt.requestBody.Amount {
					t.Errorf("Expected amount %f, got %f", tt.requestBody.Amount, transaction.Amount)
				}
				if transaction.Month != tt.requestBody.Month {
					t.Errorf("Expected month %s, got %s", tt.requestBody.Month, transaction.Month)
				}
			}
		})
	}
}

func TestUpdateTransaction(t *testing.T) {
	cleanup := setupTestDBForTransactions(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompany(t)

	// Create a test transaction
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transactionID := primitive.NewObjectID()
	_, err := database.GetCollection(collectionName).InsertOne(ctx, models.Transaction{
		ID:          transactionID,
		CompanyID:   companyID,
		Month:       "Janeiro",
		Year:        2024,
		Amount:      1000.00,
		Category:    "Salario",
		Status:      models.StatusOpen,
		Description: "Original description",
	})
	if err != nil {
		t.Fatalf("Failed to create test transaction: %v", err)
	}

	tests := []struct {
		name           string
		transactionID  string
		requestBody    models.UpdateTransactionRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name:          "Valid update",
			transactionID: transactionID.Hex(),
			requestBody: models.UpdateTransactionRequest{
				Month:       "Fevereiro",
				Year:        2024,
				Amount:      1500.00,
				Category:    "Bonus",
				Status:      models.StatusPaid,
				Description: "Updated description",
			},
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name:          "Invalid transaction ID format",
			transactionID: "invalid-id",
			requestBody: models.UpdateTransactionRequest{
				Month:       "Janeiro",
				Year:        2024,
				Amount:      1000.00,
				Category:    "Salario",
				Status:      models.StatusPaid,
				Description: "Test",
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:          "Non-existent transaction",
			transactionID: primitive.NewObjectID().Hex(),
			requestBody: models.UpdateTransactionRequest{
				Month:       "Janeiro",
				Year:        2024,
				Amount:      1000.00,
				Category:    "Salario",
				Status:      models.StatusPaid,
				Description: "Test",
			},
			expectedStatus: 404,
			expectError:    true,
		},
		{
			name:          "Invalid month in update",
			transactionID: transactionID.Hex(),
			requestBody: models.UpdateTransactionRequest{
				Month:       "InvalidMonth",
				Year:        2024,
				Amount:      1000.00,
				Category:    "Salario",
				Status:      models.StatusPaid,
				Description: "Test",
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:          "Invalid amount - zero",
			transactionID: transactionID.Hex(),
			requestBody: models.UpdateTransactionRequest{
				Month:       "Janeiro",
				Year:        2024,
				Amount:      0,
				Category:    "Salario",
				Status:      models.StatusPaid,
				Description: "Test",
			},
			expectedStatus: 400,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTransactionTestApp(userID)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/transactions/"+tt.transactionID, bytes.NewBuffer(jsonBody))
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

			if !tt.expectError && tt.expectedStatus == 200 {
				body, _ := io.ReadAll(resp.Body)
				var transaction models.Transaction
				if err := json.Unmarshal(body, &transaction); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if transaction.Amount != tt.requestBody.Amount {
					t.Errorf("Expected amount %f, got %f", tt.requestBody.Amount, transaction.Amount)
				}
			}
		})
	}
}

func TestDeleteTransaction(t *testing.T) {
	cleanup := setupTestDBForTransactions(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompany(t)

	// Create a test transaction
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transactionID := primitive.NewObjectID()
	_, err := database.GetCollection(collectionName).InsertOne(ctx, models.Transaction{
		ID:          transactionID,
		CompanyID:   companyID,
		Month:       "Janeiro",
		Year:        2024,
		Amount:      1000.00,
		Category:    "Salario",
		Status:      models.StatusOpen,
		Description: "To be deleted",
	})
	if err != nil {
		t.Fatalf("Failed to create test transaction: %v", err)
	}

	tests := []struct {
		name           string
		transactionID  string
		expectedStatus int
	}{
		{
			name:           "Valid delete",
			transactionID:  transactionID.Hex(),
			expectedStatus: 200,
		},
		{
			name:           "Invalid transaction ID format",
			transactionID:  "invalid-id",
			expectedStatus: 400,
		},
		{
			name:           "Non-existent transaction",
			transactionID:  primitive.NewObjectID().Hex(),
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTransactionTestApp(userID)

			req := httptest.NewRequest("DELETE", "/transactions/"+tt.transactionID, nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, resp.StatusCode, string(body))
			}

			if tt.expectedStatus == 200 {
				body, _ := io.ReadAll(resp.Body)
				var response map[string]string
				if err := json.Unmarshal(body, &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if _, exists := response["message"]; !exists {
					t.Errorf("Expected message field in response")
				}
			}
		})
	}
}

func TestToggleStatus(t *testing.T) {
	cleanup := setupTestDBForTransactions(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompany(t)

	// Create test transactions with different statuses
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	openTransactionID := primitive.NewObjectID()
	paidTransactionID := primitive.NewObjectID()

	_, err := database.GetCollection(collectionName).InsertMany(ctx, []interface{}{
		models.Transaction{
			ID:        openTransactionID,
			CompanyID: companyID,
			Month:     "Janeiro",
			Year:      2024,
			Amount:    1000.00,
			Category:  "Salario",
			Status:    models.StatusOpen,
		},
		models.Transaction{
			ID:        paidTransactionID,
			CompanyID: companyID,
			Month:     "Fevereiro",
			Year:      2024,
			Amount:    1500.00,
			Category:  "Bonus",
			Status:    models.StatusPaid,
		},
	})
	if err != nil {
		t.Fatalf("Failed to create test transactions: %v", err)
	}

	tests := []struct {
		name           string
		transactionID  string
		expectedStatus int
		expectedNewStatus models.Status
	}{
		{
			name:           "Toggle from OPEN to PAID",
			transactionID:  openTransactionID.Hex(),
			expectedStatus: 200,
			expectedNewStatus: models.StatusPaid,
		},
		{
			name:           "Toggle from PAID to OPEN",
			transactionID:  paidTransactionID.Hex(),
			expectedStatus: 200,
			expectedNewStatus: models.StatusOpen,
		},
		{
			name:           "Invalid transaction ID format",
			transactionID:  "invalid-id",
			expectedStatus: 400,
			expectedNewStatus: "",
		},
		{
			name:           "Non-existent transaction",
			transactionID:  primitive.NewObjectID().Hex(),
			expectedStatus: 404,
			expectedNewStatus: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTransactionTestApp(userID)

			req := httptest.NewRequest("PATCH", "/transactions/"+tt.transactionID+"/toggle-status", nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, resp.StatusCode, string(body))
			}

			if tt.expectedStatus == 200 {
				body, _ := io.ReadAll(resp.Body)
				var transaction models.Transaction
				if err := json.Unmarshal(body, &transaction); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if transaction.Status != tt.expectedNewStatus {
					t.Errorf("Expected status %s, got %s", tt.expectedNewStatus, transaction.Status)
				}
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	cleanup := setupTestDBForTransactions(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompany(t)

	// Create test transactions with different statuses
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testTransactions := []interface{}{
		models.Transaction{
			ID:        primitive.NewObjectID(),
			CompanyID: companyID,
			Month:     "Janeiro",
			Year:      2024,
			Amount:    1000.00,
			Category:  "Salario",
			Status:    models.StatusPaid,
		},
		models.Transaction{
			ID:        primitive.NewObjectID(),
			CompanyID: companyID,
			Month:     "Fevereiro",
			Year:      2024,
			Amount:    500.00,
			Category:  "Bonus",
			Status:    models.StatusPaid,
		},
		models.Transaction{
			ID:        primitive.NewObjectID(),
			CompanyID: companyID,
			Month:     "Marco",
			Year:      2024,
			Amount:    2000.00,
			Category:  "Projeto",
			Status:    models.StatusOpen,
		},
	}
	database.GetCollection(collectionName).InsertMany(ctx, testTransactions)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedPaid   float64
		expectedOpen   float64
		expectedTotal  float64
	}{
		{
			name:           "Get stats for company",
			queryParams:    "?companyId=" + companyID.Hex(),
			expectedStatus: 200,
			expectedPaid:   1500.00,
			expectedOpen:   2000.00,
			expectedTotal:  3500.00,
		},
		{
			name:           "Get stats without companyId",
			queryParams:    "",
			expectedStatus: 200,
			expectedPaid:   1500.00,
			expectedOpen:   2000.00,
			expectedTotal:  3500.00,
		},
		{
			name:           "Invalid company ID format",
			queryParams:    "?companyId=invalid-id",
			expectedStatus: 400,
			expectedPaid:   0,
			expectedOpen:   0,
			expectedTotal:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTransactionTestApp(userID)

			req := httptest.NewRequest("GET", "/stats"+tt.queryParams, nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, resp.StatusCode, string(body))
			}

			if tt.expectedStatus == 200 {
				body, _ := io.ReadAll(resp.Body)
				var stats models.Stats
				if err := json.Unmarshal(body, &stats); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if stats.Paid != tt.expectedPaid {
					t.Errorf("Expected paid %f, got %f", tt.expectedPaid, stats.Paid)
				}
				if stats.Open != tt.expectedOpen {
					t.Errorf("Expected open %f, got %f", tt.expectedOpen, stats.Open)
				}
				if stats.Total != tt.expectedTotal {
					t.Errorf("Expected total %f, got %f", tt.expectedTotal, stats.Total)
				}
			}
		})
	}
}

func TestCreateTransactionInvalidJSON(t *testing.T) {
	cleanup := setupTestDBForTransactions(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, _ := createTestUserAndCompany(t)
	app := setupTransactionTestApp(userID)

	req := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString("invalid json"))
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

func TestUpdateTransactionInvalidJSON(t *testing.T) {
	cleanup := setupTestDBForTransactions(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompany(t)

	// Create a test transaction
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transactionID := primitive.NewObjectID()
	database.GetCollection(collectionName).InsertOne(ctx, models.Transaction{
		ID:        transactionID,
		CompanyID: companyID,
		Month:     "Janeiro",
		Year:      2024,
		Amount:    1000.00,
		Category:  "Salario",
		Status:    models.StatusOpen,
	})

	app := setupTransactionTestApp(userID)

	req := httptest.NewRequest("PUT", "/transactions/"+transactionID.Hex(), bytes.NewBufferString("{invalid json}"))
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

// Benchmark tests
func BenchmarkGetAllTransactions(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		b.Skipf("Skipping benchmark: MongoDB not available - %v", err)
		return
	}
	defer client.Disconnect(ctx)

	database.DB = client.Database("m2m_benchmark_transactions")
	defer database.DB.Drop(ctx)

	// Create user and company
	userResult, _ := database.GetCollection("users").InsertOne(ctx, bson.M{
		"name":         "Benchmark User",
		"email":        "bench@example.com",
		"passwordHash": "hash",
		"createdAt":    time.Now(),
	})
	userID := userResult.InsertedID.(primitive.ObjectID)

	companyResult, _ := database.GetCollection("companies").InsertOne(ctx, bson.M{
		"userId":    userID,
		"name":      "Benchmark Company",
		"createdAt": time.Now(),
	})
	companyID := companyResult.InsertedID.(primitive.ObjectID)

	// Create 100 test transactions
	transactions := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		transactions[i] = models.Transaction{
			ID:        primitive.NewObjectID(),
			CompanyID: companyID,
			Month:     "Janeiro",
			Year:      2024,
			Amount:    float64(i * 100),
			Category:  "Benchmark",
			Status:    models.StatusPaid,
		}
	}
	database.GetCollection(collectionName).InsertMany(ctx, transactions)

	app := setupTransactionTestApp(userID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/transactions?companyId="+companyID.Hex(), nil)
		app.Test(req, -1)
	}
}
