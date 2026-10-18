import { app } from "./app";
import { env } from "./config";

export default {
  port: env.PORT,
  fetch: app.fetch,
};
