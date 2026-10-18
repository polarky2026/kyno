import { HealthResponseSchema } from "@kyno/contracts";
import { describe, expect, test } from "bun:test";
import { app } from "./app";

describe("GET /api/health", () => {
  test("returns a response that matches the contract", async () => {
    const res = await app.request("/api/health");

    expect(res.status).toBe(200);
    expect(HealthResponseSchema.parse(await res.json())).toEqual({ status: "ok" });
  });
});
