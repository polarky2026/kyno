import type { paths } from "@kyno/contracts";
import createClient from "openapi-fetch";

export const api = createClient<paths>();
