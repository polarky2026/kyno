import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useState, type FormEvent } from "react";
import { Button } from "../components/Button";
import { Field } from "../components/Field";
import { authClient } from "../lib/auth-client";

export const Route = createFileRoute("/sign-in")({
  component: SignInPage,
});

function SignInPage() {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const [isPending, setIsPending] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const email = String(form.get("email") ?? "");
    const password = String(form.get("password") ?? "");

    setIsPending(true);
    setError(null);

    const { error } = await authClient.signIn.email({ email, password });

    setIsPending(false);

    if (error) {
      setError(
        error.code === "EMAIL_NOT_VERIFIED"
          ? "Verify your email first. We just sent you a new link."
          : "Email or password is incorrect.",
      );
      return;
    }

    await navigate({ to: "/library" });
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
