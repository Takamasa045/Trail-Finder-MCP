# Trail‑Finder MCP

[English](README.md) | [日本語](README.ja.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [한국어](README.ko.md) | [Español](README.es.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Português](README.pt.md)

**バージョン:** 0.2.0（正本ドキュメントは英語と日本語です）

オープンデータから **地名**・**登山口/道標/小屋/水場**・**徒歩ルート**・**標高**・**天気** を返す MCP サーバーです。

Claude Code / Claude Desktop / Codex / Grok から使えます。駅名や山名から一括で計画するなら `plan_hike` を使ってください。

---

## Quick Start

**Go 1.23+** が必要です。

```bash
git clone https://github.com/Takamasa045/Trail-Finder-MCP.git
cd Trail-Finder-MCP
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

MCP クライアントにバイナリを登録します。

```json
{
  "mcpServers": {
    "trail-finder": {
      "command": "/absolute/path/to/trail-finder-mcp",
      "env": {
        "DEFAULT_TZ": "Asia/Tokyo",
        "DEFAULT_LANG": "ja",
        "ELEVATION_PROVIDER": "open-meteo",
        "TRAILFINDER_USER_AGENT": "trail-finder-mcp/0.2.0 (+https://github.com/Takamasa045/Trail-Finder-MCP)"
      }
    }
  }
}
```

設定ファイルの場所:

- **Claude Code**: `~/.claude.json` の `mcpServers`
- **Claude Desktop (macOS)**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Codex**: `~/.codex/config.toml` の `[mcp_servers.trail-finder]`

デバッグ用 HTTP:

```bash
./trail-finder-mcp -http :8080
curl -s localhost:8080/healthz
```

---

## ツール

| ツール | 説明 |
|--------|------|
| `geocode` | 地名 → 座標（Nominatim）。「高尾山口駅」など日本語が使えます |
| `trailheads` | 周辺の登山口・道標。小屋・峠・水場も可 |
| `route_foot` | OSRM の徒歩ルート。形状と標高差を含みます |
| `elevation` | 地点の標高（m）。デフォルトは Open-Meteo |
| `forecast` | 気温・降水・降水確率・風速 **m/s**・突風・天気コード・日の出日の入り |
| `plan_hike` | 地名または座標から、ルート・沿線 POI・標高差・短時間予報をまとめて返す |

`engine=valhalla` は未実装です。`auto` または `osrm` を使ってください。

### `plan_hike` の入力例

```json
{
  "from": { "query": "高尾山口駅" },
  "to": { "query": "高尾山" },
  "also_water": true,
  "hours": 24
}
```

座標でも渡せます: `{ "lat": 35.63, "lon": 139.27 }`。`query` と座標が両方あるときは `query` を使います。水場はデフォルトで含めます（`also_water: false` でオフ）。出発地予報は `forecast`、目的地は `forecast_goal` です。POI や天気が落ちてもルートは返し、`warnings` に理由を入れます。

### `trailheads` の include

`guidepost` / `trailhead` / `shelter` / `pass` / `entrance`（任意。ノイズが多いのでデフォルトでは使いません）。

日本向けに `highway=trailhead` と道標タグを優先しています。

---

## 使い方の例

- 「高尾山口駅から高尾山までの計画を」→ `plan_hike`
- 「高尾山口駅の座標は？」→ `geocode`
- 「その周辺の登山口と水場」→ `trailheads`
- 「高尾山の今日の風と降水確率」→ `forecast`

---

## 環境変数

`.env.example` を参照してください。公開 Overpass / OSRM / Nominatim は混みやすいので、本格運用では自前エンドポイントを推奨します。

---

## 開発

```bash
go test ./...
```

429/502/503/504 は最大 2 回リトライします。Overpass と Nominatim は短いメモリキャッシュがあります。

ライセンスは MIT。要件メモは [docs/requirements.md](docs/requirements.md) です。

返却データは参考情報です。現地の状況・装備・公式発表を優先してください。安全は保証しません。
