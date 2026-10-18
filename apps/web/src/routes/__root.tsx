import { createRootRoute, Link, Outlet } from "@tanstack/react-router";

export const Route = createRootRoute({
  component: RootLayout,
  notFoundComponent: NotFound,
});

function RootLayout() {
  return (
    <div className="flex min-h-screen flex-col">
      <header className="flex items-center justify-between border-b border-border px-6 py-4">
        <Link to="/" className="font-display text-2xl text-brand-text">
          Kyno
        </Link>
        <Link to="/sign-up" className="text-sm text-muted hover:text-foreground">
          Create account
        </Link>
      </header>
      <Outlet />
    </div>
  );
}

function NotFound() {
  return (
    <main className="grid flex-1 place-content-center gap-3 p-6 text-center">
      <h1 className="font-display text-4xl">This page isn't in the archive</h1>
      <Link to="/" className="text-brand-text underline underline-offset-4">
        Back to Kyno
      </Link>
    </main>
  );
}
