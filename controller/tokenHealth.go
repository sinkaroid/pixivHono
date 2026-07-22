package controller

import (
	"github.com/gofiber/fiber/v2"
	"pixivhono/client"
	"pixivhono/config"
)

func TokenHealthController(c *fiber.Ctx) error {
	refreshToken := config.GlobalConfig.PixivRefreshToken
	if refreshToken == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":    false,
			"valid": false,
			"error": "Missing env: PIXIV_REFRESH_TOKEN",
		})
	}

	_, err := client.RefreshPixivAccessToken(refreshToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":     false,
			"valid":  false,
			"error":  "Refresh token is invalid or expired.",
			"detail": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"ok":      true,
		"valid":   true,
		"message": "Refresh token is valid.",
	})
}
