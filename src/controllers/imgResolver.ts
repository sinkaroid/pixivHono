import { Context } from "hono";

function isAllowedPixivImageUrl(rawUrl: string): boolean {
  try {
    const parsed = new URL(rawUrl);
    return parsed.protocol === "https:" && parsed.hostname.endsWith("pximg.net");
  } catch {
    return false;
  }
}

export async function imgResolverController(c: Context) {
  const imageUrl = c.req.query("url");

  if (!imageUrl) {
    return c.json({ error: "Missing query param: url" }, 400);
  }

  if (!isAllowedPixivImageUrl(imageUrl)) {
    return c.json({ error: "Invalid or disallowed image url" }, 400);
  }

  try {
    const upstream = await fetch(imageUrl, {
      headers: {
        Referer: "https://www.pixiv.net/",
        "User-Agent":
          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
      },
    });

    if (!upstream.ok || !upstream.body) {
      return c.json(
        {
          error: "Failed to resolve image",
          status: upstream.status,
        },
        502,
      );
    }

    const contentType = upstream.headers.get("content-type") ?? "application/octet-stream";
    const cacheControl = upstream.headers.get("cache-control") ?? "public, max-age=3600";

    return c.body(upstream.body, 200, {
      "Content-Type": contentType,
      "Cache-Control": cacheControl,
    });
  } catch (error) {
    return c.json(
      { error: "Failed to resolve image", detail: (error as Error).message },
      500,
    );
  }
}
