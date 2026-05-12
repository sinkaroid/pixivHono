import { OpenAPIHono } from "@hono/zod-openapi";
import { swaggerUI } from "@hono/swagger-ui";
import { openAPISpec } from "./lib/openapi-spec";
import packageJson from "../package.json";

const app = new OpenAPIHono();

const GEO_TIMEOUT_MS = 3000;
const GEO_CACHE_MS = 1000 * 60 * 30;

let cachedLastLocation = "Unknown";
let lastLocationTimestamp = 0;

function formatMB(value: number): string {
  return `${Math.round(value * 100) / 100} MB`;
}

function cachedLocationOrUnknown(): string {
  if (Date.now() - lastLocationTimestamp < GEO_CACHE_MS) {
    return cachedLastLocation;
  }

  return "Unknown";
}

function currentProcess() {
  const memory = process.memoryUsage();

  const rss = memory.rss / 1024 / 1024;
  const heapUsed = memory.heapUsed / 1024 / 1024;
  const heapTotal = memory.heapTotal / 1024 / 1024;

  return {
    rss: formatMB(rss),
    heap: `${Math.round(heapUsed * 100) / 100}/${Math.round(heapTotal * 100) / 100} MB`,
  };
}

async function getServer(): Promise<string> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), GEO_TIMEOUT_MS);

  try {
    const raw = await fetch("https://ipwho.is/", {
      signal: controller.signal,
    });

    if (!raw.ok) {
      return cachedLocationOrUnknown();
    }

    const data = await raw.json() as {
      success?: boolean;
      country?: string;
      region?: string;
    };

    if (data.success === false) {
      return cachedLocationOrUnknown();
    }

    const country = data.country?.trim();
    const region = data.region?.trim();

    if (!country || !region) {
      return cachedLocationOrUnknown();
    }

    const location = `${country}, ${region}`;

    cachedLastLocation = location;
    lastLocationTimestamp = Date.now();

    return location;
  } catch {
    return cachedLocationOrUnknown();
  } finally {
    clearTimeout(timeout);
  }
}

app.get("/", async (c) => {
  const proc = currentProcess();
  const server = await getServer();

  return c.json({
    success: true,
    message: "Hi, I'm alive!",
    endpoint: "https://sinkaroid.github.io/pixivHono",
    date: new Date().toLocaleString(),
    rss: proc.rss,
    heap: proc.heap,
    server,
    version: packageJson.version,
  });
});

app.get("/doc", (c) => c.json(openAPISpec));

app.get(
  "/playground",
  swaggerUI({
    url: "/doc",
  })
);

export default app;