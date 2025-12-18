package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"m2m-backend/config"
	"m2m-backend/database"
	"m2m-backend/handlers"
	"m2m-backend/middleware"
	"m2m-backend/migrations"
)

func main() {
	// Initialize JWT Secret (must be first!)
	if err := config.InitializeJWTSecret(); err != nil {
		log.Fatal("[SECURITY ERROR] Failed to initialize JWT secret:", err)
	}

	// Connect to MongoDB
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Create database indexes for optimal query performance
	if err := migrations.CreateIndexes(database.DB); err != nil {
		log.Printf("Warning: Failed to create some indexes: %v", err)
		// Don't fail the application, indexes are for optimization
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "M2M Financeiro API",
	})

	// Security Headers Middleware - Protection against common web vulnerabilities
	app.Use(func(c *fiber.Ctx) error {
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
		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'")
		
		// Enforce HTTPS (Strict-Transport-Security)
		// Note: Only enable in production with valid SSL certificate
		// c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		
		return c.Next()
	})

	// Global middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
	}))

	// Global rate limiting: 100 requests per minute per IP
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"error": "Muitas requisições. Tente novamente em alguns minutos.",
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		Storage:                nil, // Uses memory storage by default
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "m2m-backend"})
	})

	// API Routes
	api := app.Group("/api")

	// Auth routes with stricter rate limiting (20 requests per minute)
	auth := api.Group("/auth")
	authLimiter := limiter.New(limiter.Config{
		Max:        20,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"error": "Muitas tentativas de autenticação. Tente novamente em alguns minutos.",
			})
		},
	})
	auth.Post("/register", authLimiter, handlers.Register)
	auth.Post("/login", authLimiter, handlers.Login)
	auth.Get("/me", middleware.Protected(), handlers.GetMe)

	// Protected routes
	api.Use(middleware.Protected())

	// Company routes
	companies := api.Group("/companies")
	companies.Get("/", handlers.GetCompanies)
	companies.Post("/", handlers.CreateCompany)
	companies.Put("/:id", handlers.UpdateCompany)
	companies.Delete("/:id", handlers.DeleteCompany)

	// Transactions routes
	transactions := api.Group("/transactions")
	transactions.Get("/", handlers.GetAllTransactions)
	transactions.Get("/:id", handlers.GetTransaction)
	transactions.Post("/", handlers.CreateTransaction)
	transactions.Put("/:id", handlers.UpdateTransaction)
	transactions.Delete("/:id", handlers.DeleteTransaction)
	transactions.Patch("/:id/toggle-status", handlers.ToggleStatus)

	// Stats route
	api.Get("/stats", handlers.GetStats)

	// Category routes
	categories := api.Group("/categories")
	categories.Get("/", handlers.GetCategories)
	categories.Post("/", handlers.CreateCategory)
	categories.Put("/:id", handlers.UpdateCategory)
	categories.Delete("/:id", handlers.DeleteCategory)

	// Recurring routes
	recurring := api.Group("/recurring")
	recurring.Get("/", handlers.GetRecurring)
	recurring.Post("/", handlers.CreateRecurring)
	recurring.Delete("/:id", handlers.DeleteRecurring)
	recurring.Post("/process", handlers.ProcessRecurring)

	// Seed route (for initial data)
	api.Post("/seed", handlers.SeedTransactions)

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

	// Close MongoDB connection
	if err := database.Disconnect(); err != nil {
		log.Printf("Error closing MongoDB connection: %v", err)
	}

	log.Println("Server shutdown complete")
}
