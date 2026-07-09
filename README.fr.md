# Trail‑Finder MCP

[English](README.md) | [日本語](README.ja.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [한국어](README.ko.md) | [Español](README.es.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Português](README.pt.md)

**Version :** [v0.1.0](https://github.com/Takamasa045/Trail-Finder-MCP/releases/tag/v0.1.0)

Serveur MCP pour Claude Code qui agrège des données ouvertes sur les **départs de sentiers / balises / points d’eau**, les **itinéraires à pied**, l’**altitude** et la **météo**.
Appelez-le directement depuis Claude Code pour obtenir des informations géographiques utiles à la randonnée et à la montagne.

---

## 🚀 Démarrage rapide

### 1) Compiler le binaire

```bash
cd /path/to/trail-finder-mcp
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

Nécessite **Go 1.23+**.

### 2) Configurer Claude Code MCP

Ajoutez ce qui suit à `~/.config/claude-code/mcp_config.json` :

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

- `command` doit être le chemin absolu du binaire compilé
- Personnalisez les variables d’environnement si besoin (voir ci-dessous)

### 3) Lancer Claude Code

Après avoir enregistré la configuration, lancez Claude Code. Le serveur MCP `trail-finder` se charge automatiquement et expose ces outils :

| Outil | Description |
|-------|-------------|
| `trailheads` | Rechercher les départs de sentiers / balises / points d’eau à proximité |
| `route_foot` | Calculer un itinéraire à pied entre deux points |
| `elevation` | Obtenir l’altitude d’un lieu |
| `forecast` | Obtenir une prévision météo à court terme |

---

## 🛠️ Outils disponibles

### `trailheads`

Recherche les départs de sentiers, balises, points d’eau et POI similaires à proximité.

**Paramètres d’entrée :**
- `lat`, `lon` — coordonnées du centre
- `radius_m` — rayon de recherche en mètres
- `include` — types de POI à inclure (ex. `["guidepost", "trailhead"]`)
- `also_water` — inclure les points d’eau (`true` / `false`)
- `limit` — nombre maximal de résultats

**Sortie :**
POI à proximité en JSON (type, nom, coordonnées, etc.)

### `route_foot`

Calcule l’itinéraire à pied le plus court entre deux points.

**Paramètres d’entrée :**
- `from` — origine `{lat, lon}`
- `to` — destination `{lat, lon}`
- `engine` — moteur de routage (`"auto"`, `"osrm"`, `"valhalla"`)
- `options` — options supplémentaires
  - `include_geometry` (défaut : `true`) — inclure la géométrie de l’itinéraire
  - `include_steps` (défaut : `false`) — inclure le détail des étapes OSRM
  - `avoid_ferry` (défaut : `false`) — éviter les ferries (OSRM `exclude=ferry`)

**Sortie :**
Distance, durée et géométrie GeoJSON LineString

### `elevation`

Obtient l’altitude d’un point.

**Paramètres d’entrée :**
- `lat`, `lon` — coordonnées

**Sortie :**
Altitude en mètres

### `forecast`

Obtient une prévision météo à court terme pour un point.

**Paramètres d’entrée :**
- `lat`, `lon` — coordonnées
- `hours` — horizon de prévision en heures

**Sortie :**
Série temporelle de température, précipitations et vitesse du vent (Open-Meteo)

---

## ⚙️ Variables d’environnement

Définissez-les dans un fichier `.env` ou dans la section `env` de la configuration MCP :

```bash
TRAILFINDER_OVERPASS_URL=https://overpass-api.de/api/interpreter
OSRM_URL=https://router.project-osrm.org
VALHALLA_URL=                                      # si vous utilisez votre propre Valhalla
ELEVATION_PROVIDER=open-elevation                  # open-elevation | open-topo
OPENMETEO_URL=https://api.open-meteo.com/v1/forecast
DEFAULT_TZ=Asia/Tokyo
TRAILFINDER_USER_AGENT=trail-finder-mcp/0.1.0 (+your-contact)
```

Voir `.env.example` pour un modèle.

---

## 📋 Exemples d’utilisation

Posez une question en langage naturel à Claude Code ; il choisira l’outil adapté :

- « Trouve les départs de sentiers près de la gare Takao-sanguchi » → `trailheads`
- « Itinéraire de la gare Takao-sanguchi au sommet du mont Takao » → `route_foot`
- « Quelle est l’altitude du mont Takao ? » → `elevation`
- « Quel temps fait-il aujourd’hui sur le mont Takao ? » → `forecast`

---

## 🖥️ Claude Desktop (optionnel)

Vous pouvez aussi utiliser ce serveur avec Claude Desktop. Chemins du fichier de configuration :

- **macOS** : `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows** : `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux** : `~/.config/Claude/claude_desktop_config.json`

Le contenu de la configuration est le même que pour Claude Code.

---

## 📦 Versions

Les versions étiquetées sont publiées sur GitHub :

- [Dernière version](https://github.com/Takamasa045/Trail-Finder-MCP/releases/latest)
- Version actuelle : **v0.1.0**

```bash
git clone https://github.com/Takamasa045/Trail-Finder-MCP.git
cd Trail-Finder-MCP
git checkout v0.1.0
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

---

## ⚠️ Avertissements

- Les API publiques OSRM / Overpass varient en charge et en latence. Pour la production, préférez des instances auto-hébergées.
- Les données renvoyées sont indicatives. Sur le terrain, privilégiez les conditions locales et la sécurité.
- Cet outil **ne garantit pas** la sécurité en randonnée. Planifiez et préparez-vous toujours.

---

## 📝 Sources de données et licence

API et sources de données externes utilisées par ce projet :

- **OpenStreetMap / Overpass API** — données cartographiques (ODbL)
- **OSRM** — moteur de routage
- **Open-Meteo** — données de prévision météo
- **Open-Elevation / OpenTopoData** — données d’altitude
