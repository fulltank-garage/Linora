# Linora

Linora is a mobile-first LIFF-style web app for analyzing Facebook Page activity through a shared LINE Official Account.

## Stack

- React, TypeScript, Vite, MUI, React Router
- Go, Gin
- pnpm workspace monorepo
- PostgreSQL via Docker Compose

## Apps And Packages

- `apps/web`: mobile LIFF web app
- `apps/api`: Go API
- `packages/ui`: MUI theme and reusable mobile components
- `packages/shared`: frontend-safe TypeScript contracts

## Local Development

```bash
pnpm install
pnpm db:up
pnpm dev:web
pnpm dev:api
```

Web app: http://localhost:5173

API health check: http://localhost:8080/health

## Verification

```bash
cd apps/api && go test ./...
cd apps/api && go build ./cmd/server
pnpm build
```

## Phase 1 Scope

This foundation includes the mobile UI, manual analysis flow, shared contracts, and a deterministic API response for manual reports. LINE webhook, Facebook OAuth, Meta Graph API fetching, persistent storage, token encryption, and real AI integration are planned for later phases.
