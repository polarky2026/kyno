# Kyno

Your cinema archive. A personal library to rate, rank and revisit the movies and series you love.

## Stack

- **Web:** React, TypeScript, Vite, TanStack Router, Tailwind CSS
- **API:** Go with `net/http`
- **Shared:** API contract and design tokens
- **Tooling:** ESLint, Prettier, gofmt, go vet, GitHub Actions

## Project structure

```
apps/
  api/          Go API
    cmd/api/      Entry point
    internal/     Application code (config, HTTP handlers, platform helpers)
  web/          React app
packages/
  contracts/    Shared API contract
  tokens/       Design tokens (colors, themes, typography) with accessibility checks
```

## Getting started

Requirements:

- [Bun](https://bun.sh) 1.4
- [Go](https://go.dev) 1.27
- [air](https://github.com/air-verse/air) for live reload: `go install github.com/air-verse/air@latest`

```bash
bun install
```

Run the API and the web app in two terminals:

```bash
bun run dev:api   # http://localhost:3000
bun run dev:web   # http://localhost:5173
```

The web app proxies `/api` requests to the API during development.

## Scripts

### Web and shared packages

| Command                | Description                    |
| ---------------------- | ------------------------------ |
| `bun run test`         | Run all tests                  |
| `bun run typecheck`    | Type-check every package       |
| `bun run lint`         | Lint the codebase              |
| `bun run format`       | Format files with Prettier     |
| `bun run format:check` | Check formatting without fixes |

### API

Run from `apps/api`:

| Command         | Description                     |
| --------------- | ------------------------------- |
| `go test ./...` | Run all tests                   |
| `go vet ./...`  | Report suspicious constructs    |
| `gofmt -l .`    | List files that need formatting |
