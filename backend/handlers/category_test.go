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

// setupTestDBForCategories initializes test database for category tests
func setupTestDBForCategories(t *testing.T) func() {
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

	database.DB = client.Database("m2m_test_categories")

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		database.DB.Drop(ctx)
		client.Disconnect(ctx)
	}

	return cleanup
}

// createTestUserAndCompanyForCategories creates a test user and company for category tests
func createTestUserAndCompanyForCategories(t *testing.T) (primitive.ObjectID, primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create test user
	userResult, err := database.GetCollection("users").InsertOne(ctx, bson.M{
		"name":         "Category Test User",
		"email":        "categoryuser@example.com",
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
		"name":      "Category Test Company",
		"createdAt": time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test company: %v", err)
	}
	companyID := companyResult.InsertedID.(primitive.ObjectID)

	return userID, companyID
}

// setupCategoryTestApp creates a Fiber app with category routes and user context
func setupCategoryTestApp(userID primitive.ObjectID) *fiber.App {
	app := fiber.New()

	// Middleware to set user context
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", userID.Hex())
		return c.Next()
	})

	app.Get("/categories", GetCategories)
	app.Post("/categories", CreateCategory)
	app.Put("/categories/:id", UpdateCategory)
	app.Delete("/categories/:id", DeleteCategory)

	return app
}

func TestGetCategories(t *testing.T) {
	cleanup := setupTestDBForCategories(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForCategories(t)

	// Create test categories
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testCategories := []interface{}{
		models.Category{
			ID:        primitive.NewObjectID(),
			CompanyID: companyID,
			Name:      "Salario",
			Type:      models.Income,
			Color:     "#00FF00",
			Budget:    5000,
			CreatedAt: time.Now(),
		},
		models.Category{
			ID:        primitive.NewObjectID(),
			CompanyID: companyID,
			Name:      "Alimentacao",
			Type:      models.Expense,
			Color:     "#FF0000",
			Budget:    1000,
			CreatedAt: time.Now(),
		},
	}
	database.GetCollection(categoryCollection).InsertMany(ctx, testCategories)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Get all categories for company",
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
			app := setupCategoryTestApp(userID)

			req := httptest.NewRequest("GET", "/categories"+tt.queryParams, nil)
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
				var categories []models.Category
				if err := json.Unmarshal(body, &categories); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if len(categories) != tt.expectedCount {
					t.Errorf("Expected %d categories, got %d", tt.expectedCount, len(categories))
				}
			}
		})
	}
}

func TestGetCategoriesEmptyResult(t *testing.T) {
	cleanup := setupTestDBForCategories(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForCategories(t)
	app := setupCategoryTestApp(userID)

	req := httptest.NewRequest("GET", "/categories?companyId="+companyID.Hex(), nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var categories []models.Category
	if err := json.Unmarshal(body, &categories); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(categories) != 0 {
		t.Errorf("Expected empty array, got %d categories", len(categories))
	}
}

func TestCreateCategory(t *testing.T) {
	cleanup := setupTestDBForCategories(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForCategories(t)

	tests := []struct {
		name           string
		queryParams    string
		requestBody    models.CreateCategoryRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name:        "Valid category creation - EXPENSE",
			queryParams: "?companyId=" + companyID.Hex(),
			requestBody: models.CreateCategoryRequest{
				Name:   "Transporte",
				Type:   models.Expense,
				Color:  "#0000FF",
				Budget: 500,
			},
			expectedStatus: 201,
			expectError:    false,
		},
		{
			name:        "Valid category creation - INCOME",
			queryParams: "?companyId=" + companyID.Hex(),
			requestBody: models.CreateCategoryRequest{
				Name:   "Freelance",
				Type:   models.Income,
				Color:  "#00FF00",
				Budget: 3000,
			},
			expectedStatus: 201,
			expectError:    false,
		},
		{
			name:        "Missing category name",
			queryParams: "?companyId=" + companyID.Hex(),
			requestBody: models.CreateCategoryRequest{
				Name:   "",
				Type:   models.Expense,
				Color:  "#FF0000",
				Budget: 100,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:        "Missing company ID",
			queryParams: "",
			requestBody: models.CreateCategoryRequest{
				Name:   "Test",
				Type:   models.Expense,
				Color:  "#FF0000",
				Budget: 100,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:        "Invalid company ID format",
			queryParams: "?companyId=invalid-id",
			requestBody: models.CreateCategoryRequest{
				Name:   "Test",
				Type:   models.Expense,
				Color:  "#FF0000",
				Budget: 100,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:        "XSS attempt in category name",
			queryParams: "?companyId=" + companyID.Hex(),
			requestBody: models.CreateCategoryRequest{
				Name:   "<script>alert('xss')</script>",
				Type:   models.Expense,
				Color:  "#FF0000",
				Budget: 100,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:        "SQL injection attempt",
			queryParams: "?companyId=" + companyID.Hex(),
			requestBody: models.CreateCategoryRequest{
				Name:   "'; DROP TABLE categories; --",
				Type:   models.Expense,
				Color:  "#FF0000",
				Budget: 100,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:        "Name too long",
			queryParams: "?companyId=" + companyID.Hex(),
			requestBody: models.CreateCategoryRequest{
				Name:   string(make([]byte, 60)),
				Type:   models.Expense,
				Color:  "#FF0000",
				Budget: 100,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:        "Invalid color format",
			queryParams: "?companyId=" + companyID.Hex(),
			requestBody: models.CreateCategoryRequest{
				Name:   "Test Category",
				Type:   models.Expense,
				Color:  "invalid-color",
				Budget: 100,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:        "Invalid category type",
			queryParams: "?companyId=" + companyID.Hex(),
			requestBody: models.CreateCategoryRequest{
				Name:   "Test Category",
				Type:   "INVALID_TYPE",
				Color:  "#FF0000",
				Budget: 100,
			},
			expectedStatus: 400,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupCategoryTestApp(userID)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/categories"+tt.queryParams, bytes.NewBuffer(jsonBody))
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
				var category models.Category
				if err := json.Unmarshal(body, &category); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if category.Name != tt.requestBody.Name {
					t.Errorf("Expected name %s, got %s", tt.requestBody.Name, category.Name)
				}
				if category.Type != tt.requestBody.Type {
					t.Errorf("Expected type %s, got %s", tt.requestBody.Type, category.Type)
				}
				if category.Color != tt.requestBody.Color {
					t.Errorf("Expected color %s, got %s", tt.requestBody.Color, category.Color)
				}
			}
		})
	}
}

func TestCreateCategoryDefaultType(t *testing.T) {
	cleanup := setupTestDBForCategories(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForCategories(t)
	app := setupCategoryTestApp(userID)

	// Create category without specifying type
	reqBody := map[string]interface{}{
		"name":   "Test Category",
		"color":  "#FF0000",
		"budget": 100,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/categories?companyId="+companyID.Hex(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	var category models.Category
	if err := json.Unmarshal(body, &category); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	// Should default to EXPENSE
	if category.Type != models.Expense {
		t.Errorf("Expected default type EXPENSE, got %s", category.Type)
	}
}

func TestUpdateCategory(t *testing.T) {
	cleanup := setupTestDBForCategories(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForCategories(t)

	// Create a test category
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	categoryID := primitive.NewObjectID()
	_, err := database.GetCollection(categoryCollection).InsertOne(ctx, models.Category{
		ID:        categoryID,
		CompanyID: companyID,
		Name:      "Original Category",
		Type:      models.Expense,
		Color:     "#FF0000",
		Budget:    1000,
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	tests := []struct {
		name           string
		categoryID     string
		requestBody    models.UpdateCategoryRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name:       "Valid update",
			categoryID: categoryID.Hex(),
			requestBody: models.UpdateCategoryRequest{
				Name:   "Updated Category",
				Type:   models.Income,
				Color:  "#00FF00",
				Budget: 2000,
			},
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name:       "Invalid category ID format",
			categoryID: "invalid-id",
			requestBody: models.UpdateCategoryRequest{
				Name:   "Test",
				Type:   models.Expense,
				Color:  "#FF0000",
				Budget: 100,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:       "Non-existent category",
			categoryID: primitive.NewObjectID().Hex(),
			requestBody: models.UpdateCategoryRequest{
				Name:   "Test",
				Type:   models.Expense,
				Color:  "#FF0000",
				Budget: 100,
			},
			expectedStatus: 404, // Category doesn't exist
			expectError:    true,
		},
		{
			name:       "Missing name in update",
			categoryID: categoryID.Hex(),
			requestBody: models.UpdateCategoryRequest{
				Name:   "",
				Type:   models.Expense,
				Color:  "#FF0000",
				Budget: 100,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:       "XSS attempt in update",
			categoryID: categoryID.Hex(),
			requestBody: models.UpdateCategoryRequest{
				Name:   "<script>alert('xss')</script>",
				Type:   models.Expense,
				Color:  "#FF0000",
				Budget: 100,
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:       "Invalid color in update",
			categoryID: categoryID.Hex(),
			requestBody: models.UpdateCategoryRequest{
				Name:   "Test Category",
				Type:   models.Expense,
				Color:  "not-a-color",
				Budget: 100,
			},
			expectedStatus: 400,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupCategoryTestApp(userID)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/categories/"+tt.categoryID, bytes.NewBuffer(jsonBody))
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
				var category models.Category
				if err := json.Unmarshal(body, &category); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if category.Name != tt.requestBody.Name {
					t.Errorf("Expected name %s, got %s", tt.requestBody.Name, category.Name)
				}
			}
		})
	}
}

func TestDeleteCategory(t *testing.T) {
	cleanup := setupTestDBForCategories(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForCategories(t)

	// Create a test category
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	categoryID := primitive.NewObjectID()
	_, err := database.GetCollection(categoryCollection).InsertOne(ctx, models.Category{
		ID:        categoryID,
		CompanyID: companyID,
		Name:      "Category to Delete",
		Type:      models.Expense,
		Color:     "#FF0000",
		Budget:    500,
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	tests := []struct {
		name           string
		categoryID     string
		expectedStatus int
	}{
		{
			name:           "Valid delete",
			categoryID:     categoryID.Hex(),
			expectedStatus: 204,
		},
		{
			name:           "Invalid category ID format",
			categoryID:     "invalid-id",
			expectedStatus: 400,
		},
		{
			name:           "Non-existent category",
			categoryID:     primitive.NewObjectID().Hex(),
			expectedStatus: 404, // Category doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupCategoryTestApp(userID)

			req := httptest.NewRequest("DELETE", "/categories/"+tt.categoryID, nil)
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

func TestCreateCategoryInvalidJSON(t *testing.T) {
	cleanup := setupTestDBForCategories(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForCategories(t)
	app := setupCategoryTestApp(userID)

	req := httptest.NewRequest("POST", "/categories?companyId="+companyID.Hex(), bytes.NewBufferString("invalid json"))
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

func TestUpdateCategoryInvalidJSON(t *testing.T) {
	cleanup := setupTestDBForCategories(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForCategories(t)

	// Create a test category
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	categoryID := primitive.NewObjectID()
	database.GetCollection(categoryCollection).InsertOne(ctx, models.Category{
		ID:        categoryID,
		CompanyID: companyID,
		Name:      "Test Category",
		Type:      models.Expense,
		Color:     "#FF0000",
		Budget:    100,
		CreatedAt: time.Now(),
	})

	app := setupCategoryTestApp(userID)

	req := httptest.NewRequest("PUT", "/categories/"+categoryID.Hex(), bytes.NewBufferString("{invalid json}"))
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
func BenchmarkGetCategories(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		b.Skipf("Skipping benchmark: MongoDB not available - %v", err)
		return
	}
	defer client.Disconnect(ctx)

	database.DB = client.Database("m2m_benchmark_categories")
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

	// Create 20 test categories
	categories := make([]interface{}, 20)
	for i := 0; i < 20; i++ {
		categories[i] = models.Category{
			ID:        primitive.NewObjectID(),
			CompanyID: companyID,
			Name:      "Category " + string(rune('A'+i)),
			Type:      models.Expense,
			Color:     "#FF0000",
			Budget:    float64(i * 100),
			CreatedAt: time.Now(),
		}
	}
	database.GetCollection(categoryCollection).InsertMany(ctx, categories)

	app := setupCategoryTestApp(userID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/categories?companyId="+companyID.Hex(), nil)
		app.Test(req, -1)
	}
}

func BenchmarkCreateCategory(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		b.Skipf("Skipping benchmark: MongoDB not available - %v", err)
		return
	}
	defer client.Disconnect(ctx)

	database.DB = client.Database("m2m_benchmark_create_category")
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

	app := setupCategoryTestApp(userID)

	reqBody := models.CreateCategoryRequest{
		Name:   "Benchmark Category",
		Type:   models.Expense,
		Color:  "#FF0000",
		Budget: 1000,
	}
	jsonBody, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/categories?companyId="+companyID.Hex(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		app.Test(req, -1)
	}
}
