import { Context } from "hono";
import { refreshPixivAccessToken } from "../clients/pixiv";

export async function tokenHealthController(c: Context) {
  const refreshToken = process.env.PIXIV_REFRESH_TOKEN;
  if (!refreshToken) {
    return c.json(
      {
        ok: false,
        valid: false,
        error: "Missing env: PIXIV_REFRESH_TOKEN",
      },
      500,
    );
  }

  try {
    await refreshPixivAccessToken(refreshToken);
    return c.json({
      ok: true,
      valid: true,
      message: "Refresh token is valid.",
    });
  } catch (error) {
    return c.json(
      {
        ok: false,
        valid: false,
        error: "Refresh token is invalid or expired.",
        detail: (error as Error).message,
      },
      500,
    );
  }
}
