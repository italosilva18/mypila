package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
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
const refreshTokenCollection = "refresh_tokens"

// JWT expiration times - shorter for production security
var (
	accessTokenExpiry  = 30 * time.Minute   // Short-lived access token
	refreshTokenExpiry = 7 * 24 * time.Hour // Longer refresh token
)

func init() {
	// Allow override via environment for development convenience
	if os.Getenv("GO_ENV") == "development" {
		accessTokenExpiry = 24 * time.Hour // Longer for development
	}
}

// hashToken creates a SHA256 hash of the token for secure storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits of entropy
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// logSecurityEvent logs security-related events for auditing
func logSecurityEvent(eventType, email, ip, userAgent, result, requestID string) {
	log.Printf("[SECURITY] %s | requestId=%s | email=%s | ip=%s | ua=%s | result=%s",
		eventType, requestID, email, ip, userAgent, result)
}

func Register(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Capture request metadata for security logging
	ip := c.IP()
	userAgent := c.Get("User-Agent")
	requestID := helpers.GetRequestID(c)

	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Normalize email to lowercase before validation
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Validations
	errors := helpers.CollectErrors(
		helpers.ValidateRequired(req.Name, "name"),
		helpers.ValidateMaxLength(req.Name, "name", 100),
		helpers.ValidateEmail(req.Email),
		helpers.ValidateMinLength(req.Password, "password", 6),
		helpers.ValidateMaxLength(req.Password, "password", 72),
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
		return helpers.DatabaseError(c, "check_user_exists", err)
	}
	if count > 0 {
		return helpers.AuthEmailExists(c)
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return helpers.InternalError(c, helpers.ErrCodeInternalError, "Falha ao processar senha", err, nil)
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
		logSecurityEvent("REGISTER_FAILED", req.Email, ip, userAgent, "db_error", requestID)
		return helpers.DatabaseError(c, "create_user", err)
	}

	// Generate access token
	accessToken, err := generateAccessToken(user)
	if err != nil {
		logSecurityEvent("REGISTER_FAILED", req.Email, ip, userAgent, "token_error", requestID)
		return helpers.AuthTokenFailed(c, err)
	}

	// Generate and store refresh token
	refreshToken, err := createRefreshToken(ctx, user.ID, ip, userAgent)
	if err != nil {
		logSecurityEvent("REGISTER_FAILED", req.Email, ip, userAgent, "refresh_token_error", requestID)
		return helpers.InternalError(c, helpers.ErrCodeAuthTokenFailed, "Falha ao gerar token de refresh", err, nil)
	}

	logSecurityEvent("REGISTER_SUCCESS", req.Email, ip, userAgent, "success", requestID)

	return c.Status(201).JSON(models.AuthResponseWithTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTokenExpiry.Seconds()),
		User:         user,
	})
}

func Login(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Capture request metadata for security logging
	ip := c.IP()
	userAgent := c.Get("User-Agent")
	requestID := helpers.GetRequestID(c)

	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Normalize email to lowercase before validation
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Validations
	errors := helpers.CollectErrors(
		helpers.ValidateEmail(req.Email),
		helpers.ValidateRequired(req.Password, "password"),
		helpers.ValidateMaxLength(req.Password, "password", 72),
	)

	if helpers.HasErrors(errors) {
		logSecurityEvent("LOGIN_FAILED", req.Email, ip, userAgent, "validation_error", requestID)
		return helpers.SendValidationErrors(c, errors)
	}

	collection := database.GetCollection(userCollection)

	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		logSecurityEvent("LOGIN_FAILED", req.Email, ip, userAgent, "user_not_found", requestID)
		return helpers.AuthInvalidCredentials(c)
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		logSecurityEvent("LOGIN_FAILED", req.Email, ip, userAgent, "invalid_password", requestID)
		return helpers.AuthInvalidCredentials(c)
	}

	// Generate access token
	accessToken, err := generateAccessToken(user)
	if err != nil {
		logSecurityEvent("LOGIN_FAILED", req.Email, ip, userAgent, "token_error", requestID)
		return helpers.AuthTokenFailed(c, err)
	}

	// Generate and store refresh token
	refreshToken, err := createRefreshToken(ctx, user.ID, ip, userAgent)
	if err != nil {
		logSecurityEvent("LOGIN_FAILED", req.Email, ip, userAgent, "refresh_token_error", requestID)
		return helpers.InternalError(c, helpers.ErrCodeAuthTokenFailed, "Falha ao gerar token de refresh", err, nil)
	}

	logSecurityEvent("LOGIN_SUCCESS", req.Email, ip, userAgent, "success", requestID)

	return c.JSON(models.AuthResponseWithTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTokenExpiry.Seconds()),
		User:         user,
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
		return helpers.AuthInvalidUserContext(c)
	}

	// Convert string ID to ObjectID
	userId, err := primitive.ObjectIDFromHex(userIdStr)
	if err != nil {
		return helpers.InvalidIDFormat(c, "userId")
	}

	// Fetch user from database
	collection := database.GetCollection(userCollection)
	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if err != nil {
		return helpers.AuthUserNotFound(c)
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

// generateAccessToken creates a new JWT access token for the user
func generateAccessToken(user models.User) (string, error) {
	// Use the secure JWT secret from config
	secret := config.GetJWTSecret()

	claims := jwt.MapClaims{
		"userId": user.ID.Hex(),
		"email":  user.Email,
		"exp":    time.Now().Add(accessTokenExpiry).Unix(),
		"iat":    time.Now().Unix(),
		"type":   "access", // Token type for validation
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// createRefreshToken generates a new refresh token and stores its hash in the database
func createRefreshToken(ctx context.Context, userID primitive.ObjectID, ipAddress, userAgent string) (string, error) {
	// Generate a secure random token
	token, err := generateSecureToken()
	if err != nil {
		return "", err
	}

	// Hash the token before storing
	tokenHash := hashToken(token)

	// Create the refresh token record
	refreshToken := models.RefreshToken{
		ID:        primitive.NewObjectID(),
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: time.Now().Add(refreshTokenExpiry),
		CreatedAt: time.Now(),
		IsRevoked: false,
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}

	// Store in database
	collection := database.GetCollection(refreshTokenCollection)
	_, err = collection.InsertOne(ctx, refreshToken)
	if err != nil {
		return "", err
	}

	log.Printf("[SECURITY] REFRESH_TOKEN_CREATED | userId=%s | ip=%s", userID.Hex(), ipAddress)

	return token, nil
}

// RefreshToken handles the token refresh endpoint
// POST /api/auth/refresh
func RefreshToken(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ip := c.IP()
	userAgent := c.Get("User-Agent")
	requestID := helpers.GetRequestID(c)

	var req models.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	if req.RefreshToken == "" {
		logSecurityEvent("REFRESH_FAILED", "", ip, userAgent, "missing_token", requestID)
		return helpers.BadRequest(c, helpers.ErrCodeBadRequest, "Token de refresh e obrigatorio", nil)
	}

	// Hash the incoming token to compare with stored hash
	tokenHash := hashToken(req.RefreshToken)

	// Find the refresh token in database
	collection := database.GetCollection(refreshTokenCollection)
	var storedToken models.RefreshToken
	err := collection.FindOne(ctx, bson.M{"tokenHash": tokenHash}).Decode(&storedToken)
	if err != nil {
		logSecurityEvent("REFRESH_FAILED", "", ip, userAgent, "token_not_found", requestID)
		return helpers.AuthInvalidToken(c)
	}

	// Check if token is revoked
	if storedToken.IsRevoked {
		// Potential token reuse attack - revoke all tokens for this user
		logSecurityEvent("REFRESH_FAILED", "", ip, userAgent, "token_already_revoked_potential_attack", requestID)
		revokeAllUserTokens(ctx, storedToken.UserID)
		return helpers.Unauthorized(c, helpers.ErrCodeAuthInvalidToken, "Token de refresh foi revogado. Por seguranca, faca login novamente.", nil)
	}

	// Check if token is expired
	if time.Now().After(storedToken.ExpiresAt) {
		logSecurityEvent("REFRESH_FAILED", "", ip, userAgent, "token_expired", requestID)
		// Mark as revoked since it's expired
		revokeToken(ctx, storedToken.ID)
		return helpers.Unauthorized(c, helpers.ErrCodeAuthInvalidToken, "Token de refresh expirado", nil)
	}

	// Get the user
	userColl := database.GetCollection("users")
	var user models.User
	err = userColl.FindOne(ctx, bson.M{"_id": storedToken.UserID}).Decode(&user)
	if err != nil {
		logSecurityEvent("REFRESH_FAILED", "", ip, userAgent, "user_not_found", requestID)
		return helpers.AuthUserNotFound(c)
	}

	// Token rotation: revoke the old token
	err = revokeToken(ctx, storedToken.ID)
	if err != nil {
		log.Printf("[SECURITY] WARNING: Failed to revoke old refresh token: %v", err)
		// Continue anyway - new token will still work
	}

	// Generate new access token
	accessToken, err := generateAccessToken(user)
	if err != nil {
		logSecurityEvent("REFRESH_FAILED", user.Email, ip, userAgent, "access_token_error", requestID)
		return helpers.AuthTokenFailed(c, err)
	}

	// Generate new refresh token (token rotation)
	newRefreshToken, err := createRefreshToken(ctx, user.ID, ip, userAgent)
	if err != nil {
		logSecurityEvent("REFRESH_FAILED", user.Email, ip, userAgent, "refresh_token_error", requestID)
		return helpers.InternalError(c, helpers.ErrCodeAuthTokenFailed, "Falha ao gerar novo token de refresh", err, nil)
	}

	logSecurityEvent("REFRESH_SUCCESS", user.Email, ip, userAgent, "success", requestID)

	return c.JSON(models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(accessTokenExpiry.Seconds()),
	})
}

// Logout handles user logout by revoking the refresh token
// POST /api/auth/logout
func Logout(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ip := c.IP()
	userAgent := c.Get("User-Agent")
	requestID := helpers.GetRequestID(c)

	var req models.LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	if req.RefreshToken == "" {
		// Still return success - client might not have the token
		return c.JSON(fiber.Map{"message": "Logout realizado com sucesso"})
	}

	// Hash the token and revoke it
	tokenHash := hashToken(req.RefreshToken)

	collection := database.GetCollection(refreshTokenCollection)
	now := time.Now()
	result, err := collection.UpdateOne(
		ctx,
		bson.M{"tokenHash": tokenHash, "isRevoked": false},
		bson.M{
			"$set": bson.M{
				"isRevoked": true,
				"revokedAt": now,
			},
		},
	)

	if err != nil {
		log.Printf("[SECURITY] WARNING: Failed to revoke token on logout: %v", err)
		// Still return success - don't leak info
	}

	if result != nil && result.ModifiedCount > 0 {
		logSecurityEvent("LOGOUT_SUCCESS", "", ip, userAgent, "token_revoked", requestID)
	} else {
		logSecurityEvent("LOGOUT_SUCCESS", "", ip, userAgent, "no_token_to_revoke", requestID)
	}

	return c.JSON(fiber.Map{"message": "Logout realizado com sucesso"})
}

// LogoutAll revokes all refresh tokens for the authenticated user
// POST /api/auth/logout-all
func LogoutAll(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ip := c.IP()
	userAgent := c.Get("User-Agent")
	requestID := helpers.GetRequestID(c)

	// Get userId from context (set by Protected middleware)
	userIdStr, ok := c.Locals("userId").(string)
	if !ok {
		return helpers.AuthInvalidUserContext(c)
	}

	userId, err := primitive.ObjectIDFromHex(userIdStr)
	if err != nil {
		return helpers.InvalidIDFormat(c, "userId")
	}

	// Revoke all tokens for this user
	count, err := revokeAllUserTokens(ctx, userId)
	if err != nil {
		log.Printf("[SECURITY] WARNING: Failed to revoke all tokens: %v", err)
		return helpers.InternalError(c, helpers.ErrCodeInternalError, "Falha ao revogar tokens", err, nil)
	}

	logSecurityEvent("LOGOUT_ALL_SUCCESS", "", ip, userAgent, "all_tokens_revoked", requestID)

	return c.JSON(fiber.Map{
		"message":       "Logout de todos os dispositivos realizado com sucesso",
		"tokensRevoked": count,
	})
}

// revokeToken revokes a single refresh token by ID
func revokeToken(ctx context.Context, tokenID primitive.ObjectID) error {
	collection := database.GetCollection(refreshTokenCollection)
	now := time.Now()

	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": tokenID},
		bson.M{
			"$set": bson.M{
				"isRevoked": true,
				"revokedAt": now,
			},
		},
	)

	return err
}

// revokeAllUserTokens revokes all refresh tokens for a user
func revokeAllUserTokens(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	collection := database.GetCollection(refreshTokenCollection)
	now := time.Now()

	result, err := collection.UpdateMany(
		ctx,
		bson.M{"userId": userID, "isRevoked": false},
		bson.M{
			"$set": bson.M{
				"isRevoked": true,
				"revokedAt": now,
			},
		},
	)

	if err != nil {
		return 0, err
	}

	if result.ModifiedCount > 0 {
		log.Printf("[SECURITY] TOKENS_REVOKED | userId=%s | count=%d", userID.Hex(), result.ModifiedCount)
	}

	return result.ModifiedCount, nil
}

// CleanupExpiredTokens removes expired refresh tokens from the database
// This should be called periodically (e.g., via cron job or scheduled task)
func CleanupExpiredTokens(ctx context.Context) (int64, error) {
	collection := database.GetCollection(refreshTokenCollection)

	result, err := collection.DeleteMany(
		ctx,
		bson.M{
			"$or": []bson.M{
				{"expiresAt": bson.M{"$lt": time.Now()}},
				{"isRevoked": true, "revokedAt": bson.M{"$lt": time.Now().Add(-24 * time.Hour)}}, // Keep revoked for 24h for audit
			},
		},
	)

	if err != nil {
		return 0, err
	}

	if result.DeletedCount > 0 {
		log.Printf("[SECURITY] CLEANUP_TOKENS | deleted=%d", result.DeletedCount)
	}

	return result.DeletedCount, nil
}
