import { UserRoleSchema, type MeResponse, type UserRole } from "@kyno/contracts";
import { Hono } from "hono";
import { requireAuth, type AuthVariables } from "../../middleware/require-auth";

function toUserRole(value: unknown): UserRole {
  const result = UserRoleSchema.safeParse(value);
  return result.success ? result.data : "user";
}

export const accountRoutes = new Hono<{ Variables: AuthVariables }>().get(
  "/me",
  requireAuth,
  (c) => {
    const user = c.get("user");

    const body: MeResponse = {
      id: user.id,
      name: user.name,
      email: user.email,
      role: toUserRole(user.role),
    };

    return c.json(body);
  },
);
