import { drizzleAdapter } from "@better-auth/drizzle-adapter";
import { betterAuth } from "better-auth";
import { env } from "../config";
import { db } from "../db/client";
import * as authSchema from "../db/schema/auth";
import { sendEmail } from "./email";
import { verifyEmailTemplate } from "./emails/verify-email";

export const auth = betterAuth({
  secret: env.BETTER_AUTH_SECRET,
  baseURL: env.BETTER_AUTH_URL,
  database: drizzleAdapter(db, { provider: "pg", schema: authSchema }),
  emailAndPassword: {
    enabled: true,
    minPasswordLength: 10,
    requireEmailVerification: true,
  },
  emailVerification: {
    sendOnSignUp: true,
    sendOnSignIn: true,
    autoSignInAfterVerification: true,
    sendVerificationEmail: async ({ user, url }) => {
      const template = verifyEmailTemplate({ name: user.name, url });

      sendEmail({ to: user.email, ...template }).catch((error: unknown) => {
        console.error("Verification email failed", error);
      });
    },
  },
  rateLimit: {
    enabled: true,
  },
  user: {
    additionalFields: {
      role: {
        type: "string",
        required: false,
        defaultValue: "user",
        input: false,
      },
    },
  },
});
