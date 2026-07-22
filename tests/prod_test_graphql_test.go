package tests

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"pixivhono/app"
	"pixivhono/cache"
	"pixivhono/client"
	"pixivhono/config"
)

func TestProdGraphQLIllust(t *testing.T) {
	client.ClearTokenCache()
	cfg := config.Load()
	if cfg.PixivRefreshToken == "" {
		t.Skip("Skipping production test: PIXIV_REFRESH_TOKEN not configured in .env")
	}
	cfg.EnableGraphQL = true

	config.GlobalConfig = cfg
	cache.Init(cfg.RedisURL)

	fiberApp := app.SetupApp(cfg)

	body := `{"query":"{ illust(id: 147501814) { id title image_urls { medium } } }"}`
	req := httptest.NewRequest("POST", "/api/graphql", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	resp, err := fiberApp.Test(req, 15000)
	if err != nil {
		t.Fatalf("GraphQL illust request failed: %v", err)
	}

	b, _ := io.ReadAll(resp.Body)
	t.Logf("--- POST /api/graphql (Illust) ---")
	t.Logf("Status Code: %d", resp.StatusCode)
	t.Logf("Response Body:\n%s\n", string(b))

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if errs, ok := data["errors"]; ok {
		t.Fatalf("GraphQL errors: %v", errs)
	}
	if _, ok := data["data"]; !ok {
		t.Fatal("Missing 'data' key in response")
	}
}

func TestProdGraphQLSearch(t *testing.T) {
	client.ClearTokenCache()
	cfg := config.Load()
	if cfg.PixivRefreshToken == "" {
		t.Skip("Skipping production test: PIXIV_REFRESH_TOKEN not configured in .env")
	}
	cfg.EnableGraphQL = true

	config.GlobalConfig = cfg
	cache.Init(cfg.RedisURL)

	fiberApp := app.SetupApp(cfg)

	body := `{"query":"{ search(query: \"yuri\", page: 1, limit: 5) { id title } }"}`
	req := httptest.NewRequest("POST", "/api/graphql", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	resp, err := fiberApp.Test(req, 15000)
	if err != nil {
		t.Fatalf("GraphQL search request failed: %v", err)
	}

	b, _ := io.ReadAll(resp.Body)
	t.Logf("--- POST /api/graphql (Search) ---")
	t.Logf("Status Code: %d", resp.StatusCode)
	t.Logf("Response Body:\n%s\n", string(b))

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if errs, ok := data["errors"]; ok {
		t.Fatalf("GraphQL errors: %v", errs)
	}
	if _, ok := data["data"]; !ok {
		t.Fatal("Missing 'data' key in response")
	}
}
