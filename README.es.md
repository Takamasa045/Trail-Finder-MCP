# Trail‑Finder MCP

[English](README.md) | [日本語](README.ja.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [한국어](README.ko.md) | [Español](README.es.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Português](README.pt.md)

**Versión:** [v0.1.0](https://github.com/Takamasa045/Trail-Finder-MCP/releases/tag/v0.1.0)

Servidor MCP para Claude Code que agrega datos abiertos de **accesos a senderos / postes indicadores / fuentes de agua**, **rutas a pie**, **elevación** y **clima**.
Llámalo directamente desde Claude Code para obtener información geográfica útil para senderismo y montañismo.

---

## 🚀 Inicio rápido

### 1) Compilar el binario

```bash
cd /path/to/trail-finder-mcp
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

Requiere **Go 1.23+**.

### 2) Configurar Claude Code MCP

Añade lo siguiente a `~/.config/claude-code/mcp_config.json`:

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

- `command` debe ser la ruta absoluta del binario compilado
- Personaliza las variables de entorno según necesites (ver abajo)

### 3) Iniciar Claude Code

Tras guardar la configuración, inicia Claude Code. El servidor MCP `trail-finder` se carga automáticamente y expone estas herramientas:

| Herramienta | Descripción |
|-------------|-------------|
| `trailheads` | Buscar accesos a senderos / postes / fuentes de agua cercanos |
| `route_foot` | Calcular una ruta a pie entre dos puntos |
| `elevation` | Obtener la elevación de un lugar |
| `forecast` | Obtener un pronóstico del tiempo a corto plazo |

---

## 🛠️ Herramientas disponibles

### `trailheads`

Busca accesos a senderos, postes indicadores, fuentes de agua y POIs similares.

**Parámetros de entrada:**
- `lat`, `lon` — coordenadas del centro
- `radius_m` — radio de búsqueda en metros
- `include` — tipos de POI a incluir (p. ej. `["guidepost", "trailhead"]`)
- `also_water` — incluir fuentes de agua (`true` / `false`)
- `limit` — número máximo de resultados

**Salida:**
POIs cercanos en JSON (tipo, nombre, coordenadas, etc.)

### `route_foot`

Calcula la ruta a pie más corta entre dos puntos.

**Parámetros de entrada:**
- `from` — origen `{lat, lon}`
- `to` — destino `{lat, lon}`
- `engine` — motor de enrutamiento (`"auto"`, `"osrm"`, `"valhalla"`)
- `options` — opciones adicionales
  - `include_geometry` (predeterminado: `true`) — incluir geometría de la ruta
  - `include_steps` (predeterminado: `false`) — incluir detalles de pasos OSRM
  - `avoid_ferry` (predeterminado: `false`) — evitar ferris (OSRM `exclude=ferry`)

**Salida:**
Distancia, duración y geometría GeoJSON LineString

### `elevation`

Obtiene la elevación de un punto.

**Parámetros de entrada:**
- `lat`, `lon` — coordenadas

**Salida:**
Elevación en metros

### `forecast`

Obtiene un pronóstico del tiempo a corto plazo para un punto.

**Parámetros de entrada:**
- `lat`, `lon` — coordenadas
- `hours` — horizonte del pronóstico en horas

**Salida:**
Serie temporal de temperatura, precipitación y velocidad del viento (Open-Meteo)

---

## ⚙️ Variables de entorno

Configúralas en un archivo `.env` o en la sección `env` de la configuración MCP:

```bash
TRAILFINDER_OVERPASS_URL=https://overpass-api.de/api/interpreter
OSRM_URL=https://router.project-osrm.org
VALHALLA_URL=                                      # si usas tu propio Valhalla
ELEVATION_PROVIDER=open-elevation                  # open-elevation | open-topo
OPENMETEO_URL=https://api.open-meteo.com/v1/forecast
DEFAULT_TZ=Asia/Tokyo
TRAILFINDER_USER_AGENT=trail-finder-mcp/0.1.0 (+your-contact)
```

Consulta `.env.example` como plantilla.

---

## 📋 Ejemplos de uso

Pregunta a Claude Code en lenguaje natural; elegirá la herramienta adecuada:

- “Encuentra accesos a senderos cerca de la estación Takao-sanguchi” → `trailheads`
- “Ruta desde la estación Takao-sanguchi hasta la cima del monte Takao” → `route_foot`
- “¿Cuál es la elevación del monte Takao?” → `elevation`
- “¿Qué tiempo hace hoy en el monte Takao?” → `forecast`

---

## 🖥️ Claude Desktop (opcional)

También puedes usar este servidor con Claude Desktop. Rutas del archivo de configuración:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

El contenido de la configuración es el mismo que para Claude Code.

---

## 📦 Publicaciones

Las versiones etiquetadas se publican en GitHub:

- [Última versión](https://github.com/Takamasa045/Trail-Finder-MCP/releases/latest)
- Versión actual: **v0.1.0**

```bash
git clone https://github.com/Takamasa045/Trail-Finder-MCP.git
cd Trail-Finder-MCP
git checkout v0.1.0
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

---

## ⚠️ Avisos

- Las API públicas de OSRM / Overpass varían en carga y latencia. Para producción, prefiere instancias propias.
- Los datos devueltos son solo de referencia. En la montaña, prioriza las condiciones locales y la seguridad.
- Esta herramienta **no garantiza** la seguridad en el senderismo. Planifica y prepárate siempre.

---

## 📝 Fuentes de datos y licencia

API y fuentes de datos externas usadas por este proyecto:

- **OpenStreetMap / Overpass API** — datos cartográficos (ODbL)
- **OSRM** — motor de enrutamiento
- **Open-Meteo** — datos de pronóstico del tiempo
- **Open-Elevation / OpenTopoData** — datos de elevación
