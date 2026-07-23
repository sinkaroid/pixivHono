<div align="center">
<a href="https://sinkaroid.github.io/pixivHono/"><img width="500" src="resources/project/images/pixivhono.png" alt="pixivhono"></a>

<h4 align="center">Unified REST + GraphQL gateway for Pixiv API with image resolver</h4>
<p align="center">
	<a href="https://github.com/sinkaroid/pixivHono/actions/workflows/playground.yml"><img src="https://github.com/sinkaroid/pixivHono/workflows/Playground/badge.svg"></a>
	<a href="https://qlty.sh/gh/sinkaroid/projects/pixivHono"><img src="https://qlty.sh/gh/sinkaroid/projects/pixivHono/maintainability.png" alt="Maintainability" /></a>
</p>

PixivHono was originally named **Pixiv + Hono** (legacy name) and now runs on **Go + Fiber** with near-zero GC overhead, sub-10 MB baseline memory, and native concurrency. The motivation behind this project is: one gateway, one config, no more building proxies, wrestling 4XX responses, or reinventing OAuth refresh.

<a href="https://sinkaroid.github.io/pixivhono">Playground</a> •
<a href="https://github.com/sinkaroid/pixivhono/blob/master/CONTRIBUTING.md">Contributing</a> •
<a href="https://github.com/sinkaroid/pixivhono/issues/new/choose">Report Issues</a>

</div>

---

<a href="https://sinkaroid.github.io/pixivHono/"><img align="right" src="resources/project/images/pixivhono-docs.png" width="300"></a>

- [pixivHono](#)
  - [The problems](#the-problems)
  - [The solutions](#the-solutions)
  - [Prerequisites](#prerequisites)
    - [Installation](#installation)
      - [Docker](#docker)
      - [Manual](#manual)
  - [Features](#features)
    - [Rest](#rest)
    - [GraphQL](#graphql)
  - [Playground](#playground)
    - [Routing](#routing)
    - [Status response](#status-response)
  - [Running tests](#running-tests)
  - [Credentials](#credentials)
    - [Refresh token](#pixiv-refresh-token)
    - [OAuth Credentials & Fallback](#pixiv-oauth-credentials--fallback)
  - [Pronunciation](#pronunciation)
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

## Prerequisites

> **Redis is required** for caching to function. Without a Redis connection, all requests will fall through to upstream on every call — no caching, no deduplication, and full upstream rate limit exposure. For self-hosted production use, a running Redis instance is expected.

- Go 1.24+
- Redis
  - If just small usage or experimenting, You can get [redis.io/try-free](https://redis.io/try-free/) for demo and free tier available.

## Installation

Rename `.env.schema` to `.env` and fill the value with your own

```bash
# GraphQL API (enable: set to true)
PIXIV_GRAPHQL=true

## Pixiv refresh token used to authenticate Pixiv API requests.
PIXIV_REFRESH_TOKEN=your_pixiv_refresh_token

## API key required by protected endpoints.
API_KEY=your_secret_key

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

# 3. Download dependencies
go mod tidy

# 4. Run (dev with hot-reload)
task dev
# or without task: air
# or plain: go run main.go

# 5. Build & run (production)
task build
task start
# or manually:
#   go build -o build/server main.go
#   ./build/server

# 6. Print OpenAPI spec
go run . -spec
```

---

## Features

The current feature set is intentionally minimal.

## Rest

Enabled by default

### `GET /pixiv/search`

https://sinkaroid.github.io/pixivHono/#GET/pixiv/search  
Proxy search to Pixiv's internal illust search API. Requires `PIXIV_REFRESH_TOKEN`.

**Query parameters:**

| Param   | Type   | Required | Default | Description                     |
| ------- | ------ | -------- | ------- | ------------------------------- |
| `query` | string | yes      | —       | Search term (tag, title, etc.)  |
| `page`  | int    | no       | `1`     | Page number (30 items per page) |

**What it does:**

1. Validates input — `query` must be non-empty, `page` must be positive integer.
2. Acquires an access token from the configured `PIXIV_REFRESH_TOKEN` (cached internally).
3. Calls Pixiv's `/v1/search/illust` with `search_target=partial_match_for_tags`, `sort=date_desc`, `filter=for_ios`.
4. Enriches every illust in the response with `_resolved` image URLs (see img_resolver below).
5. Caches the full response in Redis using `SEARCH_CACHE_TTL_MS` (default 1h).
6. Returns cached results on subsequent identical requests — zero upstream hits.

**Response:** Paginated illust list with enriched URL metadata.

**Auth:** Requires `Authorization: Bearer <API_KEY>` header.

---

### `GET /pixiv/artworks`

https://sinkaroid.github.io/pixivHono/#GET/pixiv/artworks  
Fetch a single illust detail by ID. Requires `PIXIV_REFRESH_TOKEN`.

**Query parameters:**

| Param | Type | Required | Default | Description     |
| ----- | ---- | -------- | ------- | --------------- |
| `id`  | int  | yes      | —       | Pixiv illust ID |

**What it does:**

1. Validates `id` — must be a positive integer.
2. Acquires an access token from `PIXIV_REFRESH_TOKEN`.
3. Calls Pixiv's `/v1/illust/detail` with the given `illust_id`.
4. Enriches the illust with `_resolved` image URLs and `original_image_url_resolved`.

**Response:** Single illust object with full metadata (tags, image URLs, stats, etc.).

**Auth:** Requires `Authorization: Bearer <API_KEY>` header.

---

### `GET /pixiv/img_resolver`

https://sinkaroid.github.io/pixivHono/#GET/pixiv/img_resolver  
Proxy `i.pximg.net` images through the gateway, bypassing Pixiv's hotlink protection.

**Query parameters:**

| Param | Type   | Required | Default | Description                                 |
| ----- | ------ | -------- | ------- | ------------------------------------------- |
| `url` | string | yes      | —       | Full `https://i.pximg.net/...` URL to proxy |

**What it does:**

1. Validates the URL is `https://*.pximg.net` — rejects anything else.
2. Forwards the request with `Referer: https://www.pixiv.net/` and a browser `User-Agent`.
3. Streams the image response back with original `Content-Type` and a sensible `Cache-Control: public, max-age=3600`.
4. 30-second upstream timeout. Non-2xx upstream status returns `502 Bad Gateway`.

**Why this exists:** Pixiv's CDN checks the `Referer` header — requests from any origin other than `pixiv.net` return `403 Forbidden`. This endpoint sets the correct `Referer` so images render in your app.

**Auth:** ⚠️ **No API key required.** This endpoint is publicly accessible (CORS-enabled, no auth). It is intentionally excluded from API key and rate-limit middleware to allow direct `<img>` tag embedding from browser clients.

**Alternative — skip the proxy entirely:** If you have your own reverse proxy domain, set `PIXIV_IMG_RESOLVER` in `.env` and image URLs are rewritten server-side. The `_resolved` fields in search/artwork responses will point to your resolver instead of proxying through pixivHono.

## GraphQL

Enable: `PIXIV_GRAPHQL=true` in `.env`.

When enabled, pixivHono exposes a GraphQL API at `/api/graphql` (auth-required) with a built-in GraphiQL-style **Playground** at `/graphql` (no auth, local dev only).

### Schema

Two root queries, matching the REST endpoints:

| Query    | Args                                        | Returns    | Description                        |
| -------- | ------------------------------------------- | ---------- | ---------------------------------- |
| `illust` | `id: Int!`                                  | `Illust`   | Single illust by ID                |
| `search` | `query: String!`, `page: Int`, `limit: Int` | `[Illust]` | Paginated search (max 100 results) |

### `Illust` type (selected fields)

```
id, title, caption, type, image_urls { square_medium, medium, large, original, _resolved },
width, height, page_count, sanity_level, x_restrict,
total_view, total_bookmarks, create_date,
tags { name, translated_name }, user
```

The `image_urls._resolved` field contains the proxy-rewritten URLs (set via `PIXIV_IMG_RESOLVER` or auto-routed through `img_resolver`).

### Examples

**Fetch illust metadata:**

```bash
curl -X POST http://localhost:3000/api/graphql \
  -H "Authorization: Bearer your_secret_key" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ illust(id: 147501814) { id title image_urls { medium } } }"}'
```

**Search illusts:**

```bash
curl -X POST http://localhost:3000/api/graphql \
  -H "Authorization: Bearer your_secret_key" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ search(query: \"yuri\", page: 1, limit: 5) { id title } }"}'
```

**Get resolved image URLs:**

```bash
curl -X POST http://localhost:3000/api/graphql \
  -H "Authorization: Bearer your_secret_key" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ illust(id: 147501814) { image_urls { medium _resolved } } }"}'
```

**Playground (browser):** Open `http://localhost:3000/graphql` to explore the schema interactively — no auth required for local access.

## Playground

https://sinkaroid.github.io/pixivhono

### Routing

| Method | Path                  | Auth Required | Description                      |
| ------ | --------------------- | ------------- | -------------------------------- |
| `GET`  | `/`                   | No            | System status (memory, server)   |
| `GET`  | `/doc`                | No            | OpenAPI v3 JSON spec             |
| `GET`  | `/playground`         | No            | Swagger UI (browser)             |
| `GET`  | `/metrics`            | No            | Prometheus metrics               |
| `GET`  | `/pixiv/search`       | Yes           | Search illusts                   |
| `GET`  | `/pixiv/artworks`     | Yes           | Illust detail by ID              |
| `GET`  | `/pixiv/img_resolver` | No            | Image proxy (public CORS)        |
| `GET`  | `/pixiv/token_health` | Yes           | Token health check               |
| `POST` | `/api/graphql`        | Yes           | GraphQL API (if enabled)         |
| `GET`  | `/graphql`            | No            | GraphiQL playground (if enabled) |

### Status response

All endpoints follow a consistent HTTP status code convention:

| Status | Code                  | When                                | Body                                                                         |
| ------ | --------------------- | ----------------------------------- | ---------------------------------------------------------------------------- |
| 200    | OK                    | Request succeeds                    | Response payload (varies by endpoint)                                        |
| 400    | Bad Request           | Missing/invalid query parameter     | `{\"error\": \"...\"}`                                                       |
| 401    | Unauthorized          | Missing or invalid `Authorization`  | `\"Unauthorized\"` (Bearer) or `{\"error\": \"Unauthorized\"}` (`x-api-key`) |
| 429    | Too Many Requests     | Rate limit exceeded                 | `{\"error\": \"Too many requests, please try again later.\"}`                |
| 500    | Internal Server Error | Missing env var or upstream failure | `{\"error\": \"...\", \"detail\": \"...\"}`                                  |
| 502    | Bad Gateway           | Image upstream returned non-2xx     | `{\"error\": \"Failed to resolve image\", \"status\": 502}`                  |

**Success responses** return the Pixiv API payload enriched with `_resolved` image URL fields (see [img_resolver](#get-pixivimg_resolver)).  
**Error responses** always contain an `"error"` key, with optional `"detail"` for upstream failure context.

## Running tests

See `task test`, `task test:prod`, `task test:graphql` in [Taskfile.yml](/Taskfile.yml).

## Credentials

Two sets of credentials are needed to run pixivHono:

- **Pixiv refresh token** — authenticates API requests to Pixiv's private endpoints
- **Pixiv OAuth client credentials** — built-in fallback (see below)

### Pixiv refresh token

Pixiv does not expose a public API registration portal. To get your `PIXIV_REFRESH_TOKEN`:

1. Reference: [eggplants/get-pixivpy-token](https://github.com/eggplants/get-pixivpy-token)
2. Paste the token into `.env`:
   ```env
   PIXIV_REFRESH_TOKEN=your_token_here
   ```

This token is used on every request to acquire a short-lived access token. It is cached in Redis and refreshed automatically.

### Pixiv OAuth Credentials & Fallback

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

## Pronunciation

PixivHono was originally named **Pixiv + Hono** (legacy name)

## Legal

This tool can be freely copied, modified, altered, distributed without any attribution whatsoever. However, if you feel
like this tool deserves an attribution, mention it. It won't hurt anybody.

> Licence: WTF.

## Microservices

Microservices and subprojects is part of a broader ecosystem of specialized services, each focused on a specific platform or content source while sharing a common design philosophy maintained by [ScathachGrip](https://github.com/ScathachGrip)

- **sinkaroid/pixivHono — Unified REST and GraphQL gateway API for Pixiv**
- [sinkaroid/matoi](https://github.com/sinkaroid/matoi) — Unified REST and GraphQL gateway for booru-based imageboards
- [sinkaroid/jandapress](https://github.com/sinkaroid/jandapress) — Unified REST and GraphQL API for nhentai and other doujinshi
- [sinkaroid/lustpress](https://github.com/sinkaroid/lustpress) — Unified REST and GraphQL API for PornHub and other R18 platforms

Each service is developed independently, enabling modular deployments, isolated maintenance, and platform-specific optimizations while remaining interoperable within the ecosystem.
