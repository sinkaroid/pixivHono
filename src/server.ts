import { logger } from "hono/logger";
import app from "./index";
import { artworkController } from "./controllers/artwork";
import { imgResolverController } from "./controllers/imgResolver";
import { searchController } from "./controllers/search";
import { tokenHealthController } from "./controllers/tokenHealth";
import { apiKeyMiddleware } from "./middleware/apiKey";
import {
  corsMiddleware,
  rateLimitMiddleware,
  slowDownMiddleware,
} from "./middleware/trafficControl";

// Bun automatically loads .env on startup
const enableAccessLog = (process.env.ENABLE_ACCESS_LOG ?? "false").toLowerCase() === "true";
const enableUserAgentLog = (process.env.ENABLE_USER_AGENT_LOG ?? "false").toLowerCase() === "true";

if (enableAccessLog) {
  app.use("*", logger());
}
if (enableUserAgentLog) {
  app.use("*", async (c, next) => {
    const userAgent = c.req.header("user-agent") ?? "unknown";
    console.log(`[UA] ${c.req.method} ${c.req.path} :: ${userAgent}`);
    await next();
  });
}
app.use("*", corsMiddleware);
app.use("*", async (c, next) => {
  if (c.req.path === "/pixiv/img_resolver") {
    await next();
    return;
  }

  return apiKeyMiddleware(c, next);
});
app.use("*", slowDownMiddleware);
app.use("*", rateLimitMiddleware);

// API routes with documentation tags
app.get("/pixiv/search", (c) => searchController(c));
app.get("/pixiv/artworks", (c) => artworkController(c));
app.get("/pixiv/img_resolver", (c) => imgResolverController(c));
app.get("/pixiv/token_health", (c) => tokenHealthController(c));

const port = Number(process.env.PORT ?? 3000);

export default {
  fetch: app.fetch,
  port,
};

console.log(`Server running on http://localhost:${port}`);


