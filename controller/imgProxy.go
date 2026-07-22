package controller

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func isAllowedPixivImageUrl(rawUrl string) bool {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return false
	}
	return u.Scheme == "https" && strings.HasSuffix(u.Hostname(), "pximg.net")
}

func ImgProxyController(c *fiber.Ctx) error {
	imageUrl := c.Query("url")

	if imageUrl == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing query param: url",
		})
	}

	if !isAllowedPixivImageUrl(imageUrl) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid or disallowed image url",
		})
	}

	req, err := http.NewRequest("GET", imageUrl, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to resolve image",
			"detail": err.Error(),
		})
	}

	req.Header.Set("Referer", "https://www.pixiv.net/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")

	// Upstream fetch timeout
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to resolve image",
			"detail": err.Error(),
		})
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = resp.Body.Close()
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error":  "Failed to resolve image",
			"status": resp.StatusCode,
		})
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	cacheControl := resp.Header.Get("Cache-Control")
	if cacheControl == "" {
		cacheControl = "public, max-age=3600"
	}

	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", cacheControl)

	// Fiber's SendStream will read and automatically close the body if it implements io.Closer
	return c.SendStream(resp.Body, int(resp.ContentLength))
}
