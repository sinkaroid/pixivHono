import type { Context } from "hono";

function getFirstHeaderValue(value: string | undefined): string | undefined {
  if (!value) return undefined;
  const first = value.split(",")[0]?.trim();
  return first || undefined;
}

function parseForwardedHeader(value: string | undefined): { proto?: string; host?: string } {
  const first = getFirstHeaderValue(value);
  if (!first) return {};

  const result: { proto?: string; host?: string } = {};

  for (const part of first.split(";")) {
    const [rawKey, ...rest] = part.split("=");
    const key = rawKey?.trim().toLowerCase();
    const rawValue = rest.join("=").trim();
    if (!key || !rawValue) continue;

    const normalizedValue = rawValue.replace(/^"|"$/g, "");
    if (key === "proto") result.proto = normalizedValue;
    if (key === "host") result.host = normalizedValue;
  }

  return result;
}

export function getPublicBaseUrl(c: Context): string {
  const forwarded = parseForwardedHeader(c.req.header("forwarded"));
  const forwardedProto = getFirstHeaderValue(c.req.header("x-forwarded-proto"));
  const forwardedHost = getFirstHeaderValue(c.req.header("x-forwarded-host"));

  const proto = (forwarded.proto ?? forwardedProto)?.toLowerCase();
  const host = forwarded.host ?? forwardedHost ?? c.req.header("host");

  if (host && proto && (proto === "http" || proto === "https")) {
    try {
      return new URL(`${proto}://${host}`).origin;
    } catch {
      // Fall back to request URL origin when proxy headers are malformed.
    }
  }

  return new URL(c.req.url).origin;
}
