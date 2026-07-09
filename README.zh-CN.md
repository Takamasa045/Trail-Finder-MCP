# Trail‑Finder MCP

[English](README.md) | [日本語](README.ja.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [한국어](README.ko.md) | [Español](README.es.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Português](README.pt.md)

**版本:** [v0.1.0](https://github.com/Takamasa045/Trail-Finder-MCP/releases/tag/v0.1.0)

面向 Claude Code 的 MCP 服务器，聚合开放数据，提供 **登山口 / 路标 / 水源**、**步行路线**、**海拔** 与 **天气** 信息。
可直接在 Claude Code 中调用，轻松获取徒步与登山所需的地理信息。

---

## 🚀 快速开始

### 1) 编译二进制

```bash
cd /path/to/trail-finder-mcp
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

需要 **Go 1.23+**。

### 2) 配置 Claude Code MCP

在 `~/.config/claude-code/mcp_config.json` 中添加：

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

- `command` 请填写编译后二进制的绝对路径
- 可按需自定义环境变量（详见下文）

### 3) 启动 Claude Code

保存配置后启动 Claude Code，`trail-finder` MCP 服务器会自动加载并提供以下工具：

| 工具 | 说明 |
|------|------|
| `trailheads` | 搜索附近登山口 / 路标 / 水源 |
| `route_foot` | 计算两点之间的步行路线 |
| `elevation` | 获取指定地点的海拔 |
| `forecast` | 获取指定地点的天气预报 |

---

## 🛠️ 可用工具

### `trailheads`

搜索附近的登山口、路标、水源等 POI。

**输入参数:**
- `lat`, `lon` — 搜索中心坐标
- `radius_m` — 搜索半径（米）
- `include` — 要包含的 POI 类型（如 `["guidepost", "trailhead"]`）
- `also_water` — 是否包含水源（`true` / `false`）
- `limit` — 最大返回数量

**输出:**
以 JSON 返回附近 POI（类型、名称、坐标等）

### `route_foot`

计算两点之间的最短步行路线。

**输入参数:**
- `from` — 起点 `{lat, lon}`
- `to` — 终点 `{lat, lon}`
- `engine` — 路由引擎（`"auto"`, `"osrm"`, `"valhalla"`）
- `options` — 额外选项
  - `include_geometry`（默认: `true`）是否包含路线几何
  - `include_steps`（默认: `false`）是否包含 OSRM 步骤详情
  - `avoid_ferry`（默认: `false`）避开轮渡（OSRM `exclude=ferry`）

**输出:**
距离、用时，以及 GeoJSON LineString 路线

### `elevation`

获取指定点的海拔。

**输入参数:**
- `lat`, `lon` — 坐标

**输出:**
海拔（米）

### `forecast`

获取指定点的短时天气预报。

**输入参数:**
- `lat`, `lon` — 坐标
- `hours` — 预报小时数

**输出:**
温度、降水量、风速的时间序列数据（Open-Meteo）

---

## ⚙️ 环境变量

可在 `.env` 文件或 MCP 配置的 `env` 中设置：

```bash
TRAILFINDER_OVERPASS_URL=https://overpass-api.de/api/interpreter
OSRM_URL=https://router.project-osrm.org
VALHALLA_URL=                                      # 使用自建 Valhalla 时填写
ELEVATION_PROVIDER=open-elevation                  # open-elevation | open-topo
OPENMETEO_URL=https://api.open-meteo.com/v1/forecast
DEFAULT_TZ=Asia/Tokyo
TRAILFINDER_USER_AGENT=trail-finder-mcp/0.1.0 (+your-contact)
```

模板见 `.env.example`。

---

## 📋 使用示例

用自然语言向 Claude Code 提问，会自动选择合适的工具：

- “高尾山口站附近的登山口有哪些” → `trailheads`
- “从高尾山口站到高尾山山顶的路线” → `route_foot`
- “高尾山的海拔是多少？” → `elevation`
- “高尾山今天天气怎么样？” → `forecast`

---

## 🖥️ Claude Desktop（可选）

也可在 Claude Desktop 中使用。配置文件路径：

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

配置内容与 Claude Code 相同。

---

## 📦 发布

GitHub 上提供带标签的发布版本：

- [最新版本](https://github.com/Takamasa045/Trail-Finder-MCP/releases/latest)
- 当前版本: **v0.1.0**

```bash
git clone https://github.com/Takamasa045/Trail-Finder-MCP.git
cd Trail-Finder-MCP
git checkout v0.1.0
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

---

## ⚠️ 注意事项

- 公开的 OSRM / Overpass API 负载与速度会波动，正式使用建议自建实例。
- 返回数据仅供参考。实际登山请以现场情况与安全为首要考量。
- 本工具**不保证**登山安全，请务必提前规划与准备。

---

## 📝 数据来源与许可

本项目使用的外部 API 与数据源：

- **OpenStreetMap / Overpass API** — 地图数据（ODbL）
- **OSRM** — 路由引擎
- **Open-Meteo** — 天气预报数据
- **Open-Elevation / OpenTopoData** — 海拔数据
