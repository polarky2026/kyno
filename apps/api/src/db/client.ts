import { drizzle } from "drizzle-orm/postgres-js";
import postgres from "postgres";
import { env } from "../config";

const connection = postgres(env.DATABASE_URL);

export const db = drizzle({ client: connection });
