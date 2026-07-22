package controller

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"pixivhono/cache"
	"pixivhono/client"
	"pixivhono/config"
	"pixivhono/lib"
)

func SearchController(c *fiber.Ctx) error {
	q := c.Query("query")
	pageRaw := c.Query("page", "1")

	if q == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing query param: query",
		})
	}

	page, err := strconv.Atoi(pageRaw)
	if err != nil || page <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid query param: page (must be positive integer)",
		})
	}

	refreshToken := config.GlobalConfig.PixivRefreshToken
	if refreshToken == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Missing env: PIXIV_REFRESH_TOKEN",
		})
	}

	cacheKey := cache.GetSearchCacheKey(q, page)
	ctx := context.Background()

	// 1. Try to read from cache
	if cachedBytes, err := cache.GlobalCache.Get(ctx, cacheKey); err == nil {
		c.Type("json")
		return c.Send(cachedBytes)
	}

	// 2. Fetch from Pixiv API
	accessToken, err := client.GetPixivAccessToken(refreshToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to search artworks",
			"detail": err.Error(),
		})
	}

	const perPage = 30
	offset := strconv.Itoa((page - 1) * perPage)

	params := map[string]string{
		"word":                           q,
		"search_target":                  "partial_match_for_tags",
		"sort":                           "date_desc",
		"filter":                         "for_ios",
		"merge_plain_keyword_results":    "true",
		"include_translated_tag_results": "true",
		"offset":                         offset,
	}

	resp, err := client.PixivGet("/v1/search/illust", params, accessToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to search artworks",
			"detail": err.Error(),
		})
	}

	// 3. Enrich response
	baseUrl := c.BaseURL()
	enriched := lib.EnrichSearchResponseWithResolvedUrls(resp, baseUrl)

	// 4. Marshall & Cache result
	enrichedBytes, err := json.Marshal(enriched)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to search artworks",
			"detail": err.Error(),
		})
	}

	ttlMs := config.GlobalConfig.SearchCacheTTLMs
	if ttlMs <= 0 {
		ttlMs = 3600000
	}
	_ = cache.GlobalCache.Set(ctx, cacheKey, enrichedBytes, time.Duration(ttlMs)*time.Millisecond)

	c.Type("json")
	return c.Send(enrichedBytes)
}
