# PostHog Data Source for Grafana

Query [PostHog](https://posthog.com) data directly from Grafana using **HogQL** (PostHog's SQL dialect).

## Features

- HogQL query editor with Monaco SQL highlighting
- Time range macros (`$__timeFrom`, `$__timeTo`) integrate with Grafana's time picker
- Grafana dashboard variable support
- Go backend — API keys stored encrypted, queries execute server-side
- Alerting support

## Installation

### From GitHub Release

```bash
grafana-cli \
  --pluginUrl https://github.com/mewc/posthog-grafana-plugin/releases/download/v1.0.0/chartcastr-posthog-datasource-v1.0.0.zip \
  plugins install chartcastr-posthog-datasource
```

Since this plugin is not yet signed via the Grafana catalog, you need to allow it explicitly.

Add to `grafana.ini`:

```ini
[plugins]
allow_loading_unsigned_plugins = chartcastr-posthog-datasource
```

Or as an environment variable (useful for Docker/Kubernetes):

```bash
GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=chartcastr-posthog-datasource
```

Restart Grafana after installing.

### Manual install

1. Download the latest zip from [Releases](https://github.com/mewc/posthog-grafana-plugin/releases)
2. Extract into your Grafana plugins directory (default: `/var/lib/grafana/plugins/`)
3. Allow the unsigned plugin (see above)
4. Restart Grafana

## Configuration

1. **Connections → Data sources → Add data source** → search "PostHog"
2. Configure:
   - **PostHog Instance** — US Cloud, EU Cloud, or Custom (self-hosted) URL
   - **Project ID** — found in PostHog → Settings → Project
   - **API Key** — a **Personal API Key** (PostHog → Settings → Personal API Keys). This is _not_ the project API key.
3. **Save & Test** — should show a green checkmark

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

### Time range macros

| Macro | Replaced with |
|-------|---------------|
| `$__timeFrom` | Start of selected Grafana time range, e.g. `'2024-01-15 10:30:00'` |
| `$__timeTo` | End of selected Grafana time range |

## Development

### Prerequisites

- Node.js >= 22
- Go >= 1.23
- [Mage](https://magefile.org/) — `go install github.com/magefile/mage@latest`
- Docker & Docker Compose

### Quick start

```bash
# Install dependencies
npm install && go mod tidy

# Build frontend + backend
npm run build && mage -v buildAll

# Start Grafana with the plugin loaded
docker compose up --build
```

Open http://localhost:3000 — anonymous admin access is enabled in dev mode.

### Development mode (hot reload)

```bash
npm run dev                    # terminal 1: frontend watcher
docker compose up --build      # terminal 2: Grafana + backend auto-rebuild
```

The Docker Compose setup uses [supervisord](https://github.com/grafana/grafana-plugin-sdk-go) to watch for backend changes, rebuild via `mage`, and attach a [Delve](https://github.com/go-delve/delve) debugger on port 2345.

### Auto-provisioning

The plugin is auto-provisioned via `provisioning/datasources/default.yaml`. Set these env vars before `docker compose up` to skip manual configuration:

```bash
export POSTHOG_URL=https://us.posthog.com       # or eu.posthog.com, or self-hosted
export POSTHOG_PROJECT_ID=12345
export POSTHOG_API_KEY=phx_your_personal_api_key
```

Or just configure the datasource manually in the Grafana UI.

### Tests

```bash
npm run typecheck       # TypeScript type checking
npm run lint            # ESLint
npm run test:ci         # Jest (frontend)
go test ./pkg/...       # Go (backend)
```

## Distribution

### GitHub Release (recommended)

Push a version tag to trigger the automated release workflow:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The [release workflow](https://github.com/mewc/posthog-grafana-plugin/blob/main/.github/workflows/release.yml) builds frontend + backend, packages a zip, and creates a GitHub Release. Users install with:

```bash
grafana-cli --pluginUrl <zip-url> plugins install chartcastr-posthog-datasource
```

### Manual packaging

```bash
npm run build && mage -v buildAll
cp -r dist chartcastr-posthog-datasource
zip -r chartcastr-posthog-datasource-1.0.0.zip chartcastr-posthog-datasource/
```

Install the zip on any Grafana instance by extracting into the plugins directory and setting `allow_loading_unsigned_plugins`.

### Optional: Plugin signing

Signing is **not required** for self-hosted Grafana with `allow_loading_unsigned_plugins`. It is only required if you want to:

- Distribute without requiring users to modify `grafana.ini`
- Submit to the [Grafana plugin catalog](https://grafana.com/grafana/plugins/)

To sign, you need a free [Grafana Cloud](https://grafana.com) account (just for the signing token — no Grafana Cloud instance needed):

1. Sign up at [grafana.com](https://grafana.com)
2. Go to **My Account → Security → Access Policies**
3. Create a policy with scope `plugins:write`, then create a token
4. Add the token as a GitHub repo secret: `GRAFANA_ACCESS_POLICY_TOKEN`

The release workflow will automatically sign the plugin when this secret is present.

### Optional: Grafana plugin catalog submission

To publish to the official Grafana plugin catalog (so users can install via `grafana-cli plugins install` without a URL):

1. Ensure the repo is **public** and the plugin is **signed** (see above)
2. Validate: `npx -y @grafana/plugin-validator@latest -sourceCodeUri https://github.com/mewc/posthog-grafana-plugin chartcastr-posthog-datasource-1.0.0.zip`
3. Submit at [grafana.com](https://grafana.com) → **Org Settings → My Plugins → Submit New Plugin**
4. Provide the release zip URL, source repo URL, and SHA1 hash (`sha1sum *.zip`)
5. Grafana team reviews (automated checks + manual code review)

## License

Apache-2.0
