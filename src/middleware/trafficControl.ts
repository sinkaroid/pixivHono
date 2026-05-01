import { MiddlewareHandler } from "hono";
import { cors } from "hono/cors";
import type { Bucket } from "../types/trafficControl";

const buckets = new Map<string, Bucket>();
const RATE_LIMIT_WINDOW_MS = Number(process.env.RATE_LIMIT_WINDOW_MS ?? 60_000);
const RATE_LIMIT_MAX = Number(process.env.RATE_LIMIT_MAX ?? 60);
const SLOW_DOWN_WINDOW_MS = Number(process.env.SLOW_DOWN_WINDOW_MS ?? 60_000);
const SLOW_DOWN_DELAY_AFTER = Number(process.env.SLOW_DOWN_DELAY_AFTER ?? 20);
const SLOW_DOWN_DELAY_MS = Number(process.env.SLOW_DOWN_DELAY_MS ?? 250);
const SLOW_DOWN_MAX_DELAY_MS = Number(process.env.SLOW_DOWN_MAX_DELAY_MS ?? 2_000);
const BUCKET_MAX_SIZE = Number(process.env.RATE_LIMIT_BUCKET_MAX_SIZE ?? 50_000);
const BUCKET_SWEEP_INTERVAL_MS = Number(process.env.RATE_LIMIT_SWEEP_INTERVAL_MS ?? 30_000);

function getClientKey(req: Request): string {
  const forwarded = req.headers.get("x-forwarded-for");
  if (forwarded) return forwarded.split(",")[0].trim();
  const realIp = req.headers.get("x-real-ip");
  if (realIp) return realIp;
  return "unknown";
}

function touchBucket(key: string, windowMs: number): Bucket {
  const now = Date.now();
  const current = buckets.get(key);

  if (!current || current.resetAt <= now) {
    const fresh = { count: 1, resetAt: now + windowMs };
    buckets.set(key, fresh);
    return fresh;
  }

  current.count += 1;
  return current;
}

function sweepExpiredBuckets() {
  const now = Date.now();
  for (const [key, bucket] of buckets) {
    if (bucket.resetAt <= now) {
      buckets.delete(key);
    }
  }
}

setInterval(sweepExpiredBuckets, BUCKET_SWEEP_INTERVAL_MS).unref();

export const corsMiddleware = cors({
  origin: process.env.CORS_ORIGIN ?? "*",
  allowMethods: ["GET", "OPTIONS"],
  allowHeaders: ["Content-Type", "Authorization"],
});

export const slowDownMiddleware: MiddlewareHandler = async (c, next) => {
  if (c.req.method === "OPTIONS") return next();

  if (buckets.size > BUCKET_MAX_SIZE) {
    sweepExpiredBuckets();
  }
  const key = `slow:${getClientKey(c.req.raw)}:${c.req.path}`;
  const bucket = touchBucket(key, SLOW_DOWN_WINDOW_MS);

  if (bucket.count > SLOW_DOWN_DELAY_AFTER) {
    const steps = bucket.count - SLOW_DOWN_DELAY_AFTER;
    const wait = Math.min(steps * SLOW_DOWN_DELAY_MS, SLOW_DOWN_MAX_DELAY_MS);
    await new Promise((resolve) => setTimeout(resolve, wait));
  }

  await next();
};

export const rateLimitMiddleware: MiddlewareHandler = async (c, next) => {
  if (c.req.method === "OPTIONS") return next();

  if (buckets.size > BUCKET_MAX_SIZE) {
    sweepExpiredBuckets();
  }
  const key = `limit:${getClientKey(c.req.raw)}:${c.req.path}`;
  const bucket = touchBucket(key, RATE_LIMIT_WINDOW_MS);

  const remaining = Math.max(0, RATE_LIMIT_MAX - bucket.count);
  c.header("X-RateLimit-Limit", String(RATE_LIMIT_MAX));
  c.header("X-RateLimit-Remaining", String(remaining));
  c.header("X-RateLimit-Reset", String(Math.floor(bucket.resetAt / 1000)));

  if (bucket.count > RATE_LIMIT_MAX) {
    const retryAfterSec = Math.max(1, Math.ceil((bucket.resetAt - Date.now()) / 1000));
    c.header("Retry-After", String(retryAfterSec));
    return c.json({ error: "Too many requests, please try again later." }, 429);
  }

  await next();
};
