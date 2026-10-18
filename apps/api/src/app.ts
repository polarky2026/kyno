import type { HealthResponse } from "@kyno/contracts";
import { Hono } from "hono";

export const app = new Hono();

app.get("/api/health", (c) => {
  const body: HealthResponse = { status: "ok" };
  return c.json(body);
});
