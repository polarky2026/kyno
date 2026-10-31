import { createFileRoute, Link, redirect, useNavigate } from "@tanstack/react-router";
import { useState, type SubmitEvent } from "react";
import { Button } from "../components/Button";
import { Field } from "../components/Field";
import { api } from "../lib/api";

export const Route = createFileRoute("/sign-in")({
  beforeLoad: ({ context }) => {
    if (context.user) {
      throw redirect({ to: "/library" });
    }
  },
  component: SignInPage,
});

function SignInPage() {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [isPending, setIsPending] = useState(false);

  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const email = String(form.get("email") ?? "");
    const password = String(form.get("password") ?? "");

    setIsPending(true);
    setError(null);

    try {
      const { error } = await api.POST("/api/auth/sign-in", { body: { email, password } });

      if (error) {
        setError(error.error);
        return;
      }

      await navigate({ to: "/library" });
    } catch {
      setError("Couldn't reach Kyno. Check your connection and try again.");
    } finally {
      setIsPending(false);
    }
  }

  return (
    <main className="grid flex-1 place-items-center p-6">
      <form onSubmit={(event) => void handleSubmit(event)} className="grid w-full max-w-sm gap-5">
        <div className="grid gap-1 text-center">
          <h1 className="font-display text-4xl">Welcome back</h1>
          <p className="text-sm text-muted">Sign in to your cinema archive.</p>
        </div>

        <Field label="Email" name="email" type="email" autoComplete="email" required />
        <Field
          label="Password"
          name="password"
          type="password"
          autoComplete="current-password"
          required
        />

        {error && (
          <p role="alert" className="text-sm text-brand-text">
            {error}
          </p>
        )}

        <Button type="submit" disabled={isPending}>
          {isPending ? "Signing in…" : "Sign in"}
        </Button>

        <p className="text-center text-sm text-muted">
          New to Kyno?{" "}
          <Link to="/sign-up" className="text-brand-text underline underline-offset-4">
            Create an account
          </Link>
        </p>
      </form>
    </main>
  );
}
