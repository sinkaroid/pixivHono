package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"pixivhono/app"
	"pixivhono/config"
)

func TestMetricsEndpoint(t *testing.T) {
	cfg := config.Load()
	fiberApp := app.SetupApp(cfg)

	// Perform a request with a standard method
	req1 := httptest.NewRequest("GET", "/", nil)
	resp1, err := fiberApp.Test(req1, 1000)
	if err != nil {
		t.Fatalf("Request 1 failed: %v", err)
	}
	_ = resp1.Body.Close()

	// Perform a request with custom method GETT
	req2 := httptest.NewRequest("GETT", "/", nil)
	resp2, err := fiberApp.Test(req2, 1000)
	if err != nil {
		t.Fatalf("Request 2 failed: %v", err)
	}
	body2, _ := io.ReadAll(resp2.Body)
	t.Logf("GETT request status: %d, body: %s", resp2.StatusCode, string(body2))
	_ = resp2.Body.Close()

	// Scrape metrics
	reqMetrics := httptest.NewRequest("GET", "/metrics", nil)
	respMetrics, err := fiberApp.Test(reqMetrics, 1000)
	if err != nil {
		t.Fatalf("Metrics request failed: %v", err)
	}
	defer func() { _ = respMetrics.Body.Close() }()

	body, err := io.ReadAll(respMetrics.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	t.Logf("Metrics Status Code: %d", respMetrics.StatusCode)
	t.Logf("Metrics Body:\n%s", string(body))

	if respMetrics.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", respMetrics.StatusCode)
	}
}
