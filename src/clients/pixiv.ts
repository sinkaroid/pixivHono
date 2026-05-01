import {
  PIXIV_APP_API_BASE,
  PIXIV_CLIENT_ID,
  PIXIV_CLIENT_SECRET,
  PIXIV_HASH_SECRET,
  PIXIV_OAUTH_URL,
} from "../config/pixiv";
import type { PixivToken } from "../types/pixiv";

const ACCESS_TOKEN_TTL_MS = Number(
  process.env.PIXIV_REFRESHED_ACCESS_TOKEN_CACHE_TTL_MS
    ?? process.env.PIXIV_ACCESS_TOKEN_TTL_MS
    ?? 3_000_000,
);
let cachedAccessToken: string | null = null;
let cachedAccessTokenExpiresAt = 0;
let inflightTokenRefresh: Promise<string> | null = null;

async function makeClientHash(clientTime: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(clientTime + PIXIV_HASH_SECRET);
  const hash = await crypto.subtle.digest("SHA-256", data);
  return Array.from(new Uint8Array(hash))
    .map(b => b.toString(16).padStart(2, "0"))
    .join("");
}

export async function refreshPixivAccessToken(refreshToken: string): Promise<PixivToken> {
  const clientTime = new Date().toISOString().replace("Z", "+00:00");
  const clientHash = await makeClientHash(clientTime);
  const form = new URLSearchParams({
    client_id: PIXIV_CLIENT_ID,
    client_secret: PIXIV_CLIENT_SECRET,
    get_secure_url: "1",
    grant_type: "refresh_token",
    refresh_token: refreshToken,
  });

  const response = await fetch(PIXIV_OAUTH_URL, {
    method: "POST",
    headers: {
      "content-type": "application/x-www-form-urlencoded",
      "x-client-time": clientTime,
      "x-client-hash": clientHash,
      "app-os": "ios",
      "app-os-version": "16.4.1",
      "user-agent": "PixivIOSApp/7.16.9 (iOS 16.4.1; iPad13,4)",
    },
    body: form.toString(),
  });

  const data = (await response.json()) as {
    response?: { access_token?: string }
  };

  if (!response.ok || !data.response?.access_token) {
    throw new Error(`Failed to refresh token (status ${response.status})`);
  }

  return { accessToken: data.response.access_token };
}

export async function getPixivAccessToken(refreshToken: string): Promise<string> {
  const now = Date.now();
  if (cachedAccessToken && now < cachedAccessTokenExpiresAt) {
    return cachedAccessToken;
  }

  if (inflightTokenRefresh) {
    return inflightTokenRefresh;
  }

  inflightTokenRefresh = (async () => {
    const { accessToken } = await refreshPixivAccessToken(refreshToken);
    const ttlMs = Number.isFinite(ACCESS_TOKEN_TTL_MS) && ACCESS_TOKEN_TTL_MS > 0
      ? ACCESS_TOKEN_TTL_MS
      : 3_000_000;
    cachedAccessToken = accessToken;
    cachedAccessTokenExpiresAt = Date.now() + ttlMs;
    return accessToken;
  })();

  try {
    return await inflightTokenRefresh;
  } finally {
    inflightTokenRefresh = null;
  }
}

export async function pixivGet(
  path: string,
  params: Record<string, string>,
  accessToken: string,
) {
  const url = new URL(path, PIXIV_APP_API_BASE);
  for (const [key, value] of Object.entries(params)) {
    url.searchParams.set(key, value);
  }

  const response = await fetch(url, {
    method: "GET",
    headers: {
      Host: "app-api.pixiv.net",
      "App-OS": "ios",
      "App-OS-Version": "14.6",
      "User-Agent": "PixivIOSApp/7.13.3 (iOS 14.6; iPhone13,2)",
      Authorization: `Bearer ${accessToken}`,
    },
  });

  const data = await response.json();
  if (!response.ok) {
    throw new Error(`Pixiv API failed (status ${response.status})`);
  }

  return data;
}


