package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"m2m-backend/config"
)

// AdminMiddleware verifica se o usuário é o administrador principal (Italo)
func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Pegar token do header Authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Token não fornecido",
			})
		}

		// Extrair token (Bearer <token>)
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Formato de token inválido",
			})
		}

		// Validar token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Método de assinatura inválido")
			}
			return []byte(config.GetJWTSecret()), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Token inválido ou expirado",
			})
		}

		// Extrair claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Unauthorized",
				"message": "Claims inválidas",
			})
		}

		// Verificar se o email é do administrador principal
		email, _ := claims["email"].(string)
		adminEmail := "italosilva14@hotmail.com"
		
		if email != adminEmail {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Forbidden",
				"message": "Acesso restrito apenas ao administrador principal.",
			})
		}

		// Adicionar user ID ao contexto
		c.Locals("userId", claims["user_id"])
		c.Locals("userEmail", email)

		return c.Next()
	}
}
