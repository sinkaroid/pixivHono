import { Context } from "hono";
import { getPixivAccessToken, pixivGet } from "../clients/pixiv";
import { enrichArtworkResponseWithResolvedUrls } from "../lib/pixivImageResolver";

export async function artworkController(c: Context) {
  const idRaw = c.req.query("id");
  const illustId = Number(idRaw);

  if (!Number.isInteger(illustId) || illustId <= 0) {
    return c.json({ error: "Invalid artwork id" }, 400);
  }

  const refreshToken = process.env.PIXIV_REFRESH_TOKEN;
  if (!refreshToken) {
    return c.json({ error: "Missing env: PIXIV_REFRESH_TOKEN" }, 500);
  }

  try {
    const accessToken = await getPixivAccessToken(refreshToken);
    const response = await pixivGet(
      "/v1/illust/detail",
      { illust_id: String(illustId) },
      accessToken,
    );
    const baseUrl = new URL(c.req.url).origin;
    const enriched = enrichArtworkResponseWithResolvedUrls(response, baseUrl);
    return c.json(enriched as object);
  } catch (error) {
    return c.json(
      { error: "Failed to get artwork", detail: (error as Error).message },
      500,
    );
  }
}



