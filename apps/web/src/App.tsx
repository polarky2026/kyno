import { HealthResponseSchema } from "@kyno/contracts";
import { useEffect, useState } from "react";

type ApiStatus = "checking" | "online" | "offline";

export function App() {
  const [status, setStatus] = useState<ApiStatus>("checking");

  useEffect(() => {
    fetch("/api/health")
      .then((response) => response.json())
      .then((data) => {
        HealthResponseSchema.parse(data);
        setStatus("online");
      })
      .catch(() => setStatus("offline"));
  }, []);

  return (
    <main className="shell">
      <p className="eyebrow">Your Cinema Archive</p>
      <h1 className="wordmark">Kyno</h1>
      <p className="status" data-status={status}>
        API {status}
      </p>
    </main>
  );
}
