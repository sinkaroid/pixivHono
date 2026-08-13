package main

import (
	"flag"
	"fmt"
	"strconv"

	"pixivhono/app"
	"pixivhono/cache"
	"pixivhono/config"
	"pixivhono/lib"
	"pixivhono/middleware"
	"pixivhono/utils"
)

var Version = "1.3.3-alpha"

func main() {
	spec := flag.Bool("spec", false, "print OpenAPI spec JSON and exit")
	flag.Parse()
	if *spec {
		fmt.Println(lib.OpenAPISpecJSON)
		return
	}

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
