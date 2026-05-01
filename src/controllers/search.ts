import { Context } from "hono";
import { getPixivAccessToken, pixivGet } from "../clients/pixiv";
import { getSearchCacheKey, getSearchCacheTtlMs, searchCache } from "../cache/search";
import { enrichSearchResponseWithResolvedUrls } from "../lib/pixivImageResolver";
import { getPublicBaseUrl } from "../lib/requestOrigin";

export async function searchController(c: Context) {
  const q = c.req.query("query");
  const pageRaw = c.req.query("page") ?? "1";
  const page = Number(pageRaw);
  const perPage = 30;

  if (!q) {
    return c.json({ error: "Missing query param: query" }, 400);
  }
  if (!Number.isInteger(page) || page <= 0) {
    return c.json({ error: "Invalid query param: page (must be positive integer)" }, 400);
  }

  const refreshToken = process.env.PIXIV_REFRESH_TOKEN;
  if (!refreshToken) {
    return c.json({ error: "Missing env: PIXIV_REFRESH_TOKEN" }, 500);
  }

  try {
    const baseUrl = getPublicBaseUrl(c);
    const cacheKey = getSearchCacheKey(q, page);
    const cached = await searchCache.get(cacheKey);
    if (cached) {
      const enrichedCached = enrichSearchResponseWithResolvedUrls(cached, baseUrl);
      return c.json(enrichedCached as object);
    }

    const accessToken = await getPixivAccessToken(refreshToken);
    const response = await pixivGet(
      "/v1/search/illust",
      {
        word: q,
        search_target: "partial_match_for_tags",
        sort: "date_desc",
        filter: "for_ios",
        merge_plain_keyword_results: "true",
        include_translated_tag_results: "true",
        offset: String((page - 1) * perPage),
      },
      accessToken,
    );

    const enriched = enrichSearchResponseWithResolvedUrls(response, baseUrl);

    await searchCache.set(cacheKey, response, getSearchCacheTtlMs());
    return c.json(enriched as object);
  } catch (error) {
    return c.json(
      { error: "Failed to search artworks", detail: (error as Error).message },
      500,
    );
  }
}



