package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/joho/godotenv"
	"pixivhono/app"
	"pixivhono/cache"
	"pixivhono/client"
	"pixivhono/config"
)

func init() {
	_ = godotenv.Load("../.env")
}

func TestProdArtworkEndpoint(t *testing.T) {
	client.ClearTokenCache()
	cfg := config.Load()
	if cfg.PixivRefreshToken == "" {
		t.Skip("Skipping production test: PIXIV_REFRESH_TOKEN not configured in .env")
	}

	config.GlobalConfig = cfg
	cache.Init(cfg.RedisURL)

	fiberApp := app.SetupApp(cfg)

	req := httptest.NewRequest("GET", "/pixiv/artworks?id=147501814", nil)
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	resp, err := fiberApp.Test(req, 15000)
	if err != nil {
		t.Fatalf("Production artwork request failed: %v", err)
	}

	body, _ := io.ReadAll(resp.Body)
	t.Logf("--- GET /pixiv/artworks (Prod) ---")
	t.Logf("Status Code: %d", resp.StatusCode)
	t.Logf("Response Body:\n%s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

func TestProdSearchEndpoint(t *testing.T) {
	client.ClearTokenCache()
	cfg := config.Load()
	if cfg.PixivRefreshToken == "" {
		t.Skip("Skipping production test: PIXIV_REFRESH_TOKEN not configured in .env")
	}

	config.GlobalConfig = cfg
	cache.Init(cfg.RedisURL)

	fiberApp := app.SetupApp(cfg)

	req := httptest.NewRequest("GET", "/pixiv/search?query=yuri&page=1", nil)
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	resp, err := fiberApp.Test(req, 15000)
	if err != nil {
		t.Fatalf("Production search request failed: %v", err)
	}

	body, _ := io.ReadAll(resp.Body)
	t.Logf("--- GET /pixiv/search (Prod) ---")
	t.Logf("Status Code: %d", resp.StatusCode)
	t.Logf("Response Body:\n%s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

func TestProdImageResolverEndpoint(t *testing.T) {
	cfg := config.Load()
	config.GlobalConfig = cfg
	cache.Init(cfg.RedisURL)

	fiberApp := app.SetupApp(cfg)

	targetImgUrl := "https://i.pximg.net/img-master/img/2026/07/22/17/55/25/147501814_p0_master1200.jpg"
	req := httptest.NewRequest("GET", "/pixiv/img_resolver?url="+url.QueryEscape(targetImgUrl), nil)
	resp, err := fiberApp.Test(req, 15000)
	if err != nil {
		t.Fatalf("Production image resolver request failed: %v", err)
	}

	body, _ := io.ReadAll(resp.Body)
	t.Logf("--- GET /pixiv/img_resolver (Prod) ---")
	t.Logf("Status Code: %d", resp.StatusCode)
	t.Logf("Content-Type: %s", resp.Header.Get("Content-Type"))
	t.Logf("Cache-Control: %s", resp.Header.Get("Cache-Control"))
	t.Logf("Response Body size: %d bytes", len(body))

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}
