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
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"m2m-backend/config"
	"m2m-backend/database"
	"m2m-backend/helpers"
	"m2m-backend/models"
	"m2m-backend/services"
)

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
		helpers.ValidateSQLInjection(req.Name, "name"),
	)

	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Sanitize user input after validation (DO NOT sanitize email/password)
	req.Name = helpers.SanitizeString(req.Name)

	// Check if user exists
	var count int
	err := database.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", req.Email).Scan(&count)
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

	userID := uuid.New()
	now := time.Now()

	user := models.User{
		ID:           userID,
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err = database.Pool.Exec(ctx,
		`INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		user.ID, user.Name, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
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

	var user models.User
	err := database.QueryRow(ctx,
		`SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email = $1`,
		req.Email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
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
func GetMe(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get userId from context (set by Protected middleware)
	userIdStr, ok := c.Locals("userId").(string)
	if !ok {
		return helpers.AuthInvalidUserContext(c)
	}

	// Convert string ID to UUID
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		return helpers.InvalidIDFormat(c, "userId")
	}

	// Fetch user from database
	var user models.User
	err = database.QueryRow(ctx,
		`SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE id = $1`,
		userId).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
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
		"userId": user.ID.String(),
		"email":  user.Email,
		"exp":    time.Now().Add(accessTokenExpiry).Unix(),
		"iat":    time.Now().Unix(),
		"type":   "access", // Token type for validation
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// createRefreshToken generates a new refresh token and stores its hash in the database
func createRefreshToken(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) (string, error) {
	// Generate a secure random token
	token, err := generateSecureToken()
	if err != nil {
		return "", err
	}

	// Hash the token before storing
	tokenHash := hashToken(token)

	tokenID := uuid.New()
	now := time.Now()
	expiresAt := now.Add(refreshTokenExpiry)

	// Store in database
	_, err = database.Pool.Exec(ctx,
		`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, is_revoked, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		tokenID, userID, tokenHash, expiresAt, false, now)
	if err != nil {
		return "", err
	}

	log.Printf("[SECURITY] REFRESH_TOKEN_CREATED | userId=%s | ip=%s", userID.String(), ipAddress)

	return token, nil
}

// RefreshToken handles the token refresh endpoint
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
	var storedToken models.RefreshToken
	err := database.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, is_revoked, created_at FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash).Scan(&storedToken.ID, &storedToken.UserID, &storedToken.TokenHash, &storedToken.ExpiresAt, &storedToken.IsRevoked, &storedToken.CreatedAt)
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
	var user models.User
	err = database.QueryRow(ctx,
		`SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE id = $1`,
		storedToken.UserID).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
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

	now := time.Now()
	result, err := database.Pool.Exec(ctx,
		`UPDATE refresh_tokens SET is_revoked = true WHERE token_hash = $1 AND is_revoked = false`,
		tokenHash)

	if err != nil {
		log.Printf("[SECURITY] WARNING: Failed to revoke token on logout: %v", err)
		// Still return success - don't leak info
	}

	if result.RowsAffected() > 0 {
		logSecurityEvent("LOGOUT_SUCCESS", "", ip, userAgent, "token_revoked", requestID)
	} else {
		logSecurityEvent("LOGOUT_SUCCESS", "", ip, userAgent, "no_token_to_revoke", requestID)
	}

	_ = now // unused but keeping for consistency

	return c.JSON(fiber.Map{"message": "Logout realizado com sucesso"})
}

// LogoutAll revokes all refresh tokens for the authenticated user
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

	userId, err := uuid.Parse(userIdStr)
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
func revokeToken(ctx context.Context, tokenID uuid.UUID) error {
	_, err := database.Pool.Exec(ctx,
		`UPDATE refresh_tokens SET is_revoked = true WHERE id = $1`,
		tokenID)
	return err
}

// revokeAllUserTokens revokes all refresh tokens for a user
func revokeAllUserTokens(ctx context.Context, userID uuid.UUID) (int64, error) {
	result, err := database.Pool.Exec(ctx,
		`UPDATE refresh_tokens SET is_revoked = true WHERE user_id = $1 AND is_revoked = false`,
		userID)

	if err != nil {
		return 0, err
	}

	count := result.RowsAffected()
	if count > 0 {
		log.Printf("[SECURITY] TOKENS_REVOKED | userId=%s | count=%d", userID.String(), count)
	}

	return count, nil
}

// CleanupExpiredTokens removes expired refresh tokens from the database
func CleanupExpiredTokens(ctx context.Context) (int64, error) {
	result, err := database.Pool.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE expires_at < $1 OR (is_revoked = true AND created_at < $2)`,
		time.Now(), time.Now().Add(-24*time.Hour))

	if err != nil {
		return 0, err
	}

	count := result.RowsAffected()
	if count > 0 {
		log.Printf("[SECURITY] CLEANUP_TOKENS | deleted=%d", count)
	}

	return count, nil
}

// ForgotPassword handles password reset requests
func ForgotPassword(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ip := c.IP()
	userAgent := c.Get("User-Agent")
	requestID := helpers.GetRequestID(c)

	var req models.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Validate email
	errors := helpers.CollectErrors(
		helpers.ValidateEmail(req.Email),
	)
	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	// Check if user exists
	var user models.User
	err := database.QueryRow(ctx,
		`SELECT id, name, email FROM users WHERE email = $1`,
		req.Email).Scan(&user.ID, &user.Name, &user.Email)

	// Always return success message to prevent email enumeration
	successResponse := models.ForgotPasswordResponse{
		Message: "Se o email estiver cadastrado, voce recebera um link para redefinir sua senha.",
	}

	if err != nil {
		logSecurityEvent("FORGOT_PASSWORD", req.Email, ip, userAgent, "user_not_found", requestID)
		return c.JSON(successResponse)
	}

	// Generate reset token
	resetToken, err := generateSecureToken()
	if err != nil {
		logSecurityEvent("FORGOT_PASSWORD", req.Email, ip, userAgent, "token_generation_failed", requestID)
		return helpers.InternalError(c, helpers.ErrCodeInternalError, "Falha ao gerar token", err, nil)
	}

	tokenHash := hashToken(resetToken)
	tokenID := uuid.New()
	expiresAt := time.Now().Add(1 * time.Hour) // Token expires in 1 hour

	// Delete any existing reset tokens for this user
	_, _ = database.Pool.Exec(ctx,
		`DELETE FROM password_reset_tokens WHERE user_id = $1`,
		user.ID)

	// Store the reset token
	_, err = database.Pool.Exec(ctx,
		`INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, used, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		tokenID, user.ID, tokenHash, expiresAt, false, time.Now())
	if err != nil {
		logSecurityEvent("FORGOT_PASSWORD", req.Email, ip, userAgent, "db_error", requestID)
		return helpers.DatabaseError(c, "create_reset_token", err)
	}

	// Send email
	emailService := services.NewEmailService()
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:8007"
	}

	err = emailService.SendPasswordResetEmail(user.Email, user.Name, resetToken, frontendURL)
	if err != nil {
		log.Printf("[ERROR] Failed to send password reset email: %v", err)
		logSecurityEvent("FORGOT_PASSWORD", req.Email, ip, userAgent, "email_send_failed", requestID)
		// Still return success to prevent email enumeration
		return c.JSON(successResponse)
	}

	logSecurityEvent("FORGOT_PASSWORD", req.Email, ip, userAgent, "email_sent", requestID)
	return c.JSON(successResponse)
}

// ResetPassword handles the actual password reset
func ResetPassword(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ip := c.IP()
	userAgent := c.Get("User-Agent")
	requestID := helpers.GetRequestID(c)

	var req models.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.InvalidRequestBody(c)
	}

	// Validate request
	if req.Token == "" {
		return helpers.BadRequest(c, helpers.ErrCodeBadRequest, "Token e obrigatorio", nil)
	}

	errors := helpers.CollectErrors(
		helpers.ValidateMinLength(req.NewPassword, "newPassword", 6),
		helpers.ValidateMaxLength(req.NewPassword, "newPassword", 72),
	)
	if helpers.HasErrors(errors) {
		return helpers.SendValidationErrors(c, errors)
	}

	tokenHash := hashToken(req.Token)

	// Find the reset token
	var resetToken models.PasswordResetToken
	err := database.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, used, created_at
		 FROM password_reset_tokens WHERE token_hash = $1`,
		tokenHash).Scan(&resetToken.ID, &resetToken.UserID, &resetToken.TokenHash,
		&resetToken.ExpiresAt, &resetToken.Used, &resetToken.CreatedAt)

	if err != nil {
		logSecurityEvent("RESET_PASSWORD", "", ip, userAgent, "token_not_found", requestID)
		return helpers.BadRequest(c, helpers.ErrCodeBadRequest, "Token invalido ou expirado", nil)
	}

	// Check if token is already used
	if resetToken.Used {
		logSecurityEvent("RESET_PASSWORD", "", ip, userAgent, "token_already_used", requestID)
		return helpers.BadRequest(c, helpers.ErrCodeBadRequest, "Token ja foi utilizado", nil)
	}

	// Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		logSecurityEvent("RESET_PASSWORD", "", ip, userAgent, "token_expired", requestID)
		return helpers.BadRequest(c, helpers.ErrCodeBadRequest, "Token expirado. Solicite um novo link.", nil)
	}

	// Hash the new password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return helpers.InternalError(c, helpers.ErrCodeInternalError, "Falha ao processar senha", err, nil)
	}

	// Update user password
	_, err = database.Pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`,
		string(hash), time.Now(), resetToken.UserID)
	if err != nil {
		logSecurityEvent("RESET_PASSWORD", "", ip, userAgent, "db_error", requestID)
		return helpers.DatabaseError(c, "update_password", err)
	}

	// Mark token as used
	_, _ = database.Pool.Exec(ctx,
		`UPDATE password_reset_tokens SET used = true WHERE id = $1`,
		resetToken.ID)

	// Revoke all refresh tokens for security
	_, _ = revokeAllUserTokens(ctx, resetToken.UserID)

	logSecurityEvent("RESET_PASSWORD", "", ip, userAgent, "success", requestID)

	return c.JSON(fiber.Map{
		"message": "Senha redefinida com sucesso! Faca login com sua nova senha.",
	})
}
