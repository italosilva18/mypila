package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"m2m-backend/helpers"
)

// RateLimitConfig defines the configuration for rate limiting
type RateLimitConfig struct {
	Max        int
	Expiration time.Duration
	Message    string
}

// Predefined rate limit configurations for different operation types
var (
	// StrictLimit: 10 req/min - For destructive operations (DELETE)
	StrictLimit = RateLimitConfig{
		Max:        10,
		Expiration: 1 * time.Minute,
		Message:    "Limite de operacoes destrutivas atingido. Aguarde um minuto antes de tentar novamente.",
	}

	// ModerateLimit: 30 req/min - For creation operations (POST on quotes, etc.)
	ModerateLimit = RateLimitConfig{
		Max:        30,
		Expiration: 1 * time.Minute,
		Message:    "Limite de criacao de recursos atingido. Aguarde um minuto antes de tentar novamente.",
	}

	// HeavyOperationLimit: 5 req/min - For heavy processing operations
	HeavyOperationLimit = RateLimitConfig{
		Max:        5,
		Expiration: 1 * time.Minute,
		Message:    "Operacao pesada limitada. Aguarde um minuto antes de processar novamente.",
	}

	// AuthLimit: 20 req/min - For authentication operations
	AuthLimit = RateLimitConfig{
		Max:        20,
		Expiration: 1 * time.Minute,
		Message:    "Muitas tentativas de autenticacao. Tente novamente em alguns minutos.",
	}
)

// createLimiter creates a Fiber limiter middleware with the given configuration
func createLimiter(config RateLimitConfig) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        config.Max,
		Expiration: config.Expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP + user ID for authenticated routes for more accurate limiting
			userId := c.Locals("userId")
			if userId != nil {
				return c.IP() + "-" + userId.(string)
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return helpers.RateLimited(c, config.Message, nil)
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
	})
}

// StrictLimiter returns a rate limiter for destructive operations (10 req/min)
// Use for: DELETE operations that could cause mass data loss
func StrictLimiter() fiber.Handler {
	return createLimiter(StrictLimit)
}

// ModerateLimiter returns a rate limiter for creation operations (30 req/min)
// Use for: POST operations on quotes, transactions, etc.
func ModerateLimiter() fiber.Handler {
	return createLimiter(ModerateLimit)
}

// HeavyOperationLimiter returns a rate limiter for heavy operations (5 req/min)
// Use for: Processing operations like recurring transaction processing, PDF generation, etc.
func HeavyOperationLimiter() fiber.Handler {
	return createLimiter(HeavyOperationLimit)
}

// AuthLimiter returns a rate limiter for authentication operations (20 req/min)
// Use for: Login, register, token refresh operations
func AuthLimiter() fiber.Handler {
	return createLimiter(AuthLimit)
}

// CustomLimiter allows creating a rate limiter with custom parameters
func CustomLimiter(max int, expiration time.Duration, message string) fiber.Handler {
	return createLimiter(RateLimitConfig{
		Max:        max,
		Expiration: expiration,
		Message:    message,
	})
}
