import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { api } from "../lib/api";

export const Route = createFileRoute("/")({
  component: HomePage,
});

type ApiStatus = "checking" | "online" | "offline";

function HomePage() {
  const [status, setStatus] = useState<ApiStatus>("checking");

  useEffect(() => {
    api
      .GET("/api/health")
      .then(({ data }) => setStatus(data?.status === "ok" ? "online" : "offline"))
      .catch(() => setStatus("offline"));
  }, []);

  return (
    <main className="grid flex-1 place-content-center gap-3 p-6 text-center">
      <p className="text-xs tracking-[0.2em] text-muted uppercase">Your cinema archive</p>
      <h1 className="font-display text-7xl text-brand-text sm:text-8xl">Kyno</h1>
      <p className="text-sm text-muted">
        {status === "online" && (
          <span aria-hidden="true" className="text-highlight">
            ●{" "}
          </span>
        )}
        API {status}
      </p>
    </main>
  );
}
