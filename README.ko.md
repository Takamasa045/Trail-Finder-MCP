# Trail‑Finder MCP

[English](README.md) | [日本語](README.ja.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [한국어](README.ko.md) | [Español](README.es.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Português](README.pt.md)

**버전:** [v0.1.0](https://github.com/Takamasa045/Trail-Finder-MCP/releases/tag/v0.1.0)

**등산로 입구 / 이정표 / 급수원**, **도보 경로**, **고도**, **날씨** 등 오픈 데이터를 모아 제공하는 Claude Code 전용 MCP 서버입니다.
Claude Code에서 바로 호출해 등산·하이킹에 필요한 지리 정보를 쉽게 얻을 수 있습니다.

---

## 🚀 빠른 시작

### 1) 바이너리 빌드

```bash
cd /path/to/trail-finder-mcp
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

**Go 1.23+** 가 필요합니다.

### 2) Claude Code MCP 설정

`~/.config/claude-code/mcp_config.json` 에 다음을 추가합니다.

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

- `command` 에는 빌드한 바이너리의 절대 경로를 지정하세요
- 환경 변수는 필요에 따라 커스터마이즈할 수 있습니다 (아래 참고)

### 3) Claude Code 실행

설정 저장 후 Claude Code를 실행하면 `trail-finder` MCP 서버가 자동으로 로드되며 다음 도구를 사용할 수 있습니다.

| 도구 | 설명 |
|------|------|
| `trailheads` | 주변 등산로 입구 / 이정표 / 급수원 검색 |
| `route_foot` | 두 지점 간 도보 경로 계산 |
| `elevation` | 지정 지점 고도 조회 |
| `forecast` | 지정 지점 일기 예보 조회 |

---

## 🛠️ 사용 가능한 도구

### `trailheads`

주변 등산로 입구, 이정표, 급수원 등 POI를 검색합니다.

**입력 파라미터:**
- `lat`, `lon` — 검색 중심 좌표
- `radius_m` — 검색 반경(미터)
- `include` — 포함할 POI 유형 (예: `["guidepost", "trailhead"]`)
- `also_water` — 급수원 포함 여부 (`true` / `false`)
- `limit` — 최대 결과 수

**출력:**
인근 POI를 JSON으로 반환 (유형, 이름, 좌표 등)

### `route_foot`

두 지점 간 최단 도보 경로를 계산합니다.

**입력 파라미터:**
- `from` — 출발지 `{lat, lon}`
- `to` — 목적지 `{lat, lon}`
- `engine` — 라우팅 엔진 (`"auto"`, `"osrm"`, `"valhalla"`)
- `options` — 추가 옵션
  - `include_geometry` (기본: `true`) 경로 지오메트리 포함 여부
  - `include_steps` (기본: `false`) OSRM 단계 상세 포함 여부
  - `avoid_ferry` (기본: `false`) 페리 회피 (OSRM `exclude=ferry`)

**출력:**
거리, 소요 시간, GeoJSON LineString 경로

### `elevation`

지정 지점의 고도를 조회합니다.

**입력 파라미터:**
- `lat`, `lon` — 좌표

**출력:**
고도(미터)

### `forecast`

지정 지점의 단기 일기 예보를 조회합니다.

**입력 파라미터:**
- `lat`, `lon` — 좌표
- `hours` — 예보 시간 수

**출력:**
온도, 강수량, 풍속의 시계열 데이터 (Open-Meteo)

---

## ⚙️ 환경 변수

`.env` 파일 또는 MCP 설정의 `env` 섹션에서 설정할 수 있습니다.

```bash
TRAILFINDER_OVERPASS_URL=https://overpass-api.de/api/interpreter
OSRM_URL=https://router.project-osrm.org
VALHALLA_URL=                                      # 자체 Valhalla 사용 시
ELEVATION_PROVIDER=open-elevation                  # open-elevation | open-topo
OPENMETEO_URL=https://api.open-meteo.com/v1/forecast
DEFAULT_TZ=Asia/Tokyo
TRAILFINDER_USER_AGENT=trail-finder-mcp/0.1.0 (+your-contact)
```

템플릿은 `.env.example` 을 참고하세요.

---

## 📋 사용 예시

Claude Code에 자연어로 질문하면 적절한 도구가 자동으로 호출됩니다.

- “다카오산구치 역 근처 등산로 입구 알려줘” → `trailheads`
- “다카오산구치 역에서 다카오산 정상까지 경로” → `route_foot`
- “다카오산 고도는?” → `elevation`
- “다카오산 오늘 날씨는?” → `forecast`

---

## 🖥️ Claude Desktop (선택)

Claude Desktop에서도 동일하게 사용할 수 있습니다. 설정 파일 경로:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

설정 내용은 Claude Code와 동일합니다.

---

## 📦 릴리스

GitHub에서 태그된 릴리스를 제공합니다.

- [최신 릴리스](https://github.com/Takamasa045/Trail-Finder-MCP/releases/latest)
- 현재 버전: **v0.1.0**

```bash
git clone https://github.com/Takamasa045/Trail-Finder-MCP.git
cd Trail-Finder-MCP
git checkout v0.1.0
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

---

## ⚠️ 주의사항

- 공개 OSRM / Overpass API는 부하와 속도가 변동합니다. 본격 운용 시 자체 인스턴스를 권장합니다.
- 반환 정보는 참고용입니다. 실제 등산에서는 현장 판단과 안전을 최우선으로 하세요.
- 이 도구는 등산 안전을 보장하지 않습니다. 반드시 사전 계획과 준비를 하세요.

---

## 📝 데이터 소스 및 라이선스

이 프로젝트에서 사용하는 외부 API 및 데이터 소스:

- **OpenStreetMap / Overpass API** — 지도 데이터 (ODbL)
- **OSRM** — 라우팅 엔진
- **Open-Meteo** — 일기 예보 데이터
- **Open-Elevation / OpenTopoData** — 고도 데이터
