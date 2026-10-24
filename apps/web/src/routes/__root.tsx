import { createRootRoute, Link, Outlet, useNavigate } from "@tanstack/react-router";
import { authClient } from "../lib/auth-client";

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
        <AccountNav />
      </header>
      <Outlet />
    </div>
  );
}

function AccountNav() {
  const navigate = useNavigate();
  const { data: session, isPending } = authClient.useSession();

  async function signOut() {
    await authClient.signOut();
    await navigate({ to: "/" });
  }

  if (isPending) {
    return null;
  }

  if (!session) {
    return (
      <nav className="flex items-center gap-5 text-sm">
        <Link to="/sign-in" className="text-muted hover:text-foreground">
          Sign in
        </Link>
        <Link to="/sign-up" className="text-muted hover:text-foreground">
          Create account
        </Link>
      </nav>
    );
  }

  return (
    <nav className="flex items-center gap-5 text-sm">
      <Link to="/library" className="text-muted hover:text-foreground">
        Library
      </Link>
      <button
        type="button"
        onClick={() => void signOut()}
        className="text-muted hover:text-foreground"
      >
        Sign out
      </button>
    </nav>
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
