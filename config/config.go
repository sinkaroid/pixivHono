package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	APIKey                     string
	PixivRefreshToken          string
	PixivImgResolver           string
	CORSOrigin                 string
	RedisURL                   string
	Version                    string
	RateLimitBucketMaxSize     int
	RateLimitMax               int
	SlowDownDelayMs            int
	SlowDownWindowMs           int
	SlowDownMaxDelayMs         int
	SlowDownDelayAfter         int
	RateLimitSweepIntervalMs   int
	RateLimitWindowMs          int
	SearchCacheTTLMs           int
	PixivAccessTokenTTLMs      int
	Port                       int
	EnableUserAgentLog         bool
	EnableAccessLog            bool
	AllowQueryAPIKeyInDev      bool
	ForceHttpsPixivImgResolver bool
}

var (
	PIXIV_APP_API_BASE  = "https://app-api.pixiv.net"
	PIXIV_OAUTH_URL     = "https://oauth.secure.pixiv.net/auth/token"
	PIXIV_CLIENT_ID     = "MOBrBDS8blbauoSck0ZfDbtuzpyT"
	PIXIV_CLIENT_SECRET = "lsACyCD94FhDUtGTXi3QzcFE2uU1hqtDaKeqrdwj"
	PIXIV_HASH_SECRET   = "28c1fdd170a5204386cb1313c7077b34f83e4aaf4aa829ce78c231e05b0bae2c"
)

var GlobalConfig *Config

func Load() *Config {
	// Attempt to load .env file if it exists, matching Bun's auto-load behavior
	_ = godotenv.Load()

	if val := os.Getenv("PIXIV_APP_API_BASE"); val != "" {
		PIXIV_APP_API_BASE = val
	}
	if val := os.Getenv("PIXIV_OAUTH_URL"); val != "" {
		PIXIV_OAUTH_URL = val
	}
	if val := os.Getenv("PIXIV_CLIENT_ID"); val != "" {
		PIXIV_CLIENT_ID = val
	}
	if val := os.Getenv("PIXIV_CLIENT_SECRET"); val != "" {
		PIXIV_CLIENT_SECRET = val
	}
	if val := os.Getenv("PIXIV_HASH_SECRET"); val != "" {
		PIXIV_HASH_SECRET = val
	}

	cfg := &Config{
		APIKey:                     os.Getenv("API_KEY"),
		RateLimitBucketMaxSize:     getEnvInt("RATE_LIMIT_BUCKET_MAX_SIZE", 50000),
		PixivRefreshToken:          os.Getenv("PIXIV_REFRESH_TOKEN"),
		EnableUserAgentLog:         getEnvBool("ENABLE_USER_AGENT_LOG", false),
		RateLimitMax:               getEnvInt("RATE_LIMIT_MAX", 60),
		SlowDownDelayMs:            getEnvInt("SLOW_DOWN_DELAY_MS", 250),
		RedisURL:                   os.Getenv("REDIS_URL"),
		SearchCacheTTLMs:           getEnvInt("SEARCH_CACHE_TTL_MS", 3600000),
		CORSOrigin:                 getEnvString("CORS_ORIGIN", "*"),
		SlowDownWindowMs:           getEnvInt("SLOW_DOWN_WINDOW_MS", 60000),
		EnableAccessLog:            getEnvBool("ENABLE_ACCESS_LOG", false),
		SlowDownMaxDelayMs:         getEnvInt("SLOW_DOWN_MAX_DELAY_MS", 2000),
		RateLimitSweepIntervalMs:   getEnvInt("RATE_LIMIT_SWEEP_INTERVAL_MS", 30000),
		RateLimitWindowMs:          getEnvInt("RATE_LIMIT_WINDOW_MS", 60000),
		SlowDownDelayAfter:         getEnvInt("SLOW_DOWN_DELAY_AFTER", 20),
		AllowQueryAPIKeyInDev:      getEnvBool("ALLOW_QUERY_API_KEY_IN_DEV", false),
		ForceHttpsPixivImgResolver: getEnvBool("FORCE_HTTPS_PIXIV_IMG_RESOLVER", false),
		Port:                       getEnvInt("PORT", 3000),
		PixivImgResolver:           os.Getenv("PIXIV_IMG_RESOLVER"),
	}

	// Cache TTL checks PIXIV_REFRESHED_ACCESS_TOKEN_CACHE_TTL_MS first, then PIXIV_ACCESS_TOKEN_TTL_MS, then defaults to 3,000,000
	tokenTTL := getEnvInt("PIXIV_REFRESHED_ACCESS_TOKEN_CACHE_TTL_MS", -1)
	if tokenTTL == -1 {
		tokenTTL = getEnvInt("PIXIV_ACCESS_TOKEN_TTL_MS", 3000000)
	}
	cfg.PixivAccessTokenTTLMs = tokenTTL

	GlobalConfig = cfg
	return cfg
}

func getEnvString(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		v := strings.ToLower(val)
		return v == "true" || v == "1" || v == "yes" || v == "on"
	}
	return fallback
}
