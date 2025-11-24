# Trail‑Finder MCP

オープンデータを横断して **登山口/道標/水場**・**徒歩ルート**・**標高**・**天気** を返す Claude Code 専用 MCP サーバーです。
Claude Code から直接呼び出して、登山やハイキングに必要な地理情報を簡単に取得できます。

---

## 🚀 Quick Start

### 1) バイナリのビルド

```bash
cd /path/to/trail-finder-mcp
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

### 2) Claude Code の MCP 設定

Claude Code の設定ファイル `~/.config/claude-code/mcp_config.json` に以下を追加します。

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

- `command` にはビルドしたバイナリの絶対パスを指定してください
- 環境変数は必要に応じてカスタマイズできます（詳細は後述）

### 3) Claude Code を起動

設定ファイルを保存後、Claude Code を起動すると「trail-finder」MCP サーバーが自動的に読み込まれ、以下の 4 つのツールが利用可能になります:

- `trailheads` - 周辺の登山口/道標/水場を検索
- `route_foot` - 2 地点間の徒歩ルートを計算
- `elevation` - 指定地点の標高を取得
- `forecast` - 指定地点の天気予報を取得

---

## 🛠️ 利用可能なツール

### `trailheads`
周辺の登山口、道標、水場などを検索します。

**入力パラメータ:**
- `lat`, `lon` - 検索中心の緯度経度
- `radius_m` - 検索半径（メートル）
- `include` - 含めるPOIタイプの配列 (`["guidepost", "trailhead"]` など)
- `also_water` - 水場も含めるか（true/false）
- `limit` - 最大取得件数

**出力:**
近傍の POI を JSON で返却（タイプ、名称、座標など）

### `route_foot`
2 地点間の徒歩最短ルートを計算します。

**入力パラメータ:**
- `from` - 出発地点 `{lat, lon}`
- `to` - 目的地点 `{lat, lon}`
- `engine` - ルーティングエンジン (`"auto"`, `"osrm"`, `"valhalla"`)
- `options` - 追加オプション
  - `include_geometry` (デフォルト: true) ルート形状を含めるか
  - `include_steps` (デフォルト: false) OSRM のステップ詳細を含めるか
  - `avoid_ferry` (デフォルト: false) フェリーを避ける（OSRM の exclude=ferry）

**出力:**
距離、所要時間、GeoJSON LineString 形式のルート

### `elevation`
指定地点の標高を取得します。

**入力パラメータ:**
- `lat`, `lon` - 緯度経度

**出力:**
標高（メートル）

### `forecast`
指定地点の短時間天気予報を取得します。

**入力パラメータ:**
- `lat`, `lon` - 緯度経度
- `hours` - 予報時間数

**出力:**
時系列の温度、降水量、風速データ（Open-Meteo）

---

## ⚙️ 環境変数（カスタマイズ）

`.env` ファイルまたは MCP 設定の `env` セクションで以下の環境変数を設定できます:

```bash
TRAILFINDER_OVERPASS_URL=https://overpass-api.de/api/interpreter
OSRM_URL=https://router.project-osrm.org
VALHALLA_URL=                                      # 自前の Valhalla を使う場合
ELEVATION_PROVIDER=open-elevation                  # open-elevation | open-topo
OPENMETEO_URL=https://api.open-meteo.com/v1/forecast
DEFAULT_TZ=Asia/Tokyo
TRAILFINDER_USER_AGENT=trail-finder-mcp/0.1.0 (+your-contact)
```

---

## 📋 使用例

Claude Code で以下のように質問すると、自動的に適切なツールが呼び出されます:

- 「高尾山口駅周辺の登山口を教えて」→ `trailheads` ツールを使用
- 「高尾山口駅から高尾山山頂までのルートを調べて」→ `route_foot` ツールを使用
- 「高尾山の標高は？」→ `elevation` ツールを使用
- 「高尾山の今日の天気は？」→ `forecast` ツールを使用

---

## 🖥️ Claude Desktop での使用（オプション）

Claude Desktop でも同様に使用できます。設定ファイルのパスは以下の通りです:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

設定内容は Claude Code と同じです。

---

## ⚠️ 注意事項

- 公開 OSRM/Overpass API は負荷や速度が変動します。本格的な運用では自前インスタンスの利用を推奨します。
- 返却される情報は参考データです。実際の登山やハイキングでは、現地の状況判断と安全配慮を最優先してください。
- このツールは登山の安全を保証するものではありません。必ず事前の計画と準備を行ってください。

---

## 📝 ライセンス

このプロジェクトで使用している外部 API やデータソース:
- **OpenStreetMap / Overpass API** - 地図データ（ODbL ライセンス）
- **OSRM** - ルーティングエンジン
- **Open-Meteo** - 天気予報データ
- **Open-Elevation / OpenTopoData** - 標高データ
