import { z } from "zod";

const EnvSchema = z.object({
  NODE_ENV: z.enum(["development", "test", "production"]).default("development"),
  PORT: z.coerce.number().int().positive().default(3000),
  DATABASE_URL: z.string().regex(/^postgres(ql)?:\/\//, "Must be a PostgreSQL connection URL"),
  BETTER_AUTH_SECRET: z.string().min(32, "Must be at least 32 characters"),
  BETTER_AUTH_URL: z.url(),
  RESEND_API_KEY: z.string().startsWith("re_", "Must be a Resend API key"),
  EMAIL_FROM: z.string().min(1),
});

export type Env = z.infer<typeof EnvSchema>;

export function parseEnv(source: Record<string, string | undefined>): Env {
  const result = EnvSchema.safeParse(source);

  if (!result.success) {
    throw new Error(`Invalid environment variables:\n${z.prettifyError(result.error)}`);
  }

  return result.data;
}
