# Kyno

Your cinema archive. A personal library to rate, rank and revisit the movies and series you love.

## Stack

- **Web:** React, TypeScript, Vite
- **API:** Hono on Bun
- **Shared:** Zod contracts and design tokens
- **Tooling:** TypeScript, ESLint, Prettier, GitHub Actions

## Project structure

```
apps/
  api/          Hono API
  web/          React app
packages/
  contracts/    Request and response schemas shared by the API and the web
  tokens/       Design tokens (colors and themes) with accessibility checks
```

## Getting started

Requirements: [Bun](https://bun.sh) 1.4.

```bash
bun install
cp apps/api/.env.example apps/api/.env
```

Run the API and the web app in two terminals:

```bash
bun run dev:api   # http://localhost:3000
bun run dev:web   # http://localhost:5173
```

## Scripts

| Command                | Description                    |
| ---------------------- | ------------------------------ |
| `bun run test`         | Run all tests                  |
| `bun run typecheck`    | Type-check every package       |
| `bun run lint`         | Lint the codebase              |
| `bun run format`       | Format files with Prettier     |
| `bun run format:check` | Check formatting without fixes |
