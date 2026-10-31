import type { User } from "@kyno/contracts";
import { api } from "./api";

/**
 * Returns the signed-in user, or null when there is no valid session.
 * The session lives in an HttpOnly cookie, so this is the only way the
 * web app can know who is signed in.
 */
export async function fetchCurrentUser(): Promise<User | null> {
  try {
    const { data } = await api.GET("/api/me");
    return data?.user ?? null;
  } catch {
    return null;
  }
}
