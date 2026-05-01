import Keyv from "keyv";
import KeyvRedis from "@keyv/redis";

const redisUrl = process.env.REDIS_URL;
const ttlMs = Number(process.env.SEARCH_CACHE_TTL_MS ?? 3_600_000);

export const searchCache = redisUrl
  ? new Keyv({ store: new KeyvRedis(redisUrl), namespace: "pixiv-search" })
  : new Keyv({ namespace: "pixiv-search" });

searchCache.on("error", (error) => {
  console.error("[search-cache] error", error);
});

export function getSearchCacheKey(query: string, page: number): string {
  return `q:${query.trim().toLowerCase()}:p:${page}`;
}

export function getSearchCacheTtlMs(): number {
  return Number.isFinite(ttlMs) && ttlMs > 0 ? ttlMs : 3_600_000;
}
