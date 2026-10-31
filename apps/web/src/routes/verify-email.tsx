import { createFileRoute, Link, redirect } from "@tanstack/react-router";
import { api } from "../lib/api";

type VerifyEmailSearch = {
  token: string;
};

export const Route = createFileRoute("/verify-email")({
  validateSearch: (search: Record<string, unknown>): VerifyEmailSearch => ({
    token: typeof search.token === "string" ? search.token : "",
  }),
  loaderDeps: ({ search }) => ({ token: search.token }),
  // Verification links work once, so the token is spent in the loader, which
  // runs once per navigation, rather than in an effect, which StrictMode runs twice.
  loader: async ({ deps }) => {
    const { error } = await api.POST("/api/auth/verify-email", { body: { token: deps.token } });

    if (error) {
      return { error: error.error };
    }

    // The redirect is a new navigation, so the root route reloads the user
    // from the session cookie the API just set.
    throw redirect({ to: "/library" });
  },
  component: VerifyEmailPage,
});

function VerifyEmailPage() {
  const { error } = Route.useLoaderData();

  return (
    <main className="grid flex-1 place-content-center gap-3 p-6 text-center">
      <h1 className="font-display text-4xl">We couldn't verify your email</h1>
      <p className="max-w-sm text-muted">{error}</p>
      <Link to="/sign-in" className="text-brand-text underline underline-offset-4">
        Sign in to get a new link
      </Link>
    </main>
  );
}
