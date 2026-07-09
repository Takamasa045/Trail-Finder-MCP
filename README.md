# Trail‑Finder MCP

[English](README.md) | [日本語](README.ja.md)

**Version:** [v0.1.0](https://github.com/Takamasa045/Trail-Finder-MCP/releases/tag/v0.1.0)

Claude Code MCP server that aggregates open data for **trailheads / guideposts / water sources**, **walking routes**, **elevation**, and **weather**.
Call it directly from Claude Code to fetch geo information useful for hiking and mountain trips.

---

## 🚀 Quick Start

### 1) Build the binary

```bash
cd /path/to/trail-finder-mcp
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

Requires **Go 1.23+**.

### 2) Configure Claude Code MCP

Add the following to `~/.config/claude-code/mcp_config.json`:

```jsonc
{
  "mcpServers": {
    "trail-finder": {
      "command": "/absolute/path/to/trail-finder-mcp",
      "env": {
        "TRAILFINDER_OVERPASS_URL": "https://overpass-api.de/api/interpreter",
        "OSRM_URL": "https://router.project-osrm.org",
        "OPENMETEO_URL": "https://api.open-meteo.com/v1/forecast",
        "DEFAULT_TZ": "Asia/Tokyo"
      }
    }
  }
}
```

- Set `command` to the absolute path of the built binary
- Customize environment variables as needed (see below)

### 3) Start Claude Code

After saving the config, start Claude Code. The `trail-finder` MCP server loads automatically and exposes these tools:

| Tool | Description |
|------|-------------|
| `trailheads` | Search nearby trailheads / guideposts / water sources |
| `route_foot` | Compute a walking route between two points |
| `elevation` | Get elevation for a location |
| `forecast` | Get a short-range weather forecast |

---

## 🛠️ Available Tools

### `trailheads`

Search nearby trailheads, guideposts, water sources, and similar POIs.

**Input:**
- `lat`, `lon` — center coordinates
- `radius_m` — search radius in meters
- `include` — POI types to include (e.g. `["guidepost", "trailhead"]`)
- `also_water` — whether to include water sources (`true` / `false`)
- `limit` — max number of results

**Output:**
Nearby POIs as JSON (type, name, coordinates, etc.)

### `route_foot`

Compute the shortest walking route between two points.

**Input:**
- `from` — origin `{lat, lon}`
- `to` — destination `{lat, lon}`
- `engine` — routing engine (`"auto"`, `"osrm"`, `"valhalla"`)
- `options` — extra options
  - `include_geometry` (default: `true`) — include route geometry
  - `include_steps` (default: `false`) — include OSRM step details
  - `avoid_ferry` (default: `false`) — avoid ferries (OSRM `exclude=ferry`)

**Output:**
Distance, duration, and GeoJSON LineString geometry

### `elevation`

Get elevation for a point.

**Input:**
- `lat`, `lon` — coordinates

**Output:**
Elevation in meters

### `forecast`

Get a short-range weather forecast for a point.

**Input:**
- `lat`, `lon` — coordinates
- `hours` — forecast horizon in hours

**Output:**
Time series of temperature, precipitation, and wind speed (Open-Meteo)

---

## ⚙️ Environment Variables

Set these in a `.env` file or in the MCP config `env` section:

```bash
TRAILFINDER_OVERPASS_URL=https://overpass-api.de/api/interpreter
OSRM_URL=https://router.project-osrm.org
VALHALLA_URL=                                      # set if you run your own Valhalla
ELEVATION_PROVIDER=open-elevation                  # open-elevation | open-topo
OPENMETEO_URL=https://api.open-meteo.com/v1/forecast
DEFAULT_TZ=Asia/Tokyo
TRAILFINDER_USER_AGENT=trail-finder-mcp/0.1.0 (+your-contact)
```

See `.env.example` for a starter template.

---

## 📋 Example Prompts

Ask Claude Code in natural language; it will pick the right tool:

- “Find trailheads near Takao-sanguchi Station” → `trailheads`
- “Route from Takao-sanguchi Station to the summit of Mt. Takao” → `route_foot`
- “What is the elevation of Mt. Takao?” → `elevation`
- “What’s the weather on Mt. Takao today?” → `forecast`

---

## 🖥️ Claude Desktop (optional)

You can also use this server with Claude Desktop. Config file paths:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

The config content is the same as for Claude Code.

---

## 📦 Releases

Tagged releases are published on GitHub:

- [Latest release](https://github.com/Takamasa045/Trail-Finder-MCP/releases/latest)
- Current version: **v0.1.0**

```bash
git clone https://github.com/Takamasa045/Trail-Finder-MCP.git
cd Trail-Finder-MCP
git checkout v0.1.0
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

---

## ⚠️ Notes

- Public OSRM / Overpass APIs vary in load and latency. For production use, prefer self-hosted instances.
- Returned data is for reference only. On the trail, prioritize local conditions and safety.
- This tool does **not** guarantee hiking safety. Always plan and prepare properly.

---

## 📝 Data Sources & License

External APIs and data sources used by this project:

- **OpenStreetMap / Overpass API** — map data (ODbL)
- **OSRM** — routing engine
- **Open-Meteo** — weather forecast data
- **Open-Elevation / OpenTopoData** — elevation data
