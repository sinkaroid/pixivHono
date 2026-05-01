import { z } from "zod";

export const ArtworkSchema = z.object({
  id: z.number().describe("Artwork ID"),
  title: z.string().describe("Artwork title"),
  description: z.string().optional().describe("Artwork description"),
  userId: z.number().describe("Creator user ID"),
  userName: z.string().describe("Creator user name"),
  imageUrl: z.string().url().describe("Artwork image URL"),
});

export const SearchResultSchema = z.object({
  artworks: z.array(ArtworkSchema).describe("List of artworks"),
  total: z.number().describe("Total results"),
  offset: z.number().describe("Current offset"),
});

export const HealthCheckSchema = z.object({
  status: z.string().describe("Health status"),
  timestamp: z.string().datetime().describe("Check timestamp"),
});
