package main

import (
	"fmt"
	"strconv"

	"pixivhono/app"
	"pixivhono/cache"
	"pixivhono/config"
	"pixivhono/middleware"
	"pixivhono/utils"
)

var Version = "1.2.0-alpha"

func main() {
	cfg := config.Load()
	cfg.Version = Version

	cache.Init(cfg.RedisURL)
	utils.StartSystemMetrics()
	middleware.StartSweepTimer()

	fiberApp := app.SetupApp(cfg)

	port := strconv.Itoa(cfg.Port)
	fmt.Printf("Server running on http://localhost:%s\n", port)
	if err := fiberApp.Listen(":" + port); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
