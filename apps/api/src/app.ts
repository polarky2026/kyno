import type { HealthResponse } from "@kyno/contracts";
import { Hono } from "hono";
import { auth } from "./lib/auth";

export const app = new Hono();

app.get("/api/health", (c) => {
  const body: HealthResponse = { status: "ok" };
  return c.json(body);
});

app.on(["GET", "POST"], "/api/auth/*", (c) => auth.handler(c.req.raw));
