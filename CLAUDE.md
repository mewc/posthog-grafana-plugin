# PostHog Grafana Data Source Plugin

## Architecture

- **Go backend** (`pkg/`): Handles PostHog API auth, executes HogQL queries, converts results to Grafana data frames
- **TypeScript frontend** (`src/`): Config UI, HogQL query editor with Monaco SQL editor
- **Plugin ID**: `chartcastr-posthog-datasource`
- **Executable**: `gpx_posthog`

## Key Files

```
pkg/main.go                    → Go entry point
pkg/plugin/datasource.go       → QueryData + CheckHealth handlers, time macro expansion, type mapping
pkg/plugin/posthog_client.go   → HTTP client for PostHog /api/projects/:id/query/
pkg/plugin/models.go           → Settings, query, API request/response structs
src/module.ts                  → Frontend plugin registration
src/datasource.ts              → DataSourceWithBackend subclass, template variable support
src/components/ConfigEditor.tsx → URL/Project ID/API Key config form
src/components/QueryEditor.tsx  → Monaco SQL editor for HogQL
src/plugin.json                → Plugin manifest
```

## Commands

```bash
npm install              # install frontend deps
npm run build            # build frontend (webpack production)
npm run dev              # frontend dev watcher
npm run typecheck        # TypeScript type check
npm run lint             # ESLint
npm run test:ci          # Jest tests

go test ./pkg/...        # backend unit tests
mage -v buildAll         # build Go backend for all platforms
docker compose up --build # run Grafana dev environment
```

## PostHog API

Queries use `POST /api/projects/{projectId}/query/` with body:
```json
{"query": {"kind": "HogQLQuery", "query": "SELECT ..."}}
```

Auth: `Authorization: Bearer {personalApiKey}`

Response: `{columns: string[], types: string[], results: any[][]}`

## Time Macros

`$__timeFrom` and `$__timeTo` are replaced with `'YYYY-MM-DD HH:MM:SS'` formatted UTC timestamps from Grafana's time range picker.

## Type Mapping

ClickHouse types from PostHog → Grafana field types:
- DateTime, Date → `time.Time` (nullable)
- Int*, UInt*, Float*, Decimal → `float64` (nullable)
- Bool → `bool` (nullable)
- Everything else → `string` (nullable)
