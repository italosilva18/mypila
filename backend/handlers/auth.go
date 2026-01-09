package handlers

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"m2m-backend/config"
	"m2m-backend/database"
	"m2m-backend/helpers"
	"m2m-backend/models"
)

const userCollection = "users"

// JWT expiration times - shorter for production security
var (
	accessTokenExpiry  = 30 * time.Minute  // Short-lived access token
	refreshTokenExpiry = 7 * 24 * time.Hour // Longer refresh token (future use)
)

func init() {
	// Allow override via environment for development convenience
	if os.Getenv("GO_ENV") == "development" {
		accessTokenExpiry = 24 * time.Hour // Longer for development
	}
}

// logSecurityEvent logs security-related events for auditing
func logSecurityEvent(eventType, email, ip, userAgent, result string) {
	log.Printf("[SECURITY] %s | email=%s | ip=%s | ua=%s | result=%s",
		eventType, email, ip, userAgent, result)
}

func Register(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Capture request metadata for security logging
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Corpo da requisição inválido"})
	}

	// Normalizar email para lowercase antes de validar
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Validações
	errors := helpers.CollectErrors(
		helpers.ValidateRequired(req.Name, "name"),
		helpers.ValidateMaxLength(req.Name, "name", 100), // Limite razoável para nomes
		helpers.ValidateEmail(req.Email),
		helpers.ValidateMinLength(req.Password, "password", 6),
		helpers.ValidateMaxLength(req.Password, "password", 72), // Limite do bcrypt
		helpers.ValidateNoScriptTags(req.Name, "name"),
		helpers.ValidateMongoInjection(req.Name, "name"),
		helpers.ValidateSQLInjection(req.Name, "name"),
	)

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Sanitize user input after validation (DO NOT sanitize email/password)
	req.Name = helpers.SanitizeString(req.Name)

	collection := database.GetCollection(userCollection)

	// Check if user exists
	count, err := collection.CountDocuments(ctx, bson.M{"email": req.Email})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Erro no banco de dados"})
	}
	if count > 0 {
		return c.Status(409).JSON(fiber.Map{"error": "Email já cadastrado"})
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao processar senha"})
	}

	user := models.User{
		ID:           primitive.NewObjectID(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		logSecurityEvent("REGISTER_FAILED", req.Email, ip, userAgent, "db_error")
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao criar usuário"})
	}

	// Generate Token
	token, err := generateToken(user)
	if err != nil {
		logSecurityEvent("REGISTER_FAILED", req.Email, ip, userAgent, "token_error")
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao gerar token"})
	}

	logSecurityEvent("REGISTER_SUCCESS", req.Email, ip, userAgent, "success")

	return c.Status(201).JSON(models.AuthResponse{
		Token: token,
		User:  user,
	})
}

func Login(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Capture request metadata for security logging
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Corpo da requisição inválido"})
	}

	// Normalizar email para lowercase antes de validar
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Validações
	errors := helpers.CollectErrors(
		helpers.ValidateEmail(req.Email),
		helpers.ValidateRequired(req.Password, "password"),
		helpers.ValidateMaxLength(req.Password, "password", 72), // Limite do bcrypt
	)

	if helpers.HasErrors(errors) {
		logSecurityEvent("LOGIN_FAILED", req.Email, ip, userAgent, "validation_error")
		return helpers.SendValidationErrors(c, errors)
	}

	collection := database.GetCollection(userCollection)

	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		logSecurityEvent("LOGIN_FAILED", req.Email, ip, userAgent, "user_not_found")
		return c.Status(401).JSON(fiber.Map{"error": "Credenciais inválidas"})
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		logSecurityEvent("LOGIN_FAILED", req.Email, ip, userAgent, "invalid_password")
		return c.Status(401).JSON(fiber.Map{"error": "Credenciais inválidas"})
	}

	// Generate Token
	token, err := generateToken(user)
	if err != nil {
		logSecurityEvent("LOGIN_FAILED", req.Email, ip, userAgent, "token_error")
		return c.Status(500).JSON(fiber.Map{"error": "Falha ao gerar token"})
	}

	logSecurityEvent("LOGIN_SUCCESS", req.Email, ip, userAgent, "success")

	return c.JSON(models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// GetMe returns the current authenticated user's information
// This endpoint is useful for:
// - Validating if a token is still valid
// - Getting fresh user data on app load
// - Checking authentication state in the frontend
func GetMe(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get userId from context (set by Protected middleware)
	userIdStr, ok := c.Locals("userId").(string)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid user context"})
	}

	// Convert string ID to ObjectID
	userId, err := primitive.ObjectIDFromHex(userIdStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	// Fetch user from database
	collection := database.GetCollection(userCollection)
	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

func generateToken(user models.User) (string, error) {
	// Use the secure JWT secret from config
	secret := config.GetJWTSecret()

	claims := jwt.MapClaims{
		"userId": user.ID.Hex(),
		"email":  user.Email,
		"exp":    time.Now().Add(accessTokenExpiry).Unix(),
		"iat":    time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
