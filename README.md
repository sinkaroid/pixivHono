## pixivHono

Rename `.env.schema` to `.env` and fill the value with your own

```bash
# GraphQL API (enable: set to true)
PIXIV_GRAPHQL=true

## Pixiv refresh token used to authenticate Pixiv API requests.
PIXIV_REFRESH_TOKEN=your_pixiv_refresh_token

## API key required by protected endpoints.
API_KEY=ScathachGrip/bot

## If true, the API key can also be provided via `?api_key=` query parameter.
## Intended for local development convenience.
ALLOW_QUERY_API_KEY_IN_DEV=false

## HTTP server port.
PORT=3000

## Pixiv API and OAuth endpoints.
PIXIV_APP_API_BASE=https://app-api.pixiv.net
PIXIV_OAUTH_URL=https://oauth.secure.pixiv.net/auth/token

## Optional Pixiv OAuth credentials and hash secret override.
PIXIV_CLIENT_ID=
PIXIV_CLIENT_SECRET=
PIXIV_HASH_SECRET=

## Redis connection URL for caching.
REDIS_URL=redis://localhost:6379

## Cache TTL for search responses in milliseconds.
SEARCH_CACHE_TTL_MS=60000

## CORS allowed origin (`*` allows all origins).
CORS_ORIGIN=*

## Rate limit settings.
RATE_LIMIT_WINDOW_MS=60000
RATE_LIMIT_MAX=60
RATE_LIMIT_BUCKET_MAX_SIZE=50000
RATE_LIMIT_SWEEP_INTERVAL_MS=30000

## Request slowdown settings.
SLOW_DOWN_WINDOW_MS=60000
SLOW_DOWN_DELAY_AFTER=20
SLOW_DOWN_DELAY_MS=250
SLOW_DOWN_MAX_DELAY_MS=2000

## Cache TTL (ms) for access token generated from PIXIV_REFRESH_TOKEN.
PIXIV_REFRESHED_ACCESS_TOKEN_CACHE_TTL_MS=3000000

## Optional logging controls (set true only when debugging).
ENABLE_ACCESS_LOG=false
ENABLE_USER_AGENT_LOG=false

## Optional external domain used to resolve Pixiv image URLs.
PIXIV_IMG_RESOLVER=SOME_RESOLVER_IF_ALREADY
FORCE_HTTPS_PIXIV_IMG_RESOLVER=true

```

### Docker

    docker pull ghcr.io/sinkaroid/pixivhono:latest
    docker run -p 3000:3000 -d ghcr.io/sinkaroid/pixivhono:latest

### Docker (adjust your own)

```bash
docker run -d \
  --name=pixivhono \
  -p 3047:3000 \
  -e PIXIV_GRAPHQL="true" \
  -e API_KEY="SOME_CREDS" \
  -e ALLOW_QUERY_API_KEY_IN_DEV="true" \
  -e PIXIV_REFRESH_TOKEN="SOME_CREDS" \
  -e REDIS_URL="redis://default:somenicepassword@redis-666.c10.us-east-6-6.ec666.cloud.redislabs.com:1337" \
  -e SEARCH_CACHE_TTL_MS="7200000" \
  -e CORS_ORIGIN="*" \
  -e RATE_LIMIT_WINDOW_MS="60000" \
  -e RATE_LIMIT_MAX="60" \
  -e RATE_LIMIT_BUCKET_MAX_SIZE="50000" \
  -e RATE_LIMIT_SWEEP_INTERVAL_MS="30000" \
  -e SLOW_DOWN_WINDOW_MS="60000" \
  -e SLOW_DOWN_DELAY_AFTER="20" \
  -e SLOW_DOWN_DELAY_MS="250" \
  -e SLOW_DOWN_MAX_DELAY_MS="2000" \
  -e ENABLE_ACCESS_LOG="true" \
  -e ENABLE_USER_AGENT_LOG="true" \
  -e FORCE_HTTPS_PIXIV_IMG_RESOLVER="true" \
  ghcr.io/sinkaroid/pixivhono:latest
```

### Manual

```bash
# 1. Clone & enter
git clone https://github.com/sinkaroid/pixivhono.git
cd pixivhono

# 2. Configure
cp .env.schema .env
# edit .env — set PIXIV_REFRESH_TOKEN, API_KEY, etc.

# 3. Run (dev with hot-reload)
task dev
# or without task: air
# or plain: go run main.go

# 4. Build & run (production)
task build
task start
# or manually:
#   go build -o build/server main.go
#   ./build/server

# 5. Print OpenAPI spec
go run . -spec
```
