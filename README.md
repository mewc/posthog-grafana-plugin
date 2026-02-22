# PostHog Data Source for Grafana

Query [PostHog](https://posthog.com) data directly from Grafana using **HogQL** (PostHog's SQL dialect).

## Features

- Write HogQL queries with a Monaco SQL editor
- Time range macros (`$__timeFrom`, `$__timeTo`) integrate with Grafana's time picker
- Supports Grafana dashboard variables
- Backend plugin — API keys are stored encrypted, queries run server-side
- Supports Grafana alerting

## Installation

### From GitHub Release

```bash
grafana-cli --pluginUrl https://github.com/mewc/posthog-grafana-plugin/releases/download/v1.0.0/mewc-posthog-datasource-v1.0.0.zip plugins install mewc-posthog-datasource
```

### Manual

1. Download the latest release zip
2. Extract to your Grafana plugins directory (e.g. `/var/lib/grafana/plugins/`)
3. Add to `grafana.ini`: `allow_loading_unsigned_plugins = mewc-posthog-datasource`
4. Restart Grafana

## Configuration

1. Go to **Connections → Data sources → Add data source**
2. Search for "PostHog"
3. Configure:
   - **PostHog Instance**: US Cloud, EU Cloud, or Custom URL
   - **Project ID**: Found in PostHog → Settings → Project
   - **API Key**: A Personal API Key (create at Settings → Personal API Keys)
4. Click **Save & Test**

## Query Examples

### Events over time

```sql
SELECT
  toStartOfDay(timestamp) AS day,
  count() AS event_count
FROM events
WHERE timestamp >= $__timeFrom AND timestamp < $__timeTo
GROUP BY day
ORDER BY day
```

### Top events

```sql
SELECT
  event,
  count() AS count
FROM events
WHERE timestamp >= $__timeFrom AND timestamp < $__timeTo
GROUP BY event
ORDER BY count DESC
LIMIT 10
```

### Unique users per day

```sql
SELECT
  toStartOfDay(timestamp) AS day,
  count(DISTINCT distinct_id) AS unique_users
FROM events
WHERE timestamp >= $__timeFrom AND timestamp < $__timeTo
GROUP BY day
ORDER BY day
```

## Time Range Macros

| Macro | Description |
|-------|-------------|
| `$__timeFrom` | Start of the selected Grafana time range |
| `$__timeTo` | End of the selected Grafana time range |

## Development

### Prerequisites

- Node.js >= 22
- Go >= 1.23
- [Mage](https://magefile.org/)
- Docker & Docker Compose

### Setup

```bash
npm install
go mod tidy
```

### Build

```bash
# Frontend
npm run build

# Backend (all platforms)
mage -v buildAll

# Development mode with hot reload
npm run dev    # frontend watcher
docker compose up --build  # Grafana + backend
```

### Test

```bash
# Frontend
npm run test:ci

# Backend
go test ./pkg/...
```

Grafana runs at http://localhost:3000 (anonymous admin access enabled in dev mode).

## License

Apache-2.0
