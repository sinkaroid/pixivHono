package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"pixivhono/config"
)

func APIKeyMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		expectedApiKey := config.GlobalConfig.APIKey
		if expectedApiKey == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Server misconfigured: missing API_KEY",
			})
		}

		authHeader := c.Get("Authorization")
		if authHeader != "" {
			// Extract token from "Bearer <token>"
			const prefix = "Bearer "
			if len(authHeader) > len(prefix) && strings.EqualFold(authHeader[:len(prefix)], prefix) {
				token := authHeader[len(prefix):]
				if token == expectedApiKey {
					return c.Next()
				}
			}
			// Hono bearerAuth middleware returns plain text "Unauthorized" on failed verify
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
		}

		headerApiKey := c.Get("x-api-key")
		var queryApiKey string
		if config.GlobalConfig.AllowQueryAPIKeyInDev {
			queryApiKey = c.Query("api_key")
		}

		providedApiKey := headerApiKey
		if providedApiKey == "" {
			providedApiKey = queryApiKey
		}

		if providedApiKey == "" || providedApiKey != expectedApiKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		return c.Next()
	}
}
