# Trail‑Finder MCP

オープンデータを横断して **登山口/道標/水場**・**徒歩ルート**・**標高**・**天気** を返す MCP サーバーです。  
HTTP ツールとしての利用に加えて、Claude Code / Claude Desktop などの MCP クライアントから直接呼び出せる **stdio MCP サーバー** モードを提供します。

---

## 🚀 Quick Start

### 1) 環境変数（任意）
`.env` 例：
```
TRAILFINDER_OVERPASS_URL=https://overpass-api.de/api/interpreter
OSRM_URL=https://router.project-osrm.org          # OSRM（徒歩）既定はこの公開エンドポイント
VALHALLA_URL=                                      # (任意) 自前の Valhalla を使う場合
ELEVATION_PROVIDER=open-elevation                  # open-elevation | open-topo
OPENMETEO_URL=https://api.open-meteo.com/v1/forecast
DEFAULT_TZ=Asia/Tokyo
PORT=8080
```

### 2) 実行モード

#### HTTP モード（デフォルト）
```bash
go run ./cmd/trail-finder-mcp --mode http
# or
PORT=8080 go run ./cmd/trail-finder-mcp --mode http
```

#### MCP (stdio) モード
Claude Code / Claude Desktop から直接呼び出す場合はこちらを使用します。
```bash
go run ./cmd/trail-finder-mcp --mode mcp
# 環境変数で指定する場合
TRAILFINDER_MODE=mcp go run ./cmd/trail-finder-mcp
```

### 3) 動作確認（例）
- 周辺の登山口/道標/水場：
```bash
curl -s http://localhost:8080/tools/trailheads -H 'Content-Type: application/json' -d '{
  "lat": 35.6251,
  "lon": 139.2430,
  "radius_m": 2000,
  "include": ["guidepost","trailhead"],
  "also_water": true,
  "limit": 200
}' | jq .
```

- 徒歩ルート（高尾山口駅→高尾山 山頂 付近）：
```bash
curl -s http://localhost:8080/tools/route_foot -H 'Content-Type: application/json' -d '{
  "from": {"lat": 35.6251, "lon": 139.2430},
  "to":   {"lat": 35.6258, "lon": 139.2439},
  "engine": "auto",
  "options": {"include_geometry": true}
}' | jq .
```

- 標高：
```bash
curl -s http://localhost:8080/tools/elevation -H 'Content-Type: application/json' -d '{
  "lat": 35.6251, "lon": 139.2430
}' | jq .
```

- 予報（Open‑Meteo）：
```bash
curl -s http://localhost:8080/tools/forecast -H 'Content-Type: application/json' -d '{
  "lat": 35.6251, "lon": 139.2430, "hours": 12
}' | jq .
```

---

## 📦 HTTP ツール仕様（簡易）

- `POST /tools/trailheads`  
  入力: `lat, lon, radius_m, include[], also_water, limit`  
  出力: 近傍の `guidepost/trailhead/drinking_water/spring` を JSON で返却。

- `POST /tools/route_foot`  
  入力: `from{lat,lon}, to{lat,lon}, engine("auto"|"osrm"|"valhalla")`  
  出力: 徒歩最短ルート（距離・所要秒・GeoJSON LineString）。MVPは OSRM を優先。

- `POST /tools/elevation`  
  入力: `lat,lon`  
  出力: 単点標高（m）。`ELEVATION_PROVIDER` が `open-topo` のとき OpenTopoData。

- `POST /tools/forecast`  
  入力: `lat,lon,hours`  
  出力: 時系列（hourly）で温度/降水/風速を返却（Open‑Meteo）。

---

## 🧩 HTTP ツール登録例（MCP/Skills 等）
```jsonc
{
  "tools": [ {
    "name": "trailheads",
    "type": "http",
    "url": "http://localhost:8080/tools/trailheads",
    "description": "周辺の登山口/道標/水場POI"
  }, {
    "name": "route_foot",
    "type": "http",
    "url": "http://localhost:8080/tools/route_foot",
    "description": "徒歩ルート（OSRM/Valhalla）"
  }, {
    "name": "elevation",
    "type": "http",
    "url": "http://localhost:8080/tools/elevation",
    "description": "単点標高(m)"
  }, {
    "name": "forecast",
    "type": "http",
    "url": "http://localhost:8080/tools/forecast",
    "description": "短時間予報（Open‑Meteo）"
  } ]
}
```

---

## 🤝 Claude Code / Claude Desktop 連携手順

1. バイナリをビルド（任意）
   ```bash
   go build -o trail-finder-mcp ./cmd/trail-finder-mcp
   ```

2. Claude Desktop の設定ファイル（例: macOS は `~/Library/Application Support/Claude/claude_desktop_config.json`）に MCP サーバーを追加します。
   ```jsonc
   {
     "mcpServers": {
       "trail-finder": {
         "command": "/absolute/path/to/trail-finder-mcp",
         "args": ["--mode", "mcp"],
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
   - `command` はビルド済みバイナリ、または `go run` をラップしたシェルスクリプトでも構いません。
   - `args` を省略する場合は `TRAILFINDER_MODE=mcp` を環境変数として設定してください。

3. Claude Code / Claude Desktop を再起動すると「trail-finder」 MCP サーバーがツール一覧に表示され、`trailheads`, `route_foot`, `elevation`, `forecast` の 4 ツールが利用可能になります。

> Windows の設定ファイル例: `%APPDATA%\\Claude\\claude_desktop_config.json`  
> Linux の設定ファイル例: `~/.config/Claude/claude_desktop_config.json`

---

## ⚠️ 注意
- 公開 OSRM/Overpass は負荷・速度が変動します。実運用は自前インスタンス推奨。
- 返却値は参考情報。実地判断・安全配慮は利用者に委ねられます。
