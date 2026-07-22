package middleware

import (
	"math"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"pixivhono/config"
)

type Bucket struct {
	Count   int
	ResetAt int64 // unix timestamp in milliseconds
}

var (
	bucketsMu sync.Mutex
	buckets   = make(map[string]*Bucket)
)

func getClientKey(c *fiber.Ctx) string {
	forwarded := c.Get("x-forwarded-for")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	realIp := c.Get("x-real-ip")
	if realIp != "" {
		return realIp
	}
	return "unknown"
}

func touchBucket(key string, windowMs int64) *Bucket {
	bucketsMu.Lock()
	defer bucketsMu.Unlock()

	now := time.Now().UnixNano() / int64(time.Millisecond)
	current, exists := buckets[key]

	if !exists || current.ResetAt <= now {
		fresh := &Bucket{Count: 1, ResetAt: now + windowMs}
		buckets[key] = fresh
		return fresh
	}

	current.Count += 1
	// Return a copy to avoid race conditions
	return &Bucket{Count: current.Count, ResetAt: current.ResetAt}
}

func sweepExpiredBuckets() {
	bucketsMu.Lock()
	defer bucketsMu.Unlock()

	now := time.Now().UnixNano() / int64(time.Millisecond)
	for k, bucket := range buckets {
		if bucket.ResetAt <= now {
			delete(buckets, k)
		}
	}
}

func StartSweepTimer() {
	interval := config.GlobalConfig.RateLimitSweepIntervalMs
	if interval <= 0 {
		interval = 30000
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	go func() {
		for range ticker.C {
			sweepExpiredBuckets()
		}
	}()
}

func CORSMiddleware() fiber.Handler {
	origin := config.GlobalConfig.CORSOrigin
	if origin == "" {
		origin = "*"
	}
	return cors.New(cors.Config{
		AllowOrigins: origin,
		AllowMethods: "GET,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	})
}

func SlowDownMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == "OPTIONS" {
			return c.Next()
		}

		bucketsMu.Lock()
		size := len(buckets)
		bucketsMu.Unlock()
		if size > config.GlobalConfig.RateLimitBucketMaxSize {
			sweepExpiredBuckets()
		}

		key := "slow:" + getClientKey(c) + ":" + c.Path()
		bucket := touchBucket(key, int64(config.GlobalConfig.SlowDownWindowMs))

		if bucket.Count > config.GlobalConfig.SlowDownDelayAfter {
			steps := bucket.Count - config.GlobalConfig.SlowDownDelayAfter
			waitMs := int64(steps) * int64(config.GlobalConfig.SlowDownDelayMs)
			maxDelayMs := int64(config.GlobalConfig.SlowDownMaxDelayMs)
			if waitMs > maxDelayMs {
				waitMs = maxDelayMs
			}
			time.Sleep(time.Duration(waitMs) * time.Millisecond)
		}

		return c.Next()
	}
}

func RateLimitMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == "OPTIONS" {
			return c.Next()
		}

		bucketsMu.Lock()
		size := len(buckets)
		bucketsMu.Unlock()
		if size > config.GlobalConfig.RateLimitBucketMaxSize {
			sweepExpiredBuckets()
		}

		key := "limit:" + getClientKey(c) + ":" + c.Path()
		bucket := touchBucket(key, int64(config.GlobalConfig.RateLimitWindowMs))

		limit := config.GlobalConfig.RateLimitMax
		remaining := limit - bucket.Count
		if remaining < 0 {
			remaining = 0
		}

		c.Set("X-RateLimit-Limit", strconvItoa(limit))
		c.Set("X-RateLimit-Remaining", strconvItoa(remaining))
		c.Set("X-RateLimit-Reset", strconvItoa64(bucket.ResetAt/1000))

		if bucket.Count > limit {
			nowMs := time.Now().UnixNano() / int64(time.Millisecond)
			diffSec := float64(bucket.ResetAt-nowMs) / 1000.0
			retryAfterSec := int(math.Ceil(diffSec))
			if retryAfterSec < 1 {
				retryAfterSec = 1
			}

			c.Set("Retry-After", strconvItoa(retryAfterSec))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests, please try again later.",
			})
		}

		return c.Next()
	}
}

// custom integer conversions to avoid external dependencies
func strconvItoa(val int) string {
	if val == 0 {
		return "0"
	}
	var buf [32]byte
	i := len(buf) - 1
	neg := val < 0
	if neg {
		val = -val
	}
	for val > 0 {
		buf[i] = byte('0' + (val % 10))
		i--
		val /= 10
	}
	if neg {
		buf[i] = '-'
		i--
	}
	return string(buf[i+1:])
}

func strconvItoa64(val int64) string {
	if val == 0 {
		return "0"
	}
	var buf [32]byte
	i := len(buf) - 1
	neg := val < 0
	if neg {
		val = -val
	}
	for val > 0 {
		buf[i] = byte('0' + (val % 10))
		i--
		val /= 10
	}
	if neg {
		buf[i] = '-'
		i--
	}
	return string(buf[i+1:])
}
