# Linora Phase 1 Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first runnable Linora monorepo with mobile LIFF-style React UI, shared packages, and a Go API manual analysis MVP.

**Architecture:** Use `pnpm` workspaces for frontend packages and keep Go isolated in `apps/api`. `apps/web` renders the mobile app and calls backend routes; `packages/ui` owns MUI theme/components; `packages/shared` owns frontend-safe contracts; `apps/api` owns validation and analysis behavior.

**Tech Stack:** React, TypeScript, Vite, MUI, React Router, Axios, LIFF SDK, Go, Gin, PostgreSQL Docker Compose.

---

### Task 1: Workspace Foundation

**Files:**
- Create: `package.json`
- Create: `pnpm-workspace.yaml`
- Create: `.gitignore`
- Create: `.env.example`
- Create: `docker-compose.yml`
- Create: `Makefile`
- Create: `README.md`

- [ ] Create the root workspace files with scripts for `dev`, `dev:web`, `dev:api`, `build`, `lint`, `format`, `db:up`, and `db:down`.
- [ ] Add environment examples for APP, API, LINE, Facebook, AI, and Postgres settings.
- [ ] Add Docker Compose with a `postgres:16` service named `linora-db`.
- [ ] Add README commands for installing, running, and verifying the app.

### Task 2: Shared And UI Packages

**Files:**
- Create: `packages/shared/package.json`
- Create: `packages/shared/src/index.ts`
- Create: `packages/ui/package.json`
- Create: `packages/ui/src/index.ts`
- Create: `packages/ui/src/theme/index.ts`
- Create: `packages/ui/src/components/index.ts`
- Create: `packages/ui/src/components/MobileAppShell.tsx`
- Create: `packages/ui/src/components/InsightCard.tsx`

- [ ] Define frontend-safe TypeScript contracts for customer profile, Facebook page summary, manual analysis request, top posts, important comments, and analysis reports.
- [ ] Define the Emerald/Charcoal/Ivory MUI theme once in `@linora/ui`.
- [ ] Create reusable mobile shell and card primitives so `apps/web` does not duplicate theme or layout rules.

### Task 3: Go API With Test-First Manual Analysis

**Files:**
- Create: `apps/api/go.mod`
- Create: `apps/api/cmd/server/main.go`
- Create: `apps/api/internal/analysis/service.go`
- Create: `apps/api/internal/analysis/service_test.go`
- Create: `apps/api/internal/httpapi/router.go`
- Create: `apps/api/internal/httpapi/router_test.go`

- [ ] Write service tests for valid manual analysis, missing page name, and empty post content.
- [ ] Implement deterministic manual report generation with a bounded health score and Thai summary text.
- [ ] Write HTTP router tests for `GET /health` and `POST /api/analysis/manual`.
- [ ] Implement Gin routes that return consistent JSON success and error payloads.

### Task 4: React Mobile App

**Files:**
- Create/modify: `apps/web/package.json`
- Create/modify: `apps/web/src/main.tsx`
- Create/modify: `apps/web/src/App.tsx`
- Create: `apps/web/src/api/client.ts`
- Create: `apps/web/src/data/demo.ts`
- Create: `apps/web/src/pages/DashboardPage.tsx`
- Create: `apps/web/src/pages/ConnectFacebookPage.tsx`
- Create: `apps/web/src/pages/PageSelectPage.tsx`
- Create: `apps/web/src/pages/AnalyzingPage.tsx`
- Create: `apps/web/src/pages/ReportsPage.tsx`
- Create: `apps/web/src/pages/CommentsPage.tsx`
- Create: `apps/web/src/pages/SettingsPage.tsx`
- Create: `apps/web/src/pages/ManualAnalyzePage.tsx`

- [ ] Scaffold Vite React TypeScript app.
- [ ] Install MUI, Emotion, icons, React Router, Axios, and LIFF SDK.
- [ ] Wire `@linora/ui` and `@linora/shared` via workspace dependencies.
- [ ] Implement mobile shell, bottom navigation, route pages, demo report state, and manual analysis submission.
- [ ] Keep the first screen as the dashboard and use realistic Thai UI copy from the spec.

### Task 5: Verification

**Files:**
- Modify as needed only when tests or builds expose defects.

- [ ] Run `go test ./...` inside `apps/api` and fix failures.
- [ ] Run `go build ./cmd/server` inside `apps/api` and fix failures.
- [ ] Run `pnpm install`.
- [ ] Run `pnpm build` and fix TypeScript or bundling failures.
- [ ] Start the web dev server and visually inspect the mobile UI against the approved concept.
