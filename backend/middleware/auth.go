package middleware

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"m2m-backend/config"
)

// logAuthEvent logs authentication-related events for security auditing
func logAuthEvent(eventType, ip, path, reason string) {
	log.Printf("[AUTH] %s | ip=%s | path=%s | reason=%s", eventType, ip, path, reason)
}

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		path := c.Path()

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logAuthEvent("UNAUTHORIZED", ip, path, "missing_token")
			return c.Status(401).JSON(fiber.Map{"error": "Missing authorization token"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logAuthEvent("UNAUTHORIZED", ip, path, "invalid_format")
			return c.Status(401).JSON(fiber.Map{"error": "Invalid token format"})
		}

		tokenString := parts[1]

		// Use the secure JWT secret from config
		secret := config.GetJWTSecret()

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			logAuthEvent("UNAUTHORIZED", ip, path, "invalid_or_expired_token")
			return c.Status(401).JSON(fiber.Map{"error": "Invalid or expired token"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logAuthEvent("UNAUTHORIZED", ip, path, "invalid_claims")
			return c.Status(401).JSON(fiber.Map{"error": "Invalid token claims"})
		}

		// Store user info in context for use in handlers and rate limiting
		userId := claims["userId"]
		email := claims["email"]
		c.Locals("userId", userId)
		c.Locals("email", email)

		return c.Next()
	}
}
