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

// setupTestDBForCompanies initializes test database for company tests
func setupTestDBForCompanies(t *testing.T) func() {
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

	database.DB = client.Database("m2m_test_companies")

	// Reset the company service singleton to use the new database
	ResetCompanyService()

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		database.DB.Drop(ctx)
		client.Disconnect(ctx)
		// Reset again after cleanup
		ResetCompanyService()
	}

	return cleanup
}

// createTestUserForCompanies creates a test user for company tests
func createTestUserForCompanies(t *testing.T) primitive.ObjectID {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userResult, err := database.GetCollection("users").InsertOne(ctx, bson.M{
		"name":         "Company Test User",
		"email":        "companyuser@example.com",
		"passwordHash": "hashedpassword",
		"createdAt":    time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return userResult.InsertedID.(primitive.ObjectID)
}

// setupCompanyTestApp creates a Fiber app with company routes and user context
func setupCompanyTestApp(userID primitive.ObjectID) *fiber.App {
	app := fiber.New()

	// Middleware to set user context
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", userID.Hex())
		return c.Next()
	})

	app.Get("/companies", GetCompanies)
	app.Post("/companies", CreateCompany)
	app.Put("/companies/:id", UpdateCompany)
	app.Delete("/companies/:id", DeleteCompany)

	return app
}

func TestGetCompanies(t *testing.T) {
	cleanup := setupTestDBForCompanies(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID := createTestUserForCompanies(t)

	// Create test companies
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testCompanies := []interface{}{
		models.Company{
			ID:        primitive.NewObjectID(),
			UserID:    userID,
			Name:      "Company A",
			CreatedAt: time.Now(),
		},
		models.Company{
			ID:        primitive.NewObjectID(),
			UserID:    userID,
			Name:      "Company B",
			CreatedAt: time.Now(),
		},
	}
	database.GetCollection("companies").InsertMany(ctx, testCompanies)

	app := setupCompanyTestApp(userID)

	req := httptest.NewRequest("GET", "/companies", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	var companies []models.Company
	if err := json.Unmarshal(body, &companies); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(companies) != 2 {
		t.Errorf("Expected 2 companies, got %d", len(companies))
	}
}

func TestGetCompaniesEmpty(t *testing.T) {
	cleanup := setupTestDBForCompanies(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID := createTestUserForCompanies(t)
	app := setupCompanyTestApp(userID)

	req := httptest.NewRequest("GET", "/companies", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var companies []models.Company
	if err := json.Unmarshal(body, &companies); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(companies) != 0 {
		t.Errorf("Expected empty array, got %d companies", len(companies))
	}
}

func TestGetCompaniesOnlyUserCompanies(t *testing.T) {
	cleanup := setupTestDBForCompanies(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID := createTestUserForCompanies(t)
	otherUserID := primitive.NewObjectID()

	// Create companies for both users
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testCompanies := []interface{}{
		models.Company{
			ID:        primitive.NewObjectID(),
			UserID:    userID,
			Name:      "My Company",
			CreatedAt: time.Now(),
		},
		models.Company{
			ID:        primitive.NewObjectID(),
			UserID:    otherUserID,
			Name:      "Other User Company",
			CreatedAt: time.Now(),
		},
	}
	database.GetCollection("companies").InsertMany(ctx, testCompanies)

	app := setupCompanyTestApp(userID)

	req := httptest.NewRequest("GET", "/companies", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var companies []models.Company
	json.Unmarshal(body, &companies)

	// Should only get the user's own company
	if len(companies) != 1 {
		t.Errorf("Expected 1 company, got %d", len(companies))
	}

	if len(companies) > 0 && companies[0].Name != "My Company" {
		t.Errorf("Expected 'My Company', got %s", companies[0].Name)
	}
}

func TestCreateCompany(t *testing.T) {
	cleanup := setupTestDBForCompanies(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID := createTestUserForCompanies(t)

	tests := []struct {
		name           string
		requestBody    models.CreateCompanyRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid company creation",
			requestBody: models.CreateCompanyRequest{
				Name: "New Company",
			},
			expectedStatus: 201,
			expectError:    false,
		},
		{
			name: "Missing company name",
			requestBody: models.CreateCompanyRequest{
				Name: "",
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "XSS attempt in company name",
			requestBody: models.CreateCompanyRequest{
				Name: "<script>alert('xss')</script>",
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "SQL injection attempt",
			requestBody: models.CreateCompanyRequest{
				Name: "'; DROP TABLE companies; --",
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name: "Company name too long",
			requestBody: models.CreateCompanyRequest{
				Name: string(make([]byte, 110)),
			},
			expectedStatus: 400,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupCompanyTestApp(userID)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/companies", bytes.NewBuffer(jsonBody))
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
				var company models.Company
				if err := json.Unmarshal(body, &company); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if company.Name != tt.requestBody.Name {
					t.Errorf("Expected name %s, got %s", tt.requestBody.Name, company.Name)
				}
				if company.UserID != userID {
					t.Errorf("Expected userID %s, got %s", userID.Hex(), company.UserID.Hex())
				}
			}
		})
	}
}

func TestCreateCompanyInvalidJSON(t *testing.T) {
	cleanup := setupTestDBForCompanies(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID := createTestUserForCompanies(t)
	app := setupCompanyTestApp(userID)

	req := httptest.NewRequest("POST", "/companies", bytes.NewBufferString("invalid json"))
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

func TestUpdateCompany(t *testing.T) {
	cleanup := setupTestDBForCompanies(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID := createTestUserForCompanies(t)

	// Create a test company
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	companyID := primitive.NewObjectID()
	_, err := database.GetCollection("companies").InsertOne(ctx, models.Company{
		ID:        companyID,
		UserID:    userID,
		Name:      "Original Company",
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test company: %v", err)
	}

	tests := []struct {
		name           string
		companyID      string
		requestBody    models.CreateCompanyRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name:      "Valid update",
			companyID: companyID.Hex(),
			requestBody: models.CreateCompanyRequest{
				Name: "Updated Company",
			},
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name:      "Invalid company ID format",
			companyID: "invalid-id",
			requestBody: models.CreateCompanyRequest{
				Name: "Test",
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:      "Non-existent company",
			companyID: primitive.NewObjectID().Hex(),
			requestBody: models.CreateCompanyRequest{
				Name: "Test",
			},
			expectedStatus: 404,
			expectError:    true,
		},
		{
			name:      "Missing name in update",
			companyID: companyID.Hex(),
			requestBody: models.CreateCompanyRequest{
				Name: "",
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:      "XSS attempt in update",
			companyID: companyID.Hex(),
			requestBody: models.CreateCompanyRequest{
				Name: "<script>alert('xss')</script>",
			},
			expectedStatus: 400,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupCompanyTestApp(userID)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/companies/"+tt.companyID, bytes.NewBuffer(jsonBody))
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
				var company models.Company
				if err := json.Unmarshal(body, &company); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if company.Name != tt.requestBody.Name {
					t.Errorf("Expected name %s, got %s", tt.requestBody.Name, company.Name)
				}
			}
		})
	}
}

func TestUpdateCompanyUnauthorized(t *testing.T) {
	cleanup := setupTestDBForCompanies(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID := createTestUserForCompanies(t)
	otherUserID := primitive.NewObjectID()

	// Create a company owned by another user
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	companyID := primitive.NewObjectID()
	database.GetCollection("companies").InsertOne(ctx, models.Company{
		ID:        companyID,
		UserID:    otherUserID,
		Name:      "Other User Company",
		CreatedAt: time.Now(),
	})

	app := setupCompanyTestApp(userID)

	reqBody := models.CreateCompanyRequest{Name: "Hacked"}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/companies/"+companyID.Hex(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 403 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 403 (Forbidden), got %d. Body: %s", resp.StatusCode, string(body))
	}
}

func TestUpdateCompanyInvalidJSON(t *testing.T) {
	cleanup := setupTestDBForCompanies(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID := createTestUserForCompanies(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	companyID := primitive.NewObjectID()
	database.GetCollection("companies").InsertOne(ctx, models.Company{
		ID:        companyID,
		UserID:    userID,
		Name:      "Test Company",
		CreatedAt: time.Now(),
	})

	app := setupCompanyTestApp(userID)

	req := httptest.NewRequest("PUT", "/companies/"+companyID.Hex(), bytes.NewBufferString("{invalid json}"))
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

func TestDeleteCompany(t *testing.T) {
	cleanup := setupTestDBForCompanies(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID := createTestUserForCompanies(t)

	// Create a test company
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	companyID := primitive.NewObjectID()
	_, err := database.GetCollection("companies").InsertOne(ctx, models.Company{
		ID:        companyID,
		UserID:    userID,
		Name:      "Company to Delete",
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test company: %v", err)
	}

	tests := []struct {
		name           string
		companyID      string
		expectedStatus int
	}{
		{
			name:           "Valid delete",
			companyID:      companyID.Hex(),
			expectedStatus: 200,
		},
		{
			name:           "Invalid company ID format",
			companyID:      "invalid-id",
			expectedStatus: 400,
		},
		{
			name:           "Non-existent company",
			companyID:      primitive.NewObjectID().Hex(),
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupCompanyTestApp(userID)

			req := httptest.NewRequest("DELETE", "/companies/"+tt.companyID, nil)
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

func TestDeleteCompanyUnauthorized(t *testing.T) {
	cleanup := setupTestDBForCompanies(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID := createTestUserForCompanies(t)
	otherUserID := primitive.NewObjectID()

	// Create a company owned by another user
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	companyID := primitive.NewObjectID()
	database.GetCollection("companies").InsertOne(ctx, models.Company{
		ID:        companyID,
		UserID:    otherUserID,
		Name:      "Other User Company",
		CreatedAt: time.Now(),
	})

	app := setupCompanyTestApp(userID)

	req := httptest.NewRequest("DELETE", "/companies/"+companyID.Hex(), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 403 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 403 (Forbidden), got %d. Body: %s", resp.StatusCode, string(body))
	}
}

func TestDeleteCompanyCascade(t *testing.T) {
	cleanup := setupTestDBForCompanies(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID := createTestUserForCompanies(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create company
	companyID := primitive.NewObjectID()
	database.GetCollection("companies").InsertOne(ctx, models.Company{
		ID:        companyID,
		UserID:    userID,
		Name:      "Company with Data",
		CreatedAt: time.Now(),
	})

	// Create related transactions
	database.GetCollection("transactions").InsertMany(ctx, []interface{}{
		bson.M{"companyId": companyID, "amount": 100},
		bson.M{"companyId": companyID, "amount": 200},
	})

	// Create related categories
	database.GetCollection("categories").InsertMany(ctx, []interface{}{
		bson.M{"companyId": companyID, "name": "Category 1"},
		bson.M{"companyId": companyID, "name": "Category 2"},
	})

	app := setupCompanyTestApp(userID)

	req := httptest.NewRequest("DELETE", "/companies/"+companyID.Hex(), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	// Verify cascade delete - check that related data was deleted
	transCount, _ := database.GetCollection("transactions").CountDocuments(ctx, bson.M{"companyId": companyID})
	if transCount != 0 {
		t.Errorf("Expected 0 transactions after cascade delete, got %d", transCount)
	}

	catCount, _ := database.GetCollection("categories").CountDocuments(ctx, bson.M{"companyId": companyID})
	if catCount != 0 {
		t.Errorf("Expected 0 categories after cascade delete, got %d", catCount)
	}
}

// Benchmark tests
func BenchmarkGetCompanies(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		b.Skipf("Skipping benchmark: MongoDB not available - %v", err)
		return
	}
	defer client.Disconnect(ctx)

	database.DB = client.Database("m2m_benchmark_companies")
	defer database.DB.Drop(ctx)

	// Create user
	userResult, _ := database.GetCollection("users").InsertOne(ctx, bson.M{
		"name":         "Benchmark User",
		"email":        "bench@example.com",
		"passwordHash": "hash",
		"createdAt":    time.Now(),
	})
	userID := userResult.InsertedID.(primitive.ObjectID)

	// Create 10 test companies
	companies := make([]interface{}, 10)
	for i := 0; i < 10; i++ {
		companies[i] = models.Company{
			ID:        primitive.NewObjectID(),
			UserID:    userID,
			Name:      "Company " + string(rune('A'+i)),
			CreatedAt: time.Now(),
		}
	}
	database.GetCollection("companies").InsertMany(ctx, companies)

	app := setupCompanyTestApp(userID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/companies", nil)
		app.Test(req, -1)
	}
}

func BenchmarkCreateCompany(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		b.Skipf("Skipping benchmark: MongoDB not available - %v", err)
		return
	}
	defer client.Disconnect(ctx)

	database.DB = client.Database("m2m_benchmark_create_company")
	defer database.DB.Drop(ctx)

	// Create user
	userResult, _ := database.GetCollection("users").InsertOne(ctx, bson.M{
		"name":         "Benchmark User",
		"email":        "bench@example.com",
		"passwordHash": "hash",
		"createdAt":    time.Now(),
	})
	userID := userResult.InsertedID.(primitive.ObjectID)

	app := setupCompanyTestApp(userID)

	reqBody := models.CreateCompanyRequest{Name: "Benchmark Company"}
	jsonBody, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/companies", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		app.Test(req, -1)
	}
}
