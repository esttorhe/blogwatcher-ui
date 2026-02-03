# Domain Pitfalls: Adding Masonry Layout, Thumbnails, and Search

**Domain:** Adding UI polish and search to existing Go/HTMX web application
**Project:** BlogWatcher UI v1.1
**Researched:** 2026-02-03
**Confidence:** HIGH (verified with official docs and real-world implementations)

## Executive Summary

Adding masonry layout, thumbnail extraction, and search to an existing Go/HTMX application introduces integration challenges around state management, dynamic content coordination, and performance optimization. The most critical risks are: (1) HTMX swap animations breaking masonry layouts without explicit recalculation, (2) thumbnail extraction creating N+1 query problems and cache misses, (3) LIKE search performance degrading rapidly beyond 10K articles, and (4) filter state getting out of sync between URL parameters and server state.

---

## Critical Pitfalls

These mistakes cause rewrites, major performance degradation, or broken user experience.

### Pitfall 1: Masonry Layout Breaks After HTMX Swaps

**What goes wrong:**

When HTMX swaps content into a masonry grid container, the layout algorithm doesn't recalculate positions automatically. Cards overlap, leave gaps, or stack incorrectly because the masonry layout library (or CSS Grid masonry) doesn't know the DOM changed. This is particularly problematic with:
- Partial updates (filtering articles)
- Infinite scroll pagination
- Individual card state changes (marking as read)

**Why it happens:**

Masonry layouts require explicit layout recalculation when:
- Container dimensions change
- Items are added/removed
- Item heights change (images load, thumbnails appear)

HTMX's `hx-swap` replaces DOM content but doesn't notify layout libraries. CSS Grid masonry (still experimental, `display: grid-lanes` proposal) has no JavaScript API to trigger recalculation.

**Consequences:**

- Cards overlap making content unreadable
- Gaps in layout look broken
- User loses confidence in the UI quality
- Fixing this post-launch requires adding JavaScript hooks everywhere HTMX swaps content

**Prevention:**

1. **If using CSS Grid masonry:** Be aware this is still experimental (Firefox, Safari TP). Limited browser support as of 2026. Must provide JavaScript-based fallback for Chrome/Edge.

2. **If using JavaScript masonry library (e.g., Masonry.js, Isotope):**
   - Listen for HTMX `htmx:afterSwap` event
   - Call layout recalculation on swap target
   - Example pattern:
   ```javascript
   document.body.addEventListener('htmx:afterSwap', function(evt) {
     if (evt.detail.target.classList.contains('masonry-grid')) {
       masonryInstance.layout();
     }
   });
   ```

3. **Reserve space for thumbnails to prevent layout shift:**
   - Set `min-height` on card containers matching thumbnail height
   - Use aspect-ratio CSS property: `aspect-ratio: 16/9; width: 100%;`
   - This prevents masonry recalculation when images load

4. **Test swap scenarios explicitly:**
   - Filter changes (status, blog, search)
   - Pagination loads
   - Individual card updates (mark as read)
   - Theme toggle (dark/light affects heights due to different text rendering)

**Detection:**

Early warning signs:
- Cards overlap when switching filters
- Layout looks correct on initial page load but breaks after interaction
- Different card heights in inspector than visual layout
- Console errors about missing elements during masonry calculation

**Phase impact:** Phase 1 (Masonry Layout) — Block merging until recalculation hooks proven working

**Sources:**
- [Native CSS Masonry Layout In CSS Grid — Smashing Magazine](https://www.smashingmagazine.com/native-css-masonry-layout-css-grid/)
- [Resizable grid items with dynamic content · Issue #1000 · desandro/masonry](https://github.com/desandro/masonry/issues/1000)
- [HTMX Examples - Animations](https://htmx.org/examples/animations/)

---

### Pitfall 2: Thumbnail Extraction Creates N+1 Query Problem

**What goes wrong:**

Naive thumbnail extraction hits external URLs (Open Graph, favicons) on every article render, causing:
- Page load times of 5-30 seconds for 50 articles
- HTTP request storms (50+ concurrent requests)
- RSS feeds/websites rate-limiting or blocking your IP
- Timeouts leaving blank thumbnails

**Why it happens:**

The fallback chain (RSS media → Open Graph → favicon) requires HTTP requests. Without caching, the template rendering code makes these requests synchronously for each article. In Go templates, you can't easily do concurrent fetching without pre-processing.

The existing code only stores `thumbnail_url` (if any) in the database — it doesn't cache the extracted/fallback thumbnails.

**Consequences:**

- Unacceptable page load performance (5-30s for article list)
- User bounces before page finishes rendering
- External sites rate-limit or ban your server IP
- Inconsistent thumbnails (some load, some timeout)
- Server CPU/memory spikes from concurrent HTTP client pool exhaustion

**Prevention:**

1. **Add `thumbnail_url` column to articles table if not exists:**
   ```sql
   ALTER TABLE articles ADD COLUMN thumbnail_url TEXT;
   ```

2. **Extract thumbnails during scan/sync, not during render:**
   - When scanner discovers articles, immediately extract thumbnails
   - Store result in `thumbnail_url` column
   - Template just renders cached URL, no HTTP requests at render time

3. **Implement fallback chain with timeouts:**
   ```go
   func extractThumbnail(article model.Article, blog model.Blog) string {
     // 1. RSS media:content or media:thumbnail (already in feed parse)
     if article.MediaURL != "" {
       return article.MediaURL
     }

     // 2. Open Graph (fetch article URL with 2s timeout)
     if ogImage := fetchOpenGraph(article.URL, 2*time.Second); ogImage != "" {
       return ogImage
     }

     // 3. Favicon fallback
     return faviconURL(blog.URL) // existing function
   }
   ```

4. **Use HTTP client with proper configuration:**
   ```go
   var httpClient = &http.Client{
     Timeout: 2 * time.Second,
     Transport: &http.Transport{
       MaxIdleConns:        100,
       MaxIdleConnsPerHost: 10,
       IdleConnTimeout:     30 * time.Second,
     },
   }
   ```

5. **Batch process thumbnails with concurrency control:**
   - Use worker pool pattern (10 goroutines max)
   - Process thumbnails for all new articles in parallel during sync
   - Don't block the main sync operation on thumbnail failures

6. **Serve thumbnails through proxy if needed:**
   - Some sites block hotlinking of images
   - Add `/thumbnail-proxy?url=` endpoint that caches and serves images
   - Avoids CORS issues and provides fallback serving

**Detection:**

Warning signs:
- `ListArticlesWithBlog` handler takes >1s to respond
- Network tab shows 50+ concurrent favicon/image requests
- RSS feed providers send abuse warnings
- Blank thumbnails appearing intermittently
- Server log shows HTTP client timeout errors

**Phase impact:** Phase 2 (Thumbnails) — Critical to implement during scanner integration, not as template helper

**Sources:**
- [How to Set Up Feed to Post's Image Options - WP RSS Aggregator](https://kb.wprssaggregator.com/article/308-how-to-set-up-feed-to-posts-image-options)
- [Getting those thumbnails from Medium RSS Feed](https://medium.com/@kartikyathakur/getting-those-thumbnails-from-medium-rss-feed-183f74aefa8c)
- [Imagor - Fast, secure image processing server](https://github.com/cshum/imagor)

---

### Pitfall 3: LIKE Search Performance Degrades Rapidly

**What goes wrong:**

Implementing title search with SQLite `LIKE '%term%'` works fine with 100 articles but becomes unusably slow (2-5s queries) at 10K articles. Users type in search box expecting instant results but page hangs, browser shows loading spinner indefinitely.

**Why it happens:**

LIKE with wildcards on both sides (`%term%`) cannot use indexes efficiently. SQLite must scan every row and perform string matching. The query complexity is O(n * m) where n = article count, m = title length.

Prefix-only LIKE (`term%`) can use indexes, but infix search (`%term%`) cannot. Users expect Google-style "search anywhere in title" functionality.

**Consequences:**

- Search feels broken (multi-second delays)
- Users abandon search feature
- Server CPU spikes during searches
- Concurrent searches can exhaust SQLite connection pool (max 1 connection)
- Poor user experience compared to v1.0 (which had no search)

**Prevention:**

1. **Start with simple LIKE for MVP, document limitations:**
   - Works acceptably up to ~5K articles
   - Query: `WHERE title LIKE '%' || ? || '%'`
   - Add note in UI: "Searching titles only"

2. **Plan migration to FTS5 when article count grows:**
   - SQLite FTS5 is **750x faster** than LIKE for full-text search
   - Create virtual table: `CREATE VIRTUAL TABLE articles_fts USING fts5(title, content=articles, content_rowid=id);`
   - Query: `SELECT * FROM articles_fts WHERE title MATCH ?`
   - Requires maintaining FTS index (triggers or manual sync)

3. **Implement query timeout protection:**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   rows, err := db.QueryContext(ctx, query, args...)
   ```

4. **Add search result limits:**
   - Limit to 100 results: `LIMIT 100`
   - Show "100+ results, refine search" message
   - Prevents massive result sets from compounding slow query

5. **Debounce search input on client side:**
   - Don't trigger HTMX request on every keystroke
   - Wait 300ms after user stops typing
   - Use `hx-trigger="keyup changed delay:300ms"`

6. **Consider client-side filtering for small datasets:**
   - If article count < 1000, fetch all and filter with JavaScript
   - Avoids server round-trip for every keystroke
   - Trade-off: larger initial payload vs. instant search

**Migration path to FTS5:**

When article count > 5K or search performance complaints:
1. Create FTS5 virtual table
2. Populate from existing articles
3. Add triggers to keep FTS5 in sync with articles table
4. Update search query to use MATCH instead of LIKE
5. Zero downtime: check article count, route to FTS5 vs LIKE automatically

**Detection:**

Warning signs:
- Search queries taking >500ms in logs
- Users reporting "search is slow"
- Database CPU usage spikes when search is used
- SQLite lock timeouts during concurrent searches
- Query plans show "SCAN TABLE articles" (no index usage)

**Phase impact:** Phase 3 (Search) — Start with LIKE, flag for FTS5 migration in Phase 5+

**Sources:**
- [Tuning SQL LIKE using indexes](https://use-the-index-luke.com/sql/where-clause/searching-for-ranges/like-performance-tuning)
- [SQLite FTS5 Extension](https://sqlite.org/fts5.html)
- [Full-Text Search in SQLite: A Practical Guide](https://medium.com/@johnidouglasmarangon/full-text-search-in-sqlite-a-practical-guide-80a69c3f42a4)

---

### Pitfall 4: Filter State Sync Issues Between URL and Server

**What goes wrong:**

Combined filters (blog + status + search + date) get out of sync between URL query params and server-rendered state. User bookmarks a filtered view, returns later, sees different results or broken state. Back button shows wrong filters. Sharing URL with colleague shows different articles.

Examples:
- URL shows `?blog=5&status=unread&search=golang` but sidebar shows "All articles"
- User clicks back button, URL changes but article list doesn't update
- Filter state lost when HTMX swaps main content but not sidebar
- Date filter applied but not reflected in URL (can't bookmark/share)

**Why it happens:**

HTMX's partial page updates can swap main content without updating sidebar state. Filter controls (dropdowns, inputs) live in different DOM sections than the article list. Without careful state coordination:
- URL params diverge from rendered state
- Multiple HTMX requests race, last-write-wins
- hx-push-url only updates URL, doesn't re-read it server-side

The existing v1.0 code has simple Inbox/Archived views but v1.1 adds multiple filter dimensions that must compose.

**Consequences:**

- Bookmarked/shared URLs show wrong results
- Back button breaks filtering
- User loses trust in filter accuracy
- Debug complaints like "search doesn't work" when actually it's state desync
- Confusion between what user sees and what URL indicates

**Prevention:**

1. **Treat URL as single source of truth (SSOT):**
   - Server always reads filters from query params
   - Never trust hidden form values or session state
   - Every filter change updates URL

2. **Use hx-push-url consistently:**
   ```html
   <form hx-get="/articles" hx-target="#article-list" hx-push-url="true">
     <select name="blog_id">...</select>
     <select name="status">...</select>
     <input name="search" hx-trigger="keyup changed delay:300ms">
     <input name="date_from">
     <input name="date_to">
   </form>
   ```

3. **Synchronize filter UI state from URL on page load:**
   ```go
   func parseFilters(r *http.Request) Filters {
     return Filters{
       BlogID:   parseIntParam(r, "blog_id"),
       Status:   r.URL.Query().Get("status"),
       Search:   r.URL.Query().Get("search"),
       DateFrom: r.URL.Query().Get("date_from"),
       DateTo:   r.URL.Query().Get("date_to"),
     }
   }

   // Template renders form with current values pre-selected
   tmpl.Execute(w, data{
     Filters:  parseFilters(r),
     Articles: filteredArticles,
   })
   ```

4. **Handle URL length limits:**
   - Browser URL limit: ~2000 characters
   - Long search terms can exceed this
   - For complex filters, consider abbreviated param names: `b=5&s=unread&q=term`
   - Or move some state server-side with session tokens

5. **Validate and sanitize all URL params:**
   ```go
   func parseIntParam(r *http.Request, name string) *int64 {
     val := r.URL.Query().Get(name)
     if val == "" {
       return nil
     }
     i, err := strconv.ParseInt(val, 10, 64)
     if err != nil {
       return nil // Invalid param = ignore
     }
     return &i
   }
   ```

6. **Test filter combinations explicitly:**
   - All filters applied together
   - Removing one filter at a time
   - Bookmark URL, close browser, reopen
   - Share URL in incognito window
   - Use back/forward buttons

7. **Make filter form changes update everything:**
   - Option 1: Entire page swap on filter change (simpler, more reliable)
   - Option 2: Swap both article list AND sidebar active states (complex, requires multiple hx-swap-oob)
   - Prefer option 1 for correctness, optimize later if needed

**Detection:**

Warning signs:
- URL shows filter but UI doesn't reflect it
- Back button changes URL but not article list
- Shared URLs work differently for different users
- Filter controls show default state but articles are filtered
- Browser console shows HTMX errors about missing swap targets

**Phase impact:** Phase 4 (Combined Filters) — Critical pattern to establish early, affects all filter features

**Sources:**
- [Bookmarkable by Design: URL-Driven State in HTMX](https://www.lorenstew.art/blog/bookmarkable-by-design-url-state-htmx/)
- [HTMX Simplifies State Management by Using URL Parameters](https://thedigipress.com/articles/1119)
- [HTMX hx-params attribute](https://htmx.org/attributes/hx-params/)

---

## Moderate Pitfalls

These cause delays, technical debt, or workarounds but are recoverable.

### Pitfall 5: Cumulative Layout Shift (CLS) from Thumbnails

**What goes wrong:**

Article cards render without thumbnails, then thumbnails load and cards shift down, pushing content. User tries to click article title but thumbnail appears and they click wrong card. Google penalizes site in search rankings for poor CLS score.

**Why it happens:**

HTML renders with unknown image dimensions. Browser allocates 0 height initially, then reflows when `<img>` loads and actual dimensions are known. In masonry layout, this triggers full layout recalculation.

**Prevention:**

1. **Reserve space with CSS:**
   ```css
   .article-card-thumbnail {
     aspect-ratio: 16 / 9;
     width: 100%;
     min-height: 200px;
     background: var(--card-bg);
   }
   ```

2. **Set width/height attributes on `<img>` tags:**
   ```html
   <img src="{{.ThumbnailURL}}" alt="" width="320" height="180" loading="lazy">
   ```
   Browser uses these for layout even before image loads.

3. **Use skeleton screens or placeholder:**
   - Show gray box matching final thumbnail size
   - Swap to real thumbnail when loaded
   - No layout shift because dimensions match

4. **Lazy load thumbnails below fold:**
   ```html
   <img src="{{.ThumbnailURL}}" loading="lazy">
   ```

**Phase impact:** Phase 2 (Thumbnails) — Address during implementation, test with throttled network

**Sources:**
- [Cumulative Layout Shift (CLS): The Most Misunderstood Core Web Vital (2026 Guide)](https://medium.com/@sahoo.arpan7/cumulative-layout-shift-cls-guide-to-one-of-the-most-misunderstood-core-web-vitals-5f135c68cb6f)
- [Optimize Cumulative Layout Shift](https://web.dev/articles/optimize-cls)

---

### Pitfall 6: RSS Feed Media Enclosure Parsing Edge Cases

**What goes wrong:**

Some RSS feeds have thumbnails in unexpected formats:
- Multiple `<enclosure>` tags (spec says max 1, some feeds ignore this)
- Media RSS (`<media:thumbnail>`) vs. standard enclosure
- Enclosure URL is relative not absolute
- Enclosure type is `audio/mpeg` when only images wanted
- Thumbnail URL requires authentication/cookies

**Why it happens:**

RSS spec allows 1 enclosure per item but Atom allows multiple. Media RSS extends RSS with richer media metadata. Feed publishers don't always follow specs. Real-world feeds are messy.

The existing `internal/rss/rss.go` may not parse all media element variants.

**Prevention:**

1. **Use robust RSS parser library:**
   - `gofeed` (github.com/mmcdole/gofeed) handles RSS, Atom, JSON feeds
   - Unified feed model abstracts differences
   - Handles media:content, media:thumbnail, enclosures

2. **Implement cascading extraction:**
   ```go
   func extractMediaURL(item *gofeed.Item) string {
     // 1. Check media:thumbnail extension
     if item.Extensions != nil {
       if media := item.Extensions["media"]; media != nil {
         if thumb := media["thumbnail"]; thumb != nil && len(thumb) > 0 {
           if url := thumb[0].Attrs["url"]; url != "" {
             return url
           }
         }
       }
     }

     // 2. Check enclosures for image types
     for _, enc := range item.Enclosures {
       if strings.HasPrefix(enc.Type, "image/") {
         return enc.URL
       }
     }

     // 3. Check item.Image (Atom/RSS 2.0)
     if item.Image != nil && item.Image.URL != "" {
       return item.Image.URL
     }

     return "" // No media found, will fall back to OG/favicon
   }
   ```

3. **Normalize URLs to absolute:**
   ```go
   func makeAbsoluteURL(baseURL, mediaURL string) string {
     base, err := url.Parse(baseURL)
     if err != nil {
       return mediaURL
     }
     media, err := url.Parse(mediaURL)
     if err != nil {
       return mediaURL
     }
     return base.ResolveReference(media).String()
   }
   ```

4. **Filter by MIME type:**
   - Only accept `image/*` enclosures
   - Ignore `audio/*`, `video/*` unless also has image thumbnail

5. **Test with diverse RSS feeds:**
   - WordPress blogs (use media:content)
   - Medium (thumbnails in description HTML)
   - YouTube (video enclosures + thumbnail)
   - Podcast feeds (audio enclosures, no images)
   - Atom feeds (multiple enclosures)

**Phase impact:** Phase 2 (Thumbnails) — Verify during RSS parser integration

**Sources:**
- [RSS enclosure Element](https://www.w3schools.com/xml/rss_tag_enclosure.asp)
- [Media RSS Specification](https://www.rssboard.org/media-rss)
- [gofeed - robust RSS and Atom Parser for Go](https://github.com/mmcdole/gofeed)

---

### Pitfall 7: Date Filtering Edge Cases (Timezone, Precision)

**What goes wrong:**

Date filtering for "last week" or "custom range" shows wrong articles:
- "Last week" includes articles from 8 days ago (off-by-one)
- Custom date range misses articles published at 11:59 PM
- Timezone confusion: Server time vs. user time vs. article published time
- Articles with null `published_date` disappear from filtered results

**Why it happens:**

SQLite stores timestamps as strings (RFC3339Nano format). Date arithmetic and comparisons need careful handling:
- "Last week" = past 7 days including today? Or Monday-Sunday of last week?
- Date range filtering: inclusive or exclusive bounds?
- `published_date` can be NULL (article lacks metadata)
- Article times may be UTC, user may be in different timezone

**Prevention:**

1. **Define "last week/month" semantics clearly:**
   ```go
   // "Last week" = past 7 days from now
   func lastWeek() time.Time {
     return time.Now().AddDate(0, 0, -7)
   }

   // "Last month" = past 30 days
   func lastMonth() time.Time {
     return time.Now().AddDate(0, 0, -30)
   }
   ```

2. **Use inclusive date ranges:**
   ```sql
   WHERE published_date >= ? AND published_date < ?
   ```
   - From date: inclusive (>=)
   - To date: exclusive (<) or use end of day (23:59:59)

3. **Handle NULL published dates:**
   ```sql
   WHERE (published_date >= ? OR published_date IS NULL)
   ```
   - Include articles without published date in results
   - Or explicitly filter them: checkbox "Include articles without dates"

4. **Store and compare in UTC:**
   - SQLite stores RFC3339 with timezone
   - Go `time.Time` handles timezone conversions
   - Format for query: `time.Now().UTC().Format(time.RFC3339)`

5. **Test edge cases:**
   - Articles published exactly at midnight
   - Articles from different timezones
   - NULL published dates
   - Future published dates (scheduled posts)
   - Date range spanning months/years

6. **Show timezone in UI:**
   - "Last 7 days (based on your system time)"
   - Or let user pick timezone in settings

**Phase impact:** Phase 4 (Date Filtering) — Define semantics early, document in code comments

**Sources:**
- [SQLite Date and Time Functions](https://www.sqlite.org/lang_datefunc.html)
- Go time package documentation

---

### Pitfall 8: Favicon Fallback Fails for Modern Sites

**What goes wrong:**

Fallback to `/favicon.ico` returns 404 for many modern sites. Sites use `/icon.svg`, `/apple-touch-icon.png`, or define favicon in `<link rel="icon">` HTML. Blogs without proper favicons show broken image icons.

**Why it happens:**

Modern favicon best practices (2026):
- SVG favicons with dark mode support
- Multiple sizes in site.webmanifest
- `<link rel="icon">` in HTML, not just `/favicon.ico` convention

Simply requesting `https://example.com/favicon.ico` works for legacy sites but not modern ones.

**Prevention:**

1. **Implement smart favicon extraction:**
   ```go
   func extractFavicon(siteURL string) string {
     // 1. Try standard locations
     standardPaths := []string{
       "/favicon.ico",
       "/icon.svg",
       "/favicon.svg",
       "/apple-touch-icon.png",
     }

     for _, path := range standardPaths {
       url := siteURL + path
       if headCheck(url) { // HTTP HEAD request
         return url
       }
     }

     // 2. Parse HTML <link rel="icon">
     if iconURL := parseHTMLFavicon(siteURL); iconURL != "" {
       return iconURL
     }

     // 3. Use third-party service as last resort
     return "https://www.google.com/s2/favicons?domain=" + siteURL
   }
   ```

2. **Use third-party favicon services:**
   - Google: `https://www.google.com/s2/favicons?domain=example.com&sz=128`
   - DuckDuckGo: `https://icons.duckduckgo.com/ip3/example.com.ico`
   - These aggregate from multiple sources

3. **Cache favicon URLs in database:**
   - Add `favicon_url` column to `blogs` table
   - Extract once during blog discovery/sync
   - Re-check periodically (weekly) in background

4. **Provide default placeholder:**
   - Generic blog icon SVG for sites with no favicon
   - Colored based on blog name hash (visual distinction)

5. **Test with diverse sites:**
   - Old blogs (only favicon.ico)
   - Modern sites (SVG favicons)
   - Sites with custom manifest
   - Sites with no favicon at all

**Phase impact:** Phase 2 (Thumbnails) — Implement smart fallback, not just `/favicon.ico`

**Sources:**
- [How to Favicon in 2026: Three files that fit most needs](https://evilmartians.com/chronicles/how-to-favicon-in-2021-six-files-that-fit-most-needs)
- [extract-favicon library](https://github.com/AlexMili/extract_favicon)

---

### Pitfall 9: Search with Special Characters Breaks SQL Query

**What goes wrong:**

User searches for `O'Reilly` or `100%` and gets SQL syntax error or no results. Special characters in search input aren't escaped properly for SQL LIKE queries. Or worse: SQL injection vulnerability.

**Why it happens:**

LIKE has special characters: `%` (wildcard), `_` (single char wildcard). User input containing these breaks the query. SQLite parameterized queries prevent injection but don't escape LIKE wildcards.

Example:
```go
// WRONG: User input "100%" becomes LIKE '%100%%' which matches unwanted results
query := "SELECT * FROM articles WHERE title LIKE '%' || ? || '%'"
```

**Prevention:**

1. **Escape LIKE wildcards in user input:**
   ```go
   func escapeLikeWildcards(input string) string {
     input = strings.ReplaceAll(input, "%", "\\%")
     input = strings.ReplaceAll(input, "_", "\\_")
     return input
   }

   // Usage
   searchTerm := escapeLikeWildcards(r.URL.Query().Get("search"))
   query := "SELECT * FROM articles WHERE title LIKE '%' || ? || '%' ESCAPE '\\'"
   ```

2. **Use parameterized queries (always):**
   ```go
   // CORRECT: Prevents SQL injection
   rows, err := db.Query(query, searchTerm)

   // WRONG: SQL injection vulnerability
   query := fmt.Sprintf("... WHERE title LIKE '%%%s%%'", userInput)
   ```

3. **Validate search input:**
   - Maximum length (e.g., 200 chars)
   - Reject malicious patterns (though parameterized queries handle this)

4. **Test with adversarial input:**
   - `O'Reilly` (single quote)
   - `100%` (percent sign)
   - `data_base` (underscore)
   - `'; DROP TABLE articles; --` (SQL injection attempt)
   - Emoji and Unicode characters

**Phase impact:** Phase 3 (Search) — Must handle from day 1 of search implementation

**Sources:**
- [SQLite LIKE operator documentation](https://www.sqlite.org/lang_expr.html#like)
- Go database/sql package (parameterized queries)

---

## Minor Pitfalls

These cause annoyance, poor UX, or small bugs but are easily fixed.

### Pitfall 10: Masonry Layout Column Count Not Responsive

**What goes wrong:**

Desktop shows 3 columns, mobile still shows 3 columns with tiny cards. Or mobile shows 1 column but tablet should show 2. Layout doesn't adapt to viewport width gracefully.

**Prevention:**

1. **Use CSS Grid with auto-fill:**
   ```css
   .masonry-grid {
     display: grid;
     grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
     gap: 1rem;
   }
   ```

2. **Or use media queries for explicit breakpoints:**
   ```css
   .masonry-grid { grid-template-columns: 1fr; } /* Mobile */
   @media (min-width: 768px) {
     .masonry-grid { grid-template-columns: repeat(2, 1fr); } /* Tablet */
   }
   @media (min-width: 1200px) {
     .masonry-grid { grid-template-columns: repeat(3, 1fr); } /* Desktop */
   }
   ```

3. **Test at all breakpoints:** Mobile, tablet portrait, tablet landscape, desktop, wide desktop

**Phase impact:** Phase 1 (Masonry Layout) — Basic responsive CSS, test during implementation

---

### Pitfall 11: Empty Search Results Show No Feedback

**What goes wrong:**

User searches, gets zero results, sees blank page. No message indicating "no results found" vs. "still loading".

**Prevention:**

1. **Show empty state:**
   ```html
   {{if eq (len .Articles) 0}}
     <div class="empty-state">
       <p>No articles found{{if .SearchTerm}} for "{{.SearchTerm}}"{{end}}</p>
       <a href="/articles">Clear filters</a>
     </div>
   {{end}}
   ```

2. **Distinguish loading vs. empty:**
   - HTMX swap with indicator: `hx-indicator="#spinner"`
   - After swap completes, show empty state or results

**Phase impact:** Phase 3 (Search) — UI polish, easy to add

---

### Pitfall 12: Search Highlights Don't Work with HTMX Swap

**What goes wrong:**

Search term is highlighted in results, user changes search, new results load but old highlights persist. Or highlights don't appear on new results.

**Prevention:**

1. **Server-side highlighting in template:**
   ```go
   func highlightSearchTerm(text, term string) template.HTML {
     if term == "" {
       return template.HTML(html.EscapeString(text))
     }
     escaped := html.EscapeString(text)
     highlighted := strings.ReplaceAll(escaped, term, "<mark>"+term+"</mark>")
     return template.HTML(highlighted)
   }
   ```

2. **Use CSS for highlight styling:**
   ```css
   mark {
     background: var(--highlight-bg);
     color: var(--highlight-text);
   }
   ```

3. **Clear highlights on swap:** No JavaScript state to clear if done server-side

**Phase impact:** Phase 3 (Search) — Nice-to-have feature, low risk

---

## Phase-Specific Warnings

Recommendations for which phases need deeper research or careful attention.

| Phase Topic | Likely Pitfall | Mitigation Strategy |
|-------------|----------------|---------------------|
| Masonry Layout | Layout breaks after HTMX swaps | Add HTMX event listeners to trigger layout recalculation; reserve space for thumbnails |
| Thumbnails | N+1 HTTP requests at render time | Extract thumbnails during sync, store in database; use timeouts and worker pools |
| Search (LIKE) | Performance degradation >5K articles | Start with LIKE + debounce; plan FTS5 migration; add query timeouts |
| Combined Filters | URL and UI state desync | Treat URL as SSOT; validate params; test bookmark/back button scenarios |
| Date Filtering | Timezone and NULL date handling | Define semantics clearly; handle NULL; store UTC; test edge cases |
| Entire Card Clickable | Nested links break (article link inside card link) | Use JavaScript click handler or CSS pointer-events; test accessibility |

---

## Testing Checklist

Before merging each phase, verify:

### Masonry Layout
- [ ] Cards don't overlap after filter changes
- [ ] Layout recalculates after HTMX swaps
- [ ] Responsive at mobile/tablet/desktop breakpoints
- [ ] Thumbnails loading don't break layout (aspect-ratio set)
- [ ] Theme toggle (dark/light) doesn't break layout

### Thumbnails
- [ ] Thumbnail extraction happens during sync, not render
- [ ] Fallback chain works: RSS → OG → favicon
- [ ] No N+1 HTTP requests in page render
- [ ] Timeouts prevent hanging on slow/dead URLs
- [ ] Placeholder shown while thumbnail loads (no CLS)
- [ ] Works with diverse RSS feeds (test 5+ different blogs)

### Search
- [ ] Debounce works (300ms delay after typing stops)
- [ ] Special characters (`O'Reilly`, `100%`) work correctly
- [ ] Empty search shows all articles
- [ ] Zero results show empty state message
- [ ] Query completes in <500ms with current article count
- [ ] Result limit (100) prevents massive queries

### Combined Filters
- [ ] All filter combinations work together
- [ ] URL params sync with UI state
- [ ] Bookmark URL works (open in incognito)
- [ ] Back/forward buttons work correctly
- [ ] Shared URL shows same results for other users
- [ ] Filter form pre-populates from URL params

### Date Filtering
- [ ] "Last week" shows correct articles
- [ ] Custom date range includes/excludes correctly
- [ ] NULL published dates handled gracefully
- [ ] Timezone displayed/documented
- [ ] Edge cases tested (midnight, month boundaries)

---

## Research Confidence

| Area | Confidence | Notes |
|------|------------|-------|
| Masonry Layout | HIGH | Official MDN docs, real GitHub issues, browser support status verified |
| HTMX Integration | HIGH | Official HTMX docs, community patterns for state management |
| Thumbnail Extraction | MEDIUM | RSS spec and libraries verified, real-world edge cases documented but not exhaustively tested in this project |
| SQLite Performance | HIGH | Official SQLite docs, FTS5 benchmarks, Go driver best practices |
| CLS/Web Vitals | HIGH | Official web.dev documentation, 2026 standards |
| Go RSS Parsers | MEDIUM | Library comparisons based on GitHub popularity, gofeed recommended but not exhaustively compared |

---

## Sources

This research draws from official documentation, real-world implementation issues, and 2026 best practices:

**Masonry Layout:**
- [Native CSS Masonry Layout In CSS Grid — Smashing Magazine](https://www.smashingmagazine.com/native-css-masonry-layout-css-grid/)
- [Masonry layout - CSS | MDN](https://developer.mozilla.org/en-US/docs/Web/CSS/Guides/Grid_layout/Masonry_layout)
- [An alternative proposal for CSS masonry | Chrome for Developers](https://developer.chrome.com/blog/masonry)
- [Resizable grid items with dynamic content? · Issue #1000 · desandro/masonry](https://github.com/desandro/masonry/issues/1000)

**HTMX & Dynamic Content:**
- [HTMX Examples - Animations](https://htmx.org/examples/animations/)
- [Bookmarkable by Design: URL-Driven State in HTMX](https://www.lorenstew.art/blog/bookmarkable-by-design-url-state-htmx/)
- [HTMX Simplifies State Management by Using URL Parameters](https://thedigipress.com/articles/1119)
- [HTMX hx-params attribute](https://htmx.org/attributes/hx-params/)

**Thumbnail Extraction:**
- [How to Set Up Feed to Post's Image Options - WP RSS Aggregator](https://kb.wprssaggregator.com/article/308-how-to-set-up-feed-to-posts-image-options)
- [Getting those thumbnails from Medium RSS Feed](https://medium.com/@kartikyathakur/getting-those-thumbnails-from-medium-rss-feed-183f74aefa8c)
- [RSS enclosure Element](https://www.w3schools.com/xml/rss_tag_enclosure.asp)
- [Media RSS Specification](https://www.rssboard.org/media-rss)
- [How to Favicon in 2026: Three files that fit most needs](https://evilmartians.com/chronicles/how-to-favicon-in-2021-six-files-that-fit-most-needs)

**SQLite Performance:**
- [Tuning SQL LIKE using indexes](https://use-the-index-luke.com/sql/where-clause/searching-for-ranges/like-performance-tuning)
- [SQLite FTS5 Extension](https://sqlite.org/fts5.html)
- [Making SQLite faster in Go by Sandro Turriate](https://turriate.com/articles/making-sqlite-faster-in-go)
- [Full-Text Search in SQLite: A Practical Guide](https://medium.com/@johnidouglasmarangon/full-text-search-in-sqlite-a-practical-guide-80a69c3f42a4)

**Web Performance:**
- [Cumulative Layout Shift (CLS): The Most Misunderstood Core Web Vital (2026 Guide)](https://medium.com/@sahoo.arpan7/cumulative-layout-shift-cls-guide-to-one-of-the-most-misunderstood-core-web-vitals-5f135c68cb6f)
- [Optimize Cumulative Layout Shift](https://web.dev/articles/optimize-cls)

**Go Libraries:**
- [gofeed - Parse RSS, Atom and JSON feeds in Go](https://github.com/mmcdole/gofeed)
- [Imagor - Fast, secure image processing server](https://github.com/cshum/imagor)

---

**Next steps:** Use these pitfalls to inform roadmap phase structure and identify which phases need deeper technical research before implementation.
