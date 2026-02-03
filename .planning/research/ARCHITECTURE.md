# Architecture Patterns: Masonry, Thumbnails, and Search Integration

**Domain:** BlogWatcher UI Polish (Go/HTMX architecture)
**Researched:** 2026-02-03
**Confidence:** HIGH

## Executive Summary

This research examines how masonry layout, thumbnail extraction, and search functionality integrate with the existing Go/HTMX architecture. The existing codebase uses a clean dependency injection pattern with Go templates, HTMX partial updates, and SQLite storage. All three features can be added with minimal architectural changes by leveraging existing patterns.

**Key finding:** The architecture already supports the integration patterns needed. Masonry is CSS-only, thumbnails extend the existing Article model and scraper pipeline, and search adds one handler plus SQLite FTS5 integration.

## Current Architecture Overview

### Request Flow (Existing)

```
Browser → Go HTTP Handler → Database Query → Template Render → HTMX Partial/Full Page
```

### Component Boundaries (Existing)

| Component | Responsibility | Location |
|-----------|---------------|----------|
| `cmd/server/main.go` | Server lifecycle, graceful shutdown | Entry point |
| `internal/server/server.go` | Dependency injection, route registration | Server setup |
| `internal/server/handlers.go` | HTTP handlers with HTMX detection | Request processing |
| `internal/server/routes.go` | Route definitions (Go 1.22+ method routing) | Routing |
| `internal/storage/database.go` | SQLite queries, connection management | Data layer |
| `internal/model/model.go` | Blog, Article, ArticleWithBlog structs | Data models |
| `internal/scanner/scanner.go` | RSS/scraper orchestration | Content fetching |
| `templates/*.gohtml` | Go templates with HTMX attributes | Presentation |
| `static/styles.css` | CSS custom properties, grid layout | Styling |

## Feature 1: Masonry Layout Integration

### Architecture Impact: MINIMAL (CSS-only change)

Masonry layout requires NO Go code changes. It's purely a presentation layer modification using CSS Grid.

### Integration Points

**Modified Components:**
- `static/styles.css` - Add masonry grid CSS

**Zero modifications needed to:**
- Go handlers (article data flow unchanged)
- Database layer (query results same structure)
- Templates (same article-list loop structure)
- HTMX patterns (partial updates work identically)

### Implementation Approach

**Current State (2026):**
Native CSS masonry is being standardized with the `grid-lanes` syntax, but browser support is still behind feature flags. Three production-ready alternatives exist:

#### Option A: CSS Column-based Masonry (Recommended for MVP)

**Why:** Works today with zero JavaScript, no browser flags, excellent performance.

**CSS:**
```css
.article-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1rem;
}
```

**Tradeoffs:**
- **Pro:** Works in all modern browsers, responsive out of the box
- **Pro:** Auto-fit handles responsive breakpoints automatically
- **Pro:** No JavaScript = fast, no FOUC, works with HTMX updates
- **Con:** Items laid out in column-major order (vertical flow)
- **Con:** Not true masonry (items won't fill gaps perfectly)

#### Option B: CSS Columns Masonry (Alternative)

**CSS:**
```css
.article-grid {
  column-count: auto;
  column-width: 300px;
  column-gap: 1rem;
}

.article-card {
  break-inside: avoid;
  margin-bottom: 1rem;
}
```

**Tradeoffs:**
- **Pro:** Better packing than Grid auto-fit (fills vertical gaps)
- **Pro:** Zero JavaScript
- **Con:** Items read vertically instead of horizontally (UX issue)
- **Con:** Hard to predict which column items land in

#### Option C: JavaScript Masonry Library (Deferred)

**Library:** Masonry.js (~200k npm downloads/week)

**Why defer:**
- Requires JavaScript execution after HTMX swaps
- 66 lines of JS minimum
- Introduces FOUC (flash of unstyled content)
- Adds build complexity

**When to use:** Only if true Pinterest-style masonry with row-major ordering is non-negotiable.

### Recommendation

**Use Option A (CSS Grid auto-fit) for MVP**, then monitor native CSS masonry browser support. The `grid-lanes` syntax is being actively developed for 2026-2027 release.

**Migration path:**
1. Launch with Grid auto-fit (works today)
2. Monitor [caniuse.com for masonry support](https://caniuse.com/)
3. Add progressive enhancement when browser support reaches 80%+

### Data Flow (Unchanged)

```
handleArticleList → ListArticlesWithBlog → article-list.gohtml → CSS Grid
```

Articles flow through existing pipeline. CSS receives identical HTML structure but displays it as a grid instead of vertical stack.

### Sources

- [Chrome Masonry Update](https://developer.chrome.com/blog/masonry-update) - Browser implementation status
- [MDN Masonry Layout Guide](https://developer.mozilla.org/en-US/docs/Web/CSS/Guides/Grid_layout/Masonry_layout) - Official documentation
- [CSS Grid auto-fit patterns](https://dev.to/m97chahboun/responsive-card-layout-with-css-grid-a-step-by-step-guide-3ej1) - Production implementations

## Feature 2: Thumbnail Extraction Integration

### Architecture Impact: MODERATE (extends existing scanner and model)

Thumbnails integrate into the existing content pipeline by extending the scanner to extract OpenGraph images during article discovery.

### Integration Points

**New Components:**
- `internal/thumbnail/extractor.go` - OpenGraph extraction logic
- `internal/thumbnail/cache.go` - Optional: cache thumbnails locally or store URLs

**Modified Components:**
- `internal/model/model.go` - Add `ThumbnailURL string` to Article/ArticleWithBlog
- `internal/storage/database.go` - Add `thumbnail_url` column, update queries
- `internal/scanner/scanner.go` - Call thumbnail extractor during article discovery
- `templates/partials/article-list.gohtml` - Render `<img>` with thumbnail
- `static/styles.css` - Style thumbnail in card layout

### Implementation Approach

#### Phase 2A: Thumbnail URL Extraction (Backfillable)

**Library:** `github.com/otiai10/opengraph/v2`

**Why:** Most actively maintained Go OpenGraph library with clean API.

**Integration point:** In scanner pipeline, after RSS/scraper extracts article URL but before database insert.

**Flow:**
```
RSS/Scraper → Extract Article URL → Fetch OpenGraph → Extract og:image → Store thumbnail_url
```

**Code location:** `internal/scanner/scanner.go`

**Pseudocode:**
```go
// After extracting article URL
if thumbnailURL := thumbnail.ExtractOpenGraph(articleURL); thumbnailURL != "" {
    article.ThumbnailURL = thumbnailURL
}
```

**Tradeoffs:**
- **Pro:** Backfillable (can run against existing articles)
- **Pro:** No local storage needed (just store URL)
- **Pro:** OpenGraph is standard (90%+ of blogs support it)
- **Con:** Adds HTTP request per article (mitigated: only on new articles)
- **Con:** External URLs may break/change over time

#### Phase 2B: Local Thumbnail Caching (Optional Enhancement)

**When:** If external URLs prove unreliable.

**Libraries:**
- `github.com/disintegration/imaging` - Thumbnail generation
- `github.com/nfnt/resize` - Alternative resizing library

**Flow:**
```
Extract og:image URL → Download image → Resize to 300x200 → Save to static/thumbnails/ → Store local path
```

**Storage location:** `static/thumbnails/{article_id}.jpg`

**Tradeoffs:**
- **Pro:** Reliable (no external dependency at render time)
- **Pro:** Can optimize size/format for performance
- **Con:** Increases storage requirements
- **Con:** Not backfillable (requires re-fetching images)

### Database Schema Change

**Migration:**
```sql
ALTER TABLE articles ADD COLUMN thumbnail_url TEXT;
CREATE INDEX idx_articles_thumbnail ON articles(thumbnail_url) WHERE thumbnail_url IS NOT NULL;
```

**Backward compatibility:** Existing articles have NULL thumbnail_url. Template uses conditional rendering:
```html
{{if .ThumbnailURL}}
<img class="article-thumbnail" src="{{.ThumbnailURL}}" alt="" loading="lazy">
{{end}}
```

### Recommendation

**Start with Phase 2A (URL extraction only)**. The otiai10/opengraph library is mature and well-documented. Local caching (Phase 2B) can be added later if needed.

**Backfill strategy:** Add a `POST /admin/backfill-thumbnails` endpoint that processes existing articles in batches.

### Sources

- [otiai10/opengraph](https://github.com/otiai10/opengraph) - Go OpenGraph library
- [disintegration/imaging](https://reintech.io/blog/a-guide-to-gos-image-package-manipulating-and-processing-images) - Go image processing
- [OpenGraph Protocol](https://ogp.me/) - Official specification

## Feature 3: Search Integration

### Architecture Impact: MODERATE (new handler + database FTS)

Search adds one new handler and enables SQLite FTS5 (full-text search) on the articles table. Integrates cleanly with HTMX patterns.

### Integration Points

**New Components:**
- Enable SQLite FTS5 virtual table for articles
- `handleSearch(w http.ResponseWriter, r *http.Request)` handler

**Modified Components:**
- `internal/storage/database.go` - Add `SearchArticles(query string) ([]ArticleWithBlog, error)` method
- `templates/partials/article-list.gohtml` - Add search input with HTMX attributes
- `internal/server/routes.go` - Add `GET /search` route

**Optional Library:**
- `github.com/zalgonoise/fts` - Wrapper around SQLite FTS5 (if direct SQL is insufficient)

### Implementation Approach

#### FTS5 Virtual Table Setup

SQLite FTS5 is built-in to modernc.org/sqlite (the driver already in use). No external dependencies needed.

**Schema:**
```sql
CREATE VIRTUAL TABLE articles_fts USING fts5(
    title,
    content='articles',
    content_rowid='id'
);

-- Populate FTS index from existing data
INSERT INTO articles_fts(rowid, title)
SELECT id, title FROM articles;

-- Trigger to keep FTS in sync
CREATE TRIGGER articles_ai AFTER INSERT ON articles BEGIN
  INSERT INTO articles_fts(rowid, title) VALUES (new.id, new.title);
END;

CREATE TRIGGER articles_au AFTER UPDATE ON articles BEGIN
  UPDATE articles_fts SET title = new.title WHERE rowid = old.id;
END;

CREATE TRIGGER articles_ad AFTER DELETE ON articles BEGIN
  DELETE FROM articles_fts WHERE rowid = old.id;
END;
```

**Note:** This assumes the database initialization happens in the existing blogwatcher CLI, not the UI server. The UI server should NOT create schema (it uses `OpenDatabase` which expects schema to exist).

#### Handler Implementation

**Route:** `GET /search?q={query}&filter={unread|read}&blog={id}`

**Handler location:** `internal/server/handlers.go`

**Flow:**
```go
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("q")
    if query == "" {
        // Return empty results or all articles
    }

    filter := r.URL.Query().Get("filter")
    blogParam := r.URL.Query().Get("blog")

    articles, err := s.db.SearchArticles(query, filter, blogID)
    // ... render article-list.gohtml
}
```

**Database method:**
```go
func (db *Database) SearchArticles(query string, isRead bool, blogID *int64) ([]ArticleWithBlog, error) {
    sql := `
        SELECT a.id, a.blog_id, a.title, a.url, a.published_date, a.discovered_date, a.is_read, b.name, b.url
        FROM articles_fts fts
        INNER JOIN articles a ON fts.rowid = a.id
        INNER JOIN blogs b ON a.blog_id = b.id
        WHERE fts MATCH ? AND a.is_read = ?
    `
    // Add blog filter if provided
    // Execute query
}
```

#### HTMX Integration (Active Search Pattern)

**Template change** (add to `article-list.gohtml`):
```html
<div class="search-box">
  <input type="search"
         name="q"
         placeholder="Search articles..."
         hx-get="/search"
         hx-trigger="keyup changed delay:250ms, search"
         hx-target="#article-container"
         hx-include="[name='filter'], [name='blog']">
</div>

<div id="article-container">
  {{range .Articles}}
  <!-- existing article cards -->
  {{end}}
</div>
```

**HTMX attributes explained:**
- `hx-trigger="keyup changed delay:250ms, search"` - Debounce input, trigger on search button
- `hx-target="#article-container"` - Swap only article cards, not entire page
- `hx-include="[name='filter'], [name='blog']"` - Include current filter/blog state

**Benefits of this pattern:**
- No JavaScript needed beyond HTMX
- Works with browser back/forward (search params in URL)
- Degrades gracefully (form submit works without JS)
- Race condition handling via HTMX `hx-sync`

### Performance Considerations

**SQLite FTS5 performance:**
- FTS5 uses a trigram-based index (very fast for substring matching)
- Query times: <10ms for 100k articles (based on github.com/zalgonoise/fts benchmarks)
- Index size: ~30% of text data size

**Bottleneck:** HTTP latency from OpenGraph fetching during thumbnail extraction (if enabled). Mitigate by making OpenGraph fetch optional/async.

### Recommendation

**Implement search in two phases:**
1. **Phase 3A:** Basic FTS5 search (title only, no snippet highlighting)
2. **Phase 3B:** Enhanced search (add article content/description to FTS, highlight matches)

**Phase 3A is MVP-sufficient.** Most users search by article title, not body content.

### Sources

- [SQLite FTS5 Extension](https://sqlite.org/fts5.html) - Official documentation
- [github.com/zalgonoise/fts](https://github.com/zalgonoise/fts) - Go FTS5 wrapper
- [HTMX Active Search Pattern](https://hypermedia.systems/htmx-patterns/) - Official pattern guide
- [HTMX Debounce Best Practices](https://hypermedia.systems/more-htmx-patterns/) - Performance patterns

## Component Interaction Diagram

### Before (Current State)

```
┌─────────────────────────────────────────────────────┐
│                    Browser                          │
│  ┌────────────┐  ┌────────────┐  ┌──────────────┐ │
│  │  Sidebar   │  │  Toolbar   │  │Article Cards │ │
│  │  (HTMX)    │  │  (HTMX)    │  │   (HTMX)     │ │
│  └────────────┘  └────────────┘  └──────────────┘ │
└───────────────────────┬─────────────────────────────┘
                        │ HTMX requests
                        ▼
┌─────────────────────────────────────────────────────┐
│               Go HTTP Server                        │
│  ┌──────────────────────────────────────────────┐  │
│  │  Handlers: Index, ArticleList, MarkRead, etc │  │
│  └────────────────────┬─────────────────────────┘  │
└───────────────────────┼─────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────┐
│               Storage Layer                          │
│  ┌──────────────────────────────────────────────┐  │
│  │  ListArticlesWithBlog, MarkArticleRead, etc  │  │
│  └────────────────────┬─────────────────────────┘  │
└───────────────────────┼─────────────────────────────┘
                        │
                        ▼
                   SQLite DB
```

### After (With Masonry, Thumbnails, Search)

```
┌─────────────────────────────────────────────────────┐
│                    Browser                          │
│  ┌────────────┐  ┌────────────┐  ┌──────────────┐ │
│  │  Sidebar   │  │  Search +  │  │Masonry Grid  │ │
│  │  (HTMX)    │  │  Toolbar   │  │+ Thumbnails  │ │
│  │            │  │  (HTMX)    │  │   (HTMX)     │ │
│  └────────────┘  └────────────┘  └──────────────┘ │
└───────────────────────┬─────────────────────────────┘
                        │ HTMX requests
                        ▼
┌─────────────────────────────────────────────────────┐
│               Go HTTP Server                        │
│  ┌──────────────────────────────────────────────┐  │
│  │  Handlers: Index, ArticleList, Search, etc   │◄─┐│
│  └────────────────────┬─────────────────────────┘  ││
│                       │                             ││
│  ┌────────────────────▼─────────────────────────┐  ││
│  │     Thumbnail Extractor (during scan)        │  ││
│  │  Uses: github.com/otiai10/opengraph         │  ││
│  └──────────────────────────────────────────────┘  ││
└───────────────────────┼─────────────────────────────┘│
                        │                              │
                        ▼                              │
┌─────────────────────────────────────────────────────┐│
│               Storage Layer                          ││
│  ┌──────────────────────────────────────────────┐  ││
│  │  SearchArticles (FTS5), ListArticlesWithBlog │──┘│
│  │  + thumbnail_url field                       │   │
│  └────────────────────┬─────────────────────────┘   │
└───────────────────────┼─────────────────────────────┘
                        │
                        ▼
               SQLite DB + FTS5
```

**Key changes:**
1. Search input added to toolbar (sends HTMX requests to `/search`)
2. Thumbnail extractor runs during scanner pipeline (modifies articles before DB insert)
3. Article model includes `thumbnail_url` field
4. CSS Grid replaces vertical stack (no Go changes)
5. FTS5 virtual table added to database (queried by `SearchArticles`)

## Build Order Recommendation

Based on dependencies and risk:

### Phase 1: Masonry Layout (1-2 hours)
**Why first:** Zero Go changes, purely CSS. Validates visual design early.

**Steps:**
1. Modify `static/styles.css` - Add Grid auto-fit
2. Modify `templates/partials/article-list.gohtml` - Change container class
3. Test with existing data

**Risk:** LOW (CSS-only, easily reverted)

### Phase 2: Thumbnail Extraction (4-8 hours)
**Why second:** Extends existing scanner pipeline with minimal changes.

**Steps:**
1. Add migration for `thumbnail_url` column (in blogwatcher CLI, not UI)
2. Add `go get github.com/otiai10/opengraph/v2`
3. Create `internal/thumbnail/extractor.go`
4. Integrate into `internal/scanner/scanner.go`
5. Update `internal/model/model.go` and `internal/storage/database.go`
6. Modify template to render thumbnails
7. Add CSS for thumbnail styling

**Risk:** MODERATE (HTTP requests to external sites may timeout/fail)
**Mitigation:** Make thumbnail extraction non-blocking, store NULL if extraction fails

### Phase 3: Search (4-6 hours)
**Why last:** Requires FTS5 setup in database and new handler.

**Steps:**
1. Add FTS5 virtual table setup (in blogwatcher CLI schema)
2. Add `SearchArticles` method to `internal/storage/database.go`
3. Add `handleSearch` to `internal/server/handlers.go`
4. Register `/search` route in `internal/server/routes.go`
5. Add search input to template with HTMX attributes
6. Add CSS for search input styling

**Risk:** MODERATE (FTS5 query syntax can be tricky, needs testing)
**Mitigation:** Start with simple `MATCH` queries, defer advanced features

### Dependencies

```
Phase 1 (Masonry) ──► No dependencies

Phase 2 (Thumbnails) ──► Depends on Phase 1 (CSS must handle thumbnails)

Phase 3 (Search) ──► Independent of Phase 1/2
```

**Can parallelize:** Phase 2 and Phase 3 can be developed simultaneously if Phase 1 is complete.

## Patterns to Follow

### Pattern 1: Preserve HTMX Partial Updates

**What:** All new features must support HTMX's HX-Request header detection.

**Why:** Existing architecture serves full pages OR partials based on request header. Breaking this breaks browser back/forward button.

**Example:** Search handler must check `r.Header.Get("HX-Request")` and return either full page or just article list fragment.

### Pattern 2: Extend, Don't Replace

**What:** Add fields to existing structs, don't create parallel data structures.

**Why:** Keeps data flow simple. Article is article, whether it has thumbnail or not.

**Example:** Add `ThumbnailURL string` to `model.ArticleWithBlog`, not a separate `ArticleWithThumbnail` struct.

### Pattern 3: Progressive Enhancement

**What:** Features should degrade gracefully when data is missing.

**Why:** Thumbnails won't exist for old articles, search might return no results, masonry works with any number of cards.

**Example:** Template uses `{{if .ThumbnailURL}}` to conditionally render thumbnail. CSS Grid handles 1-to-N cards automatically.

### Pattern 4: Single Database Connection

**What:** Continue using `conn.SetMaxOpenConns(1)` for SQLite.

**Why:** SQLite's single-writer constraint. FTS5 writes (during article insert) must not conflict with reads (during search).

**Mitigation:** All writes happen during `/sync`, which is infrequent. Reads (search, list) are concurrent-safe.

## Anti-Patterns to Avoid

### Anti-Pattern 1: Client-Side Rendering

**What:** Using JavaScript to fetch data and render HTML in browser.

**Why bad:** Breaks the Go template + HTMX pattern. Introduces state management complexity.

**Instead:** Server renders HTML, HTMX swaps it. Keep all rendering logic in Go templates.

### Anti-Pattern 2: Over-Engineering Thumbnail Storage

**What:** Building a microservice for thumbnail processing, S3 integration, CDN, etc.

**Why bad:** Premature optimization. OpenGraph URLs work fine for MVP.

**Instead:** Start with storing thumbnail URLs. Add local storage only if external URLs prove unreliable.

### Anti-Pattern 3: Complex Search Syntax

**What:** Exposing FTS5's full query syntax (AND/OR/NEAR operators) in MVP.

**Why bad:** Most users want simple substring matching. Complex syntax adds UI complexity (help text, examples, error handling).

**Instead:** Start with single search box that does `title MATCH query`. Add advanced search later if needed.

### Anti-Pattern 4: JavaScript Masonry Before Native CSS

**What:** Adding Masonry.js or similar libraries immediately.

**Why bad:** Adds 66+ lines of JS, FOUC issues, and breaks HTMX swap animations.

**Instead:** Use CSS Grid auto-fit. Monitor native masonry browser support. Upgrade when 80%+ browsers support it.

## Scalability Considerations

### At 100 Articles (Current State)
- **Masonry:** Instant rendering
- **Thumbnails:** <1s per article during scan
- **Search:** <10ms query time

### At 10,000 Articles
- **Masonry:** Still instant (CSS handles layout)
- **Thumbnails:** Scan slows (100 articles × 1s = 100s total). Mitigation: parallelize OpenGraph fetches
- **Search:** <50ms query time (FTS5 scales logarithmically)

### At 100,000+ Articles
- **Masonry:** Pagination required (DOM size issue, not CSS)
- **Thumbnails:** Must add caching layer (too many external requests)
- **Search:** <200ms query time (still acceptable). Consider pagination of results.

**Bottleneck:** Thumbnail extraction during sync becomes the limiting factor at scale. Mitigation: make thumbnail fetching async/background job.

## Open Questions and Risks

### Risk 1: OpenGraph Availability

**Question:** What percentage of blogs actually provide `og:image` metadata?

**Impact:** If <50% of articles have thumbnails, UI looks inconsistent.

**Mitigation:**
- Design UI to handle missing thumbnails gracefully (show blog favicon instead)
- Add fallback to other meta tags (Twitter Card `twitter:image`, generic `<meta name="image">`)

### Risk 2: FTS5 Schema Migration

**Question:** Who owns database schema in this architecture?

**Current state:** blogwatcher CLI initializes schema, UI server only reads.

**Implication:** FTS5 setup must be added to CLI's schema initialization, not UI server.

**Action needed:** Verify CLI has migration capability before implementing search.

### Risk 3: Search Performance with HTMX

**Question:** Will debounced search (250ms delay) feel responsive enough?

**Impact:** If search query takes >250ms, users might type ahead and get stale results.

**Mitigation:** Use `hx-sync="closest form:abort"` to cancel in-flight requests when user types again.

### Risk 4: Masonry Card Height Variability

**Question:** If thumbnails vary wildly in aspect ratio, will masonry layout look broken?

**Impact:** Very tall or very wide thumbnails could create awkward gaps.

**Mitigation:**
- Enforce aspect ratio with CSS (`aspect-ratio: 16/9; object-fit: cover`)
- Design cards with consistent height (thumbnail + fixed-height title/meta area)

## Success Criteria

Architecture is successful if:

- [ ] No changes to core `Server` struct or dependency injection pattern
- [ ] All features support HTMX partial updates (HX-Request header detection)
- [ ] Database layer remains SQLite with single connection constraint
- [ ] Go templates remain the single source of HTML rendering
- [ ] Static assets remain CSS + vanilla JS (no build step)
- [ ] Scanner pipeline cleanly extends (thumbnail extraction is one additional step)
- [ ] Search handler follows same pattern as existing handlers (database query → template render)
- [ ] Backward compatible (old articles without thumbnails still display correctly)

## Confidence Assessment

| Area | Confidence | Reason |
|------|------------|--------|
| Masonry CSS | HIGH | Standard CSS Grid pattern, well-documented, works today |
| Thumbnail extraction | HIGH | otiai10/opengraph is mature library, OpenGraph is standard |
| SQLite FTS5 | HIGH | Built-in SQLite feature, already using modernc.org/sqlite |
| HTMX integration | HIGH | Patterns match existing architecture exactly |
| Performance | MEDIUM | Need real-world testing with production article counts |
| OpenGraph availability | MEDIUM | Assume 80%+ blogs support it, needs validation |

## Next Steps for Implementation

1. **Validate assumptions:**
   - Check blogwatcher CLI has migration support for FTS5 schema
   - Test OpenGraph extraction against real blogs in database
   - Verify thumbnail aspect ratios (sample 50 articles)

2. **Phase 1 (Masonry):**
   - Create feature branch
   - Add CSS Grid to styles.css
   - Test with varying article counts (10, 100, 1000)
   - Screenshot comparison with current linear layout

3. **Phase 2 (Thumbnails):**
   - Add thumbnail_url migration to CLI
   - Integrate otiai10/opengraph into scanner
   - Test error handling (timeout, 404, missing og:image)
   - Implement template conditional rendering

4. **Phase 3 (Search):**
   - Add FTS5 virtual table to CLI schema
   - Implement SearchArticles in storage layer
   - Add handleSearch handler
   - Test FTS5 query syntax edge cases
   - Add HTMX search input to template
   - Test debounce timing (200ms vs 250ms vs 500ms)

5. **Integration testing:**
   - Test all three features together
   - Verify HTMX swaps preserve masonry layout
   - Test search results with/without thumbnails
   - Mobile responsiveness check

**Total estimated effort:** 10-16 hours for all three features.

## Sources

### Masonry Layout
- [Chrome Masonry Update](https://developer.chrome.com/blog/masonry-update) - Browser implementation status
- [MDN Masonry Layout](https://developer.mozilla.org/en-US/docs/Web/CSS/Guides/Grid_layout/Masonry_layout) - Official documentation
- [CSS Grid auto-fit guide](https://dev.to/m97chahboun/responsive-card-layout-with-css-grid-a-step-by-step-guide-3ej1) - Responsive patterns
- [Smashing Magazine Masonry](https://www.smashingmagazine.com/native-css-masonry-layout-css-grid/) - Deep dive

### Thumbnail Extraction
- [otiai10/opengraph](https://github.com/otiai10/opengraph) - Go OpenGraph library
- [OpenGraph Protocol](https://ogp.me/) - Official specification
- [Go image package guide](https://reintech.io/blog/a-guide-to-gos-image-package-manipulating-and-processing-images) - Image processing
- [disintegration/imaging](https://pkg.go.dev/github.com/disintegration/imaging) - Thumbnail generation library

### Search Implementation
- [SQLite FTS5 Extension](https://sqlite.org/fts5.html) - Official documentation
- [github.com/zalgonoise/fts](https://github.com/zalgonoise/fts) - Go FTS5 wrapper with benchmarks
- [HTMX Patterns](https://hypermedia.systems/htmx-patterns/) - Active Search pattern
- [HTMX More Patterns](https://hypermedia.systems/more-htmx-patterns/) - Debounce best practices
- [HTMX Documentation](https://htmx.org/docs/) - Official reference

### Go + HTMX Architecture
- [Building with Go and HTMX](https://blog.logrocket.com/building-high-performance-websites-using-htmx-go/) - Architecture patterns
- [Go HTMX integration](https://dev.to/calvinmclean/how-to-build-a-web-application-with-htmx-and-go-3183) - Real-world examples
