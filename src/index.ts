import { OpenAPIHono } from "@hono/zod-openapi";
import { swaggerUI } from "@hono/swagger-ui";
import { openAPISpec } from "./lib/openapi-spec";

const app = new OpenAPIHono();

app.get("/", (c) => c.json({ ok: true, message: "Pixiv Hono API ready" }));

app.get("/doc", (c) => c.json(openAPISpec));

app.get("/playground", swaggerUI({ 
  url: "/doc",
}));

export default app;
