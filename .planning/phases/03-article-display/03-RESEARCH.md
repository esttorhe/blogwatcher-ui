# Phase 3: Article Display - Research

**Researched:** 2026-02-02
**Domain:** Article card UI, relative time formatting, favicon services, Go template functions
**Confidence:** HIGH

## Summary

Phase 3 transforms the basic article list from Phase 2 into rich article cards displaying thumbnails/favicons, blog names, and relative timestamps ("2 hours ago"). This builds directly on the existing infrastructure: Go 1.24 stdlib, HTMX 2.0.8, dark theme CSS variables, and the database layer with Blog and Article models.

The recommended approach involves:
1. Adding a custom Go template function (`timeAgo`) for relative time formatting
2. Using Google's S2 favicon service (`google.com/s2/favicons`) for blog site icons
3. Modifying the database query to JOIN blogs with articles (getting blog name)
4. Creating CSS card components using Flexbox within existing dark theme
5. Ensuring proper security with `rel="noopener noreferrer"` on external links

The stack remains minimal: no external dependencies needed. Relative time formatting can be implemented in ~30 lines of Go using the standard library. Favicon fetching is offloaded to Google's service, avoiding caching complexity.

**Primary recommendation:** Implement a custom `timeAgo` template function using Go's standard library time.Duration, use Google's S2 favicon API with a fallback, and create a JOIN query returning ArticleWithBlog structs that include the blog name. Style article cards using Flexbox with the existing CSS variables.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| html/template FuncMap | stdlib | Custom time formatting function | Built-in, no dependencies needed |
| time.Duration | stdlib | Calculate relative time differences | Standard library time operations |
| Google S2 Favicon API | External API | Fetch site favicons | Reliable, no caching needed, handles edge cases |
| CSS Flexbox | Native | Article card internal layout | Already used in Phase 2, universal support |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| DuckDuckGo Favicon API | External API | Alternative favicon source | If Google S2 is blocked/unreliable |
| dustin/go-humanize | 3.x | Pre-built relative time | Only if custom implementation proves insufficient |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Custom timeAgo function | go-humanize or go-prettytime | External dependency vs 30 lines of stdlib code |
| Google S2 favicons | Self-hosted favicon scraping | Complexity of caching, handling edge cases |
| JOIN query | Separate blog lookup | Extra database round-trip, more code |

**Installation:**
```bash
# No new dependencies required
# Favicon images loaded via <img> tag from Google's API
# Template functions added via FuncMap before parsing
```

## Architecture Patterns

### Recommended Project Structure
```
internal/
  server/
    server.go        # UPDATE: Add FuncMap before template parsing
    handlers.go      # UPDATE: Pass blog name to templates
  storage/
    database.go      # UPDATE: Add ListArticlesWithBlog() with JOIN
  model/
    model.go         # UPDATE: Add ArticleWithBlog struct
templates/
  partials/
    article-list.gohtml    # UPDATE: Rich card layout with timeAgo
static/
  styles.css               # UPDATE: Article card CSS
```

### Pattern 1: Template FuncMap for Time Formatting
**What:** Register custom functions before template parsing
**When to use:** Any custom template function (timeAgo, formatDate, etc.)
**Example:**
```go
// Source: https://pkg.go.dev/text/template + https://www.calhoun.io/intro-to-templates-p3-functions/
// CRITICAL: Funcs() must be called BEFORE ParseGlob()
func NewServer(db *storage.Database) (http.Handler, error) {
    funcMap := template.FuncMap{
        "timeAgo": timeAgo,
    }

    // Create template with FuncMap FIRST, then parse
    tmpl := template.New("").Funcs(funcMap)
    tmpl, err := tmpl.ParseGlob("templates/*.gohtml")
    if err != nil {
        return nil, fmt.Errorf("failed to parse templates: %w", err)
    }
    // ... continue parsing other templates
}
```

### Pattern 2: Custom timeAgo Function (No Dependencies)
**What:** Convert time.Time to human-readable relative string
**When to use:** Display article publish/discover dates
**Example:**
```go
// Source: Standard library patterns, similar to go-humanize implementation
func timeAgo(t *time.Time) string {
    if t == nil {
        return ""
    }

    now := time.Now()
    diff := now.Sub(*t)

    switch {
    case diff < time.Minute:
        return "just now"
    case diff < time.Hour:
        mins := int(diff.Minutes())
        if mins == 1 {
            return "1 minute ago"
        }
        return fmt.Sprintf("%d minutes ago", mins)
    case diff < 24*time.Hour:
        hours := int(diff.Hours())
        if hours == 1 {
            return "1 hour ago"
        }
        return fmt.Sprintf("%d hours ago", hours)
    case diff < 7*24*time.Hour:
        days := int(diff.Hours() / 24)
        if days == 1 {
            return "yesterday"
        }
        return fmt.Sprintf("%d days ago", days)
    case diff < 30*24*time.Hour:
        weeks := int(diff.Hours() / 24 / 7)
        if weeks == 1 {
            return "1 week ago"
        }
        return fmt.Sprintf("%d weeks ago", weeks)
    case diff < 365*24*time.Hour:
        months := int(diff.Hours() / 24 / 30)
        if months == 1 {
            return "1 month ago"
        }
        return fmt.Sprintf("%d months ago", months)
    default:
        years := int(diff.Hours() / 24 / 365)
        if years == 1 {
            return "1 year ago"
        }
        return fmt.Sprintf("%d years ago", years)
    }
}
```

### Pattern 3: ArticleWithBlog Struct for JOIN Results
**What:** Struct that includes both article fields and blog name
**When to use:** Displaying articles with their source blog name
**Example:**
```go
// Source: Go database patterns, https://go.dev/doc/tutorial/database-access
type ArticleWithBlog struct {
    ID             int64
    BlogID         int64
    Title          string
    URL            string
    PublishedDate  *time.Time
    DiscoveredDate *time.Time
    IsRead         bool
    BlogName       string  // From JOIN with blogs table
    BlogURL        string  // For favicon lookup
}
```

### Pattern 4: SQL JOIN Query for Articles with Blog Info
**What:** Single query returns articles with their blog names
**When to use:** Article list display
**Example:**
```go
// Source: https://go.dev/doc/database/querying + https://www.alexedwards.net/blog/introduction-to-using-sql-databases-in-go
func (db *Database) ListArticlesWithBlog(isRead bool, blogID *int64) ([]model.ArticleWithBlog, error) {
    query := `
        SELECT
            a.id, a.blog_id, a.title, a.url,
            a.published_date, a.discovered_date, a.is_read,
            b.name AS blog_name, b.url AS blog_url
        FROM articles a
        INNER JOIN blogs b ON a.blog_id = b.id
        WHERE a.is_read = ?`
    args := []interface{}{isRead}

    if blogID != nil {
        query += " AND a.blog_id = ?"
        args = append(args, *blogID)
    }
    query += " ORDER BY a.discovered_date DESC"

    rows, err := db.conn.Query(query, args...)
    // ... scanning logic
}
```

### Pattern 5: Favicon URL Construction
**What:** Build Google S2 favicon URL from blog domain
**When to use:** Article card image display
**Example:**
```go
// Source: https://dev.to/derlin/get-favicons-from-any-website-using-a-hidden-google-api-3p1e
// Template function to extract domain and build favicon URL
func faviconURL(blogURL string) string {
    u, err := url.Parse(blogURL)
    if err != nil || u.Host == "" {
        return "/static/default-favicon.svg"  // Fallback
    }
    return fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s&sz=32", u.Host)
}
```

### Anti-Patterns to Avoid
- **Fetching favicons server-side:** Let browser fetch via `<img src>` from Google's API
- **Caching favicons locally:** Google's API handles caching, adds complexity without benefit
- **N+1 queries for blog names:** Use JOIN, not separate lookup per article
- **Hardcoded time zones:** Use UTC internally, let browser localize if needed
- **Missing nil checks in timeAgo:** PublishedDate can be nil, always handle

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Favicon scraping | Custom scraper with caching | Google S2 Favicon API | Handles redirects, 404s, CDNs, updates automatically |
| URL parsing for domain | Regex | net/url.Parse | Handles edge cases (ports, subdomains, protocols) |
| Relative time i18n | Custom translation maps | Accept English-only initially | Add i18n later if needed, over-engineering otherwise |
| Article card hover states | JavaScript | CSS :hover pseudo-class | Pure CSS, already in Phase 2 patterns |
| Link security | Manual attribute handling | Always add `rel="noopener noreferrer"` | Template can include by default |

**Key insight:** Favicon services exist because favicon fetching is deceptively complex (CORS, multiple formats, redirect chains, failures). Offload this to Google's proven infrastructure.

## Common Pitfalls

### Pitfall 1: FuncMap Added After Template Parsing
**What goes wrong:** Custom functions not available in templates, panics or undefined function errors
**Why it happens:** `Funcs()` must be called before `Parse*()` methods
**How to avoid:**
```go
// WRONG - will panic
tmpl := template.Must(template.ParseGlob("*.gohtml"))
tmpl.Funcs(funcMap) // Too late!

// CORRECT
tmpl := template.New("").Funcs(funcMap)
tmpl = template.Must(tmpl.ParseGlob("*.gohtml"))
```
**Warning signs:** "function undefined" errors, panics mentioning FuncMap

### Pitfall 2: Nil Pointer in Time Formatting
**What goes wrong:** Panic when PublishedDate is nil
**Why it happens:** Articles may have nil dates if scraping failed
**How to avoid:** Check for nil in timeAgo function, return empty string or fallback
**Warning signs:** Panic stack traces pointing to time formatting code

### Pitfall 3: Missing rel="noopener noreferrer"
**What goes wrong:** Security vulnerability - opened page can access window.opener
**Why it happens:** target="_blank" without protection enables reverse tabnapping
**How to avoid:** Always include both attributes on external links
```html
<!-- BAD -->
<a href="{{.URL}}" target="_blank">{{.Title}}</a>

<!-- GOOD -->
<a href="{{.URL}}" target="_blank" rel="noopener noreferrer">{{.Title}}</a>
```
**Warning signs:** Security audit flags, browser console warnings in modern browsers

### Pitfall 4: Favicon Service Blocked
**What goes wrong:** Broken images if Google S2 API is blocked (corporate firewalls, China)
**Why it happens:** External dependency on Google service
**How to avoid:** Use CSS fallback for broken images, or provide default favicon
```css
.article-favicon {
  background-color: var(--bg-elevated);  /* Fallback color */
}
.article-favicon[src=""] {
  visibility: hidden;  /* Hide if empty src */
}
```
**Warning signs:** Broken image icons in article cards

### Pitfall 5: Time Zone Confusion
**What goes wrong:** "2 hours ago" shows wrong relative time
**Why it happens:** Comparing UTC database time with local time, or vice versa
**How to avoid:** Store UTC in database, convert to UTC before comparison in timeAgo
```go
// Use UTC for consistent comparisons
now := time.Now().UTC()
diff := now.Sub(t.UTC())
```
**Warning signs:** Times off by hours, different across users

### Pitfall 6: JOIN Returns No Rows for Orphaned Articles
**What goes wrong:** Articles without a matching blog disappear
**Why it happens:** INNER JOIN excludes rows without matches
**How to avoid:** Use LEFT JOIN if orphaned articles are possible, or ensure foreign key integrity
```sql
-- Use LEFT JOIN if articles might have invalid blog_id
SELECT a.*, COALESCE(b.name, 'Unknown') AS blog_name
FROM articles a
LEFT JOIN blogs b ON a.blog_id = b.id
```
**Warning signs:** Article count mismatch between list and individual queries

## Code Examples

Verified patterns from official sources:

### Complete timeAgo Template Function
```go
// Source: Derived from https://pkg.go.dev/github.com/dustin/go-humanize patterns
// Using only standard library
func timeAgo(t *time.Time) string {
    if t == nil {
        return ""
    }

    now := time.Now()
    diff := now.Sub(*t)

    // Handle future times (shouldn't happen, but be safe)
    if diff < 0 {
        return "in the future"
    }

    seconds := int(diff.Seconds())
    minutes := int(diff.Minutes())
    hours := int(diff.Hours())
    days := hours / 24
    weeks := days / 7
    months := days / 30
    years := days / 365

    switch {
    case seconds < 60:
        return "just now"
    case minutes < 60:
        if minutes == 1 {
            return "1 minute ago"
        }
        return fmt.Sprintf("%d minutes ago", minutes)
    case hours < 24:
        if hours == 1 {
            return "1 hour ago"
        }
        return fmt.Sprintf("%d hours ago", hours)
    case days < 7:
        if days == 1 {
            return "yesterday"
        }
        return fmt.Sprintf("%d days ago", days)
    case weeks < 5:
        if weeks == 1 {
            return "1 week ago"
        }
        return fmt.Sprintf("%d weeks ago", weeks)
    case months < 12:
        if months == 1 {
            return "1 month ago"
        }
        return fmt.Sprintf("%d months ago", months)
    default:
        if years == 1 {
            return "1 year ago"
        }
        return fmt.Sprintf("%d years ago", years)
    }
}
```

### FuncMap Integration in NewServer
```go
// Source: https://pkg.go.dev/html/template + project patterns from Phase 1
func NewServer(db *storage.Database) (http.Handler, error) {
    // Define custom template functions
    funcMap := template.FuncMap{
        "timeAgo": timeAgo,
        "faviconURL": func(blogURL string) string {
            u, err := url.Parse(blogURL)
            if err != nil || u.Host == "" {
                return ""
            }
            return fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s&sz=32", u.Host)
        },
    }

    // CRITICAL: Funcs() BEFORE Parse*()
    tmpl := template.New("").Funcs(funcMap)

    // Parse templates in correct order
    tmpl, err := tmpl.ParseGlob("templates/*.gohtml")
    if err != nil {
        return nil, fmt.Errorf("failed to parse base templates: %w", err)
    }
    tmpl, err = tmpl.ParseGlob("templates/pages/*.gohtml")
    if err != nil {
        return nil, fmt.Errorf("failed to parse page templates: %w", err)
    }
    tmpl, err = tmpl.ParseGlob("templates/partials/*.gohtml")
    if err != nil {
        return nil, fmt.Errorf("failed to parse partial templates: %w", err)
    }

    s := &Server{
        db:        db,
        templates: tmpl,
        mux:       http.NewServeMux(),
    }

    s.registerRoutes()
    return s, nil
}
```

### ArticleWithBlog Model
```go
// Source: Go database patterns
type ArticleWithBlog struct {
    ID             int64
    BlogID         int64
    Title          string
    URL            string
    PublishedDate  *time.Time
    DiscoveredDate *time.Time
    IsRead         bool
    BlogName       string
    BlogURL        string
}
```

### Database Query with JOIN
```go
// Source: https://go.dev/doc/database/querying
func (db *Database) ListArticlesWithBlog(isRead bool, blogID *int64) ([]model.ArticleWithBlog, error) {
    query := `
        SELECT
            a.id, a.blog_id, a.title, a.url,
            a.published_date, a.discovered_date, a.is_read,
            b.name, b.url
        FROM articles a
        INNER JOIN blogs b ON a.blog_id = b.id
        WHERE a.is_read = ?`
    args := []interface{}{isRead}

    if blogID != nil {
        query += " AND a.blog_id = ?"
        args = append(args, *blogID)
    }
    query += " ORDER BY a.discovered_date DESC"

    rows, err := db.conn.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var articles []model.ArticleWithBlog
    for rows.Next() {
        var a model.ArticleWithBlog
        var pubDate, discDate sql.NullString

        err := rows.Scan(
            &a.ID, &a.BlogID, &a.Title, &a.URL,
            &pubDate, &discDate, &a.IsRead,
            &a.BlogName, &a.BlogURL,
        )
        if err != nil {
            return nil, err
        }

        if pubDate.Valid {
            if t, err := parseTime(pubDate.String); err == nil {
                a.PublishedDate = &t
            }
        }
        if discDate.Valid {
            if t, err := parseTime(discDate.String); err == nil {
                a.DiscoveredDate = &t
            }
        }

        articles = append(articles, a)
    }
    return articles, rows.Err()
}
```

### Article Card Template
```html
{{/* Source: HTMX patterns + CSS best practices */}}
{{define "article-list.gohtml"}}
<h1>{{if eq .CurrentFilter "read"}}Archived{{else}}Inbox{{end}}</h1>
{{range .Articles}}
<article class="article-card">
    <img class="article-favicon"
         src="{{faviconURL .BlogURL}}"
         alt=""
         width="32"
         height="32"
         loading="lazy"
         onerror="this.style.visibility='hidden'">
    <div class="article-content">
        <a href="{{.URL}}"
           target="_blank"
           rel="noopener noreferrer"
           class="article-title">
            {{.Title}}
        </a>
        <div class="article-meta">
            <span class="article-source">{{.BlogName}}</span>
            {{if .PublishedDate}}
            <span class="article-time">{{timeAgo .PublishedDate}}</span>
            {{end}}
        </div>
    </div>
</article>
{{else}}
<p class="empty-state">No articles to display.</p>
{{end}}
{{end}}
```

### Article Card CSS
```css
/* Source: CSS best practices for card layouts */
.article-card {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    padding: 1rem;
    margin-bottom: 0.5rem;
    background-color: var(--bg-surface);
    border-radius: 8px;
    border: 1px solid var(--border);
    transition: background-color var(--transition-speed) ease;
}

.article-card:hover {
    background-color: var(--bg-elevated);
}

.article-favicon {
    flex-shrink: 0;
    width: 32px;
    height: 32px;
    border-radius: 4px;
    background-color: var(--bg-elevated);
}

.article-content {
    flex: 1;
    min-width: 0;  /* Enables text truncation */
}

.article-title {
    display: block;
    font-size: 1rem;
    font-weight: 500;
    color: var(--text-primary);
    text-decoration: none;
    margin-bottom: 0.25rem;
    /* Truncate long titles */
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.article-title:hover {
    color: var(--accent);
}

.article-meta {
    display: flex;
    gap: 0.5rem;
    font-size: 0.875rem;
    color: var(--text-secondary);
}

.article-source {
    font-weight: 500;
}

.article-time::before {
    content: "·";
    margin-right: 0.5rem;
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| External time formatting libraries | Custom ~30 line stdlib implementation | Preference for minimal dependencies | No external deps for simple use case |
| Self-hosted favicon scraping | External favicon APIs (Google S2) | API availability improved | Offload complexity, reliable fallbacks |
| Separate queries for articles/blogs | SQL JOIN returning combined struct | Always best practice | Single round-trip, simpler code |
| target="_blank" alone | target="_blank" rel="noopener noreferrer" | Security standard ~2018 | Prevents reverse tabnapping |

**Deprecated/outdated:**
- **Icon Horse favicon service:** Appears to be dead as of April 2025
- **target="_blank" without rel attributes:** Modern browsers add noopener implicitly, but explicit is safer for legacy support

## Open Questions

Things that couldn't be fully resolved:

1. **Favicon service reliability in restricted networks**
   - What we know: Google S2 works globally but may be blocked by firewalls
   - What's unclear: Whether DuckDuckGo API is more accessible in restricted networks
   - Recommendation: Use Google S2 with CSS fallback; add configuration option for alternative service if users report issues

2. **Time formatting granularity preferences**
   - What we know: "7 hours ago" is standard, some apps show "today at 2:30 PM" instead
   - What's unclear: User preference for BlogWatcher
   - Recommendation: Start with relative time ("7 hours ago"), can add tooltip with absolute time later

3. **Article card click behavior**
   - What we know: Requirement says "open in new tab"
   - What's unclear: Should clicking card (not just title) open article?
   - Recommendation: Make entire card clickable for better UX on mobile, using CSS to style anchor as block

## Sources

### Primary (HIGH confidence)
- [Go html/template FuncMap documentation](https://pkg.go.dev/html/template) - Official stdlib docs
- [Go text/template function guidelines](https://pkg.go.dev/text/template) - Function return value rules
- [HTMX hx-target documentation](https://htmx.org/docs/) - Already used in Phase 2
- [MDN rel=noopener documentation](https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Attributes/rel/noopener) - Security best practices

### Secondary (MEDIUM confidence)
- [Google S2 Favicon API](https://dev.to/derlin/get-favicons-from-any-website-using-a-hidden-google-api-3p1e) - Undocumented but widely used
- [Using Functions Inside Go Templates](https://www.calhoun.io/intro-to-templates-p3-functions/) - FuncMap patterns, Jon Calhoun
- [go-humanize package](https://pkg.go.dev/github.com/dustin/go-humanize) - Reference for timeAgo output format
- [DuckDuckGo Favicon API](https://docs.logo.dev/duckduckgo-favicon-api) - Alternative favicon service
- [CSS article card patterns](https://webdesign.tutsplus.com/solving-problems-with-css-grid-and-flexbox-the-card-ui--cms-27468t) - Card layout best practices

### Tertiary (LOW confidence)
- [go-prettytime](https://github.com/andanhm/go-prettytime) - Alternative relative time library, reference only

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All stdlib except favicon API which is widely documented
- Architecture: HIGH - FuncMap patterns well-documented in official Go docs
- Pitfalls: HIGH - Security considerations from MDN, time nil handling from experience
- Favicon service: MEDIUM - Undocumented Google API but stable for years

**Research date:** 2026-02-02
**Valid until:** 2026-03-04 (30 days - stable patterns, favicon API has been unchanged for years)
