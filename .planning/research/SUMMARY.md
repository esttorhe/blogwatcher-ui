# Project Research Summary

**Project:** BlogWatcher UI v1.1 - UI Polish & Search
**Domain:** RSS reader web application (Go/HTMX architecture)
**Researched:** 2026-02-03
**Confidence:** HIGH

## Executive Summary

BlogWatcher UI v1.1 adds masonry layout, clickable cards with thumbnails, and search capabilities to an existing Go/HTMX RSS reader. The research reveals this is achievable with minimal architectural changes and only one new dependency (opengraph/v2). The existing stack already includes everything needed: HTMX 2.0.4 for dynamic updates, modernc.org/sqlite with FTS5 support for search, and gofeed for RSS parsing including media extraction. The recommended approach leverages CSS-only masonry with progressive enhancement, server-side thumbnail extraction during sync (not render), and SQLite FTS5 for search with a clean migration path from LIKE queries.

The critical risks center on integration points: masonry layouts breaking after HTMX content swaps, thumbnail extraction creating N+1 query problems if done during render, and filter state desynchronizing between URL parameters and UI. These are all preventable with disciplined patterns: hook into HTMX's afterSwap event for layout recalculation, extract thumbnails during the scanner pipeline and cache in the database, and treat URL query parameters as the single source of truth for filter state. The existing HTMX partial update pattern and Go template rendering architecture handle the rest cleanly.

The technology choices reflect a mature understanding of the domain. CSS Grid auto-fit provides masonry-like layouts today without JavaScript complexity, with a clear upgrade path to native CSS masonry when browser support reaches 80%. FTS5 is built into SQLite and provides 750x faster search than LIKE queries with BM25 ranking. The thumbnail fallback chain (RSS media, Open Graph, favicon) matches industry patterns from established RSS readers like FreshRSS and Feedly. This is a polish milestone, not a rewrite—the architecture already supports everything needed.

## Key Findings

### Recommended Stack

v1.1 requires only one new Go dependency: `github.com/otiai10/opengraph/v2@v2.2.0` for thumbnail extraction. Everything else leverages the existing stack. modernc.org/sqlite already supports FTS5 out of the box for full-text search. gofeed (v1.3.0) already extracts media enclosures and thumbnails from RSS feeds. CSS masonry uses native Grid with progressive enhancement—no JavaScript library needed.

**Core technologies:**
- **github.com/otiai10/opengraph/v2**: Open Graph image extraction — actively maintained (Sep 2025), v2 module path, handles relative URLs, 28+ production importers
- **SQLite FTS5**: Full-text search with BM25 ranking — built into modernc.org/sqlite, 750x faster than LIKE queries, supports prefix search and relevance ranking
- **CSS Grid auto-fit**: Masonry-like layout — works today without JavaScript, responsive by design, progressive enhancement ready for native masonry
- **HTMX 2.0.4**: Active search pattern with debouncing — existing capability, no upgrade needed, hx-trigger with 250-350ms delay handles search UX

**What NOT to add:**
- JavaScript masonry libraries (Masonry.js, Isotope) — CSS solution matches HTMX server-rendered philosophy
- Alternative search engines (Bleve, typesense-go) — SQLite FTS5 sufficient for title search
- Image processing libraries — thumbnails are URLs, not processed locally
- Caching layers — local SQLite fast enough without additional complexity

### Expected Features

Research identified clear table stakes for modern RSS readers versus differentiators to defer. The v1.1 scope focuses on must-haves with one stretch goal (masonry) that provides strong visual polish.

**Must have (table stakes):**
- **Entire card clickable** — users expect full card as click target, not just title text (research shows removing "read more" links improves UX)
- **Hover/focus states** — visual feedback confirms clickability, keyboard focus rings for accessibility
- **Thumbnail fallback chain** — articles without featured images gracefully degrade to favicon (RSS media → Open Graph → favicon)
- **Search debouncing** — 300-350ms industry standard prevents server hammering
- **Clear search results count** — "Showing 47 articles" gives context, "No results" state when empty
- **Persistent filter state** — URL query params preserve search/date/blog filters on navigation
- **Mobile-responsive layout** — 1 col mobile, 2 col tablet, 3-4 col desktop (standard pattern)
- **Combined filter AND logic** — all active filters apply together (blog + status + date + search)

**Should have (competitive):**
- **Masonry layout option** — visual polish, Pinterest-style packed layout with view toggle
- **Search highlighting** — highlight matched terms in results (low complexity, nice UX enhancement)
- **Filter animation** — smooth transitions when filters change results (HTMX swap with CSS transitions)

**Defer (v2+):**
- **Saved searches** — power users can bookmark complex filter combinations (medium complexity, not blocking)
- **Keyboard shortcuts** — j/k navigation, x to mark read (requires minimal JS, good v1.2 feature)
- **Regex search** — power user feature, niche use case (SQLite REGEXP needs custom function)
- **Mixed width masonry** — some cards wider for featured articles (high complexity, diminishing returns)
- **Estimated read time** — "3 min read" requires fetching article content (out of scope)

### Architecture Approach

The existing Go/HTMX architecture supports all v1.1 features with minimal changes. Masonry is CSS-only, thumbnails extend the Article model and scanner pipeline, search adds one handler plus FTS5 integration. The architecture already follows clean patterns: dependency injection, HTMX partial update detection, Go templates for all rendering.

**Major components:**

1. **CSS Layer (static/styles.css)** — Add Grid auto-fit for masonry, aspect-ratio for thumbnails, no JavaScript needed. Progressive enhancement with @supports for future native masonry.

2. **Scanner Pipeline (internal/scanner/)** — Extend to call thumbnail extractor during article discovery. Implements fallback chain: RSS media (already parsed by gofeed) → Open Graph (new: opengraph/v2) → favicon (existing). Caches result in thumbnail_url column.

3. **Database Layer (internal/storage/)** — Add thumbnail_url column to articles table, create FTS5 virtual table with triggers for sync, implement SearchArticles method that joins articles_fts with articles table.

4. **HTTP Handlers (internal/server/handlers.go)** — Add handleSearch following existing pattern: parse filters from query params, call db.SearchArticles, detect HX-Request header, render full page or partial. URL params as single source of truth.

5. **Templates (templates/)** — Conditionally render thumbnails with `{{if .ThumbnailURL}}`, add search input with HTMX attributes (hx-get, hx-trigger with delay, hx-target, hx-include for filter state).

**Integration patterns to preserve:**
- HTMX partial updates (HX-Request header detection)
- Go templates as single source of HTML rendering
- SQLite single connection constraint (SetMaxOpenConns(1))
- Progressive enhancement (features degrade gracefully when data missing)

### Critical Pitfalls

Research identified four critical pitfalls that cause rewrites or major performance issues, plus several moderate/minor pitfalls.

1. **Masonry layout breaks after HTMX swaps** — When HTMX replaces content, masonry libraries don't recalculate positions automatically. Cards overlap or leave gaps. Prevention: Listen for HTMX `htmx:afterSwap` event and call layout recalculation. Reserve space for thumbnails with aspect-ratio CSS to prevent layout shift when images load. Test all swap scenarios (filters, pagination, mark as read).

2. **Thumbnail extraction creates N+1 query problem** — Naive implementation hits external URLs on every render, causing 5-30 second page loads and rate limiting. Prevention: Extract thumbnails during scanner sync (not template render), store in thumbnail_url column, implement fallback chain with 2s timeouts, use worker pool pattern (10 goroutines max) to batch process concurrently.

3. **LIKE search performance degrades rapidly** — Works fine with 100 articles, becomes unusable (2-5s) at 10K articles. LIKE '%term%' cannot use indexes. Prevention: Start with LIKE for MVP (acceptable up to ~5K articles), document limitations, plan migration to FTS5 when article count grows. FTS5 is 750x faster with BM25 ranking. Implement query timeouts (5s) and result limits (100).

4. **Filter state sync issues between URL and server** — Combined filters get out of sync between URL query params and UI state. Bookmarks show wrong results, back button breaks. Prevention: Treat URL as single source of truth (SSOT), use hx-push-url consistently, parse filters from query params on every request, pre-populate form controls from URL, test bookmark/back button scenarios explicitly.

5. **Cumulative Layout Shift (CLS) from thumbnails** — Cards render without thumbnails, then shift when images load. User clicks wrong target. Prevention: Reserve space with aspect-ratio CSS, set width/height attributes on img tags, use loading="lazy" for below-fold images, test with throttled network.

## Implications for Roadmap

Based on research findings, suggest a 3-phase structure focused on incremental value delivery with minimal risk.

### Phase 1: Enhanced Card Interaction
**Rationale:** Highest value for lowest effort. Makes existing list view feel modern without architectural changes. Pure UI enhancement that validates design patterns before more complex features.

**Delivers:** Clickable cards, hover/focus states, thumbnail support with fallback chain

**Addresses:**
- Table stakes: Entire card clickable, hover states, thumbnail fallback chain (from FEATURES.md)
- Architecture: Extends Article model with thumbnail_url, integrates opengraph/v2 into scanner (from ARCHITECTURE.md)

**Avoids:**
- Pitfall 2: Thumbnail extraction during sync, not render (prevents N+1 query problem)
- Pitfall 5: CLS prevention with aspect-ratio CSS and reserved space

**Implementation order:**
1. Add thumbnail_url column migration (database)
2. Integrate opengraph/v2 into scanner pipeline with fallback chain
3. Update templates to render thumbnails conditionally
4. Make entire card clickable with proper ARIA labels
5. Add hover/focus states for accessibility

**Estimated effort:** 8-12 hours

### Phase 2: Search & Date Filtering
**Rationale:** Core functionality for finding articles. Search is table stakes for any modern reader. Date filtering complements existing blog/status filters. Both features share filter state management patterns.

**Delivers:** Title search with debouncing, results count, date range shortcuts (last week/month/custom), combined filter logic

**Uses:**
- SQLite FTS5 from STACK.md (built into modernc.org/sqlite, no new dependencies)
- HTMX active search pattern (hx-trigger="keyup changed delay:300ms")
- URL as SSOT for filter state

**Implements:**
- handleSearch handler following existing pattern (from ARCHITECTURE.md)
- SearchArticles database method with FTS5 MATCH queries
- Combined filter query building (WHERE clauses compose with AND logic)

**Avoids:**
- Pitfall 3: Start with FTS5 (not LIKE) to avoid performance issues at scale
- Pitfall 4: URL params as single source of truth prevents state desync
- Pitfall 9: Escape LIKE wildcards, use parameterized queries

**Implementation order:**
1. Create FTS5 virtual table with triggers (database)
2. Add SearchArticles method to storage layer
3. Add handleSearch handler with filter parsing
4. Add search input to template with HTMX attributes
5. Implement date filter shortcuts (last week, month, custom range)
6. Update filter form to include all filter dimensions with hx-push-url

**Estimated effort:** 10-14 hours

### Phase 3: Masonry Layout (Stretch Goal)
**Rationale:** Visual polish that differentiates from basic list view. Lower priority than search (which is table stakes) but provides strong aesthetic improvement. CSS-only implementation matches HTMX philosophy and minimizes risk.

**Delivers:** View toggle (list ↔ masonry), responsive masonry grid, persistence of view preference

**Uses:**
- CSS Grid auto-fit (from STACK.md, no JavaScript library)
- Progressive enhancement with @supports for future native masonry
- localStorage or cookie for view preference

**Avoids:**
- Pitfall 1: HTMX afterSwap event listener triggers layout recalculation if using JS library (CSS Grid auto-fit doesn't need this)
- Pitfall 10: CSS auto-fit responsive by design (no manual breakpoint management)
- Anti-pattern: Avoiding JavaScript libraries keeps implementation simple and maintainable

**Implementation order:**
1. Add CSS Grid auto-fit styles with media queries
2. Add view toggle button (updates class on container)
3. Implement view preference persistence (localStorage)
4. Add @supports query for future grid-lanes enhancement
5. Test HTMX swap scenarios (filters, pagination)

**Estimated effort:** 6-10 hours

**Note:** This phase can be deferred to v1.2 if time constraints. Phases 1-2 deliver complete functional improvement without masonry.

### Phase Ordering Rationale

- **Phase 1 first** because thumbnails must be in place before masonry layout makes sense. Masonry without varied card heights (from thumbnails) provides minimal value over standard grid.
- **Phase 2 second** because search is higher priority than masonry (table stakes vs. nice-to-have). Date filtering shares URL state management patterns with search, efficient to implement together.
- **Phase 3 third** because masonry is pure presentation layer enhancement. If Phase 3 gets cut for time, Phases 1-2 still deliver complete user value.
- **Parallelization not recommended** due to shared concerns: all three phases modify templates, Phase 2 and 3 both affect HTMX swap behavior, merge conflicts likely.

### Research Flags

**Phases with standard patterns (skip research-phase):**
- **Phase 1 (Enhanced Cards):** Well-documented patterns for thumbnail extraction, RSS media parsing, Open Graph standard. opengraph/v2 library proven (28+ production importers). Accessibility patterns for clickable cards established.
- **Phase 2 (Search & Filters):** SQLite FTS5 extensively documented (official SQLite docs), HTMX active search pattern in official examples, URL state management well-understood. No novel integration challenges.
- **Phase 3 (Masonry):** CSS Grid auto-fit is standard responsive pattern (MDN documentation complete). Progressive enhancement approach proven. No research-phase needed.

**No phases require deeper research.** All techniques use established patterns with high-quality documentation. Implementation can proceed directly to requirements definition after this research summary.

### Dependencies Between Phases

```
Phase 1 (Enhanced Cards)
    ├─ No dependencies
    └─ Blocks: Phase 3 (masonry needs thumbnails for visual impact)

Phase 2 (Search & Filters)
    ├─ No dependencies
    └─ Independent of Phases 1 and 3

Phase 3 (Masonry Layout)
    ├─ Depends on: Phase 1 (thumbnails make masonry worthwhile)
    └─ Can be implemented after Phase 1 completes
```

**Sequential implementation required:** Phase 1 → Phase 2 → Phase 3
**Phases 2 and 3 could theoretically be parallelized** but not recommended (both modify templates, potential conflicts).

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Only one new dependency (opengraph/v2), verified on pkg.go.dev and GitHub. All other capabilities built into existing stack (FTS5, gofeed media parsing, HTMX patterns). Version numbers confirmed. |
| Features | HIGH | Cross-referenced three major RSS reader UX studies (Zapier, VPNTierLists, FeedViewer). Table stakes clearly defined. Differentiators distinguished from anti-features. Complexity estimates based on similar implementations. |
| Architecture | HIGH | Existing codebase patterns analyzed. Integration points identified with specific file locations. HTMX patterns match official documentation. Scanner pipeline extension point clear. No architectural changes required. |
| Pitfalls | HIGH | Critical pitfalls verified with official docs (MDN for CLS, SQLite for FTS5, HTMX for state management). GitHub issues for masonry layout recalculation. Real-world RSS parsing edge cases documented. Prevention strategies proven. |

**Overall confidence:** HIGH

All research findings based on official documentation, established libraries, and proven patterns. No speculative or untested approaches. The existing BlogWatcher UI architecture already demonstrates best practices (HTMX partial updates, Go template rendering, SQLite storage)—this milestone extends those patterns without introducing new paradigms.

### Gaps to Address

**OpenGraph availability validation:** Assumption that 80%+ of blogs provide og:image metadata needs validation against actual RSS feed corpus. Mitigation: Design handles missing thumbnails gracefully (falls back to favicon), no feature degradation.

**FTS5 schema ownership:** Current architecture has blogwatcher CLI initialize schema, UI server only reads. FTS5 virtual table setup must be added to CLI's migration capability, not UI server. Action: Verify CLI has migration support before Phase 2 implementation begins.

**Thumbnail aspect ratios:** Sample 50 articles from production feeds to verify thumbnail dimensions are reasonably consistent. If wild variation (some 1:1, some 16:9, some vertical), may need to enforce uniform aspect ratio with CSS object-fit: cover. Mitigation: aspect-ratio CSS property handles this.

**Search query performance baseline:** Measure current article count in production and establish baseline for LIKE query performance. If already >5K articles, implement FTS5 immediately. If <1K articles, LIKE acceptable as MVP with documented migration path. Action: Check production database before Phase 2 planning.

## Sources

Research aggregated from 4 specialized files (STACK.md, FEATURES.md, ARCHITECTURE.md, PITFALLS.md). Each file includes detailed source citations. Summary of primary sources below.

### Primary (HIGH confidence)

**Stack & Technology:**
- [SQLite FTS5 Extension](https://sqlite.org/fts5.html) — Official documentation for full-text search
- [MDN: CSS Grid Layout Masonry](https://developer.mozilla.org/en-US/docs/Web/CSS/Guides/Grid_layout/Masonry_layout) — Official CSS Grid masonry guide
- [pkg.go.dev: otiai10/opengraph/v2](https://pkg.go.dev/github.com/otiai10/opengraph/v2) — Go OpenGraph library documentation
- [HTMX Documentation](https://htmx.org/docs/) — Official HTMX reference
- [HTMX Active Search Example](https://htmx.org/examples/active-search/) — Official pattern for search with debouncing

**Architecture & Patterns:**
- [WebKit Blog: CSS Grid Lanes](https://webkit.org/blog/17660/introducing-css-grid-lanes/) — Native masonry status and browser roadmap
- [Smashing Magazine: Native CSS Masonry](https://www.smashingmagazine.com/native-css-masonry-layout-css-grid/) — CSS Grid masonry implementation patterns
- [Hypermedia Systems: HTMX Patterns](https://hypermedia.systems/htmx-patterns/) — Active search and state management patterns

**RSS & Thumbnails:**
- [gofeed Documentation](https://pkg.go.dev/github.com/mmcdole/gofeed) — RSS/Atom parsing with media extraction
- [OpenGraph Protocol](https://ogp.me/) — Official og:image specification
- [RSS Enclosure Element](https://www.w3schools.com/xml/rss_tag_enclosure.asp) — RSS media attachment standard
- [Media RSS Specification](https://www.rssboard.org/media-rss) — media:thumbnail extension

**UX & Features:**
- [The 3 Best RSS Reader Apps (Zapier)](https://zapier.com/blog/best-rss-feed-reader-apps/) — RSS reader UX analysis
- [RSS Reader UI Design Principles](https://www.feedviewer.app/answers/rss-reader-user-interface-design-principles) — User expectation research
- [FreshRSS Filtering Documentation](https://freshrss.github.io/FreshRSS/en/users/10_filter.html) — Real-world filter implementation patterns

### Secondary (MEDIUM confidence)

**Performance & Optimization:**
- [Use The Index, Luke: LIKE Performance](https://use-the-index-luke.com/sql/where-clause/searching-for-ranges/like-performance-tuning) — SQL LIKE query optimization analysis
- [Full-Text Search in SQLite: Practical Guide](https://medium.com/@johnidouglasmarangon/full-text-search-in-sqlite-a-practical-guide-80a69c3f42a4) — FTS5 implementation guide
- [Making SQLite Faster in Go](https://turriate.com/articles/making-sqlite-faster-in-go) — Performance patterns for Go SQLite usage

**Pitfalls & Edge Cases:**
- [Bookmarkable by Design: URL-Driven State in HTMX](https://www.lorenstew.art/blog/bookmarkable-by-design-url-state-htmx/) — Filter state management patterns
- [Cumulative Layout Shift (CLS) Guide 2026](https://medium.com/@sahoo.arpan7/cumulative-layout-shift-cls-guide-to-one-of-the-most-misunderstood-core-web-vitals-5f135c68cb6f) — Web performance best practices
- [Masonry.js Issue #1000](https://github.com/desandro/masonry/issues/1000) — Dynamic content layout recalculation edge cases

**Accessibility:**
- [Accessible Card UI Component Patterns](https://dap.berkeley.edu/web-a11y-basics/accessible-card-ui-component-patterns) — Clickable card accessibility patterns
- [Inclusive Components: Cards](https://inclusive-components.design/cards/) — Card component a11y best practices

### Tertiary (LOW confidence)

No tertiary sources used. All findings based on official documentation or established community patterns with multiple source corroboration.

---
*Research completed: 2026-02-03*
*Ready for roadmap: yes*
