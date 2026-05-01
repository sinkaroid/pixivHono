export const openAPISpec = {
  openapi: "3.0.0",
  info: {
    title: "Pixiv Hono API",
    version: "1.0.1",
    description: "A high-performance Pixiv API proxy built with Hono and Bun",
    contact: {
      name: "@sinkaroid-alter",
    },
  },
  servers: [
    {
      url: "http://localhost:3000",
      description: "Development server",
    },
  ],
  paths: {
    "/pixiv/search": {
      get: {
        summary: "Search artworks",
        operationId: "searchArtworks",
        parameters: [
          {
            name: "query",
            in: "query",
            required: true,
            description: "Search query (e.g., artist name, tags)",
            schema: {
              type: "string",
            },
          },
          {
            name: "page",
            in: "query",
            required: false,
            description: "Page number (default: 1)",
            schema: {
              type: "string",
            },
          },
        ],
        responses: {
          "200": {
            description: "Search results from Pixiv",
            content: {
              "application/json": {
                schema: {
                  type: "object",
                  properties: {
                    illusts: {
                      type: "array",
                      items: {
                        type: "object",
                      },
                    },
                    total: {
                      type: "number",
                    },
                    offset: {
                      type: "number",
                    },
                  },
                },
              },
            },
          },
          "400": {
            description: "Invalid query parameters",
            content: {
              "application/json": {
                schema: {
                  type: "object",
                  properties: {
                    error: {
                      type: "string",
                    },
                  },
                },
              },
            },
          },
          "500": {
            description: "Server error",
            content: {
              "application/json": {
                schema: {
                  type: "object",
                  properties: {
                    error: {
                      type: "string",
                    },
                    detail: {
                      type: "string",
                    },
                  },
                },
              },
            },
          },
        },
      },
    },
    "/pixiv/artworks": {
      get: {
        summary: "Get artwork details",
        operationId: "getArtwork",
        parameters: [
          {
            name: "id",
            in: "query",
            required: true,
            description: "Artwork ID",
            schema: {
              type: "string",
            },
          },
        ],
        responses: {
          "200": {
            description: "Artwork details from Pixiv",
            content: {
              "application/json": {
                schema: {
                  type: "object",
                },
              },
            },
          },
          "400": {
            description: "Invalid artwork ID",
            content: {
              "application/json": {
                schema: {
                  type: "object",
                  properties: {
                    error: {
                      type: "string",
                    },
                  },
                },
              },
            },
          },
          "500": {
            description: "Server error",
            content: {
              "application/json": {
                schema: {
                  type: "object",
                  properties: {
                    error: {
                      type: "string",
                    },
                    detail: {
                      type: "string",
                    },
                  },
                },
              },
            },
          },
        },
      },
    },
    "/pixiv/img_resolver": {
      get: {
        summary: "Resolve Pixiv image URL",
        operationId: "resolveImage",
        parameters: [
          {
            name: "url",
            in: "query",
            required: true,
            description: "Pixiv image URL (pximg.net only)",
            schema: {
              type: "string",
              format: "uri",
            },
          },
        ],
        responses: {
          "200": {
            description: "Resolved image from Pixiv",
            content: {
              "image/*": {
                schema: {
                  type: "string",
                  format: "binary",
                },
              },
            },
          },
          "400": {
            description: "Invalid image URL",
            content: {
              "application/json": {
                schema: {
                  type: "object",
                  properties: {
                    error: {
                      type: "string",
                    },
                  },
                },
              },
            },
          },
          "502": {
            description: "Failed to resolve image",
            content: {
              "application/json": {
                schema: {
                  type: "object",
                  properties: {
                    error: {
                      type: "string",
                    },
                    detail: {
                      type: "string",
                    },
                  },
                },
              },
            },
          },
        },
      },
    },
    "/pixiv/token_health": {
      get: {
        summary: "Check token health status",
        operationId: "checkTokenHealth",
        responses: {
          "200": {
            description: "Token health status",
            content: {
              "application/json": {
                schema: {
                  type: "object",
                  properties: {
                    status: {
                      type: "string",
                    },
                    ok: {
                      type: "boolean",
                    },
                    valid: {
                      type: "boolean",
                    },
                    message: {
                      type: "string",
                    },
                    error: {
                      type: "string",
                    },
                  },
                },
              },
            },
          },
          "500": {
            description: "Server error",
            content: {
              "application/json": {
                schema: {
                  type: "object",
                  properties: {
                    error: {
                      type: "string",
                    },
                    detail: {
                      type: "string",
                    },
                  },
                },
              },
            },
          },
        },
      },
    },
  },
};
