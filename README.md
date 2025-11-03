# Trail‑Finder MCP (HTTP)

オープンデータを横断して **登山口/道標/水場**・**徒歩ルート**・**標高**・**天気** を返す HTTP MCP サーバー（ツール実装）。  
MVPは **OSRM**（徒歩ルーティング）/ **Overpass**（OSM POI）/ **Open‑Elevation or OpenTopoData**（標高）/ **Open‑Meteo**（天気）。

> 参考：コードは単体で http サーバーとして起動し、`/tools/*` エンドポイントをツールとして叩けます。MCP/Skills 側では HTTP ツールとして登録してください。

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

### 2) 実行
```bash
go run ./cmd/trail-finder-mcp
# or
PORT=8080 go run ./cmd/trail-finder-mcp
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

## 🧩 MCP/Skills 側登録ヒント（例）
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

## ⚠️ 注意
- 公開 OSRM/Overpass は負荷・速度が変動します。実運用は自前インスタンス推奨。
- 返却値は参考情報。実地判断・安全配慮は利用者に委ねられます。
