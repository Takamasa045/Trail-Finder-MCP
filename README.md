# Trail‑Finder MCP

[English](README.md) | [日本語](README.ja.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [한국어](README.ko.md) | [Español](README.es.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Português](README.pt.md)

**Version:** 0.2.1 (English and Japanese are the canonical docs)

Claude / Codex MCP server that aggregates open data for **place names**, **trailheads / guideposts / huts / water**, **walking routes**, **elevation**, and **weather**.

Call it from Claude Code, Claude Desktop, Codex, or Grok. For a full hiking plan from station or mountain names, use `plan_hike`.

---

## Quick Start

Requires **Go 1.23+**.

```bash
git clone https://github.com/Takamasa045/Trail-Finder-MCP.git
cd Trail-Finder-MCP
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

Add the binary to your MCP client. Example (Claude Code / Claude Desktop / Grok-style JSON):

```json
{
  "mcpServers": {
    "trail-finder": {
      "command": "/absolute/path/to/trail-finder-mcp",
      "env": {
        "DEFAULT_TZ": "Asia/Tokyo",
        "DEFAULT_LANG": "ja",
        "ELEVATION_PROVIDER": "open-meteo",
        "TRAILFINDER_USER_AGENT": "trail-finder-mcp/0.2.1 (+https://github.com/Takamasa045/Trail-Finder-MCP)"
      }
    }
  }
}
```

Config file locations:

- **Claude Code**: `~/.claude.json` (`mcpServers`)
- **Claude Desktop (macOS)**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Codex**: `~/.codex/config.toml` under `[mcp_servers.trail-finder]`

Codex uses TOML rather than the JSON format above:

```toml
[mcp_servers.trail-finder]
command = "/absolute/path/to/trail-finder-mcp"

[mcp_servers.trail-finder.env]
DEFAULT_TZ = "Asia/Tokyo"
DEFAULT_LANG = "ja"
ELEVATION_PROVIDER = "open-meteo"
```

Replace `command` with the absolute path to the built binary. Restart the client and check that all six tools appear. Running the binary without arguments waits for MCP messages on stdin/stdout; it does not open a browser UI. `.env` files are not loaded automatically: set variables in the client's `env` configuration or your shell. The repository's `mcp.json` metadata assumes `go run` from the repository root; use the absolute binary path above when launching from another directory.

Optional HTTP JSON API (debug only, binds to 127.0.0.1). This is not an MCP HTTP transport; configure MCP clients using stdio as above:

```bash
./trail-finder-mcp -http :8080
curl -s localhost:8080/healthz
curl -s -X POST localhost:8080/tools/geocode -H 'Content-Type: application/json' -d '{"query":"高尾山口駅"}'
```

---

## Tools

| Tool | Description |
|------|-------------|
| `geocode` | Place name → coordinates (Nominatim). Japanese names such as 高尾山口駅 work. |
| `trailheads` | Nearby trailheads, guideposts, optional huts/passes/water (Overpass / OSM). |
| `route_foot` | Walking route via OSRM, with geometry and elevation gain/loss when available. |
| `elevation` | Elevation in meters (default: Open-Meteo). |
| `forecast` | Hourly hiking forecast: temperature, rain, rain probability, wind **m/s**, gusts, weather code, sunrise/sunset. |
| `plan_hike` | Combined plan: geocode from/to, route, POIs along the path, gain/loss, forecast at start and goal. Water is included unless `also_water` is false. |

`engine=valhalla` is **not** implemented. Use `auto` or `osrm`.

### `plan_hike` input

```json
{
  "from": { "query": "高尾山口駅" },
  "to": { "query": "高尾山" },
  "also_water": true,
  "hours": 24
}
```

Coordinates also work: `{ "lat": 35.63, "lon": 139.27 }`. If both `query` and coordinates are set, `query` wins.

`forecast` is for the start; `forecast_goal` is for the destination. Partial failures (POIs or weather) show up in `warnings` instead of failing the whole plan.

### `trailheads` include values

`guidepost`, `trailhead`, `shelter`, `pass`, `entrance` (opt-in; noisy). Default is guidepost + trailhead. Japan-oriented tags include `highway=trailhead` and `tourism=information` + `information=guidepost`. `entrance=yes` is **not** queried unless requested.

---

## Reading results and recovering from errors

- For ambiguous place names, inspect `geocode` results (`display_name` and coordinates) and pass the selected coordinates to `plan_hike`. A name passed directly to `plan_hike` uses the first match.
- `trailheads.items` are sorted by distance. An empty array means the search completed without matching mapped data; OSM coverage may be incomplete. Try a wider radius or different `include` values.
- Overpass timeouts/runtime errors are reported as errors rather than empty results. Reduce the radius/corridor or retry later.
- Empty or incomplete weather data is an error, never zero wind/rain. `plan_hike` keeps the route and adds separate start/goal forecast failures to `warnings`. Do not treat a forecast with a failure warning as available weather data.
- MCP execution failures return `isError: true` with explanatory text; malformed arguments may return a protocol error. The debug HTTP API returns 400 for invalid input and 502 for upstream failures. POST bodies must be a single JSON object, at most 1 MiB.

---

## Example prompts

- “高尾山口駅から高尾山までの計画を” → `plan_hike`
- “高尾山口駅の座標は？” → `geocode`
- “その周辺の登山口と水場” → `trailheads`
- “高尾山の今日の風と降水確率” → `forecast`

---

## Environment variables

See `.env.example`. Important ones:

| Variable | Default |
|----------|---------|
| `TRAILFINDER_OVERPASS_URL` | `https://overpass-api.de/api/interpreter` |
| `OSRM_URL` | `https://router.project-osrm.org` |
| `NOMINATIM_URL` | `https://nominatim.openstreetmap.org/search` |
| `ELEVATION_PROVIDER` | `open-meteo` (`open-elevation`, `open-topo`, `off`) |
| `OPENMETEO_URL` | `https://api.open-meteo.com/v1/forecast` |
| `OPENMETEO_ELEVATION_URL` | `https://api.open-meteo.com/v1/elevation` |
| `DEFAULT_TZ` | `Asia/Tokyo` |
| `DEFAULT_LANG` | `ja` |
| `TRAILFINDER_USER_AGENT` | identifies this client (required politeness for OSM) |

Public Overpass / OSRM / Nominatim instances vary in load. For production, prefer self-hosted endpoints.

---

## Develop

```bash
go test ./...
```

HTTP retries (429/502/503/504, 2 extra attempts) and a short in-memory cache for Overpass / Nominatim are built in.

License: MIT. Spec notes: [docs/requirements.md](docs/requirements.md).

---

## Notes

Returned data is for **reference only**. On the trail, prioritize local conditions, gear, and official notices. This tool does **not** guarantee hiking safety.

## Data sources

- OpenStreetMap / Overpass / Nominatim (ODbL)
- OSRM
- Open-Meteo (forecast + elevation)
- Optional Open-Elevation / OpenTopoData
