# Trail‑Finder MCP — 要件定義 v0.1

## 1. 概要（What / Why）

登山・ハイキングの計画や現地ナビに必要な「登山口・道標（guidepost）・水場・徒歩ルート・標高・天気」を、オープンデータ系API（OSM/Overpass, Valhalla or OSRM, Open‑Topo/Elevation, Open‑Meteo）から横断取得して統合返却する **MCPサーバー**。Claude Code / Codex / Skills からのツール呼び出しで、自然言語→地理情報クエリ→整形レスポンスの一貫処理を実現する。

---

## 2. スコープ

* **対象地物**: 登山口（trailhead相当）、道標（guidepost）、水場（飲用水/湧水）。
* **機能**: ①周辺POI検索、②徒歩ルーティング、③任意地点の標高、④地点の天気（短時間予報）。
* **出力形式**: 構造化JSON（GeoJSON互換フィールドを含む）。
* **単位系**: デフォルトは SI（距離m、標高m、速度m/s、温度℃）。
* **座標参照系**: WGS84（EPSG:4326, lat/lon）

---

## 3. ユースケース

* 例: 「高尾山口駅から高尾山山頂までの徒歩ルートと途中の水場を列挙、出発時の気温も」

  * `route_foot` で徒歩ルートを計算
  * ルート形状に沿って `trailheads` 相当のPOI・水場を抽出（バッファ内検索）
  * 出発点座標で `forecast` を取得し、直近の気温を抽出
  * まとめて整形して提示

---

## 4. 採用APIと役割

* **Overpass API (OSM)**: 周辺POI（guidepost, trailhead, drinking_water, spring など）の取得。
* **Valhalla or OSRM**: 徒歩経路（距離/所要時間/ジオメトリ）。
* **Open‑TopoData / Open‑Elevation**: 単点標高の取得（バックアップ/フォールバックを用意）。
* **Open‑Meteo**: 短時間予報（hourly温度・降水・風など）。

> 注記: 具体のエンドポイントURLやAPIキー要否はデプロイ環境の設定値（ENV）で切替可能にする。

---

## 5. MCPツール仕様（外部インターフェース）

### 5.1 共通

* **エンドポイント命名**: `/v1/tools/<tool-name>` （内部実装）
* **レスポンス**: `application/json; charset=utf-8`
* **エラー**: 後述のエラーフォーマットに準拠
* **タイムアウト**: それぞれ 8–15s 程度（サービス毎に設定）。
* **リトライ**: 冪等GETに限り指数バックオフ 2回まで。

### 5.2 `trailheads(lat, lon, radius_m=2000)`

* **説明**: 指定中心から半径 `radius_m` 以内の **登山口/道標**（必要に応じて水場も同APIで取得可能）を返す。
* **入力**:

  ```json
  {
    "lat": 35.0,
    "lon": 139.0,
    "radius_m": 2000,
    "include": ["guidepost", "trailhead"],
    "also_water": true,
    "limit": 200
  }
  ```

  * `include` は OSMタグにマッピング（デフォルト: guidepost + trailhead）
  * `also_water` が true の場合、`amenity=drinking_water` と `natural=spring` も併せて返却
* **出力**（抜粋）:

  ```json
  {
    "center": {"lat": 35.0, "lon": 139.0},
    "radius_m": 2000,
    "items": [
      {
        "id": "node/123456789",
        "type": "guidepost",        
        "name": "◯◯分岐 道標",
        "tags": {"information": "guidepost"},
        "location": {"lat": 35.001, "lon": 139.001},
        "distance_m": 153,
        "source": "overpass/osm"
      }
    ]
  }
  ```

### 5.3 `route_foot(from_lat, from_lon, to_lat, to_lon)`

* **説明**: 徒歩プロファイルでの経路探索（ValhallaまたはOSRM）。
* **入力**:

  ```json
  {
    "from": {"lat": 35.0, "lon": 139.0},
    "to":   {"lat": 35.01, "lon": 139.02},
    "engine": "auto",            
    "options": {
      "avoid_ferry": true,
      "include_geometry": true,
      "include_steps": false
    }
  }
  ```

  * `engine`: `auto|valhalla|osrm`（`auto`は優先順fallback）
* **出力**（抜粋）:

  ```json
  {
    "engine": "osrm",
    "distance_m": 4820,
    "duration_s": 4300,
    "elevation_gain_m": 420,   
    "elevation_loss_m": 10,
    "geometry": {
      "type": "LineString",
      "coordinates": [[139.0,35.0],[139.001,35.0008], ...]
    },
    "steps": []
  }
  ```

  * 標高差は `elevation()` とルートサンプリングの合成で算出（任意）

### 5.4 `elevation(lat, lon)`

* **説明**: 単点標高（m）を返す。
* **入力**: `{ "lat": 35.0, "lon": 139.0 }`
* **出力**: `{ "elevation_m": 564, "source": "open-elevation" }`
* **備考**: 1st/2ndとして Open‑TopoData と Open‑Elevation を切替可能。

### 5.5 `forecast(lat, lon, hours=24)`

* **説明**: 指定地点の短時間予報（温度・降水・風など）を返す。
* **入力**:

  ```json
  { "lat": 35.0, "lon": 139.0, "hours": 24 }
  ```
* **出力**（抜粋）:

  ```json
  {
    "timezone": "Asia/Tokyo",
    "hourly": [
      {"time": "2025-11-03T09:00:00+09:00", "temperature_c": 14.2, "precip_mm": 0.0, "wind_mps": 1.8},
      {"time": "2025-11-03T10:00:00+09:00", "temperature_c": 15.0, ...}
    ]
  }
  ```

---

## 6. 追加（オプション）ツール

### `route_features(from, to, buffer_m=120)`

* ルートを計算 → 線形状を 120m 程度でバッファ → バッファ範囲内の **guidepost / trailhead / drinking_water / spring** を集約。
* 返却: ルート + 立寄りPOI一覧（距離順並び替え、重複除去）。

---

## 7. 内部実装方針

### 7.1 Overpassクエリ方針（例）

* 周辺検索（中心+半径）:

  ```
  [out:json][timeout:25];
  (
    node(around:{radius},{lat},{lon})[information=guidepost];
    node(around:{radius},{lat},{lon})[information=trailhead];
    node(around:{radius},{lat},{lon})[amenity=drinking_water];
    node(around:{radius},{lat},{lon})[natural=spring];
  );
  out tags geom;
  ```
* ルート沿い検索: ルートを200mピッチで点列化→各点に対し `around:80` で同様クエリを分割実行→OSM `id` で重複除去。

### 7.2 ルーティング

* **Valhalla**: costing `pedestrian`、`alternates`不要、`shape`返却を標準。有向グラフの歩行可否（`access`）に従う。
* **OSRM**: profile `foot`、`geometries=geojson`、`overview=full`。
* **高低差**: ルートを一定間隔でサンプリングし `elevation()` を並列呼出→累積上昇/下降を算出。

### 7.3 キャッシュ/フェイルオーバ

* **キャッシュTTL**:

  * Overpass: 1時間（座標+半径+タグのキー）
  * ルート: 1時間（from/toの座標丸め＋エンジン）
  * 標高: 1日（座標丸め 1e‑4）
  * 予報: 30分（座標丸め 1e‑2 + hours）
* **フェイルオーバ**: Valhalla→OSRM、Open‑TopoData→Open‑Elevation の順で自動切替。

### 7.4 並列/レート制御

* Overpassは礼儀的スロットリング（例: 2–3 req/s上限 + `User-Agent`/`Contact`明示）。
* ルーティング/標高/天気は最大5並列程度に制限。

---

## 8. エラー設計

* **HTTP 4xx**: バリデーション/不正引数
* **HTTP 429**: レート超過（外部/内部）
* **HTTP 5xx**: 外部API障害/タイムアウト
* **ボディ共通**:

  ```json
  { "error": { "code": "EXTERNAL_TIMEOUT", "message": "Valhalla timeout", "hint": "engine=osrmで再試行可", "retryable": true } }
  ```
* **入力バリデーション**:

  * lat: −90〜90, lon: −180〜180
  * radius_m: 50〜20000
  * hours: 1〜168

---

## 9. 品質要件（NFR）

* **可用性**: 99.5%/月
* **P95レイテンシ**: trailheads ≤ 3.5s、route ≤ 4.0s、elevation ≤ 1.0s、forecast ≤ 1.5s
* **一貫性**: 同一入力に対して同一出力（外部データ更新を除く）
* **拡張性**: ツール/タグ種別の追加が容易（設定駆動）

---

## 10. セキュリティ/プライバシ

* **API鍵**: 必要なサービスはENVで注入。平文保存禁止。
* **CORS**: MCPサーバー内で無効（サーバーtoサーバー）。
* **位置情報**: ログ上は座標を千分位で丸め（PII低減）可。保存は最小限。

---

## 11. ライセンス/帰属

* **OSM/Overpass**: ODbL/帰属要件に準拠。返却JSONに `attribution` フィールドを付与。
* **予報**: Open‑Meteoのクレジット表記を返却に含める。

---

## 12. ロギング/監視（Observability）

* **構造化ログ**: toolName, latency_ms, cacheHit, upstream, status, bbox/radiusなど。
* **メトリクス**: QPS, エラー率, 外部毎のP95, キャッシュHIT率。
* **トレース**: 外部API呼び出しをspan化し相関ID伝搬。

---

## 13. デプロイ/運用

* **言語/実装**: Go（既存Unified Serverに沿う）推奨。HTTPツールで公開。
* **デプロイ**: Cloud Run（最小インスタンス=0、同時実行=80）想定。
* **設定**: `TRAILFINDER_OVERPASS_URL`, `ROUTER_ENGINE=valhalla|osrm|auto`, `VALHALLA_URL`, `OSRM_URL`, `ELEVATION_PRIMARY=open-topo|open-elev`, `OPENMETEO_URL`, `DEFAULT_TZ=Asia/Tokyo` など。
* **CI/CD**: GitHub Actionsでビルド&デプロイ、Smoke Test実行。

---

## 14. テスト計画

* **ユニット**: 入力検証、タグマッピング、距離計算、バッファ生成、重複除去。
* **結合**: サンドボックス地点（例: 都市公園/低山/アルプス裾野）で、trailheads/route/elevation/forecastの整合。
* **回帰**: 既知地点でのゴールデンファイル比較（距離±1%、時間±5% など許容差）。
* **負荷**: 30QPS/5分バーストでのP95監視、スロットリング確認。

---

## 15. 実装タスク（MVP）

1. ルーティング実装（Valhalla/OSRM切替）
2. Overpassラッパ（周辺検索 + ルート沿い分割検索）
3. 標高単点APIラッパ + ルートサンプリング高低差
4. Open‑Meteoラッパ（hoursフィルタ・TZ補正）
5. 共通キャッシュ層（LRU + TTL）
6. 共通エラーハンドラ/スキーマ検証
7. ログ/メトリクス/簡易ダッシュボード

---

## 16. MCP設定例（Claude/Codex）

```json
{
  "mcpServers": {
    "trail-finder-mcp": {
      "command": "trail-finder-mcp",  
      "args": ["--router=auto", "--tz=Asia/Tokyo"],
      "env": {
        "TRAILFINDER_OVERPASS_URL": "https://overpass-api.de/api/interpreter",
        "VALHALLA_URL": "https://valhalla.example.com",
        "OSRM_URL": "https://router.project-osrm.org",
        "OPENMETEO_URL": "https://api.open-meteo.com/v1/forecast"
      }
    }
  },
  "tools": [
    {"name": "trailheads", "description": "周辺の登山口/道標/水場POI", "inputSchema": {"type": "object", "properties": {"lat": {"type": "number"}, "lon": {"type": "number"}, "radius_m": {"type": "integer", "default": 2000}, "include": {"type": "array", "items": {"type": "string"}}, "also_water": {"type": "boolean", "default": true}, "limit": {"type": "integer", "default": 200}}, "required": ["lat","lon"]}},
    {"name": "route_foot", "description": "徒歩ルート", "inputSchema": {"type": "object", "properties": {"from": {"type": "object", "properties": {"lat": {"type": "number"}, "lon": {"type": "number"}}, "required": ["lat","lon"]}, "to": {"type": "object", "properties": {"lat": {"type": "number"}, "lon": {"type": "number"}}, "required": ["lat","lon"]}, "engine": {"type": "string", "enum": ["auto","valhalla","osrm"], "default": "auto"}, "options": {"type": "object"}}, "required": ["from","to"]}},
    {"name": "elevation", "description": "単点標高(m)", "inputSchema": {"type": "object", "properties": {"lat": {"type": "number"}, "lon": {"type": "number"}}, "required": ["lat","lon"]}},
    {"name": "forecast", "description": "短時間予報（℃/mm/風）", "inputSchema": {"type": "object", "properties": {"lat": {"type": "number"}, "lon": {"type": "number"}, "hours": {"type": "integer", "default": 24}}, "required": ["lat","lon"]}}
  ]
}
```

---

## 17. 返却サンプル（統合シナリオ）

```json
{
  "route": {
    "distance_m": 4820,
    "duration_s": 4300,
    "geometry": {"type": "LineString", "coordinates": [[139.0,35.0], ...]}
  },
  "pois": {
    "guideposts": [ {"id": "node/123", "name": "◯◯道標", "lat": 35.001, "lon": 139.001, "dist_to_route_m": 20} ],
    "trailheads": [ {"id": "node/456", "name": "△△登山口", "lat": 35.002, "lon": 139.002, "dist_to_start_m": 80} ],
    "waters": [ {"id": "node/789", "type": "drinking_water", "lat": 35.003, "lon": 139.003, "dist_to_route_m": 45} ]
  },
  "weather": {"start_time": "2025-11-03T09:00:00+09:00", "temperature_c": 14.2}
}
```

---

## 18. エッジケース/注意事項

* OSMタグの地域差: `information=trailhead` が無い地域がある → `guidepost` + 名前/タグ補助で推定。
* 季節閉鎖/通行規制: ルーティングが**現地状況**に追随しない場合あり → 返却にディスクレーマ付与。
* ルート外POIのノイズ: バッファ幅を狭める/距離閾値でフィルタ。
* 標高データのグリッド粗さ: 位置丸めで誤差増 → 表示は小数点0–1桁に丸め。

---

## 19. 安全配慮（ディスクレーマ）

本ツールは参考情報を提供するもので、現地の通行可否・天候・危険箇所の最終判断はユーザーに委ねられます。必ず最新の現地情報・気象情報・装備・届出等を確認してください。

---

## 20. 今後の拡張

* GPX/KML入出力、ルート断面図（標高プロファイル）
* 積雪/積雪期予報データの統合、日照/日の出入
* 混雑推定（SNS/天気/休日連動）
* オフラインキャッシュ、モバイル最適化エンドポイント
