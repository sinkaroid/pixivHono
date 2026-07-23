<div align="center">
<a href="https://sinkaroid.github.io/pixivHono/"><img width="500" src="resources/project/images/pixivhono.png" alt="pixivhono"></a>

<h4 align="center">Unified REST + GraphQL gateway for Pixiv API + image resolver</h4>
<p align="center">
	<a href="https://github.com/sinkaroid/pixivHono/actions/workflows/playground.yml"><img src="https://github.com/sinkaroid/pixivHono/workflows/Playground/badge.svg"></a>
	<a href="https://qlty.sh/gh/sinkaroid/projects/pixivHono"><img src="https://qlty.sh/gh/sinkaroid/projects/pixivHono/maintainability.png" alt="Maintainability" /></a>
</p>

pixivHono was originally built with **Pixiv + Hono** + TypeScript (legacy name) and now runs on **Go + Fiber** with near-zero GC overhead, sub-10 MB baseline memory, and native concurrency.  
The motivation behind this project is: one gateway, one config, no more building Pixiv proxy, wrestling 403s, or reinventing OAuth refresh.

<a href="https://sinkaroid.github.io/pixivhono">Playground</a> •
<a href="https://github.com/sinkaroid/pixivhono/blob/master/CONTRIBUTING.md">Contributing</a> •
<a href="https://github.com/sinkaroid/pixivhono/issues/new/choose">Report Issues</a>

</div>

---

<a href="https://sinkaroid.github.io/pixivHono/"><img align="right" src="resources/project/images/pixivhono-docs.png" width="300"></a>

- [Jandapress](#)
  - [The problems](#the-problems)
  - [The solutions](#the-solutions)
  - [Features](#features)
  - [Prerequisites](#prerequisites)
    - [Installation](#installation)
      - [Docker](#docker)
      - [Manual](#manual)
    - [Tests](#tests)
    - [Nhentai Guide](#nhentai-guide)
  - [Playground](https://sinkaroid.github.io/jandapress)
    - [Routing](#playground)
    - [Status response](#status-response)
  - [Running tests](#running-tests)
    - [Tests](#tests)
  - [CLosing remarks](https://github.com/sinkaroid/jandapress/blob/master/CLOSING_REMARKS.md)
    - [Alternative links](https://github.com/sinkaroid/jandapress/blob/master/CLOSING_REMARKS.md#alternative-links)
  - [Pronunciation](#Pronunciation)
  - [Legal](#legal)
  - [Microservices](#microservices)

## The problems

Pixiv's official API has severe limitations when consumed directly from client-side applications:

- **CORS restrictions** — Pixiv's API endpoints (`app-api.pixiv.net`, `oauth.secure.pixiv.net`) do not permit browser-origin requests, making direct AJAX calls impossible from web apps.
- **Image hotlink protection** — Pixiv image hosts (`i.pximg.net`) returns `403 Forbidden` when the `Referer` header is missing or doesn't match their expected origin. This breaks all embedded image rendering outside pixiv.net.
- **OAuth complexity** — The refresh-token flow requires multiple round-trips, client-side secrets, and manual token refresh logic — none of which belong in a browser context.
- **No public search API** — Pixiv exposes no official search or discovery endpoint for third-party clients.

Pixiv-powered frontend (gallery, wallpaper app, image board) must solve all four problems themselves — every single time.

## The solutions

`pixivHono` is a **unified REST + GraphQL gateway** that sits between your application and Pixiv, solving every problem above:

| Problem             | Solution                                                                                                                                      |
| ------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| CORS & AJAX blocked | Proxy all Pixiv API calls through a CORS-enabled gateway                                                                                      |
| Image 403 errors    | Built-in `img_resolver` rewrites `i.pximg.net` URLs through a configurable proxy domain — set `PIXIV_IMG_RESOLVER` and images render anywhere |
| OAuth tedium        | Automatic token refresh behind a single `PIXIV_REFRESH_TOKEN` env var; access tokens cached and rotated transparently                         |
| No search API       | REST `/search` and GraphQL `search()` query powered by Pixiv's internal search                                                                |

Beyond the proxy layer, you get:

- **REST** + **GraphQL** (opt-in via `PIXIV_GRAPHQL=true`) — query exactly the data you need.
- **Rate limiting & slow-down** — protect upstream from abuse.
- **Redis caching** — configurable TTLs for search responses and access tokens.
- **Prometheus metrics** — track request rates, cache hits, and upstream latency.
- **OpenAPI / Swagger** — interactive playground at `/playground`.

No scraping. No reverse-engineering auth flows. No client-side secrets. One gateway, one `.env`, done.

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

---

## Pixiv OAuth Credentials & Fallback

By default, `pixivHono` includes built-in fallback OAuth client credentials obtained via reverse engineering the official Pixiv Mobile App (iOS/Android), widely used across open-source Pixiv API clients (reference: [`upbit/pixivpy`](https://github.com/upbit/pixivpy/blob/master/pixivpy3/api.py#L26-L28)):

- **Default Client ID**: `MOBrBDS8blbauoSck0ZfDbtuzpyT`
- **Default Client Secret**: `lsACyCD94FhDUtGTXi3QzcFE2uU1hqtDaKeqrdwj`
- **Default Hash Secret**: `28c1fdd170a5204386cb1313c7077b34f83e4aaf4aa829ce78c231e05b0bae2c`

### What happens if Pixiv updates their credentials?

If Pixiv rotates or revokes these official app credentials in a future update, OAuth token authentication will fail (e.g. returning `400 Bad Request` / `invalid_client`).

**You do NOT need to modify or re-compile the source code.** You can override any of these secrets directly in your `.env` file or environment variables:

```env
PIXIV_CLIENT_ID=your_new_client_id
PIXIV_CLIENT_SECRET=your_new_client_secret
PIXIV_HASH_SECRET=your_new_hash_secret
PIXIV_OAUTH_URL=https://oauth.secure.pixiv.net/auth/token
PIXIV_APP_API_BASE=https://app-api.pixiv.net
```
