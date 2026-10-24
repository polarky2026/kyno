import { MeResponseSchema } from "@kyno/contracts";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/library")({
  loader: async () => {
    const response = await fetch("/api/me");

    if (!response.ok) {
      throw new Error("Couldn't load your account");
    }

    return MeResponseSchema.parse(await response.json());
  },
  component: LibraryPage,
});

function LibraryPage() {
  const me = Route.useLoaderData();

  return (
    <main className="grid flex-1 place-content-center gap-3 p-6 text-center">
      <p className="text-xs tracking-[0.2em] text-muted uppercase">Your library</p>
      <h1 className="font-display text-5xl">Welcome, {me.name}</h1>
      <p className="max-w-sm text-muted">
        Your archive is empty for now. Movies arrive in the next stage.
      </p>
    </main>
  );
}
