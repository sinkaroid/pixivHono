/**
 * API Documentation
 * 
 * This file provides OpenAPI schema information for reference
 * Routes are decorated with @hono/zod-openapi for automatic documentation generation
 * 
 * Available endpoints:
 * - GET /pixiv/search?query=<string>&page=<number>
 * - GET /pixiv/artworks?id=<number>
 * - GET /pixiv/img_resolver?url=<string>
 * - GET /pixiv/token_health
 * 
 * Interactive documentation: http://localhost:3000/reference
 * OpenAPI JSON spec: http://localhost:3000/doc
 */

export interface SearchParams {
  query: string; // e.g., "girl"
  page?: string; // default: "1"
}

export interface SearchResponse {
  illusts: Array<{
    id: number;
    title: string;
    user: { id: number; name: string };
    image_urls: { square_medium: string; medium: string };
  }>;
  total: number;
  offset: number;
}

export interface ArtworkParams {
  id: string; // Artwork ID
}

export interface ArtworkResponse {
  illust: Record<string, unknown>; // Full Pixiv artwork object
}

export interface ImgResolverParams {
  url: string; // Pixiv image URL (pximg.net only)
}

export interface ImgResolverResponse {
  url?: string; // Resolved image URL
}

export interface TokenHealthResponse {
  status: string; // "healthy", "expired", or "invalid"
  ok?: boolean;
  valid?: boolean;
  message?: string;
  error?: string;
  detail?: string;
}

export interface ErrorResponse {
  error: string;
  detail?: string;
}
