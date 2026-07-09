# Trail‑Finder MCP

[English](README.md) | [日本語](README.ja.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [한국어](README.ko.md) | [Español](README.es.md) | [Français](README.fr.md) | [Deutsch](README.de.md) | [Português](README.pt.md)

**Versão:** [v0.1.0](https://github.com/Takamasa045/Trail-Finder-MCP/releases/tag/v0.1.0)

Servidor MCP para Claude Code que agrega dados abertos de **acessos a trilhas / postes de sinalização / fontes de água**, **rotas a pé**, **elevação** e **clima**.
Chame-o diretamente do Claude Code para obter informações geográficas úteis para caminhadas e montanhismo.

---

## 🚀 Início rápido

### 1) Compilar o binário

```bash
cd /path/to/trail-finder-mcp
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

Requer **Go 1.23+**.

### 2) Configurar o Claude Code MCP

Adicione o seguinte em `~/.config/claude-code/mcp_config.json`:

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

- `command` deve ser o caminho absoluto do binário compilado
- Personalize as variáveis de ambiente conforme necessário (veja abaixo)

### 3) Iniciar o Claude Code

Após salvar a configuração, inicie o Claude Code. O servidor MCP `trail-finder` carrega automaticamente e expõe estas ferramentas:

| Ferramenta | Descrição |
|------------|-----------|
| `trailheads` | Buscar acessos a trilhas / postes / fontes de água próximos |
| `route_foot` | Calcular uma rota a pé entre dois pontos |
| `elevation` | Obter a elevação de um local |
| `forecast` | Obter uma previsão do tempo de curto prazo |

---

## 🛠️ Ferramentas disponíveis

### `trailheads`

Busca acessos a trilhas, postes de sinalização, fontes de água e POIs semelhantes.

**Parâmetros de entrada:**
- `lat`, `lon` — coordenadas do centro
- `radius_m` — raio de busca em metros
- `include` — tipos de POI a incluir (ex.: `["guidepost", "trailhead"]`)
- `also_water` — incluir fontes de água (`true` / `false`)
- `limit` — número máximo de resultados

**Saída:**
POIs próximos em JSON (tipo, nome, coordenadas etc.)

### `route_foot`

Calcula a rota a pé mais curta entre dois pontos.

**Parâmetros de entrada:**
- `from` — origem `{lat, lon}`
- `to` — destino `{lat, lon}`
- `engine` — motor de roteamento (`"auto"`, `"osrm"`, `"valhalla"`)
- `options` — opções extras
  - `include_geometry` (padrão: `true`) — incluir geometria da rota
  - `include_steps` (padrão: `false`) — incluir detalhes de etapas OSRM
  - `avoid_ferry` (padrão: `false`) — evitar balsas (OSRM `exclude=ferry`)

**Saída:**
Distância, duração e geometria GeoJSON LineString

### `elevation`

Obtém a elevação de um ponto.

**Parâmetros de entrada:**
- `lat`, `lon` — coordenadas

**Saída:**
Elevação em metros

### `forecast`

Obtém uma previsão do tempo de curto prazo para um ponto.

**Parâmetros de entrada:**
- `lat`, `lon` — coordenadas
- `hours` — horizonte da previsão em horas

**Saída:**
Série temporal de temperatura, precipitação e velocidade do vento (Open-Meteo)

---

## ⚙️ Variáveis de ambiente

Defina-as em um arquivo `.env` ou na seção `env` da configuração MCP:

```bash
TRAILFINDER_OVERPASS_URL=https://overpass-api.de/api/interpreter
OSRM_URL=https://router.project-osrm.org
VALHALLA_URL=                                      # se usar Valhalla próprio
ELEVATION_PROVIDER=open-elevation                  # open-elevation | open-topo
OPENMETEO_URL=https://api.open-meteo.com/v1/forecast
DEFAULT_TZ=Asia/Tokyo
TRAILFINDER_USER_AGENT=trail-finder-mcp/0.1.0 (+your-contact)
```

Veja `.env.example` para um modelo.

---

## 📋 Exemplos de uso

Pergunte ao Claude Code em linguagem natural; ele escolherá a ferramenta certa:

- “Encontre acessos a trilhas perto da estação Takao-sanguchi” → `trailheads`
- “Rota da estação Takao-sanguchi ao cume do monte Takao” → `route_foot`
- “Qual é a elevação do monte Takao?” → `elevation`
- “Como está o tempo hoje no monte Takao?” → `forecast`

---

## 🖥️ Claude Desktop (opcional)

Você também pode usar este servidor com o Claude Desktop. Caminhos do arquivo de configuração:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

O conteúdo da configuração é o mesmo do Claude Code.

---

## 📦 Lançamentos

Lançamentos com tags são publicados no GitHub:

- [Último lançamento](https://github.com/Takamasa045/Trail-Finder-MCP/releases/latest)
- Versão atual: **v0.1.0**

```bash
git clone https://github.com/Takamasa045/Trail-Finder-MCP.git
cd Trail-Finder-MCP
git checkout v0.1.0
go build -o trail-finder-mcp ./cmd/trail-finder-mcp
```

---

## ⚠️ Avisos

- As APIs públicas OSRM / Overpass variam em carga e latência. Para produção, prefira instâncias próprias.
- Os dados retornados são apenas de referência. Na trilha, priorize as condições locais e a segurança.
- Esta ferramenta **não garante** segurança em caminhadas. Sempre planeje e prepare-se.

---

## 📝 Fontes de dados e licença

APIs e fontes de dados externas usadas por este projeto:

- **OpenStreetMap / Overpass API** — dados cartográficos (ODbL)
- **OSRM** — motor de roteamento
- **Open-Meteo** — dados de previsão do tempo
- **Open-Elevation / OpenTopoData** — dados de elevação
