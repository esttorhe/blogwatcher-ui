# Technology Stack: v1.1 Additions

**Project:** BlogWatcher UI
**Milestone:** v1.1 UI Polish & Search
**Researched:** 2026-02-03
**Confidence:** HIGH

## Executive Summary

v1.1 adds masonry layout, thumbnail extraction, and search features. **No new Go dependencies required** for core functionality. CSS-only masonry with progressive enhancement, existing `gofeed` handles RSS media extraction, `modernc.org/sqlite` supports FTS5 out of the box. Open Graph parsing is the only new dependency needed for thumbnail fallback.

## Existing Stack (v1.0 - No Changes)

These components are already validated and working. **Do not replace or upgrade.**

| Component | Version | Purpose |
|-----------|---------|---------|
| Go | 1.22+ | Server runtime |
| HTMX | 2.0.4 | Dynamic updates without full page reloads |
| modernc.org/sqlite | Latest (supports 3.51.2+) | Pure-Go SQLite driver with FTS5 |
| gofeed | v1.3.0 | RSS/Atom feed parsing with image extraction |
| goquery | Current | HTML scraping for blog content |
| CSS custom properties | Native | Theme system (light/dark/system) |

## New Additions for v1.1

### 1. Open Graph Parser

**Library:** `github.com/otiai10/opengraph/v2`
**Version:** v2.2.0
**Purpose:** Extract og:image from article URLs as thumbnail fallback
**Why this library:**
- Active maintenance (published Sep 2025)
- Clean Go v2 module with semantic versioning
- Supports custom HTTP headers and context
- Handles relative URL conversion to absolute
- MIT licensed, production-ready (28+ importers)

**Installation:**
```bash
go get github.com/otiai10/opengraph/v2@v2.2.0
```

**Usage pattern:**
```go
import "github.com/otiai10/opengraph/v2"

ogp, err := opengraph.Fetch(articleURL)
if err == nil && len(ogp.Image) > 0 {
    thumbnailURL = ogp.Image[0].URL
}
```

**Alternative considered:** `github.com/dyatlov/go-opengraph` (v1.0.1) — rejected because older (May 2022), no v2 module path, fewer features.

### 2. CSS Masonry Layout

**Technology:** Native CSS (no library)
**Approach:** Progressive enhancement with fallback
**Why CSS-only:**
- Zero dependencies, no JavaScript required
- Graceful degradation to regular grid in older browsers
- Matches HTMX philosophy (minimal JS, server-rendered)
- Future-proof as CSS Grid Lanes standardizes

**Browser support status (2026):**
- Safari Technology Preview 234: `display: grid-lanes` available
- Chrome/Edge 140+: Experimental behind flag
- Firefox Nightly: Available with flag
- CSS WG voted Jan 31, 2025 for `grid-lanes` syntax
- Expected production: Q2-Q3 2026

**Implementation strategy:**

```css
/* Progressive enhancement approach */
.article-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  /* Fallback: regular grid auto-placement */
}

/* Future enhancement when browsers ship grid-lanes */
@supports (display: grid-lanes) {
  .article-grid {
    display: grid-lanes;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  }
}

/* Interim masonry approximation using column-count fallback */
@media (min-width: 768px) {
  .article-grid.masonry-fallback {
    column-count: 2;
    column-gap: 1rem;
  }

  .article-grid.masonry-fallback > * {
    break-inside: avoid;
    margin-bottom: 1rem;
  }
}

@media (min-width: 1024px) {
  .article-grid.masonry-fallback {
    column-count: 3;
  }
}
```

**Why not JavaScript masonry libraries:**
- Contradicts HTMX server-rendered philosophy
- Adds dependency and bundle size
- CSS solution is simpler and more maintainable
- Native support coming soon

### 3. SQLite Full-Text Search (FTS5)

**Technology:** SQLite FTS5 extension (built into modernc.org/sqlite)
**Version:** Included in SQLite 3.51.2+ (supported by modernc.org/sqlite)
**Purpose:** Title search with relevance ranking
**Why FTS5:**
- Already included, zero new dependencies
- BM25 ranking algorithm built-in
- Supports prefix search with `*` wildcard
- Designed for text search, faster than `LIKE '%query%'`

**Migration pattern:**
```sql
-- Create FTS5 virtual table mirroring articles
CREATE VIRTUAL TABLE articles_fts USING fts5(
  title,
  content=articles,
  content_rowid=id
);

-- Populate from existing articles
INSERT INTO articles_fts(rowid, title)
  SELECT id, title FROM articles;

-- Triggers to keep FTS synchronized
CREATE TRIGGER articles_fts_insert AFTER INSERT ON articles BEGIN
  INSERT INTO articles_fts(rowid, title) VALUES (new.id, new.title);
END;

CREATE TRIGGER articles_fts_update AFTER UPDATE ON articles BEGIN
  UPDATE articles_fts SET title = new.title WHERE rowid = new.id;
END;

CREATE TRIGGER articles_fts_delete AFTER DELETE ON articles BEGIN
  DELETE FROM articles_fts WHERE rowid = old.id;
END;
```

**Query pattern:**
```sql
-- Basic search
SELECT articles.* FROM articles
JOIN articles_fts ON articles.id = articles_fts.rowid
WHERE articles_fts MATCH ?
ORDER BY rank;

-- With filters
SELECT articles.* FROM articles
JOIN articles_fts ON articles.id = articles_fts.rowid
WHERE articles_fts MATCH ?
  AND articles.is_read = 0
  AND articles.blog_id = ?
  AND articles.published_date >= ?
ORDER BY rank;
```

**Alternative considered:** `LIKE '%query%'` — rejected because slower (full table scan), no relevance ranking, no prefix matching.

### 4. Thumbnail Fallback Chain

**Approach:** Try sources in order, store result
**No new libraries required** (gofeed + opengraph/v2)
**Storage:** Add `thumbnail_url TEXT` column to `articles` table

**Fallback sequence:**
1. **RSS media enclosure** (`gofeed` Item.Enclosures) — already parsed
2. **RSS item image** (`gofeed` Item.Image.URL) — already parsed
3. **Open Graph image** (fetch article URL, parse og:image) — new: opengraph/v2
4. **Favicon** (`/favicon.ico` at domain root) — construct URL from domain

**Implementation rationale:**
- RSS sources are already fetched (zero latency)
- Open Graph requires HTTP request (cache result to avoid repeated fetches)
- Favicon is reliable fallback (every site has one)
- Store result in database to avoid re-fetching on every view

**Database migration:**
```sql
ALTER TABLE articles ADD COLUMN thumbnail_url TEXT;
CREATE INDEX idx_articles_thumbnail ON articles(thumbnail_url);
```

## HTMX Integration Notes

**No changes to HTMX version or patterns.** New features use existing v2.0.4 capabilities.

**Active search pattern** (for title search):
```html
<input type="text"
       name="q"
       hx-get="/api/search"
       hx-trigger="keyup changed delay:250ms"
       hx-target="#article-list"
       hx-indicator="#search-spinner"
       placeholder="Search articles...">
```

**Date filter pattern:**
```html
<select name="date_range"
        hx-get="/api/articles"
        hx-trigger="change"
        hx-target="#article-list"
        hx-include="[name='q'], [name='blog_id'], [name='status']">
  <option value="">All time</option>
  <option value="week">Last week</option>
  <option value="month">Last month</option>
  <option value="custom">Custom range...</option>
</select>
```

**Server-side considerations:**
- Return partial HTML fragments (article cards)
- Use query params for filter state
- Include HTMX response headers for client-side updates
- Implement debouncing on server if needed (250ms client-side should suffice)

## What NOT to Add

| Technology | Why Avoid |
|-----------|-----------|
| JavaScript masonry library (Masonry.js, Isotope) | CSS solution is simpler, matches server-rendered philosophy, native support coming |
| Full-text search library (Bleve, typesense-go) | SQLite FTS5 is sufficient for title search, already included |
| Image processing library (imaging, resize) | Thumbnails are URLs, not processed locally |
| Caching library (groupcache, bigcache) | Database is local SQLite, fast enough without cache layer |
| `mattn/go-sqlite3` | Already using modernc.org/sqlite (pure Go, no CGo) |
| Alternative Open Graph parsers | otiai10/opengraph/v2 is actively maintained and feature-complete |

## Installation Commands

```bash
# Only new dependency needed
go get github.com/otiai10/opengraph/v2@v2.2.0

# Existing dependencies (no changes)
# modernc.org/sqlite — already installed
# github.com/mmcdole/gofeed@v1.3.0 — already installed
# github.com/PuerkitoBio/goquery — already installed
```

## Integration Checklist

For each new capability:

**Masonry Layout:**
- [ ] Add CSS with progressive enhancement
- [ ] Implement column-count fallback for older browsers
- [ ] Add `@supports` query for future grid-lanes
- [ ] Test graceful degradation

**Thumbnail Extraction:**
- [ ] Add `thumbnail_url` column to articles table
- [ ] Create index on thumbnail_url
- [ ] Import opengraph/v2
- [ ] Implement fallback chain (RSS → OG → favicon)
- [ ] Cache results in database to avoid re-fetching

**Title Search:**
- [ ] Create FTS5 virtual table `articles_fts`
- [ ] Add triggers to keep FTS synchronized
- [ ] Populate from existing articles
- [ ] Use HTMX active search pattern with debounce
- [ ] Return ranked results with `ORDER BY rank`

**Date Filtering:**
- [ ] Add date range selector (last week, month, custom)
- [ ] Combine with existing filters (blog, status, search)
- [ ] Use HTMX `hx-include` to preserve filter state
- [ ] Server-side query building with WHERE clauses

## Confidence Assessment

| Area | Level | Reason |
|------|-------|--------|
| CSS Masonry | HIGH | Official MDN docs, browser roadmaps verified, progressive enhancement clear |
| Open Graph | HIGH | Library verified via pkg.go.dev, v2.2.0 current, production-ready |
| FTS5 | HIGH | SQLite official docs, modernc.org/sqlite includes FTS5, standard SQL syntax |
| Thumbnail Chain | HIGH | gofeed structs verified, fallback logic straightforward |
| HTMX Patterns | HIGH | Official HTMX docs, active search pattern documented |

## Migration Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|-----------|
| FTS5 table out of sync | Medium | Use triggers to maintain consistency automatically |
| Open Graph fetch timeouts | Low | Set context timeout (5s), fallback to favicon if fails |
| CSS masonry browser support | Low | Graceful degradation to regular grid, column-count fallback |
| Thumbnail URL stale/broken | Low | Display fallback favicon on 404, re-fetch on demand |

## Sources

**CSS Masonry:**
- [MDN: Masonry Layout](https://developer.mozilla.org/en-US/docs/Web/CSS/Guides/Grid_layout/Masonry_layout)
- [WebKit: Introducing CSS Grid Lanes](https://webkit.org/blog/17660/introducing-css-grid-lanes/)
- [WebKit: When will CSS Grid Lanes arrive?](https://webkit.org/blog/17758/when-will-css-grid-lanes-arrive-how-long-until-we-can-use-it/)
- [Can I use: CSS Grid Lanes](https://caniuse.com/css-grid-lanes)
- [CSS-Tricks: Masonry Layout is Now grid-lanes](https://css-tricks.com/masonry-layout-is-now-grid-lanes/)
- [Smashing Magazine: Native CSS Masonry Layout In CSS Grid](https://www.smashingmagazine.com/native-css-masonry-layout-css-grid/)

**Open Graph Parsing:**
- [pkg.go.dev: otiai10/opengraph/v2](https://pkg.go.dev/github.com/otiai10/opengraph/v2)
- [GitHub: otiai10/opengraph](https://github.com/otiai10/opengraph)
- [pkg.go.dev: dyatlov/go-opengraph](https://pkg.go.dev/github.com/dyatlov/go-opengraph)

**SQLite FTS5:**
- [SQLite: FTS5 Extension](https://sqlite.org/fts5.html)
- [SQLite Tutorial: Full-text Search By Examples](https://www.sqlitetutorial.net/sqlite-full-text-search/)
- [Medium: Full-Text Search in SQLite: A Practical Guide](https://medium.com/@johnidouglasmarangon/full-text-search-in-sqlite-a-practical-guide-80a69c3f42a4)

**RSS Image Extraction:**
- [pkg.go.dev: mmcdole/gofeed](https://pkg.go.dev/github.com/mmcdole/gofeed)
- [W3Schools: RSS enclosure Element](https://www.w3schools.com/xml/rss_tag_enclosure.asp)
- [RSS API: Can RSS Feeds have images?](https://rssapi.net/blog/can-rss-feeds-have-images)

**HTMX Patterns:**
- [HTMX: Active Search Example](https://htmx.org/examples/active-search/)
- [HTMX: Documentation](https://htmx.org/docs/)
- [Hypermedia Systems: More Htmx Patterns](https://hypermedia.systems/more-htmx-patterns/)

**Favicon Handling:**
- [Evil Martians: How to Favicon in 2026](https://evilmartians.com/chronicles/how-to-favicon-in-2021-six-files-that-fit-most-needs)
- [pkg.go.dev: go-favicon](https://pkg.go.dev/github.com/thanhpk/go-favicon)

---
*Stack research complete. Ready for roadmap phase planning.*
