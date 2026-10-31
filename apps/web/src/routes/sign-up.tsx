import { createFileRoute, Link, redirect } from "@tanstack/react-router";
import { useState, type SubmitEvent } from "react";
import { Button } from "../components/Button";
import { Field } from "../components/Field";
import { api } from "../lib/api";

export const Route = createFileRoute("/sign-up")({
  beforeLoad: ({ context }) => {
    if (context.user) {
      throw redirect({ to: "/library" });
    }
  },
  component: SignUpPage,
});

function SignUpPage() {
  const [error, setError] = useState<string | null>(null);
  const [isPending, setIsPending] = useState(false);
  const [sentTo, setSentTo] = useState<string | null>(null);

  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const name = String(form.get("name") ?? "");
    const email = String(form.get("email") ?? "");
    const password = String(form.get("password") ?? "");

    setIsPending(true);
    setError(null);

    try {
      const { error } = await api.POST("/api/auth/sign-up", { body: { name, email, password } });

      if (error) {
        setError(error.error);
        return;
      }

      setSentTo(email);
    } catch {
      setError("Couldn't reach Kyno. Check your connection and try again.");
    } finally {
      setIsPending(false);
    }
  }

  if (sentTo) {
    return <CheckYourInbox email={sentTo} />;
  }

  return (
    <main className="grid flex-1 place-items-center p-6">
      <form onSubmit={(event) => void handleSubmit(event)} className="grid w-full max-w-sm gap-5">
        <div className="grid gap-1 text-center">
          <h1 className="font-display text-4xl">Create your archive</h1>
          <p className="text-sm text-muted">Rate, rank and revisit the films you love.</p>
        </div>

        <Field label="Name" name="name" autoComplete="name" maxLength={100} required />
        <Field label="Email" name="email" type="email" autoComplete="email" required />
        <Field
          label="Password"
          name="password"
          type="password"
          autoComplete="new-password"
          minLength={10}
          maxLength={128}
          required
        />

        {error && (
          <p role="alert" className="text-sm text-brand-text">
            {error}
          </p>
        )}

        <Button type="submit" disabled={isPending}>
          {isPending ? "Creating account…" : "Create account"}
        </Button>

        <p className="text-center text-sm text-muted">
          Already have an account?{" "}
          <Link to="/sign-in" className="text-brand-text underline underline-offset-4">
            Sign in
          </Link>
        </p>
      </form>
    </main>
  );
}

function CheckYourInbox({ email }: { email: string }) {
  const [resent, setResent] = useState(false);

  async function resend() {
    await api.POST("/api/auth/resend-verification", { body: { email } }).catch(() => undefined);
    setResent(true);
  }

  return (
    <main className="grid flex-1 place-content-center gap-3 p-6 text-center">
      <h1 className="font-display text-4xl">Check your inbox</h1>
      <p className="max-w-sm text-muted">
        We sent a verification link to <span className="text-foreground">{email}</span>. Open it to
        start your archive. The link works for 24 hours.
      </p>
      {resent ? (
        <p className="text-sm text-muted">If that email needs a new link, it's on its way.</p>
      ) : (
        <button
          type="button"
          onClick={() => void resend()}
          className="text-sm text-brand-text underline underline-offset-4"
        >
          Send a new link
        </button>
      )}
    </main>
  );
}
