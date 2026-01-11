package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"m2m-backend/config"
	"m2m-backend/database"
	"m2m-backend/handlers"
	"m2m-backend/helpers"
	"m2m-backend/middleware"
)

func main() {
	// Initialize JWT Secret (must be first!)
	if err := config.InitializeJWTSecret(); err != nil {
		log.Fatal("[SECURITY ERROR] Failed to initialize JWT secret:", err)
	}

	// Connect to PostgreSQL
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "MyPila API",
	})

	// Recover middleware - previne que panics derrubem o servidor
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	// Request ID middleware - para rastreamento de requisicoes
	app.Use(requestid.New())

	// Compression middleware - reduz tamanho das respostas
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// Security Headers Middleware - Protection against common web vulnerabilities
	app.Use(func(c *fiber.Ctx) error {
		env := os.Getenv("GO_ENV")

		// Prevent clickjacking attacks by disallowing iframe embedding
		c.Set("X-Frame-Options", "DENY")

		// Prevent MIME-type sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// Enable XSS filter built into most browsers
		c.Set("X-XSS-Protection", "1; mode=block")

		// Control how much referrer information should be included with requests
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Restrict which features and APIs can be used in the browser
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Content Security Policy - restrict resources the browser can load
		// Using nonce-based CSP would be ideal, but for now we use strict policy without unsafe-inline
		if env == "production" {
			c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		} else {
			// More permissive CSP for development (allows hot reload, dev tools)
			c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' ws: wss:")
		}

		// Enforce HTTPS (Strict-Transport-Security) - ONLY in production with SSL
		if env == "production" {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		return c.Next()
	})

	// Global middleware - Logger com Request ID para rastreamento
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${locals:requestid} ${status} - ${method} ${path} ${latency}\n",
	}))

	// CORS configuration - restrict to known origins
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "http://localhost:3000,http://localhost:3333,http://localhost:5173"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// Global rate limiting: 100 requests per minute per IP
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return helpers.RateLimited(c, "Muitas requisicoes. Tente novamente em alguns minutos.", nil)
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		Storage:                nil, // Uses memory storage by default
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "mypila-backend"})
	})

	// API Routes
	api := app.Group("/api")

	// Auth routes with stricter rate limiting (20 requests per minute)
	auth := api.Group("/auth")
	authLimiter := middleware.AuthLimiter()
	auth.Post("/register", authLimiter, handlers.Register)
	auth.Post("/login", authLimiter, handlers.Login)
	auth.Post("/refresh", authLimiter, handlers.RefreshToken)
	auth.Post("/logout", handlers.Logout)
	auth.Get("/me", middleware.Protected(), handlers.GetMe)
	auth.Post("/logout-all", middleware.Protected(), handlers.LogoutAll)

	// Protected routes
	api.Use(middleware.Protected())

	// Company routes
	companies := api.Group("/companies")
	companies.Get("/", handlers.GetCompanies)
	companies.Post("/", middleware.ModerateLimiter(), handlers.CreateCompany)
	companies.Put("/:id", handlers.UpdateCompany)
	companies.Delete("/:id", middleware.StrictLimiter(), handlers.DeleteCompany)

	// Transactions routes
	transactions := api.Group("/transactions")
	transactions.Get("/", handlers.GetAllTransactions)
	transactions.Get("/:id", handlers.GetTransaction)
	transactions.Post("/", middleware.ModerateLimiter(), handlers.CreateTransaction)
	transactions.Put("/:id", handlers.UpdateTransaction)
	transactions.Delete("/:id", middleware.StrictLimiter(), handlers.DeleteTransaction)
	transactions.Patch("/:id/toggle-status", handlers.ToggleStatus)

	// Stats route
	api.Get("/stats", handlers.GetStats)

	// Category routes
	categories := api.Group("/categories")
	categories.Get("/", handlers.GetCategories)
	categories.Post("/", middleware.ModerateLimiter(), handlers.CreateCategory)
	categories.Put("/:id", handlers.UpdateCategory)
	categories.Delete("/:id", middleware.StrictLimiter(), handlers.DeleteCategory)

	// Recurring routes
	recurring := api.Group("/recurring")
	recurring.Get("/", handlers.GetRecurring)
	recurring.Post("/", middleware.ModerateLimiter(), handlers.CreateRecurring)
	recurring.Delete("/:id", middleware.StrictLimiter(), handlers.DeleteRecurring)
	recurring.Post("/process", middleware.HeavyOperationLimiter(), handlers.ProcessRecurring)

	// Quote routes
	quotes := api.Group("/quotes")
	quotes.Get("/", handlers.GetQuotes)
	quotes.Get("/:id", handlers.GetQuote)
	quotes.Post("/", middleware.ModerateLimiter(), handlers.CreateQuote)
	quotes.Put("/:id", handlers.UpdateQuote)
	quotes.Delete("/:id", middleware.StrictLimiter(), handlers.DeleteQuote)
	quotes.Post("/:id/duplicate", middleware.ModerateLimiter(), handlers.DuplicateQuote)
	quotes.Patch("/:id/status", handlers.UpdateQuoteStatus)
	quotes.Get("/:id/pdf", middleware.HeavyOperationLimiter(), handlers.GenerateQuotePDF)
	quotes.Get("/:id/comparison", handlers.GetQuoteComparison)

	// Quote Template routes
	quoteTemplates := api.Group("/quote-templates")
	quoteTemplates.Get("/", handlers.GetQuoteTemplates)
	quoteTemplates.Get("/:id", handlers.GetQuoteTemplate)
	quoteTemplates.Post("/", middleware.ModerateLimiter(), handlers.CreateQuoteTemplate)
	quoteTemplates.Put("/:id", handlers.UpdateQuoteTemplate)
	quoteTemplates.Delete("/:id", middleware.StrictLimiter(), handlers.DeleteQuoteTemplate)

	// Seed route (for initial data) - ONLY in explicit development mode
	if os.Getenv("GO_ENV") == "development" {
		api.Post("/seed", handlers.SeedTransactions)
		log.Println("[DEV] Seed endpoint enabled at POST /api/seed")
	}

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	<-quit
	log.Println("Shutting down server...")

	// Gracefully shutdown the Fiber server
	if err := app.Shutdown(); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	// Close PostgreSQL connection
	database.Disconnect()

	log.Println("Server shutdown complete")
}
