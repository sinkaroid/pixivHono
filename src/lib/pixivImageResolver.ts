import type { ObjectMap } from "../types/common";

function normalizeResolverBase(baseUrl: string): string {
  const forceHttpsEnabled = /^(1|true|yes|on)$/i.test(
    (process.env.FORCE_HTTPS_PIXIV_IMG_RESOLVER ?? "").trim(),
  );
  return forceHttpsEnabled ? baseUrl.replace(/^http:\/\//i, "https://") : baseUrl;
}

function makeResolvedUrl(baseUrl: string, raw: string): string {
  const externalResolver = process.env.PIXIV_IMG_RESOLVER?.trim();
  if (externalResolver) {
    const resolverBaseRaw = externalResolver.startsWith("http")
      ? externalResolver
      : `https://${externalResolver}`;
    const resolverBase = normalizeResolverBase(resolverBaseRaw);
    const normalizedResolver = resolverBase.replace(/\/+$/, "");
    const parsed = new URL(raw);
    return `${normalizedResolver}${parsed.pathname}${parsed.search}`;
  }

  const normalizedBaseUrl = normalizeResolverBase(baseUrl);
  return `${normalizedBaseUrl}/pixiv/img_resolver?url=${encodeURIComponent(raw)}`;
}

function addResolvedImageUrls(imageUrls: unknown, baseUrl: string): unknown {
  if (!imageUrls || typeof imageUrls !== "object") return imageUrls;

  const source = imageUrls as ObjectMap;
  const resolved: ObjectMap = {};

  for (const [key, value] of Object.entries(source)) {
    if (typeof value === "string" && value.startsWith("https://")) {
      resolved[key] = makeResolvedUrl(baseUrl, value);
    }
  }

  return {
    ...source,
    _resolved: resolved,
  };
}

function enrichIllust(illust: ObjectMap, baseUrl: string): ObjectMap {
  const next: ObjectMap = { ...illust };

  if (illust.image_urls) {
    next.image_urls = addResolvedImageUrls(illust.image_urls, baseUrl);
  }

  const metaSinglePage = illust.meta_single_page as ObjectMap | undefined;
  if (metaSinglePage?.original_image_url && typeof metaSinglePage.original_image_url === "string") {
    next.meta_single_page = {
      ...metaSinglePage,
      original_image_url_resolved: makeResolvedUrl(baseUrl, metaSinglePage.original_image_url),
    };
  }

  const metaPages = illust.meta_pages;
  if (Array.isArray(metaPages)) {
    next.meta_pages = metaPages.map((page) => {
      if (!page || typeof page !== "object") return page;
      const pageRecord = page as ObjectMap;
      return {
        ...pageRecord,
        image_urls: addResolvedImageUrls(pageRecord.image_urls, baseUrl),
      };
    });
  }

  return next;
}

export function enrichSearchResponseWithResolvedUrls(payload: unknown, baseUrl: string): unknown {
  if (!payload || typeof payload !== "object") return payload;

  const record = payload as ObjectMap;
  if (!Array.isArray(record.illusts)) return payload;

  return {
    ...record,
    illusts: record.illusts.map((item) => {
      if (!item || typeof item !== "object") return item;
      return enrichIllust(item as ObjectMap, baseUrl);
    }),
  };
}

export function enrichArtworkResponseWithResolvedUrls(payload: unknown, baseUrl: string): unknown {
  if (!payload || typeof payload !== "object") return payload;

  const record = payload as ObjectMap;
  if (!record.illust || typeof record.illust !== "object") return payload;

  return {
    ...record,
    illust: enrichIllust(record.illust as ObjectMap, baseUrl),
  };
}