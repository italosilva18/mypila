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

// setupTestDBForRecurring initializes test database for recurring tests
func setupTestDBForRecurring(t *testing.T) func() {
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

	database.DB = client.Database("m2m_test_recurring")

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		database.DB.Drop(ctx)
		client.Disconnect(ctx)
	}

	return cleanup
}

// createTestUserAndCompanyForRecurring creates a test user and company for recurring tests
func createTestUserAndCompanyForRecurring(t *testing.T) (primitive.ObjectID, primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create test user
	userResult, err := database.GetCollection("users").InsertOne(ctx, bson.M{
		"name":         "Recurring Test User",
		"email":        "recurringuser@example.com",
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
		"name":      "Recurring Test Company",
		"createdAt": time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test company: %v", err)
	}
	companyID := companyResult.InsertedID.(primitive.ObjectID)

	return userID, companyID
}

// setupRecurringTestApp creates a Fiber app with recurring routes and user context
func setupRecurringTestApp(userID primitive.ObjectID) *fiber.App {
	app := fiber.New()

	// Middleware to set user context
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", userID.Hex())
		return c.Next()
	})

	app.Get("/recurring", GetRecurring)
	app.Post("/recurring", CreateRecurring)
	app.Delete("/recurring/:id", DeleteRecurring)
	app.Post("/recurring/process", ProcessRecurring)

	return app
}

func TestGetRecurring(t *testing.T) {
	cleanup := setupTestDBForRecurring(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForRecurring(t)

	// Create test recurring rules
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testRules := []interface{}{
		models.RecurringTransaction{
			ID:          primitive.NewObjectID(),
			CompanyID:   companyID,
			Description: "Aluguel Mensal",
			Amount:      2500.00,
			Category:    "Moradia",
			DayOfMonth:  5,
			CreatedAt:   time.Now(),
		},
		models.RecurringTransaction{
			ID:          primitive.NewObjectID(),
			CompanyID:   companyID,
			Description: "Internet",
			Amount:      150.00,
			Category:    "Servicos",
			DayOfMonth:  10,
			CreatedAt:   time.Now(),
		},
	}
	database.GetCollection("recurring").InsertMany(ctx, testRules)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Get all recurring rules for company",
			queryParams:    "?companyId=" + companyID.Hex(),
			expectedStatus: 200,
			expectedCount:  2,
		},
		{
			name:           "Missing company ID",
			queryParams:    "",
			expectedStatus: 400,
			expectedCount:  0,
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
			app := setupRecurringTestApp(userID)

			req := httptest.NewRequest("GET", "/recurring"+tt.queryParams, nil)
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
				var rules []models.RecurringTransaction
				if err := json.Unmarshal(body, &rules); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if len(rules) != tt.expectedCount {
					t.Errorf("Expected %d rules, got %d", tt.expectedCount, len(rules))
				}
			}
		})
	}
}

func TestGetRecurringEmptyResult(t *testing.T) {
	cleanup := setupTestDBForRecurring(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForRecurring(t)
	app := setupRecurringTestApp(userID)

	req := httptest.NewRequest("GET", "/recurring?companyId="+companyID.Hex(), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var rules []models.RecurringTransaction
	if err := json.Unmarshal(body, &rules); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(rules) != 0 {
		t.Errorf("Expected empty array, got %d rules", len(rules))
	}
}

func TestCreateRecurring(t *testing.T) {
	cleanup := setupTestDBForRecurring(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForRecurring(t)

	tests := []struct {
		name           string
		requestBody    models.CreateRecurringRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid recurring rule creation",
			requestBody: models.CreateRecurringRequest{
				CompanyID:   companyID.Hex(),
				Description: "Salario Mensal",
				Amount:      5000.00,
				Category:    "Salario",
				DayOfMonth:  25,
			},
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name: "Valid recurring rule - day 1",
			requestBody: models.CreateRecurringRequest{
				CompanyID:   companyID.Hex(),
				Description: "Conta de Luz",
				Amount:      200.00,
				Category:    "Servicos",
				DayOfMonth:  1,
			},
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name: "Valid recurring rule - day 31",
			requestBody: models.CreateRecurringRequest{
				CompanyID:   companyID.Hex(),
				Description: "Conta de Agua",
				Amount:      100.00,
				Category:    "Servicos",
				DayOfMonth:  31,
			},
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name: "Missing description",
			requestBody: models.CreateRecurringRequest{
				CompanyID:   companyID.Hex(),
				Description: "",
				Amount:      1000.00,
				Category:    "Salario",
				DayOfMonth:  15,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "Missing category",
			requestBody: models.CreateRecurringRequest{
				CompanyID:   companyID.Hex(),
				Description: "Test",
				Amount:      1000.00,
				Category:    "",
				DayOfMonth:  15,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "Invalid amount - negative",
			requestBody: models.CreateRecurringRequest{
				CompanyID:   companyID.Hex(),
				Description: "Test",
				Amount:      -100.00,
				Category:    "Test",
				DayOfMonth:  15,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "Invalid amount - zero",
			requestBody: models.CreateRecurringRequest{
				CompanyID:   companyID.Hex(),
				Description: "Test",
				Amount:      0,
				Category:    "Test",
				DayOfMonth:  15,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "Invalid day of month - zero",
			requestBody: models.CreateRecurringRequest{
				CompanyID:   companyID.Hex(),
				Description: "Test",
				Amount:      100.00,
				Category:    "Test",
				DayOfMonth:  0,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "Invalid day of month - 32",
			requestBody: models.CreateRecurringRequest{
				CompanyID:   companyID.Hex(),
				Description: "Test",
				Amount:      100.00,
				Category:    "Test",
				DayOfMonth:  32,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "Invalid day of month - negative",
			requestBody: models.CreateRecurringRequest{
				CompanyID:   companyID.Hex(),
				Description: "Test",
				Amount:      100.00,
				Category:    "Test",
				DayOfMonth:  -1,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "Invalid company ID format",
			requestBody: models.CreateRecurringRequest{
				CompanyID:   "invalid-id",
				Description: "Test",
				Amount:      100.00,
				Category:    "Test",
				DayOfMonth:  15,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "XSS attempt in description",
			requestBody: models.CreateRecurringRequest{
				CompanyID:   companyID.Hex(),
				Description: "<script>alert('xss')</script>",
				Amount:      100.00,
				Category:    "Test",
				DayOfMonth:  15,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "SQL injection attempt in description",
			requestBody: models.CreateRecurringRequest{
				CompanyID:   companyID.Hex(),
				Description: "'; DROP TABLE recurring; --",
				Amount:      100.00,
				Category:    "Test",
				DayOfMonth:  15,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "XSS attempt in category",
			requestBody: models.CreateRecurringRequest{
				CompanyID:   companyID.Hex(),
				Description: "Test Description",
				Amount:      100.00,
				Category:    "<script>alert('xss')</script>",
				DayOfMonth:  15,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "Description too long",
			requestBody: models.CreateRecurringRequest{
				CompanyID:   companyID.Hex(),
				Description: string(make([]byte, 210)),
				Amount:      100.00,
				Category:    "Test",
				DayOfMonth:  15,
			},
			expectedStatus: 400,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupRecurringTestApp(userID)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/recurring", bytes.NewBuffer(jsonBody))
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
				var rule models.RecurringTransaction
				if err := json.Unmarshal(body, &rule); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if rule.Description != tt.requestBody.Description {
					t.Errorf("Expected description %s, got %s", tt.requestBody.Description, rule.Description)
				}
				if rule.Amount != tt.requestBody.Amount {
					t.Errorf("Expected amount %f, got %f", tt.requestBody.Amount, rule.Amount)
				}
				if rule.DayOfMonth != tt.requestBody.DayOfMonth {
					t.Errorf("Expected dayOfMonth %d, got %d", tt.requestBody.DayOfMonth, rule.DayOfMonth)
				}
			}
		})
	}
}

func TestCreateRecurringInvalidJSON(t *testing.T) {
	cleanup := setupTestDBForRecurring(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, _ := createTestUserAndCompanyForRecurring(t)
	app := setupRecurringTestApp(userID)

	req := httptest.NewRequest("POST", "/recurring", bytes.NewBufferString("invalid json"))
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

func TestDeleteRecurring(t *testing.T) {
	cleanup := setupTestDBForRecurring(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForRecurring(t)

	// Create a test recurring rule
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ruleID := primitive.NewObjectID()
	_, err := database.GetCollection("recurring").InsertOne(ctx, models.RecurringTransaction{
		ID:          ruleID,
		CompanyID:   companyID,
		Description: "Rule to Delete",
		Amount:      500.00,
		Category:    "Test",
		DayOfMonth:  15,
		CreatedAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test rule: %v", err)
	}

	tests := []struct {
		name           string
		ruleID         string
		expectedStatus int
	}{
		{
			name:           "Valid delete",
			ruleID:         ruleID.Hex(),
			expectedStatus: 204,
		},
		{
			name:           "Invalid rule ID format",
			ruleID:         "invalid-id",
			expectedStatus: 400,
		},
		{
			name:           "Non-existent rule",
			ruleID:         primitive.NewObjectID().Hex(),
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupRecurringTestApp(userID)

			req := httptest.NewRequest("DELETE", "/recurring/"+tt.ruleID, nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, resp.StatusCode, string(body))
			}
		})
	}
}

func TestProcessRecurring(t *testing.T) {
	cleanup := setupTestDBForRecurring(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForRecurring(t)

	// Create test recurring rules
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	database.GetCollection("recurring").InsertMany(ctx, []interface{}{
		models.RecurringTransaction{
			ID:          primitive.NewObjectID(),
			CompanyID:   companyID,
			Description: "Aluguel",
			Amount:      2000.00,
			Category:    "Moradia",
			DayOfMonth:  5,
			CreatedAt:   time.Now(),
		},
		models.RecurringTransaction{
			ID:          primitive.NewObjectID(),
			CompanyID:   companyID,
			Description: "Internet",
			Amount:      150.00,
			Category:    "Servicos",
			DayOfMonth:  10,
			CreatedAt:   time.Now(),
		},
	})

	tests := []struct {
		name            string
		queryParams     string
		expectedStatus  int
		expectedCreated int
	}{
		{
			name:            "Process recurring for Janeiro 2024",
			queryParams:     "?companyId=" + companyID.Hex() + "&month=Janeiro&year=2024",
			expectedStatus:  200,
			expectedCreated: 2,
		},
		{
			name:            "Process same month again - should not duplicate",
			queryParams:     "?companyId=" + companyID.Hex() + "&month=Janeiro&year=2024",
			expectedStatus:  200,
			expectedCreated: 0,
		},
		{
			name:            "Process different month",
			queryParams:     "?companyId=" + companyID.Hex() + "&month=Fevereiro&year=2024",
			expectedStatus:  200,
			expectedCreated: 2,
		},
		{
			name:            "Missing company ID",
			queryParams:     "?month=Janeiro&year=2024",
			expectedStatus:  400,
			expectedCreated: 0,
		},
		{
			name:            "Missing month",
			queryParams:     "?companyId=" + companyID.Hex() + "&year=2024",
			expectedStatus:  400,
			expectedCreated: 0,
		},
		{
			name:            "Missing year",
			queryParams:     "?companyId=" + companyID.Hex() + "&month=Janeiro",
			expectedStatus:  400,
			expectedCreated: 0,
		},
		{
			name:            "Invalid month",
			queryParams:     "?companyId=" + companyID.Hex() + "&month=InvalidMonth&year=2024",
			expectedStatus:  400,
			expectedCreated: 0,
		},
		{
			name:            "Invalid year - too low",
			queryParams:     "?companyId=" + companyID.Hex() + "&month=Janeiro&year=1999",
			expectedStatus:  400,
			expectedCreated: 0,
		},
		{
			name:            "Invalid company ID format",
			queryParams:     "?companyId=invalid-id&month=Janeiro&year=2024",
			expectedStatus:  400,
			expectedCreated: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupRecurringTestApp(userID)

			req := httptest.NewRequest("POST", "/recurring/process"+tt.queryParams, nil)
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
				var result struct {
					Message string `json:"message"`
					Created int    `json:"created"`
				}
				if err := json.Unmarshal(body, &result); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if result.Created != tt.expectedCreated {
					t.Errorf("Expected %d created, got %d", tt.expectedCreated, result.Created)
				}
			}
		})
	}
}

func TestProcessRecurringCreatesTransactions(t *testing.T) {
	cleanup := setupTestDBForRecurring(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForRecurring(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a recurring rule
	database.GetCollection("recurring").InsertOne(ctx, models.RecurringTransaction{
		ID:          primitive.NewObjectID(),
		CompanyID:   companyID,
		Description: "Test Recurring",
		Amount:      1000.00,
		Category:    "Test",
		DayOfMonth:  15,
		CreatedAt:   time.Now(),
	})

	app := setupRecurringTestApp(userID)

	// Process recurring
	req := httptest.NewRequest("POST", "/recurring/process?companyId="+companyID.Hex()+"&month=Março&year=2024", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify transaction was created
	var transaction models.Transaction
	err = database.GetCollection("transactions").FindOne(ctx, bson.M{
		"companyId":   companyID,
		"description": "Test Recurring",
		"month":       "Março",
		"year":        2024,
	}).Decode(&transaction)

	if err != nil {
		t.Errorf("Expected transaction to be created, but not found: %v", err)
	}

	if transaction.Amount != 1000.00 {
		t.Errorf("Expected amount 1000, got %f", transaction.Amount)
	}

	if transaction.Status != models.StatusOpen {
		t.Errorf("Expected status OPEN, got %s", transaction.Status)
	}
}

func TestProcessRecurringDoesNotDuplicate(t *testing.T) {
	cleanup := setupTestDBForRecurring(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForRecurring(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a recurring rule
	database.GetCollection("recurring").InsertOne(ctx, models.RecurringTransaction{
		ID:          primitive.NewObjectID(),
		CompanyID:   companyID,
		Description: "Monthly Subscription",
		Amount:      99.99,
		Category:    "Subscriptions",
		DayOfMonth:  1,
		CreatedAt:   time.Now(),
	})

	app := setupRecurringTestApp(userID)

	// Process recurring first time
	req := httptest.NewRequest("POST", "/recurring/process?companyId="+companyID.Hex()+"&month=Abril&year=2024", nil)
	resp, _ := app.Test(req, -1)
	resp.Body.Close()

	// Process recurring second time
	req = httptest.NewRequest("POST", "/recurring/process?companyId="+companyID.Hex()+"&month=Abril&year=2024", nil)
	resp, _ = app.Test(req, -1)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var result struct {
		Created int `json:"created"`
	}
	json.Unmarshal(body, &result)

	if result.Created != 0 {
		t.Errorf("Expected 0 created (duplicate prevention), got %d", result.Created)
	}

	// Verify only one transaction exists
	count, _ := database.GetCollection("transactions").CountDocuments(ctx, bson.M{
		"companyId":   companyID,
		"description": "Monthly Subscription",
		"month":       "Abril",
		"year":        2024,
	})

	if count != 1 {
		t.Errorf("Expected exactly 1 transaction, got %d", count)
	}
}

func TestGetRecurringOnlyUserRules(t *testing.T) {
	cleanup := setupTestDBForRecurring(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForRecurring(t)
	otherCompanyID := primitive.NewObjectID()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create rules for both companies
	database.GetCollection("recurring").InsertMany(ctx, []interface{}{
		models.RecurringTransaction{
			ID:          primitive.NewObjectID(),
			CompanyID:   companyID,
			Description: "My Rule",
			Amount:      100.00,
			Category:    "Test",
			DayOfMonth:  1,
			CreatedAt:   time.Now(),
		},
		models.RecurringTransaction{
			ID:          primitive.NewObjectID(),
			CompanyID:   otherCompanyID,
			Description: "Other Company Rule",
			Amount:      200.00,
			Category:    "Test",
			DayOfMonth:  1,
			CreatedAt:   time.Now(),
		},
	})

	app := setupRecurringTestApp(userID)

	req := httptest.NewRequest("GET", "/recurring?companyId="+companyID.Hex(), nil)
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var rules []models.RecurringTransaction
	json.Unmarshal(body, &rules)

	// Should only get the user's own rules
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}

	if len(rules) > 0 && rules[0].Description != "My Rule" {
		t.Errorf("Expected 'My Rule', got %s", rules[0].Description)
	}
}

// Benchmark tests
func BenchmarkGetRecurring(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		b.Skipf("Skipping benchmark: MongoDB not available - %v", err)
		return
	}
	defer client.Disconnect(ctx)

	database.DB = client.Database("m2m_benchmark_recurring")
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

	// Create 20 test recurring rules
	rules := make([]interface{}, 20)
	for i := 0; i < 20; i++ {
		rules[i] = models.RecurringTransaction{
			ID:          primitive.NewObjectID(),
			CompanyID:   companyID,
			Description: "Rule " + string(rune('A'+i)),
			Amount:      float64(i * 100),
			Category:    "Test",
			DayOfMonth:  (i % 28) + 1,
			CreatedAt:   time.Now(),
		}
	}
	database.GetCollection("recurring").InsertMany(ctx, rules)

	app := setupRecurringTestApp(userID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/recurring?companyId="+companyID.Hex(), nil)
		app.Test(req, -1)
	}
}

func BenchmarkProcessRecurring(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		b.Skipf("Skipping benchmark: MongoDB not available - %v", err)
		return
	}
	defer client.Disconnect(ctx)

	database.DB = client.Database("m2m_benchmark_process_recurring")
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

	// Create 10 recurring rules
	for i := 0; i < 10; i++ {
		database.GetCollection("recurring").InsertOne(ctx, models.RecurringTransaction{
			ID:          primitive.NewObjectID(),
			CompanyID:   companyID,
			Description: "Rule " + string(rune('A'+i)),
			Amount:      float64(i * 100),
			Category:    "Test",
			DayOfMonth:  (i % 28) + 1,
			CreatedAt:   time.Now(),
		})
	}

	app := setupRecurringTestApp(userID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use different months to avoid duplicate prevention
		month := []string{"Janeiro", "Fevereiro", "Marco", "Abril", "Maio", "Junho", "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"}[i%12]
		year := 2020 + (i / 12)
		req := httptest.NewRequest("POST", "/recurring/process?companyId="+companyID.Hex()+"&month="+month+"&year="+string(rune('0'+year/1000))+string(rune('0'+(year/100)%10))+string(rune('0'+(year/10)%10))+string(rune('0'+year%10)), nil)
		app.Test(req, -1)
	}
}
