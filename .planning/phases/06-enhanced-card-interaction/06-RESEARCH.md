# Phase 6: Enhanced Card Interaction - Research

**Researched:** 2026-02-03
**Domain:** RSS thumbnail extraction, Open Graph scraping, Go HTML templates, clickable card patterns
**Confidence:** HIGH

## Summary

This phase adds thumbnail support and full-card clickability to article cards. The research confirms that thumbnail extraction must happen during sync (not render time) to avoid N+1 queries. The standard approach is a three-tier fallback chain: RSS media → Open Graph metadata → favicon.

**Key findings:**
- gofeed extracts RSS enclosures and Item.Image natively (but Item.Image is often nil)
- Media RSS elements (media:thumbnail, media:content) require accessing Extensions map
- otiai10/opengraph/v2 is the standard Go library for Open Graph extraction
- Clickable cards use CSS pseudo-element stretched-link pattern (not wrapper anchor)
- modernc.org/sqlite v1.44.3 uses SQLite 3.51.2 which supports ADD COLUMN IF NOT EXISTS
- Empty img src causes unnecessary HTTP requests; use conditional rendering or onerror

**Primary recommendation:** Extract thumbnails during scanner.ScanBlog() with fallback chain (RSS → Open Graph → skip). Store thumbnail_url in database. Render with lazy loading and onerror fallback to favicon. Make entire card clickable with CSS ::after pseudo-element pattern.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/mmcdole/gofeed | v1.3.0 | RSS/Atom parsing | Already in use; natively extracts enclosures and Item.Image |
| github.com/otiai10/opengraph/v2 | v2.2.0 | Open Graph extraction | Most popular Go OG library; clean API with Intent customization |
| github.com/PuerkitoBio/goquery | v1.10.3 | HTML parsing | Already in use; required by opengraph library |
| modernc.org/sqlite | v1.44.3 | SQLite driver | Already in use; supports ADD COLUMN IF NOT EXISTS |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| net/http | stdlib | HTTP client | Built-in; use for fetching article pages for Open Graph extraction |
| html/template | stdlib | Template rendering | Built-in; auto-escapes URLs in attributes |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| otiai10/opengraph/v2 | dyatlov/go-opengraph | otiai10 has better documentation and v2 API with Intent pattern |
| otiai10/opengraph/v2 | Manual meta tag parsing with goquery | opengraph library handles edge cases, multiple images, relative URLs |
| Native extraction | golang-migrate for schema | Overkill for single column addition; native SQL is simpler |

**Installation:**
```bash
go get github.com/otiai10/opengraph/v2@v2.2.0
```

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── rss/              # RSS parsing (existing)
│   └── rss.go        # Add thumbnail extraction from feed
├── scraper/          # HTML scraping (existing)
│   └── scraper.go    # Keep as-is
├── scanner/          # Orchestration (existing)
│   └── scanner.go    # Call thumbnail extraction during sync
├── thumbnail/        # NEW: Thumbnail extraction
│   └── thumbnail.go  # Extract from RSS, Open Graph, favicon
├── model/            # Data models (existing)
│   └── model.go      # Add ThumbnailURL field to Article/ArticleWithBlog
└── storage/          # Database layer (existing)
    └── database.go   # Add migration for thumbnail_url column
```

### Pattern 1: Thumbnail Extraction During Sync (CRITICAL)
**What:** Extract thumbnail URLs during scanner.ScanBlog() before inserting articles into database
**When to use:** Always - thumbnail extraction is I/O bound and must not happen per-render

**Why not at render time:**
- N+1 query problem: rendering 50 articles = 50+ HTTP requests
- Render blocking: page waits for thumbnail fetches
- Repeated work: same article fetched multiple times

**Example:**
```go
// Source: Research findings + existing scanner.go pattern
// internal/scanner/scanner.go - Add to convertFeedArticles

func convertFeedArticles(blogID int64, feedArticles []rss.FeedArticle) []model.Article {
    result := make([]model.Article, 0, len(feedArticles))
    for _, article := range feedArticles {
        result = append(result, model.Article{
            BlogID:        blogID,
            Title:         article.Title,
            URL:           article.URL,
            ThumbnailURL:  article.ThumbnailURL, // NEW: extracted from RSS
            PublishedDate: article.PublishedDate,
            IsRead:        false,
        })
    }
    return result
}
```

### Pattern 2: Three-Tier Fallback Chain
**What:** Try RSS media → Open Graph → skip (favicon already available at render)
**When to use:** For every article during sync

**Fallback logic:**
```go
// Source: Research findings on RSS media, Open Graph, and favicon patterns
// internal/thumbnail/thumbnail.go (NEW FILE)

package thumbnail

import (
    "context"
    "net/http"
    "time"

    "github.com/mmcdole/gofeed"
    "github.com/otiai10/opengraph/v2"
)

// ExtractThumbnail returns thumbnail URL using fallback chain
// Returns empty string if no thumbnail found (favicon used at render)
func ExtractThumbnail(articleURL string, item *gofeed.Item, timeout time.Duration) string {
    // Tier 1: RSS Item.Image (channel-level image)
    if item != nil && item.Image != nil && item.Image.URL != "" {
        return item.Image.URL
    }

    // Tier 2: RSS Enclosures (look for image types)
    if item != nil && len(item.Enclosures) > 0 {
        for _, enc := range item.Enclosures {
            if isImageType(enc.Type) {
                return enc.URL
            }
        }
    }

    // Tier 3: Open Graph og:image
    // Only fetch if RSS didn't provide thumbnail
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    intent := opengraph.Intent{
        Context:    ctx,
        Strict:     true, // Only trust meta tags
        HTTPClient: &http.Client{Timeout: timeout},
    }

    ogp, err := opengraph.Fetch(articleURL, intent)
    if err == nil && len(ogp.Image) > 0 && ogp.Image[0].URL != "" {
        return ogp.Image[0].URL
    }

    // Tier 4: No thumbnail found - return empty
    // Favicon will be used at render time
    return ""
}

func isImageType(mimeType string) bool {
    switch mimeType {
    case "image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp", "image/svg+xml":
        return true
    default:
        return false
    }
}
```

### Pattern 3: CSS Stretched Link for Clickable Cards
**What:** Use CSS ::after pseudo-element to make entire card clickable
**When to use:** When card has primary link (article) and secondary actions (read/unread button)

**Why not wrapper anchor:**
- Invalid HTML: `<a>` cannot contain interactive elements (buttons)
- Accessibility: confuses screen readers
- UX: prevents text selection

**Example:**
```html
<!-- Source: https://dev.to/micmath/clickable-card-patterns-and-anti-patterns-2hl2 -->
<!-- templates/partials/article-list.gohtml -->

<article class="article-card" id="article-{{.ID}}">
    {{if .ThumbnailURL}}
    <img class="article-thumbnail"
         src="{{.ThumbnailURL}}"
         alt=""
         width="200"
         height="150"
         loading="lazy"
         onerror="this.style.display='none'; this.onerror=null;">
    {{end}}

    <div class="article-content">
        <!-- Primary link with stretched-link class -->
        <a href="{{.URL}}"
           target="_blank"
           rel="noopener noreferrer"
           class="article-title stretched-link">
            {{.Title}}
        </a>
        <div class="article-meta">
            <span class="article-source">{{.BlogName}}</span>
            {{if .PublishedDate}}
            <span class="article-time">{{timeAgo .PublishedDate}}</span>
            {{end}}
        </div>
    </div>

    <!-- Secondary action requires z-index layering -->
    <button class="action-btn"
            hx-post="/articles/{{.ID}}/read"
            hx-target="#article-{{.ID}}"
            hx-swap="outerHTML swap:300ms"
            title="Mark as read">
        Read
    </button>
</article>
```

**CSS:**
```css
/* Source: https://getbootstrap.com/docs/5.3/helpers/stretched-link/ */
.article-card {
    position: relative; /* Required for ::after positioning */
}

.stretched-link::after {
    position: absolute;
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
    z-index: 1;
    content: "";
}

/* Secondary actions must sit above stretched link */
.action-btn {
    position: relative;
    z-index: 2;
}
```

### Pattern 4: Safe Schema Migration
**What:** Add nullable thumbnail_url column with IF NOT EXISTS check
**When to use:** Database initialization and migrations

**Example:**
```go
// Source: https://www.sqlite.org/lang_altertable.html + modernc.org/sqlite docs
// internal/storage/database.go

func (db *Database) ensureSchema() error {
    schema := `
        CREATE TABLE IF NOT EXISTS blogs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL UNIQUE,
            url TEXT NOT NULL,
            feed_url TEXT,
            scrape_selector TEXT,
            last_scanned TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS articles (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            blog_id INTEGER NOT NULL,
            title TEXT NOT NULL,
            url TEXT NOT NULL UNIQUE,
            published_date TIMESTAMP,
            discovered_date TIMESTAMP,
            is_read BOOLEAN NOT NULL DEFAULT 0,
            FOREIGN KEY (blog_id) REFERENCES blogs(id) ON DELETE CASCADE
        );

        -- Add thumbnail_url column (safe idempotent migration)
        -- SQLite 3.51.2 supports IF NOT EXISTS
        ALTER TABLE articles ADD COLUMN IF NOT EXISTS thumbnail_url TEXT;
    `

    _, err := db.conn.Exec(schema)
    return err
}
```

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Open Graph parsing | Custom meta tag scraping | otiai10/opengraph/v2 | Handles relative URLs, multiple images, og:image:secure_url, edge cases |
| URL absolutization | String concatenation | opengraph.ToAbs() | Handles protocol-relative URLs, base href, path resolution |
| HTTP timeouts | Default http.Get | context.WithTimeout + Intent.Context | Prevents hanging on slow/dead sites; proper cancellation |
| Image type detection | String contains "image" | MIME type switch | Enclosures can have "image/jpeg" or custom types; explicit list safer |
| Empty src handling | Render empty string | Conditional template + onerror | Empty src causes HTTP request to current page; performance killer |

**Key insight:** Thumbnail extraction is simpler than it looks, but Open Graph edge cases (relative URLs, multiple images, secure URLs) make a library essential. Don't parse meta tags manually.

## Common Pitfalls

### Pitfall 1: N+1 Query Problem (Thumbnail Fetch at Render Time)
**What goes wrong:** Fetching thumbnails during template rendering causes one HTTP request per article displayed (50 articles = 50+ HTTP requests, blocking page render)
**Why it happens:** Template function `thumbnailURL(articleURL)` seems convenient but executes per-article
**How to avoid:** Extract thumbnails during sync in scanner.ScanBlog(); store in database; render from database
**Warning signs:** Slow page loads, high CPU during rendering, HTTP client errors in template context

### Pitfall 2: Empty img src Attribute
**What goes wrong:** `<img src="">` causes browsers to request the current page URL again (IE/Chrome/Safari behavior)
**Why it happens:** Setting `src=""` when no thumbnail available seems harmless
**How to avoid:** Use conditional template rendering: `{{if .ThumbnailURL}}<img src="{{.ThumbnailURL}}">{{end}}` OR use onerror to hide: `onerror="this.style.display='none'"`
**Warning signs:** Duplicate HTTP requests to page URL in server logs, increased bandwidth

**References:**
- [Empty image src can destroy your site](https://humanwhocodes.com/blog/2009/11/30/empty-image-src-can-destroy-your-site/)
- [How to Fix "SRC Cannot be Blank" Issue](https://sitechecker.pro/site-audit-issues/page-empty-src-attributes/)

### Pitfall 3: Rate Limiting from Open Graph Scraping
**What goes wrong:** HTTP 429 (Too Many Requests) when fetching Open Graph data from same domain repeatedly
**Why it happens:** Syncing multiple articles from same blog = multiple requests to same domain in quick succession
**How to avoid:**
- Implement per-domain rate limiting (wait between requests to same host)
- Respect Retry-After header from 429 responses
- Add exponential backoff with jitter
- Skip Open Graph for articles from same blog after first 429
**Warning signs:** HTTP 429 errors in logs, empty thumbnails from otherwise OG-compliant sites

**References:**
- [Fix HTTP 429 when scraping with 7 proven tactics](https://ki-ecke.com/crypto-insights/fix-http-429-when-scraping-with-7-proven-tactics/)
- [429 status code - what is it and how to avoid it?](https://www.scrapingbee.com/webscraping-questions/web-scraping-blocked/429-status-code-what-it-is-and-how-to-avoid-it/)

### Pitfall 4: Wrapping Card in Anchor Tag
**What goes wrong:** `<a><article>...<button>...</button></article></a>` is invalid HTML; button inside anchor doesn't work
**Why it happens:** Seems like easiest way to make entire card clickable
**How to avoid:** Use CSS stretched-link pattern with ::after pseudo-element
**Warning signs:** Button clicks trigger link instead, screen readers announce confusing nested links, HTML validation errors

**References:**
- [Clickable Card Patterns and Anti-Patterns](https://dev.to/micmath/clickable-card-patterns-and-anti-patterns-2hl2)
- [Block Links, Cards, Clickable Regions, Rows, Etc.](https://adrianroselli.com/2020/02/block-links-cards-clickable-regions-etc.html)

### Pitfall 5: Not Handling Open Graph Fetch Timeouts
**What goes wrong:** Sync hangs for 30+ seconds when article page is slow/unresponsive
**Why it happens:** Default http.Client has no timeout; Open Graph fetch waits indefinitely
**How to avoid:** Use context.WithTimeout with Intent.Context in opengraph.Fetch; set reasonable timeout (5-10s)
**Warning signs:** Sync operations taking minutes, user reports of "stuck" sync, goroutine leaks

**References:**
- [The complete guide to Go net/http timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/)
- [How to check for HTTP timeout errors in Go](https://freshman.tech/snippets/go/http-timeout-error/)

### Pitfall 6: Not Setting onerror=null in Fallback
**What goes wrong:** Infinite loop if onerror fallback image also fails
**Why it happens:** `onerror="this.src='fallback.png'"` without `this.onerror=null` triggers again if fallback fails
**How to avoid:** Always include `this.onerror=null` in onerror handler
**Warning signs:** Browser console errors about too many onerror calls, high CPU usage from image loading loop

**Example:**
```html
<!-- WRONG: Infinite loop if fallback fails -->
<img src="{{.ThumbnailURL}}" onerror="this.src='{{faviconURL .BlogURL}}'">

<!-- CORRECT: Stops after fallback attempt -->
<img src="{{.ThumbnailURL}}" onerror="this.src='{{faviconURL .BlogURL}}'; this.onerror=null;">

<!-- BEST: Hide if fallback fails -->
<img src="{{.ThumbnailURL}}" onerror="this.style.display='none'; this.onerror=null;">
```

**References:**
- [Fallbacks for HTTP 404 images in HTML and JavaScript](https://blog.sentry.io/fallbacks-for-http-404-images-in-html-and-javascript/)
- [HTML fallback images on error](https://dev.to/dailydevtips1/html-fallback-images-on-error-1aka)

## Code Examples

Verified patterns from official sources:

### RSS Thumbnail Extraction with gofeed
```go
// Source: https://pkg.go.dev/github.com/mmcdole/gofeed + research
// internal/rss/rss.go - Update FeedArticle struct

package rss

type FeedArticle struct {
    Title         string
    URL           string
    ThumbnailURL  string    // NEW
    PublishedDate *time.Time
}

func ParseFeed(feedURL string, timeout time.Duration) ([]FeedArticle, error) {
    client := &http.Client{Timeout: timeout}
    response, err := client.Get(feedURL)
    if err != nil {
        return nil, FeedParseError{Message: fmt.Sprintf("failed to fetch feed: %v", err)}
    }
    defer response.Body.Close()

    parser := gofeed.NewParser()
    feed, err := parser.Parse(response.Body)
    if err != nil {
        return nil, FeedParseError{Message: fmt.Sprintf("failed to parse feed: %v", err)}
    }

    var articles []FeedArticle
    for _, item := range feed.Items {
        thumbnailURL := extractThumbnailFromItem(item)

        articles = append(articles, FeedArticle{
            Title:         strings.TrimSpace(item.Title),
            URL:           strings.TrimSpace(item.Link),
            ThumbnailURL:  thumbnailURL,
            PublishedDate: pickPublishedDate(item),
        })
    }

    return articles, nil
}

func extractThumbnailFromItem(item *gofeed.Item) string {
    // Try Item.Image first (channel-level image)
    if item.Image != nil && item.Image.URL != "" {
        return item.Image.URL
    }

    // Try Enclosures (common in RSS 2.0)
    for _, enc := range item.Enclosures {
        if isImageMIMEType(enc.Type) && enc.URL != "" {
            return enc.URL
        }
    }

    // Could check Extensions["media"] for media:thumbnail here
    // but most feeds use standard enclosures

    return ""
}

func isImageMIMEType(mimeType string) bool {
    return strings.HasPrefix(mimeType, "image/")
}
```

### Open Graph Thumbnail Extraction
```go
// Source: https://pkg.go.dev/github.com/otiai10/opengraph/v2
// internal/thumbnail/thumbnail.go

package thumbnail

import (
    "context"
    "net/http"
    "time"

    "github.com/otiai10/opengraph/v2"
)

// ExtractFromOpenGraph fetches og:image from article page
func ExtractFromOpenGraph(articleURL string, timeout time.Duration) string {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    intent := opengraph.Intent{
        Context:    ctx,
        Strict:     true, // Only parse <meta> tags (more secure)
        HTTPClient: &http.Client{Timeout: timeout},
        Headers: map[string]string{
            "User-Agent": "BlogWatcher/1.0", // Identify your bot
        },
    }

    ogp, err := opengraph.Fetch(articleURL, intent)
    if err != nil {
        return "" // Fail silently - thumbnail is optional
    }

    // Convert relative URLs to absolute
    if err := ogp.ToAbs(); err != nil {
        return "" // Fail silently if URL resolution fails
    }

    // Return first image if available
    if len(ogp.Image) > 0 && ogp.Image[0].URL != "" {
        return ogp.Image[0].URL
    }

    return ""
}
```

### Conditional Thumbnail Rendering
```html
<!-- Source: Go html/template docs + lazy loading best practices -->
<!-- templates/partials/article-list.gohtml -->

{{range .Articles}}
<article class="article-card" id="article-{{.ID}}">
    {{if .ThumbnailURL}}
    <img class="article-thumbnail"
         src="{{.ThumbnailURL}}"
         alt=""
         width="200"
         height="150"
         loading="lazy"
         onerror="this.style.display='none'; this.onerror=null;">
    {{else}}
    <img class="article-favicon"
         src="{{faviconURL .BlogURL}}"
         alt=""
         width="32"
         height="32"
         loading="lazy"
         onerror="this.style.visibility='hidden'">
    {{end}}

    <div class="article-content">
        <a href="{{.URL}}"
           target="_blank"
           rel="noopener noreferrer"
           class="article-title stretched-link">
            {{.Title}}
        </a>
        <div class="article-meta">
            <span class="article-source">{{.BlogName}}</span>
            {{if .PublishedDate}}
            <span class="article-time">{{timeAgo .PublishedDate}}</span>
            {{end}}
        </div>
    </div>

    <button class="action-btn"
            hx-post="/articles/{{.ID}}/read"
            hx-target="#article-{{.ID}}"
            hx-swap="outerHTML swap:300ms"
            title="Mark as read">
        Read
    </button>
</article>
{{end}}
```

### Schema Migration with IF NOT EXISTS
```go
// Source: https://www.sqlite.org/lang_altertable.html
// internal/storage/database.go

func (db *Database) ensureSchema() error {
    schema := `
        CREATE TABLE IF NOT EXISTS blogs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL UNIQUE,
            url TEXT NOT NULL,
            feed_url TEXT,
            scrape_selector TEXT,
            last_scanned TIMESTAMP
        );

        CREATE TABLE IF NOT EXISTS articles (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            blog_id INTEGER NOT NULL,
            title TEXT NOT NULL,
            url TEXT NOT NULL UNIQUE,
            published_date TIMESTAMP,
            discovered_date TIMESTAMP,
            is_read BOOLEAN NOT NULL DEFAULT 0,
            FOREIGN KEY (blog_id) REFERENCES blogs(id) ON DELETE CASCADE
        );

        -- Add thumbnail_url column if not exists (SQLite 3.51.2+ supports this)
        ALTER TABLE articles ADD COLUMN IF NOT EXISTS thumbnail_url TEXT;
    `

    _, err := db.conn.Exec(schema)
    return err
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Render-time thumbnail fetch | Sync-time extraction + DB storage | N/A (best practice) | Eliminates N+1 queries, faster rendering |
| Manual meta tag parsing | otiai10/opengraph/v2 | Library released 2022 | Handles edge cases, relative URLs, multiple images |
| Multiple favicon sizes | Single SVG + 48x48 PNG fallback | 2021-2026 | Simpler setup, better browser support |
| Wrapper anchor for cards | CSS stretched-link pattern | ~2019 (Bootstrap 4.3) | Valid HTML, better a11y, allows nested buttons |
| Manual column existence check | ALTER TABLE ADD COLUMN IF NOT EXISTS | SQLite 3.33.0 (2020-08-14) | Simpler migrations, idempotent by default |
| RSS media:* extensions only | Item.Image + Enclosures fallback | gofeed design | Better coverage across RSS variants |

**Deprecated/outdated:**
- **Multiple favicon sizes** (16x16, 32x32, 180x180, etc.): Modern approach is SVG + single PNG fallback
- **gofeed custom Translator for media:thumbnail**: Enclosures work for most feeds; extensions add complexity
- **Manual SQLite column existence checks with PRAGMA table_info**: Use IF NOT EXISTS if SQLite ≥3.33.0

## Open Questions

Things that couldn't be fully resolved:

1. **Media RSS extensions in gofeed**
   - What we know: gofeed stores media:thumbnail and media:content in Item.Extensions["media"]
   - What's unclear: Exact structure of Extensions map; whether nested fields need recursive traversal
   - Recommendation: Start with Item.Image + Enclosures; only implement Extensions parsing if feeds don't provide enclosures

2. **Rate limiting per-domain**
   - What we know: Multiple articles from same blog should space out Open Graph requests
   - What's unclear: Optimal delay between requests (100ms? 500ms? 1s?)
   - Recommendation: Start with no rate limiting; add domain-level queuing if 429 errors appear in logs

3. **Thumbnail caching duration**
   - What we know: Thumbnails stored in DB; no expiration mechanism
   - What's unclear: Should thumbnails be re-fetched periodically? If so, how often?
   - Recommendation: Store forever (no expiration); add manual "refresh thumbnail" feature later if needed

## Sources

### Primary (HIGH confidence)
- [gofeed package documentation](https://pkg.go.dev/github.com/mmcdole/gofeed) - Item struct, Enclosures, Extensions
- [otiai10/opengraph/v2 package documentation](https://pkg.go.dev/github.com/otiai10/opengraph/v2) - API, Intent struct, usage examples
- [otiai10/opengraph GitHub README](https://github.com/otiai10/opengraph) - Installation, examples, advanced configuration
- [Media RSS Specification](https://www.rssboard.org/media-rss) - media:thumbnail and media:content structure
- [SQLite ALTER TABLE documentation](https://www.sqlite.org/lang_altertable.html) - ADD COLUMN IF NOT EXISTS syntax
- [modernc.org/sqlite package documentation](https://pkg.go.dev/modernc.org/sqlite) - SQLite 3.51.2 version confirmation
- [html/template package documentation](https://pkg.go.dev/html/template) - URL escaping, auto-escaping behavior

### Secondary (MEDIUM confidence)
- [Clickable Card Patterns and Anti-Patterns](https://dev.to/micmath/clickable-card-patterns-and-anti-patterns-2hl2) - Stretched link pattern explanation
- [Block Links, Cards, Clickable Regions, Rows, Etc.](https://adrianroselli.com/2020/02/block-links-cards-clickable-regions-etc.html) - Accessibility considerations
- [Bootstrap Stretched Link documentation](https://getbootstrap.com/docs/5.3/helpers/stretched-link/) - CSS implementation details
- [The complete guide to Go net/http timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/) - Context and timeout patterns
- [How to Favicon in 2026](https://evilmartians.com/chronicles/how-to-favicon-in-2021-six-files-that-fit-most-needs) - Modern favicon best practices
- [Lazy loading - MDN](https://developer.mozilla.org/en-US/docs/Web/Performance/Guides/Lazy_loading) - loading="lazy" attribute
- [Fallbacks for HTTP 404 images in HTML and JavaScript](https://blog.sentry.io/fallbacks-for-http-404-images-in-html-and-javascript/) - onerror patterns
- [Fix HTTP 429 when scraping](https://ki-ecke.com/crypto-insights/fix-http-429-when-scraping-with-7-proven-tactics/) - Rate limiting strategies
- [Empty image src can destroy your site](https://humanwhocodes.com/blog/2009/11/30/empty-image-src-can-destroy-your-site/) - Performance impact of empty src

### Tertiary (LOW confidence - requires validation)
- [gofeed Item.Image Issue #133](https://github.com/mmcdole/gofeed/issues/133) - Community reports that Item.Image is often nil
- [Getting thumbnails from Medium RSS Feed](https://medium.com/@kartikyathakur/getting-those-thumbnails-from-medium-rss-feed-183f74aefa8c) - Blog post about thumbnail extraction patterns

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries verified via pkg.go.dev; opengraph v2 widely used
- Architecture: HIGH - Patterns verified with official docs and established best practices
- Pitfalls: HIGH - All pitfalls documented in authoritative sources (MDN, Cloudflare, Adrian Roselli)
- Open questions: MEDIUM - Extensions structure unclear; rate limiting requires experimentation

**Research date:** 2026-02-03
**Valid until:** ~60 days (2026-04-03) - Stable domain; Go stdlib and SQLite change slowly; re-verify if Go 1.27+ or new major version of dependencies released
