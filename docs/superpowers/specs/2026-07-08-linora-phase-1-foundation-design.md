# Linora Phase 1 Foundation Design

## Goal

Create the first working Linora monorepo foundation for a mobile-first LIFF-style web app and Go API. This phase prioritizes a runnable app, shared UI/theme boundaries, manual analysis MVP screens, and backend contracts that can later be connected to LINE, Meta Graph API, and a real AI provider.

## Scope

This first build includes:

- pnpm workspace monorepo structure.
- React, TypeScript, Vite, MUI mobile web app in `apps/web`.
- Go, Gin API app in `apps/api`.
- Shared frontend types in `packages/shared`.
- MUI theme and reusable shell/card components in `packages/ui`.
- Mobile-only dashboard, connect Facebook, page selection, analyzing, reports, important comments, settings, and manual analyze screens.
- Manual analysis API stub that returns structured report JSON from submitted form data.
- Health endpoint, environment example, Docker Compose for PostgreSQL, Makefile, and README.

This first build does not include production LINE webhook validation, Meta OAuth, Meta Graph data fetching, persistent database models, token encryption, or real AI API calls. Those become later phases after the foundation is stable.

## Architecture

The repository will follow the spec's monorepo boundaries:

- `apps/web` owns the LIFF/mobile UI only.
- `apps/api` owns all backend routes and future integrations.
- `packages/ui` owns the Linora MUI theme and reusable UI primitives.
- `packages/shared` owns TypeScript contracts shared by the web app.

The frontend will use React Router and MUI. It will import theme and reusable UI from `@linora/ui`, and shared report/request/response types from `@linora/shared`. The web app will call the Go API for manual analysis instead of embedding business logic or third-party credentials in the browser.

The backend will expose:

- `GET /health`
- `POST /api/analysis/manual`

The manual endpoint will validate simple manual input, calculate a deterministic starter health score, return report data in the same shape expected from future AI analysis, and keep the route easy to replace with a real AI service later.

## UI Design

The UI will be mobile-only with a maximum content width of `430px`, fixed bottom navigation, a top app bar, and full-width cards. The visual language will use the Emerald, Charcoal, Ivory, White, Border, Gold, and Muted Text palette from the spec.

The first screen will be the actual dashboard, not a landing page. It will show connection status, analysis CTA, today summary, important comments, best posting time, and latest report preview using realistic sample data when no backend report has been created yet.

Manual analysis will be reachable from the dashboard and will collect page name, post content, likes, comments, shares, important comments, and notes. On submit, it will call the backend and route to the report view with the returned structured result.

## Data Contracts

`packages/shared` will define:

- `ManualAnalysisRequest`
- `AnalysisReport`
- `ImportantComment`
- `TopPost`
- `CustomerProfile`
- `FacebookPageSummary`

These contracts are intentionally frontend-safe. They do not include LINE secrets, Facebook access tokens, or backend-only persistence details.

## Error Handling

The web app will show friendly Thai messages for connection and analysis states. API failures in manual analysis will render a visible error state and keep the user's input on screen. The backend will return JSON errors with consistent `error` fields and appropriate HTTP status codes for invalid input.

## Testing And Verification

The first implementation will verify:

- TypeScript build for workspace packages and web app.
- Go tests for the manual analysis handler/service behavior.
- Go build for the API.
- Frontend production build.
- Manual smoke check of the health endpoint if the API server is started.

## Future Phases

After this foundation:

1. Add LINE webhook, follow event, postback handling, and reply/push messages.
2. Add Facebook OAuth and page selection persistence.
3. Add Meta Graph API data fetching.
4. Replace deterministic manual analysis with real AI structured output.
5. Add database persistence, token encryption, scheduled reports, and production hardening.
