package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"pixivhono/client"
	"pixivhono/config"
	"pixivhono/lib"
)

func ArtworkController(c *fiber.Ctx) error {
	idRaw := c.Query("id")
	illustId, err := strconv.Atoi(idRaw)

	if err != nil || illustId <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid artwork id",
		})
	}

	refreshToken := config.GlobalConfig.PixivRefreshToken
	if refreshToken == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Missing env: PIXIV_REFRESH_TOKEN",
		})
	}

	accessToken, err := client.GetPixivAccessToken(refreshToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to get artwork",
			"detail": err.Error(),
		})
	}

	params := map[string]string{
		"illust_id": strconv.Itoa(illustId),
	}

	resp, err := client.PixivGet("/v1/illust/detail", params, accessToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to get artwork",
			"detail": err.Error(),
		})
	}

	baseUrl := c.BaseURL()
	enriched := lib.EnrichArtworkResponseWithResolvedUrls(resp, baseUrl)

	return c.JSON(enriched)
}
