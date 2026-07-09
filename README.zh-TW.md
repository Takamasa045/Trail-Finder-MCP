# Trail‑Finder MCP

[English](README.md) | [日本語](README.ja.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [한국어](README.ko.md) | [Español](README.es.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Português](README.pt.md)

**版本:** [v0.1.0](https://github.com/Takamasa045/Trail-Finder-MCP/releases/tag/v0.1.0)

面向 Claude Code 的 MCP 伺服器，彙整開放資料，提供 **登山口 / 路標 / 水源**、**步行路線**、**海拔** 與 **天氣** 資訊。
可直接在 Claude Code 中呼叫，輕鬆取得健行與登山所需的地理資訊。

---

## 🚀 快速開始

### 1) 編譯二進位檔

```bash
cd /path/to/trail-finder-mcp
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

需要 **Go 1.23+**。

### 2) 設定 Claude Code MCP

在 `~/.config/claude-code/mcp_config.json` 中加入：

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

- `command` 請填寫編譯後二進位檔的絕對路徑
- 可依需求自訂環境變數（詳見下文）

### 3) 啟動 Claude Code

儲存設定後啟動 Claude Code，`trail-finder` MCP 伺服器會自動載入並提供以下工具：

| 工具 | 說明 |
|------|------|
| `trailheads` | 搜尋附近登山口 / 路標 / 水源 |
| `route_foot` | 計算兩點之間的步行路線 |
| `elevation` | 取得指定地點的海拔 |
| `forecast` | 取得指定地點的天氣預報 |

---

## 🛠️ 可用工具

### `trailheads`

搜尋附近的登山口、路標、水源等 POI。

**輸入參數:**
- `lat`, `lon` — 搜尋中心座標
- `radius_m` — 搜尋半徑（公尺）
- `include` — 要包含的 POI 類型（例如 `["guidepost", "trailhead"]`）
- `also_water` — 是否包含水源（`true` / `false`）
- `limit` — 最大回傳數量

**輸出:**
以 JSON 回傳附近 POI（類型、名稱、座標等）

### `route_foot`

計算兩點之間的最短步行路線。

**輸入參數:**
- `from` — 起點 `{lat, lon}`
- `to` — 終點 `{lat, lon}`
- `engine` — 路由引擎（`"auto"`, `"osrm"`, `"valhalla"`）
- `options` — 額外選項
  - `include_geometry`（預設: `true`）是否包含路線幾何
  - `include_steps`（預設: `false`）是否包含 OSRM 步驟詳情
  - `avoid_ferry`（預設: `false`）避開渡輪（OSRM `exclude=ferry`）

**輸出:**
距離、所需時間，以及 GeoJSON LineString 路線

### `elevation`

取得指定點的海拔。

**輸入參數:**
- `lat`, `lon` — 座標

**輸出:**
海拔（公尺）

### `forecast`

取得指定點的短時天氣預報。

**輸入參數:**
- `lat`, `lon` — 座標
- `hours` — 預報小時數

**輸出:**
溫度、降水量、風速的時間序列資料（Open-Meteo）

---

## ⚙️ 環境變數

可在 `.env` 檔案或 MCP 設定的 `env` 中設定：

```bash
TRAILFINDER_OVERPASS_URL=https://overpass-api.de/api/interpreter
OSRM_URL=https://router.project-osrm.org
VALHALLA_URL=                                      # 使用自建 Valhalla 時填寫
ELEVATION_PROVIDER=open-elevation                  # open-elevation | open-topo
OPENMETEO_URL=https://api.open-meteo.com/v1/forecast
DEFAULT_TZ=Asia/Tokyo
TRAILFINDER_USER_AGENT=trail-finder-mcp/0.1.0 (+your-contact)
```

範本請見 `.env.example`。

---

## 📋 使用範例

用自然語言向 Claude Code 提問，會自動選擇合適的工具：

- 「高尾山口站附近的登山口有哪些」→ `trailheads`
- 「從高尾山口站到高尾山山頂的路線」→ `route_foot`
- 「高尾山的海拔是多少？」→ `elevation`
- 「高尾山今天天氣如何？」→ `forecast`

---

## 🖥️ Claude Desktop（選用）

也可在 Claude Desktop 中使用。設定檔路徑：

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

設定內容與 Claude Code 相同。

---

## 📦 發行版本

GitHub 上提供帶標籤的發行版本：

- [最新版本](https://github.com/Takamasa045/Trail-Finder-MCP/releases/latest)
- 目前版本: **v0.1.0**

```bash
git clone https://github.com/Takamasa045/Trail-Finder-MCP.git
cd Trail-Finder-MCP
git checkout v0.1.0
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

---

## ⚠️ 注意事項

- 公開的 OSRM / Overpass API 負載與速度會波動，正式使用建議自建實例。
- 回傳資料僅供參考。實際登山請以現場狀況與安全為優先。
- 本工具**不保證**登山安全，請務必提前規劃與準備。

---

## 📝 資料來源與授權

本專案使用的外部 API 與資料來源：

- **OpenStreetMap / Overpass API** — 地圖資料（ODbL）
- **OSRM** — 路由引擎
- **Open-Meteo** — 天氣預報資料
- **Open-Elevation / OpenTopoData** — 海拔資料
