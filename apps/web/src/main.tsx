import "@fontsource-variable/fraunces";
import "@fontsource-variable/manrope";
import { createThemeCss } from "@kyno/tokens";
import { createRouter, RouterProvider } from "@tanstack/react-router";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { routeTree } from "./routeTree.gen";
import "./styles.css";

const router = createRouter({ routeTree });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

const themeStyle = document.createElement("style");
themeStyle.textContent = createThemeCss();
document.head.append(themeStyle);

const root = document.getElementById("root");

if (!root) {
  throw new Error("Root element not found");
}

createRoot(root).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
);
