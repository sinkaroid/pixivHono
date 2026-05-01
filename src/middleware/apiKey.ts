import { MiddlewareHandler } from "hono";
import { bearerAuth } from "hono/bearer-auth";

export const apiKeyMiddleware: MiddlewareHandler = async (c, next) => {
  const expectedApiKey = process.env.API_KEY;
  const allowQueryApiKeyInDev =
    process.env.ALLOW_QUERY_API_KEY_IN_DEV === "true";

  if (!expectedApiKey) {
    return c.json({ error: "Server misconfigured: missing API_KEY" }, 500);
  }

  const authorizationHeader = c.req.header("authorization");
  if (authorizationHeader) {
    const bearerMiddleware = bearerAuth({
      verifyToken: async (token) => token === expectedApiKey,
    });
    return bearerMiddleware(c, next);
  }

  const headerApiKey = c.req.header("x-api-key");
  const queryApiKey = allowQueryApiKeyInDev ? c.req.query("api_key") : undefined;
  const providedApiKey = headerApiKey ?? queryApiKey;

  if (!providedApiKey || providedApiKey !== expectedApiKey) {
    return c.json({ error: "Unauthorized" }, 401);
  }

  await next();
};
