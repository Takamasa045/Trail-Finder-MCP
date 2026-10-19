# compare_hikes: sourced route research

This read-only workflow compares 1–4 starts for any mountain. It combines exact-name/coordinate resolution, summit/start forecast summaries, terrain estimates, scoped source excerpts, publication/modification/retrieval dates, and routing diagnostics. It needs no new API key or paid service.

Use a browser to choose authority/operator route and notice URLs, then pass them to this tool. It does not search the web, certify a publisher, read PDFs/images, interpret map arrows, normalize official route totals, or declare current foliage/access conditions. The result contains facts it actually fetched and explicit unknowns; do not replace missing facts with OSRM estimates.

## Example 1: two sides of a mountain

These are OSM search coordinates, not certified trailhead/summit coordinates. In particular, verify the north-side lodge location against the operator's map before planning.

```json
{
  "subject": "雨飾山",
  "mountain": {"lat": 36.902028, "lon": 137.9626039},
  "routes": [
    {
      "name": "小谷側・雨飾高原キャンプ場",
      "start": {"lat": 36.882143, "lon": 137.973505},
      "official_urls": ["https://info-otari.jp/hike/"]
    },
    {
      "name": "糸魚川側・雨飾山荘候補（要照合）",
      "start": {"lat": 36.9208136, "lon": 137.9545724},
      "official_urls": ["https://www.itoigawa-kanko.net/spot/amakazariyama/"]
    }
  ],
  "information_urls": ["https://info-otari.jp/2026/10/08/18973/"],
  "as_of": "2026-10-19",
  "hours": 24
}
```

`official_route_evidence` groups direct excerpts by distance/duration/ascent/difficulty. These are source statements with context, not automatically normalized totals. An 8 km round trip and an 8 km one-way leg must not be merged. Empty evidence means the tool could not extract that fact. Course maps in PDFs/images remain verification links.

`information` holds additional notices without assigning them to every route. A parking/access walk of 60 minutes must not become the north-side hiking course time. Per-route foliage/access fields use that route's own source URLs.

## Example 2: another mountain / ambiguous names

```json
{
  "subject": "高尾山",
  "mountain": {"query": "高尾山"},
  "routes": [{
    "name": "高尾山口駅",
    "start": {"query": "高尾山口駅"},
    "official_urls": ["https://www.takaotozan.co.jp/about/"]
  }],
  "hours": 12
}
```

Multiple candidates, or a single non-exact name match, return `needs_selection` and candidate display names/coordinates. They never trigger weather/routing for a guessed point. Inspect `geocode`/candidate results, then call again with the selected coordinates. Coordinates and a query together are rejected for this workflow.

A page with several mountains is scoped to the matching subject heading; another mountain's figures are excluded. If no matching section or page title exists, the source is retained as a check link but not used for mountain-specific claims. This conservative parser may miss useful pages; open them manually instead of treating the omission as no information.

## Example 3: a missing or outdated source

```json
{
  "subject": "雨飾山",
  "mountain": {"lat": 36.902028, "lon": 137.9626039},
  "routes": [{
    "name": "小谷側",
    "start": {"lat": 36.882143, "lon": 137.973505}
  }],
  "information_urls": ["https://info-otari.jp/this-page-does-not-exist/"],
  "as_of": "2026-10-19",
  "hours": 1
}
```

A failed source remains `unavailable` with the requested URL/retrieval time/error. Forecasts and other sources still return. No source/dated evidence means `unconfirmed`, not “open” or “no restrictions”.

## Dates, weather, and estimates

- `retrieved_at` is when data was fetched, never its publication date.
- `published` and `modified` keep their origins and raw timestamps. Calendar dates are normalized to `DEFAULT_TZ`; URL-path dates are marked inferred. Copyright/footer years are ignored.
- `freshness` reports publication age relative to `as_of`: within 14 days, older, future, or undated. This is a reading aid, not proof a statement is applicable. A recently edited 2019 typhoon article stays old. Date mentions inside the subject section remain visible separately, including older-year notices on a general page.
- Foliage/access remain `unconfirmed` even with a recent notice: read its dates, weekdays, conditions and exact area. Seasonal descriptions do not confirm today's colors. An old route description can still contain useful course information; age alone does not invalidate or confirm it.
- Forecast output distinguishes requested coordinates, returned model-grid coordinates/elevation, retrieval time, and the unknown model-run time. `as_of` does not request historical weather. Summaries cover the returned interval, which may cross midnight. Set `include_hourly: true` to include all hourly rows; the default is compact summaries.
- Terrain DEM elevation is an estimate. `point_elevation_difference_m` is a two-point difference, not cumulative ascent or an official summit height.
- OSRM endpoints over 100 m from the requested points are rejected. Even an accepted estimate has an unverified pedestrian profile and is not official trail distance/course time. No route line guarantees access or safety.

Limits: 4 starts, 8 distinct source URLs (3 per route, 4 additional), 2 MiB UTF-8 HTML/text per page, bounded excerpts, 3 concurrent source fetches, and a 45-second overall budget. URL retrieval only permits public HTTP(S) endpoints, validates redirects and DNS addresses, and carries no browser cookies/authentication. Source content is untrusted quoted data, never tool instructions. No arbitrary page JavaScript is executed.

The existing six low-level tools remain available; the debug HTTP equivalent is `POST /tools/compare_hikes`. MCP returns a compact text report plus full structured content.
