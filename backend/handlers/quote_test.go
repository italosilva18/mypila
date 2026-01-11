package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// setupTestDBForQuotes initializes test database for quote tests
func setupTestDBForQuotes(t *testing.T) func() {
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

	database.DB = client.Database("m2m_test_quotes")

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		database.DB.Drop(ctx)
		client.Disconnect(ctx)
	}

	return cleanup
}

// createTestUserAndCompanyForQuotes creates a test user and company for quote tests
func createTestUserAndCompanyForQuotes(t *testing.T) (primitive.ObjectID, primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create test user
	userResult, err := database.GetCollection("users").InsertOne(ctx, bson.M{
		"name":         "Quote Test User",
		"email":        "quoteuser@example.com",
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
		"name":      "Quote Test Company",
		"createdAt": time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test company: %v", err)
	}
	companyID := companyResult.InsertedID.(primitive.ObjectID)

	return userID, companyID
}

// setupQuoteTestApp creates a Fiber app with quote routes and user context
func setupQuoteTestApp(userID primitive.ObjectID) *fiber.App {
	app := fiber.New()

	// Middleware to set user context
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", userID.Hex())
		return c.Next()
	})

	app.Get("/quotes", GetQuotes)
	app.Get("/quotes/:id", GetQuote)
	app.Post("/quotes", CreateQuote)
	app.Put("/quotes/:id", UpdateQuote)
	app.Delete("/quotes/:id", DeleteQuote)
	app.Post("/quotes/:id/duplicate", DuplicateQuote)
	app.Patch("/quotes/:id/status", UpdateQuoteStatus)
	app.Get("/quotes/:id/comparison", GetQuoteComparison)

	return app
}

func TestGenerateQuoteNumber(t *testing.T) {
	cleanup := setupTestDBForQuotes(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	_, companyID := createTestUserAndCompanyForQuotes(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	currentYear := time.Now().Year()

	tests := []struct {
		name           string
		existingQuotes []interface{}
		expectedNumber string
	}{
		{
			name:           "First quote of the year",
			existingQuotes: nil,
			expectedNumber: fmt.Sprintf("ORC-%d-001", currentYear),
		},
		{
			name: "Second quote after existing one",
			existingQuotes: []interface{}{
				models.Quote{
					ID:        primitive.NewObjectID(),
					CompanyID: companyID,
					Number:    fmt.Sprintf("ORC-%d-001", currentYear),
					CreatedAt: time.Now(),
				},
			},
			expectedNumber: fmt.Sprintf("ORC-%d-002", currentYear),
		},
		{
			name: "Continue sequence after multiple quotes",
			existingQuotes: []interface{}{
				models.Quote{
					ID:        primitive.NewObjectID(),
					CompanyID: companyID,
					Number:    fmt.Sprintf("ORC-%d-001", currentYear),
					CreatedAt: time.Now().Add(-2 * time.Hour),
				},
				models.Quote{
					ID:        primitive.NewObjectID(),
					CompanyID: companyID,
					Number:    fmt.Sprintf("ORC-%d-005", currentYear),
					CreatedAt: time.Now().Add(-1 * time.Hour),
				},
				models.Quote{
					ID:        primitive.NewObjectID(),
					CompanyID: companyID,
					Number:    fmt.Sprintf("ORC-%d-010", currentYear),
					CreatedAt: time.Now(),
				},
			},
			expectedNumber: fmt.Sprintf("ORC-%d-011", currentYear),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear quotes collection before each test
			database.GetCollection(quoteCollection).Drop(ctx)

			// Insert existing quotes if any
			if tt.existingQuotes != nil {
				database.GetCollection(quoteCollection).InsertMany(ctx, tt.existingQuotes)
			}

			// Generate new quote number
			number, err := generateQuoteNumber(ctx, companyID)
			if err != nil {
				t.Fatalf("Failed to generate quote number: %v", err)
			}

			if number != tt.expectedNumber {
				t.Errorf("Expected quote number %s, got %s", tt.expectedNumber, number)
			}
		})
	}
}

func TestGetQuotes(t *testing.T) {
	cleanup := setupTestDBForQuotes(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForQuotes(t)

	// Create test quotes
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testQuotes := []interface{}{
		models.Quote{
			ID:         primitive.NewObjectID(),
			CompanyID:  companyID,
			Number:     "ORC-2024-001",
			ClientName: "Cliente A",
			Title:      "Orcamento A",
			Status:     models.QuoteDraft,
			Items: []models.QuoteItem{
				{ID: primitive.NewObjectID(), Description: "Item 1", Quantity: 1, UnitPrice: 100, Total: 100},
			},
			Total:     100,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		models.Quote{
			ID:         primitive.NewObjectID(),
			CompanyID:  companyID,
			Number:     "ORC-2024-002",
			ClientName: "Cliente B",
			Title:      "Orcamento B",
			Status:     models.QuoteSent,
			Items: []models.QuoteItem{
				{ID: primitive.NewObjectID(), Description: "Item 2", Quantity: 2, UnitPrice: 200, Total: 400},
			},
			Total:     400,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	database.GetCollection(quoteCollection).InsertMany(ctx, testQuotes)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Get all quotes for company",
			queryParams:    "?companyId=" + companyID.Hex(),
			expectedStatus: 200,
			expectedCount:  2,
		},
		{
			name:           "Get quotes filtered by status DRAFT",
			queryParams:    "?companyId=" + companyID.Hex() + "&status=DRAFT",
			expectedStatus: 200,
			expectedCount:  1,
		},
		{
			name:           "Get quotes filtered by status SENT",
			queryParams:    "?companyId=" + companyID.Hex() + "&status=SENT",
			expectedStatus: 200,
			expectedCount:  1,
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
			app := setupQuoteTestApp(userID)

			req := httptest.NewRequest("GET", "/quotes"+tt.queryParams, nil)
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
				var quotes []models.Quote
				if err := json.Unmarshal(body, &quotes); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if len(quotes) != tt.expectedCount {
					t.Errorf("Expected %d quotes, got %d", tt.expectedCount, len(quotes))
				}
			}
		})
	}
}

func TestCreateQuote(t *testing.T) {
	cleanup := setupTestDBForQuotes(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForQuotes(t)

	tests := []struct {
		name           string
		queryParams    string
		requestBody    models.CreateQuoteRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name:        "Valid quote creation",
			queryParams: "?companyId=" + companyID.Hex(),
			requestBody: models.CreateQuoteRequest{
				ClientName:  "Cliente Teste",
				ClientEmail: "cliente@teste.com",
				ClientPhone: "(11) 99999-9999",
				Title:       "Orcamento de Teste",
				Description: "Descricao do orcamento",
				Items: []models.CreateQuoteItemRequest{
					{Description: "Servico A", Quantity: 2, UnitPrice: 500.00},
					{Description: "Servico B", Quantity: 1, UnitPrice: 1000.00},
				},
				Discount:     10,
				DiscountType: "PERCENT",
				ValidUntil:   time.Now().AddDate(0, 0, 30).Format("2006-01-02"),
				Notes:        "Notas adicionais",
			},
			expectedStatus: 201,
			expectError:    false,
		},
		{
			name:        "Missing client name",
			queryParams: "?companyId=" + companyID.Hex(),
			requestBody: models.CreateQuoteRequest{
				ClientName: "",
				Title:      "Orcamento Teste",
				Items: []models.CreateQuoteItemRequest{
					{Description: "Item", Quantity: 1, UnitPrice: 100},
				},
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:        "Missing title",
			queryParams: "?companyId=" + companyID.Hex(),
			requestBody: models.CreateQuoteRequest{
				ClientName: "Cliente",
				Title:      "",
				Items: []models.CreateQuoteItemRequest{
					{Description: "Item", Quantity: 1, UnitPrice: 100},
				},
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:        "Empty items array",
			queryParams: "?companyId=" + companyID.Hex(),
			requestBody: models.CreateQuoteRequest{
				ClientName: "Cliente",
				Title:      "Orcamento",
				Items:      []models.CreateQuoteItemRequest{},
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:        "Missing company ID",
			queryParams: "",
			requestBody: models.CreateQuoteRequest{
				ClientName: "Cliente",
				Title:      "Orcamento",
				Items: []models.CreateQuoteItemRequest{
					{Description: "Item", Quantity: 1, UnitPrice: 100},
				},
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:        "Invalid company ID format",
			queryParams: "?companyId=invalid-id",
			requestBody: models.CreateQuoteRequest{
				ClientName: "Cliente",
				Title:      "Orcamento",
				Items: []models.CreateQuoteItemRequest{
					{Description: "Item", Quantity: 1, UnitPrice: 100},
				},
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:        "XSS attempt in client name",
			queryParams: "?companyId=" + companyID.Hex(),
			requestBody: models.CreateQuoteRequest{
				ClientName: "<script>alert('xss')</script>",
				Title:      "Orcamento",
				Items: []models.CreateQuoteItemRequest{
					{Description: "Item", Quantity: 1, UnitPrice: 100},
				},
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:        "Client name too long",
			queryParams: "?companyId=" + companyID.Hex(),
			requestBody: models.CreateQuoteRequest{
				ClientName: string(make([]byte, 150)),
				Title:      "Orcamento",
				Items: []models.CreateQuoteItemRequest{
					{Description: "Item", Quantity: 1, UnitPrice: 100},
				},
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:        "Quote with value discount",
			queryParams: "?companyId=" + companyID.Hex(),
			requestBody: models.CreateQuoteRequest{
				ClientName: "Cliente",
				Title:      "Orcamento com Desconto Valor",
				Items: []models.CreateQuoteItemRequest{
					{Description: "Servico", Quantity: 1, UnitPrice: 1000.00},
				},
				Discount:     100,
				DiscountType: "VALUE",
			},
			expectedStatus: 201,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupQuoteTestApp(userID)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/quotes"+tt.queryParams, bytes.NewBuffer(jsonBody))
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
				var quote models.Quote
				if err := json.Unmarshal(body, &quote); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if quote.ClientName == "" {
					t.Errorf("Expected client name in response")
				}
				if quote.Number == "" {
					t.Errorf("Expected auto-generated quote number")
				}
				if quote.Status != models.QuoteDraft {
					t.Errorf("Expected status DRAFT, got %s", quote.Status)
				}
			}
		})
	}
}

func TestUpdateQuote(t *testing.T) {
	cleanup := setupTestDBForQuotes(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForQuotes(t)

	// Create test quotes
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	draftQuoteID := primitive.NewObjectID()
	executedQuoteID := primitive.NewObjectID()

	_, err := database.GetCollection(quoteCollection).InsertMany(ctx, []interface{}{
		models.Quote{
			ID:         draftQuoteID,
			CompanyID:  companyID,
			Number:     "ORC-2024-001",
			ClientName: "Cliente Original",
			Title:      "Titulo Original",
			Status:     models.QuoteDraft,
			Items: []models.QuoteItem{
				{ID: primitive.NewObjectID(), Description: "Item Original", Quantity: 1, UnitPrice: 100, Total: 100},
			},
			Total:      100,
			ValidUntil: time.Now().AddDate(0, 0, 30),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		models.Quote{
			ID:         executedQuoteID,
			CompanyID:  companyID,
			Number:     "ORC-2024-002",
			ClientName: "Cliente Executado",
			Title:      "Orcamento Executado",
			Status:     models.QuoteExecuted,
			Items: []models.QuoteItem{
				{ID: primitive.NewObjectID(), Description: "Item", Quantity: 1, UnitPrice: 500, Total: 500},
			},
			Total:      500,
			ValidUntil: time.Now().AddDate(0, 0, 30),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("Failed to create test quotes: %v", err)
	}

	tests := []struct {
		name           string
		quoteID        string
		requestBody    models.UpdateQuoteRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name:    "Valid update",
			quoteID: draftQuoteID.Hex(),
			requestBody: models.UpdateQuoteRequest{
				ClientName: "Cliente Atualizado",
				Title:      "Titulo Atualizado",
				Items: []models.CreateQuoteItemRequest{
					{Description: "Item Atualizado", Quantity: 2, UnitPrice: 200},
				},
				Discount:     5,
				DiscountType: "PERCENT",
			},
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name:    "Cannot update executed quote",
			quoteID: executedQuoteID.Hex(),
			requestBody: models.UpdateQuoteRequest{
				ClientName: "Tentativa",
				Title:      "Atualizar Executado",
				Items: []models.CreateQuoteItemRequest{
					{Description: "Item", Quantity: 1, UnitPrice: 100},
				},
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:    "Invalid quote ID format",
			quoteID: "invalid-id",
			requestBody: models.UpdateQuoteRequest{
				ClientName: "Cliente",
				Title:      "Titulo",
				Items: []models.CreateQuoteItemRequest{
					{Description: "Item", Quantity: 1, UnitPrice: 100},
				},
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:    "Non-existent quote",
			quoteID: primitive.NewObjectID().Hex(),
			requestBody: models.UpdateQuoteRequest{
				ClientName: "Cliente",
				Title:      "Titulo",
				Items: []models.CreateQuoteItemRequest{
					{Description: "Item", Quantity: 1, UnitPrice: 100},
				},
			},
			expectedStatus: 404,
			expectError:    true,
		},
		{
			name:    "Missing client name in update",
			quoteID: draftQuoteID.Hex(),
			requestBody: models.UpdateQuoteRequest{
				ClientName: "",
				Title:      "Titulo",
				Items: []models.CreateQuoteItemRequest{
					{Description: "Item", Quantity: 1, UnitPrice: 100},
				},
			},
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:    "Empty items in update",
			quoteID: draftQuoteID.Hex(),
			requestBody: models.UpdateQuoteRequest{
				ClientName: "Cliente",
				Title:      "Titulo",
				Items:      []models.CreateQuoteItemRequest{},
			},
			expectedStatus: 400,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupQuoteTestApp(userID)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/quotes/"+tt.quoteID, bytes.NewBuffer(jsonBody))
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
				var quote models.Quote
				if err := json.Unmarshal(body, &quote); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if quote.ClientName != tt.requestBody.ClientName {
					t.Errorf("Expected client name %s, got %s", tt.requestBody.ClientName, quote.ClientName)
				}
			}
		})
	}
}

func TestDeleteQuote(t *testing.T) {
	cleanup := setupTestDBForQuotes(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForQuotes(t)

	// Create a test quote
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	quoteID := primitive.NewObjectID()
	_, err := database.GetCollection(quoteCollection).InsertOne(ctx, models.Quote{
		ID:         quoteID,
		CompanyID:  companyID,
		Number:     "ORC-2024-001",
		ClientName: "Cliente para Deletar",
		Title:      "Orcamento para Deletar",
		Status:     models.QuoteDraft,
		Items: []models.QuoteItem{
			{ID: primitive.NewObjectID(), Description: "Item", Quantity: 1, UnitPrice: 100, Total: 100},
		},
		Total:      100,
		ValidUntil: time.Now().AddDate(0, 0, 30),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test quote: %v", err)
	}

	tests := []struct {
		name           string
		quoteID        string
		expectedStatus int
	}{
		{
			name:           "Valid delete",
			quoteID:        quoteID.Hex(),
			expectedStatus: 204,
		},
		{
			name:           "Invalid quote ID format",
			quoteID:        "invalid-id",
			expectedStatus: 400,
		},
		{
			name:           "Non-existent quote",
			quoteID:        primitive.NewObjectID().Hex(),
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupQuoteTestApp(userID)

			req := httptest.NewRequest("DELETE", "/quotes/"+tt.quoteID, nil)
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

func TestDuplicateQuote(t *testing.T) {
	cleanup := setupTestDBForQuotes(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForQuotes(t)

	// Create a test quote
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	originalQuoteID := primitive.NewObjectID()
	_, err := database.GetCollection(quoteCollection).InsertOne(ctx, models.Quote{
		ID:          originalQuoteID,
		CompanyID:   companyID,
		Number:      "ORC-2024-001",
		ClientName:  "Cliente Original",
		ClientEmail: "cliente@original.com",
		Title:       "Orcamento Original",
		Description: "Descricao original",
		Status:      models.QuoteApproved,
		Items: []models.QuoteItem{
			{ID: primitive.NewObjectID(), Description: "Item A", Quantity: 2, UnitPrice: 250, Total: 500},
			{ID: primitive.NewObjectID(), Description: "Item B", Quantity: 1, UnitPrice: 1000, Total: 1000},
		},
		Subtotal:     1500,
		Discount:     100,
		DiscountType: models.DiscountValue,
		Total:        1400,
		ValidUntil:   time.Now().AddDate(0, 0, 30),
		Notes:        "Notas originais",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test quote: %v", err)
	}

	tests := []struct {
		name           string
		quoteID        string
		expectedStatus int
	}{
		{
			name:           "Valid duplicate",
			quoteID:        originalQuoteID.Hex(),
			expectedStatus: 201,
		},
		{
			name:           "Invalid quote ID format",
			quoteID:        "invalid-id",
			expectedStatus: 400,
		},
		{
			name:           "Non-existent quote",
			quoteID:        primitive.NewObjectID().Hex(),
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupQuoteTestApp(userID)

			req := httptest.NewRequest("POST", "/quotes/"+tt.quoteID+"/duplicate", nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, resp.StatusCode, string(body))
			}

			if tt.expectedStatus == 201 {
				body, _ := io.ReadAll(resp.Body)
				var duplicatedQuote models.Quote
				if err := json.Unmarshal(body, &duplicatedQuote); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				// Verify it's a new quote
				if duplicatedQuote.ID == originalQuoteID {
					t.Errorf("Duplicated quote should have a new ID")
				}

				// Verify the title has "(Copia)" suffix
				expectedTitle := "Orcamento Original (Copia)"
				if duplicatedQuote.Title != expectedTitle {
					t.Errorf("Expected title %s, got %s", expectedTitle, duplicatedQuote.Title)
				}

				// Verify status is DRAFT
				if duplicatedQuote.Status != models.QuoteDraft {
					t.Errorf("Expected status DRAFT, got %s", duplicatedQuote.Status)
				}

				// Verify it has a new quote number
				if duplicatedQuote.Number == "ORC-2024-001" {
					t.Errorf("Duplicated quote should have a new number")
				}

				// Verify items count
				if len(duplicatedQuote.Items) != 2 {
					t.Errorf("Expected 2 items, got %d", len(duplicatedQuote.Items))
				}
			}
		})
	}
}

func TestUpdateQuoteStatus(t *testing.T) {
	cleanup := setupTestDBForQuotes(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForQuotes(t)

	// Create a test quote
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	quoteID := primitive.NewObjectID()
	_, err := database.GetCollection(quoteCollection).InsertOne(ctx, models.Quote{
		ID:         quoteID,
		CompanyID:  companyID,
		Number:     "ORC-2024-001",
		ClientName: "Cliente",
		Title:      "Orcamento Status",
		Status:     models.QuoteDraft,
		Items: []models.QuoteItem{
			{ID: primitive.NewObjectID(), Description: "Item", Quantity: 1, UnitPrice: 100, Total: 100},
		},
		Total:      100,
		ValidUntil: time.Now().AddDate(0, 0, 30),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test quote: %v", err)
	}

	tests := []struct {
		name           string
		quoteID        string
		requestBody    models.UpdateQuoteStatusRequest
		expectedStatus int
		expectedNewStatus models.QuoteStatus
	}{
		{
			name:    "Update to SENT",
			quoteID: quoteID.Hex(),
			requestBody: models.UpdateQuoteStatusRequest{
				Status: "SENT",
			},
			expectedStatus:    200,
			expectedNewStatus: models.QuoteSent,
		},
		{
			name:    "Update to APPROVED",
			quoteID: quoteID.Hex(),
			requestBody: models.UpdateQuoteStatusRequest{
				Status: "APPROVED",
			},
			expectedStatus:    200,
			expectedNewStatus: models.QuoteApproved,
		},
		{
			name:    "Invalid status value",
			quoteID: quoteID.Hex(),
			requestBody: models.UpdateQuoteStatusRequest{
				Status: "INVALID_STATUS",
			},
			expectedStatus:    400,
			expectedNewStatus: "",
		},
		{
			name:    "Invalid quote ID",
			quoteID: "invalid-id",
			requestBody: models.UpdateQuoteStatusRequest{
				Status: "SENT",
			},
			expectedStatus:    400,
			expectedNewStatus: "",
		},
		{
			name:    "Non-existent quote",
			quoteID: primitive.NewObjectID().Hex(),
			requestBody: models.UpdateQuoteStatusRequest{
				Status: "SENT",
			},
			expectedStatus:    404,
			expectedNewStatus: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupQuoteTestApp(userID)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PATCH", "/quotes/"+tt.quoteID+"/status", bytes.NewBuffer(jsonBody))
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

			if tt.expectedStatus == 200 {
				body, _ := io.ReadAll(resp.Body)
				var quote models.Quote
				if err := json.Unmarshal(body, &quote); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if quote.Status != tt.expectedNewStatus {
					t.Errorf("Expected status %s, got %s", tt.expectedNewStatus, quote.Status)
				}
			}
		})
	}
}

func TestGetQuoteComparison(t *testing.T) {
	cleanup := setupTestDBForQuotes(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForQuotes(t)

	// Create a test quote with category items
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	categoryID := primitive.NewObjectID()
	quoteID := primitive.NewObjectID()

	_, err := database.GetCollection(quoteCollection).InsertOne(ctx, models.Quote{
		ID:         quoteID,
		CompanyID:  companyID,
		Number:     "ORC-2024-001",
		ClientName: "Cliente Comparacao",
		Title:      "Orcamento Comparacao",
		Status:     models.QuoteExecuted,
		Items: []models.QuoteItem{
			{ID: primitive.NewObjectID(), Description: "Servico A", Quantity: 1, UnitPrice: 1000, Total: 1000, CategoryID: categoryID},
			{ID: primitive.NewObjectID(), Description: "Servico B", Quantity: 2, UnitPrice: 500, Total: 1000},
		},
		Subtotal:   2000,
		Total:      2000,
		ValidUntil: time.Now().AddDate(0, 0, 30),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test quote: %v", err)
	}

	// Create related transactions
	database.GetCollection("transactions").InsertMany(ctx, []interface{}{
		models.Transaction{
			ID:        primitive.NewObjectID(),
			CompanyID: companyID,
			Month:     "Janeiro",
			Year:      2024,
			Amount:    800,
			Category:  categoryID.Hex(),
			Status:    models.StatusPaid,
		},
	})

	tests := []struct {
		name           string
		quoteID        string
		expectedStatus int
	}{
		{
			name:           "Valid comparison",
			quoteID:        quoteID.Hex(),
			expectedStatus: 200,
		},
		{
			name:           "Invalid quote ID format",
			quoteID:        "invalid-id",
			expectedStatus: 400,
		},
		{
			name:           "Non-existent quote",
			quoteID:        primitive.NewObjectID().Hex(),
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupQuoteTestApp(userID)

			req := httptest.NewRequest("GET", "/quotes/"+tt.quoteID+"/comparison", nil)
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
				var comparison models.QuoteComparison
				if err := json.Unmarshal(body, &comparison); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if comparison.QuoteID != quoteID.Hex() {
					t.Errorf("Expected quote ID %s, got %s", quoteID.Hex(), comparison.QuoteID)
				}
				if comparison.QuotedTotal != 2000 {
					t.Errorf("Expected quoted total 2000, got %f", comparison.QuotedTotal)
				}
				if len(comparison.Items) != 2 {
					t.Errorf("Expected 2 comparison items, got %d", len(comparison.Items))
				}
			}
		})
	}
}

func TestCreateQuoteInvalidJSON(t *testing.T) {
	cleanup := setupTestDBForQuotes(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForQuotes(t)
	app := setupQuoteTestApp(userID)

	req := httptest.NewRequest("POST", "/quotes?companyId="+companyID.Hex(), bytes.NewBufferString("invalid json"))
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

func TestUpdateQuoteInvalidJSON(t *testing.T) {
	cleanup := setupTestDBForQuotes(t)
	if cleanup == nil {
		return
	}
	defer cleanup()

	userID, companyID := createTestUserAndCompanyForQuotes(t)

	// Create a test quote
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	quoteID := primitive.NewObjectID()
	database.GetCollection(quoteCollection).InsertOne(ctx, models.Quote{
		ID:         quoteID,
		CompanyID:  companyID,
		Number:     "ORC-2024-001",
		ClientName: "Cliente",
		Title:      "Orcamento",
		Status:     models.QuoteDraft,
		Items: []models.QuoteItem{
			{ID: primitive.NewObjectID(), Description: "Item", Quantity: 1, UnitPrice: 100, Total: 100},
		},
		Total:      100,
		ValidUntil: time.Now().AddDate(0, 0, 30),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	})

	app := setupQuoteTestApp(userID)

	req := httptest.NewRequest("PUT", "/quotes/"+quoteID.Hex(), bytes.NewBufferString("{invalid json}"))
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
func BenchmarkGenerateQuoteNumber(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		b.Skipf("Skipping benchmark: MongoDB not available - %v", err)
		return
	}
	defer client.Disconnect(ctx)

	database.DB = client.Database("m2m_benchmark_quotes")
	defer database.DB.Drop(ctx)

	companyID := primitive.NewObjectID()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateQuoteNumber(ctx, companyID)
	}
}

func BenchmarkGetQuotes(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		b.Skipf("Skipping benchmark: MongoDB not available - %v", err)
		return
	}
	defer client.Disconnect(ctx)

	database.DB = client.Database("m2m_benchmark_quotes_get")
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

	// Create 50 test quotes
	quotes := make([]interface{}, 50)
	for i := 0; i < 50; i++ {
		quotes[i] = models.Quote{
			ID:         primitive.NewObjectID(),
			CompanyID:  companyID,
			Number:     fmt.Sprintf("ORC-2024-%03d", i+1),
			ClientName: fmt.Sprintf("Cliente %d", i),
			Title:      fmt.Sprintf("Orcamento %d", i),
			Status:     models.QuoteDraft,
			Items: []models.QuoteItem{
				{ID: primitive.NewObjectID(), Description: "Item", Quantity: 1, UnitPrice: 100, Total: 100},
			},
			Total:      100,
			ValidUntil: time.Now().AddDate(0, 0, 30),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
	}
	database.GetCollection(quoteCollection).InsertMany(ctx, quotes)

	app := setupQuoteTestApp(userID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/quotes?companyId="+companyID.Hex(), nil)
		app.Test(req, -1)
	}
}
