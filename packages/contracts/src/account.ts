import { z } from "zod";

export const UserRoleSchema = z.enum(["user", "admin"]);

export type UserRole = z.infer<typeof UserRoleSchema>;

export const MeResponseSchema = z.object({
  id: z.string(),
  name: z.string(),
  email: z.email(),
  role: UserRoleSchema,
});

export type MeResponse = z.infer<typeof MeResponseSchema>;
