package middleware

import (
	"call-center-api/pkg/config"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	AgentID string `json:"agent_id"`
	jwt.RegisteredClaims
}

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Authorization header missing",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Invalid authorization header",
			})
		}

		cfg := config.Load()
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Invalid token",
				"error":   err.Error(),
			})
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			c.Locals("agent_id", claims.AgentID)
			return c.Next()
		}

		return c.Status(401).JSON(fiber.Map{
			"success": false,
			"message": "Invalid token",
		})
	}
}
