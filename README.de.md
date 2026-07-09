# Trail‑Finder MCP

[English](README.md) | [日本語](README.ja.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [한국어](README.ko.md) | [Español](README.es.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Português](README.pt.md)

**Version:** [v0.1.0](https://github.com/Takamasa045/Trail-Finder-MCP/releases/tag/v0.1.0)

Claude-Code-MCP-Server, der Open Data zu **Wanderwegeinstiegen / Wegweisern / Wasserstellen**, **Fußrouten**, **Höhe** und **Wetter** bündelt.
Direkt aus Claude Code aufrufbar, um Geoinformationen für Wandern und Bergtouren zu holen.

---

## 🚀 Schnellstart

### 1) Binary bauen

```bash
cd /path/to/trail-finder-mcp
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

Erfordert **Go 1.23+**.

### 2) Claude Code MCP konfigurieren

Füge Folgendes zu `~/.config/claude-code/mcp_config.json` hinzu:

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

- `command` muss der absolute Pfad des gebauten Binary sein
- Umgebungsvariablen bei Bedarf anpassen (siehe unten)

### 3) Claude Code starten

Nach dem Speichern der Konfiguration startet Claude Code den MCP-Server `trail-finder` automatisch mit diesen Tools:

| Tool | Beschreibung |
|------|--------------|
| `trailheads` | Nahe Wanderwegeinstiege / Wegweiser / Wasserstellen suchen |
| `route_foot` | Fußroute zwischen zwei Punkten berechnen |
| `elevation` | Höhe eines Ortes abrufen |
| `forecast` | Kurzfristige Wettervorhersage abrufen |

---

## 🛠️ Verfügbare Tools

### `trailheads`

Sucht nahe Wanderwegeinstiege, Wegweiser, Wasserstellen und ähnliche POIs.

**Eingabeparameter:**
- `lat`, `lon` — Mittelpunktskoordinaten
- `radius_m` — Suchradius in Metern
- `include` — einzuschließende POI-Typen (z. B. `["guidepost", "trailhead"]`)
- `also_water` — Wasserstellen einschließen (`true` / `false`)
- `limit` — maximale Trefferanzahl

**Ausgabe:**
Nahe POIs als JSON (Typ, Name, Koordinaten usw.)

### `route_foot`

Berechnet die kürzeste Fußroute zwischen zwei Punkten.

**Eingabeparameter:**
- `from` — Start `{lat, lon}`
- `to` — Ziel `{lat, lon}`
- `engine` — Routing-Engine (`"auto"`, `"osrm"`, `"valhalla"`)
- `options` — zusätzliche Optionen
  - `include_geometry` (Standard: `true`) — Routengeometrie einbeziehen
  - `include_steps` (Standard: `false`) — OSRM-Schrittdetails einbeziehen
  - `avoid_ferry` (Standard: `false`) — Fähren meiden (OSRM `exclude=ferry`)

**Ausgabe:**
Distanz, Dauer und GeoJSON-LineString-Geometrie

### `elevation`

Ruft die Höhe eines Punkts ab.

**Eingabeparameter:**
- `lat`, `lon` — Koordinaten

**Ausgabe:**
Höhe in Metern

### `forecast`

Ruft eine kurzfristige Wettervorhersage für einen Punkt ab.

**Eingabeparameter:**
- `lat`, `lon` — Koordinaten
- `hours` — Vorhersagehorizont in Stunden

**Ausgabe:**
Zeitreihe von Temperatur, Niederschlag und Windgeschwindigkeit (Open-Meteo)

---

## ⚙️ Umgebungsvariablen

In einer `.env`-Datei oder im `env`-Abschnitt der MCP-Konfiguration setzen:

```bash
TRAILFINDER_OVERPASS_URL=https://overpass-api.de/api/interpreter
OSRM_URL=https://router.project-osrm.org
VALHALLA_URL=                                      # bei eigenem Valhalla
ELEVATION_PROVIDER=open-elevation                  # open-elevation | open-topo
OPENMETEO_URL=https://api.open-meteo.com/v1/forecast
DEFAULT_TZ=Asia/Tokyo
TRAILFINDER_USER_AGENT=trail-finder-mcp/0.1.0 (+your-contact)
```

Vorlage siehe `.env.example`.

---

## 📋 Nutzungsbeispiele

Stelle Claude Code eine Frage in natürlicher Sprache; es wählt das passende Tool:

- „Finde Wanderwegeinstiege in der Nähe des Bahnhofs Takao-sanguchi“ → `trailheads`
- „Route vom Bahnhof Takao-sanguchi zum Gipfel des Mt. Takao“ → `route_foot`
- „Wie hoch ist der Mt. Takao?“ → `elevation`
- „Wie ist das Wetter heute am Mt. Takao?“ → `forecast`

---

## 🖥️ Claude Desktop (optional)

Du kannst diesen Server auch mit Claude Desktop nutzen. Pfade der Konfigurationsdatei:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

Der Konfigurationsinhalt entspricht dem von Claude Code.

---

## 📦 Releases

Getaggte Releases werden auf GitHub veröffentlicht:

- [Neuestes Release](https://github.com/Takamasa045/Trail-Finder-MCP/releases/latest)
- Aktuelle Version: **v0.1.0**

```bash
git clone https://github.com/Takamasa045/Trail-Finder-MCP.git
cd Trail-Finder-MCP
git checkout v0.1.0
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

---

## ⚠️ Hinweise

- Öffentliche OSRM-/Overpass-APIs schwanken in Last und Latenz. Für den Produktivbetrieb eigene Instanzen bevorzugen.
- Die zurückgegebenen Daten dienen nur zur Orientierung. Vor Ort haben lokale Bedingungen und Sicherheit Vorrang.
- Dieses Tool **garantiert keine** Wandersicherheit. Plane und bereite dich immer vor.

---

## 📝 Datenquellen & Lizenz

Externe APIs und Datenquellen dieses Projekts:

- **OpenStreetMap / Overpass API** — Kartendaten (ODbL)
- **OSRM** — Routing-Engine
- **Open-Meteo** — Wettervorhersagedaten
- **Open-Elevation / OpenTopoData** — Höhendaten
