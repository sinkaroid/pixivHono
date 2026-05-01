import { z } from "zod";

/**
 * Zod schema definitions for API validation
 */

export const searchParamsSchema = z.object({
  query: z.string().min(1).describe("Search query (e.g., artist name, tags)"),
  page: z.string().optional().describe("Page number (default: 1)"),
});

export const searchResponseSchema = z.object({
  illusts: z.array(
    z.object({
      id: z.number(),
      title: z.string(),
    })
  ),
  total: z.number(),
  offset: z.number(),
}).passthrough();

export const artworkParamsSchema = z.object({
  id: z.string().min(1).describe("Artwork ID"),
});

export const artworkResponseSchema = z.record(z.string(), z.unknown());

export const imgResolverParamsSchema = z.object({
  url: z.string().url().describe("Pixiv image URL (pximg.net only)"),
});

export const imgResolverResponseSchema = z.object({
  url: z.string().url().optional(),
}).passthrough();

export const tokenHealthResponseSchema = z.object({
  status: z.string().optional(),
  ok: z.boolean().optional(),
  valid: z.boolean().optional(),
  message: z.string().optional(),
  error: z.string().optional(),
}).passthrough();

export const errorSchema = z.object({
  error: z.string(),
  detail: z.string().optional(),
}).passthrough();
