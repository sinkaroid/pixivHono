package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"pixivhono/config"
	"pixivhono/controller"
	"pixivhono/lib"
	"pixivhono/middleware"
	"pixivhono/utils"
)

const (
	docPath        = "/doc"
	playgroundPath = "/playground"
)

var (
	locationCacheMu sync.Mutex
	cachedLocation  = "Unknown"
	lastLocTime     time.Time
)

func formatMB(val float64) string {
	return fmt.Sprintf("%.2f MB", val)
}

func getProcessMemory() (string, string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	rssBytes, err := utils.GetRSSBytes()
	if err != nil {
		rssBytes = m.Sys
	}

	rssMB := float64(rssBytes) / 1024 / 1024
	heapUsedMB := float64(m.HeapAlloc) / 1024 / 1024
	heapTotalMB := float64(m.HeapSys) / 1024 / 1024

	return formatMB(rssMB), fmt.Sprintf("%.2f/%.2f MB", heapUsedMB, heapTotalMB)
}

func getServerLocation() string {
	locationCacheMu.Lock()
	defer locationCacheMu.Unlock()

	if time.Since(lastLocTime) < 30*time.Minute && cachedLocation != "Unknown" {
		return cachedLocation
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://ipwho.is/")
	if err != nil {
		return cachedLocation
	}
	defer func() { _ = resp.Body.Close() }()

	var data struct {
		Country string `json:"country"`
		Region  string `json:"region"`
		Success bool   `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || !data.Success {
		return cachedLocation
	}

	country := strings.TrimSpace(data.Country)
	region := strings.TrimSpace(data.Region)
	if country != "" && region != "" {
		cachedLocation = country + ", " + region
		lastLocTime = time.Now()
	}

	return cachedLocation
}

func SetupApp(cfg *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// ── Access Logging ──────────────────────────────────
	if cfg.EnableAccessLog {
		app.Use(logger.New())
	}
	if cfg.EnableUserAgentLog {
		app.Use(func(c *fiber.Ctx) error {
			ua := c.Get("User-Agent")
			if ua == "" {
				ua = "unknown"
			}
			fmt.Printf("[UA] %s %s :: %s\n", c.Method(), c.Path(), ua)
			return c.Next()
		})
	}

	// ── Inflight requests tracker ───────────────────────
	app.Use(utils.InflightMiddleware())

	// ── CORS ────────────────────────────────────────────
	app.Use(middleware.CORSMiddleware())

	// ── API Key validation ──────────────────────────────
	app.Use(func(c *fiber.Ctx) error {
		path := c.Path()
		if path == "/" || path == docPath || path == playgroundPath || path == "/pixiv/img_resolver" || path == "/metrics" {
			return c.Next()
		}
		return middleware.APIKeyMiddleware()(c)
	})

	// ── Traffic Control ─────────────────────────────────
	app.Use(func(c *fiber.Ctx) error {
		path := c.Path()
		if path == "/" || path == docPath || path == playgroundPath {
			return c.Next()
		}
		return middleware.SlowDownMiddleware()(c)
	})

	app.Use(func(c *fiber.Ctx) error {
		path := c.Path()
		if path == "/" || path == docPath || path == playgroundPath {
			return c.Next()
		}
		return middleware.RateLimitMiddleware()(c)
	})

	// ── Metrics Endpoint ────────────────────────────────
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.HandlerFor(utils.Registry, promhttp.HandlerOpts{})))

	// ── System Status Endpoint ──────────────────────────
	app.Get("/", func(c *fiber.Ctx) error {
		rssStr, heapStr := getProcessMemory()
		serverLoc := getServerLocation()

		return c.JSON(fiber.Map{
			"success":  true,
			"message":  "Hi, I'm alive!",
			"endpoint": "https://sinkaroid.github.io/pixivHono",
			"date":     time.Now().Format("1/2/2006, 3:04:05 PM"),
			"rss":      rssStr,
			"heap":     heapStr,
			"server":   serverLoc,
			"version":  cfg.Version,
		})
	})

	// ── OpenAPI & Playground ────────────────────────────
	app.Get(docPath, func(c *fiber.Ctx) error {
		c.Type("json")
		return c.SendString(lib.OpenAPISpecJSON)
	})

	app.Get("/playground", func(c *fiber.Ctx) error {
		c.Type("html")
		return c.SendString(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@5/favicon-32x32.png" sizes="32x32" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" charset="UTF-8"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js" charset="UTF-8"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: '/doc',
        dom_id: '#swagger-ui',
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>`)
	})

	// ── API Routes ──────────────────────────────────────
	app.Get("/pixiv/search", controller.SearchController)
	app.Get("/pixiv/artworks", controller.ArtworkController)
	app.Get("/pixiv/img_resolver", controller.ImgProxyController)
	app.Get("/pixiv/token_health", controller.TokenHealthController)

	// ── GraphQL ─────────────────────────────────────────
	if cfg.EnableGraphQL {
		gql := controller.NewGraphQLHandler(cfg)
		app.Post("/api/graphql", gql.Handle)
		app.Get("/graphql", gql.Playground)
	}

	return app
}
