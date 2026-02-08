# Phase 9: Settings Page Foundation - Research

**Researched:** 2026-02-08
**Domain:** Go templates + HTMX navigation, SQLite queries with GROUP BY
**Confidence:** HIGH

## Summary

Phase 9 adds a settings page to BlogWatcher UI where users can view all tracked blogs with their article counts. This phase builds on the existing HTMX + Go templates architecture already established in the codebase.

The standard approach for this phase follows the existing patterns:
1. Add a settings route handler that follows the same HTMX partial/full page detection pattern used by `handleArticleList` and `handleBlogList`
2. Add a gear icon to the sidebar using Feather Icons style SVG (matching existing icon patterns)
3. Create a settings page template using the same template structure (base + partials)
4. Query blogs with article counts using SQLite `LEFT JOIN` with `COUNT()` and `GROUP BY`

The codebase already has excellent patterns established - this phase extends them rather than introducing new architectural patterns.

**Primary recommendation:** Follow existing handler patterns (HTMX header detection for partial/full page), use SQLite LEFT JOIN with GROUP BY for blog counts, add settings link to sidebar nav section, create settings templates matching the existing article-list structure.

## Standard Stack

The project already uses these libraries - no new dependencies needed:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go 1.22+ | 1.22+ | HTTP server with method routing | Built-in ServeMux with method support (GET/POST routing) |
| html/template | stdlib | Server-side templating | Go standard library, automatic XSS escaping, template inheritance |
| modernc.org/sqlite | Latest | SQLite driver | Pure Go SQLite driver, already in use |
| htmx.js | 1.x (static) | Hypermedia interactions | Already embedded in static/, enables SPA-like navigation |

### Supporting
No additional libraries needed - all functionality uses existing stack.

**Installation:**
No new packages needed. All dependencies already in go.mod.

## Architecture Patterns

### Recommended Project Structure
Based on existing codebase structure:
```
templates/
├── base.gohtml              # Layout with {{template "content" .}}
├── pages/
│   ├── index.gohtml        # Main app page
│   └── settings.gohtml     # NEW: Settings page (follows index pattern)
└── partials/
    ├── sidebar.gohtml      # Updated: Add gear icon link
    ├── article-list.gohtml
    ├── blog-list.gohtml
    └── settings-page.gohtml  # NEW: Settings content partial

internal/server/
├── routes.go               # Updated: Add GET /settings
├── handlers.go             # Updated: Add handleSettings
└── database.go (storage/)  # Updated: Add ListBlogsWithCounts method
```

### Pattern 1: HTMX Navigation with Partial Rendering
**What:** Handlers detect `HX-Request: true` header and return either a partial fragment or full page
**When to use:** All navigation between main views (Inbox, Archived, Settings)
**Example from codebase:**
```go
// Source: internal/server/handlers.go lines 112-133
func (s *Server) handleArticleList(w http.ResponseWriter, r *http.Request) {
    // ... fetch data ...

    // Check if this is an HTMX request
    if r.Header.Get("HX-Request") == "true" {
        // Return partial fragment for HTMX
        s.renderTemplate(w, "article-list.gohtml", data)
        return
    }

    // Return full page for direct navigation
    data["Title"] = "BlogWatcher"
    data["Version"] = s.version
    // ... add sidebar data ...
    s.renderTemplate(w, "index.gohtml", data)
}
```

**Apply to Settings:**
```go
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
    blogs, err := s.db.ListBlogsWithCounts()
    // ... error handling ...

    data := map[string]interface{}{
        "Blogs": blogs,
    }

    if r.Header.Get("HX-Request") == "true" {
        s.renderTemplate(w, "settings-page.gohtml", data)
        return
    }

    // Full page for direct URL access
    data["Title"] = "Settings"
    data["Version"] = s.version
    s.renderTemplate(w, "settings.gohtml", data)
}
```

### Pattern 2: Sidebar Navigation Links
**What:** HTMX-powered links that swap main content without page reload
**When to use:** Primary navigation between app sections
**Example from codebase:**
```html
<!-- Source: templates/partials/sidebar.gohtml lines 25-32 -->
<a href="/articles?filter=unread"
   hx-get="/articles?filter=unread"
   hx-target="#main-content"
   hx-push-url="true"
   hx-on::after-swap="document.getElementById('sidebar-toggle').checked = false"
   class="nav-link{{if eq .CurrentFilter "unread"}} active{{end}}">
    Inbox
</a>
```

**Apply to Settings:**
Add gear icon link to sidebar navigation section (between nav and subscriptions):
```html
<a href="/settings"
   hx-get="/settings"
   hx-target="#main-content"
   hx-push-url="true"
   hx-on::after-swap="document.getElementById('sidebar-toggle').checked = false"
   class="nav-link{{if .IsSettingsPage}} active{{end}}">
    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="3"></circle>
        <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
    </svg>
    <span>Settings</span>
</a>
```

### Pattern 3: SQLite Query with Article Counts
**What:** Use LEFT JOIN with COUNT and GROUP BY to get blog metadata with article counts
**When to use:** Displaying blogs with aggregate statistics
**Existing pattern from codebase:**
```go
// Source: internal/storage/database.go line 291
// The codebase uses COUNT(*) OVER() as window function for pagination counts
query.WriteString(`SELECT a.id, a.blog_id, a.title, a.url, a.thumbnail_url,
    a.published_date, a.discovered_date, a.is_read, b.name, b.url,
    COUNT(*) OVER() as total_count
    FROM articles a`)
```

**Apply to Settings (new method for Database):**
```go
// BlogWithCount extends Blog with article count
type BlogWithCount struct {
    model.Blog
    ArticleCount int
}

// ListBlogsWithCounts returns all blogs with their article counts
func (db *Database) ListBlogsWithCounts() ([]BlogWithCount, error) {
    query := `
        SELECT
            b.id,
            b.name,
            b.url,
            b.feed_url,
            b.scrape_selector,
            b.last_scanned,
            COUNT(a.id) as article_count
        FROM blogs b
        LEFT JOIN articles a ON b.id = a.blog_id
        GROUP BY b.id, b.name, b.url, b.feed_url, b.scrape_selector, b.last_scanned
        ORDER BY b.name
    `

    rows, err := db.conn.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var blogs []BlogWithCount
    for rows.Next() {
        var (
            id             int64
            name           string
            url            string
            feedURL        sql.NullString
            scrapeSelector sql.NullString
            lastScanned    sql.NullString
            articleCount   int
        )

        if err := rows.Scan(&id, &name, &url, &feedURL, &scrapeSelector,
            &lastScanned, &articleCount); err != nil {
            return nil, err
        }

        blog := BlogWithCount{
            Blog: model.Blog{
                ID:             id,
                Name:           name,
                URL:            url,
                FeedURL:        feedURL.String,
                ScrapeSelector: scrapeSelector.String,
            },
            ArticleCount: articleCount,
        }

        if lastScanned.Valid {
            if parsed, err := parseTime(lastScanned.String); err == nil {
                blog.LastScanned = &parsed
            }
        }

        blogs = append(blogs, blog)
    }

    return blogs, rows.Err()
}
```

**Why LEFT JOIN:** Ensures blogs with zero articles still appear in results (COUNT will be 0).

### Pattern 4: Template Page Structure
**What:** Settings page follows same structure as index.gohtml with sidebar + main content
**When to use:** All full-page views that need sidebar navigation
**Example from codebase:**
```html
<!-- Source: templates/pages/index.gohtml -->
{{define "index.gohtml"}}
{{template "base" .}}
{{end}}

{{define "title"}}BlogWatcher{{end}}

{{define "content"}}
<div class="app-layout">
    {{template "sidebar.gohtml" .}}

    <main id="main-content" class="main-content">
        {{template "article-list.gohtml" .}}
    </main>
</div>
{{end}}
```

**Apply to Settings:**
```html
<!-- templates/pages/settings.gohtml -->
{{define "settings.gohtml"}}
{{template "base" .}}
{{end}}

{{define "title"}}Settings - BlogWatcher{{end}}

{{define "content"}}
<div class="app-layout">
    {{template "sidebar.gohtml" .}}

    <main id="main-content" class="main-content">
        {{template "settings-page.gohtml" .}}
    </main>
</div>
{{end}}
```

### Pattern 5: Settings Content Layout
**What:** Settings partial displays blog list in table/card format with metadata
**When to use:** HTMX partial swap and embedded in settings page
**Structure:**
```html
<!-- templates/partials/settings-page.gohtml -->
{{define "settings-page.gohtml"}}
<div class="main-content-header">
    <header class="main-header">
        <h1>Settings</h1>
    </header>
</div>
<div class="main-content-body">
    <section class="settings-section">
        <h2>Tracked Blogs</h2>
        {{if .Blogs}}
        <div class="blog-settings-list">
            {{range .Blogs}}
            <div class="blog-settings-card">
                <div class="blog-settings-info">
                    <h3 class="blog-settings-name">{{.Name}}</h3>
                    <a href="{{.URL}}" target="_blank" rel="noopener noreferrer"
                       class="blog-settings-url">{{.URL}}</a>
                    <div class="blog-settings-meta">
                        <span class="article-count">{{.ArticleCount}} article{{if ne .ArticleCount 1}}s{{end}}</span>
                    </div>
                </div>
            </div>
            {{end}}
        </div>
        {{else}}
        <p class="empty-state">No blogs tracked yet. Use the blogwatcher CLI to add blogs.</p>
        {{end}}
    </section>
</div>
{{end}}
```

### Anti-Patterns to Avoid

- **Don't create separate settings route structure:** Settings should use the same HTMX + template pattern as articles/blogs, not introduce a new architectural pattern
- **Don't query article counts in separate queries:** Use JOIN with GROUP BY in a single query, not N+1 pattern
- **Don't bypass existing template hierarchy:** Settings page should use base.gohtml template, not inline HTML
- **Don't use JavaScript for navigation:** HTMX attributes handle all navigation, keep JavaScript for theme/view toggles only

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Active link highlighting | Custom JavaScript to track current page | Template data with conditional classes | Go templates already receive route context, use `{{if .IsSettingsPage}} active{{end}}` pattern |
| Counting articles per blog | Loop through blogs, query each for count | Single SQL query with GROUP BY | N+1 query problem - single query is orders of magnitude faster |
| Gear/settings icon | Hand-draw SVG or use image file | Feather Icons SVG inline | Feather Icons already used throughout app, consistent style |
| Partial vs full page detection | Custom query parameter or route | HTMX `HX-Request` header | Standard HTMX pattern, already implemented in codebase |

**Key insight:** The codebase has already solved all the patterns needed for this phase. Don't reinvent - extend existing patterns.

## Common Pitfalls

### Pitfall 1: N+1 Query Pattern
**What goes wrong:** Fetching blogs, then looping to count articles for each blog separately
**Why it happens:** Seems intuitive to get blogs then "enhance" with counts
**How to avoid:** Use LEFT JOIN with GROUP BY in single query. LEFT JOIN ensures blogs with 0 articles still appear.
**Warning signs:**
- Multiple database queries in loop
- "for blog in blogs { count := db.CountArticles(blog.ID) }" pattern
- Slow page load with many blogs

**Correct approach:**
```sql
SELECT b.*, COUNT(a.id) as article_count
FROM blogs b
LEFT JOIN articles a ON b.id = a.blog_id
GROUP BY b.id
```

### Pitfall 2: Breaking HTMX History Navigation
**What goes wrong:** Settings page works via HTMX but breaks on direct URL access or refresh
**Why it happens:** Handler only returns partial template, not checking HX-Request header
**How to avoid:** Follow existing handler pattern - check `r.Header.Get("HX-Request")` and return full page for direct access
**Warning signs:**
- Settings loads from sidebar but refresh shows empty page
- Direct URL `/settings` doesn't work
- Browser back button breaks

**Detection:**
```go
if r.Header.Get("HX-Request") == "true" {
    // HTMX swap - return partial
    s.renderTemplate(w, "settings-page.gohtml", data)
} else {
    // Direct navigation - return full page
    data["Title"] = "Settings"
    s.renderTemplate(w, "settings.gohtml", data)
}
```

### Pitfall 3: Inconsistent Icon Styling
**What goes wrong:** Gear icon looks different from other sidebar icons (wrong size, stroke, color)
**Why it happens:** Copy/paste icon from different source without matching existing icon attributes
**How to avoid:** Use exact same SVG attributes as existing icons (width="18" height="18", stroke-width="2", stroke="currentColor")
**Warning signs:**
- Icon larger/smaller than theme toggle icons
- Different stroke thickness
- Icon doesn't change color with theme
- Icon doesn't match Feather Icons style

**Correct attributes:**
```html
<svg xmlns="http://www.w3.org/2000/svg"
     width="18" height="18"
     viewBox="0 0 24 24"
     fill="none"
     stroke="currentColor"
     stroke-width="2"
     stroke-linecap="round"
     stroke-linejoin="round">
```

### Pitfall 4: Missing GROUP BY Columns in SQLite
**What goes wrong:** SQLite error "column must appear in GROUP BY" when selecting blog fields
**Why it happens:** SQLite strict mode requires all non-aggregated columns in GROUP BY clause
**How to avoid:** Include all selected non-aggregate columns in GROUP BY, or use `GROUP BY b.id` if blog_id uniquely identifies all blog fields
**Warning signs:**
- Query works in SQLite CLI but fails in app
- Error mentions "GROUP BY clause"
- Adding columns to SELECT breaks query

**Solution:**
```sql
-- Include all non-aggregate columns in GROUP BY
GROUP BY b.id, b.name, b.url, b.feed_url, b.scrape_selector, b.last_scanned

-- OR rely on b.id being PRIMARY KEY (SQLite allows this)
GROUP BY b.id
```

Note: The codebase already uses `modernc.org/sqlite` which may have different strictness than standard SQLite. Test both approaches.

### Pitfall 5: Forgetting to Update Sidebar Template Data
**What goes wrong:** Settings link in sidebar doesn't highlight as active when on settings page
**Why it happens:** Handler doesn't pass `IsSettingsPage` flag to template data
**How to avoid:** Add settings page detection to template data, similar to `CurrentFilter` pattern
**Warning signs:**
- Settings link never shows active state
- Template errors about missing `IsSettingsPage`

**Fix:**
```go
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
    // ...
    data := map[string]interface{}{
        "Blogs":          blogs,
        "IsSettingsPage": true,  // For sidebar active state
    }
    // ...
}
```

## Code Examples

Verified patterns from existing codebase:

### Route Registration
```go
// Source: internal/server/routes.go
// Add to registerRoutes() method
s.mux.HandleFunc("GET /settings", s.handleSettings)
```

### Settings Handler (Full Implementation)
```go
// Source: Follow pattern from handlers.go handleArticleList/handleBlogList
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
    blogs, err := s.db.ListBlogsWithCounts()
    if err != nil {
        log.Printf("Error fetching blogs with counts: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    data := map[string]interface{}{
        "Blogs":          blogs,
        "IsSettingsPage": true,
    }

    // Check if this is an HTMX request
    if r.Header.Get("HX-Request") == "true" {
        // Return partial fragment for HTMX
        s.renderTemplate(w, "settings-page.gohtml", data)
        return
    }

    // Return full page for direct navigation
    data["Title"] = "Settings - BlogWatcher"
    data["Version"] = s.version
    s.renderTemplate(w, "settings.gohtml", data)
}
```

### Sidebar Update (Add Settings Link)
```html
<!-- Source: templates/partials/sidebar.gohtml -->
<!-- Add after Archived link, before Subscriptions section -->
<a href="/settings"
   hx-get="/settings"
   hx-target="#main-content"
   hx-push-url="true"
   hx-on::after-swap="document.getElementById('sidebar-toggle').checked = false"
   class="nav-link{{if .IsSettingsPage}} active{{end}}">
    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="3"></circle>
        <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
    </svg>
    Settings
</a>
```

### CSS Styles (Add to styles.css)
```css
/* Settings Page Styles */
.settings-section {
    padding: 2rem;
}

.settings-section h2 {
    font-size: 1.25rem;
    color: var(--text-secondary);
    font-weight: 500;
    margin-bottom: 1.5rem;
}

.blog-settings-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.blog-settings-card {
    background-color: var(--bg-surface);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 1.5rem;
    transition: background-color 0.2s ease;
}

.blog-settings-card:hover {
    background-color: var(--bg-elevated);
}

.blog-settings-info {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.blog-settings-name {
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
}

.blog-settings-url {
    color: var(--accent);
    font-size: 0.875rem;
    word-break: break-all;
}

.blog-settings-meta {
    display: flex;
    gap: 1rem;
    font-size: 0.875rem;
    color: var(--text-secondary);
}

.article-count {
    display: flex;
    align-items: center;
    gap: 0.25rem;
}

/* Mobile responsive */
@media (max-width: 768px) {
    .settings-section {
        padding: 1rem;
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Full page reload on navigation | HTMX partial swaps | HTMX 1.x (2020+) | SPA-like UX without JavaScript frameworks |
| jQuery for DOM manipulation | HTMX attributes only | HTMX adoption | Declarative HTML over imperative JS |
| text/template for HTML | html/template | Go 1.0+ stdlib | Automatic XSS prevention via context-aware escaping |
| Multiple queries for counts | Window functions or GROUP BY | SQLite 3.25+ (2018) | Single query for aggregates |

**Current best practices (2025):**
- HTMX for hypermedia-driven UX (not REST APIs + client rendering)
- Server-side rendering with html/template (not client-side SPAs)
- Feather Icons inline SVG (not icon fonts or image sprites)
- Go 1.22+ ServeMux method routing (not third-party routers for simple cases)

**Deprecated/outdated:**
- Icon fonts (Font Awesome, etc.): Security concerns, loading overhead - use inline SVG instead
- jQuery: HTMX provides cleaner declarative approach for server interactions
- Third-party template engines (Pongo2, Jet): html/template is mature, performant, and integrated

## Open Questions

1. **Should settings page support filtering/searching blogs?**
   - What we know: Current requirements only specify viewing all blogs
   - What's unclear: Future phases may add blog editing/deletion
   - Recommendation: Start simple (display only), add search in future phase if needed

2. **Should article count be total or unread only?**
   - What we know: Requirements specify "count of articles" without qualification
   - What's unclear: More useful to show total count or unread count?
   - Recommendation: Show total count (matches database reality), can add unread count in parentheses later

3. **Should settings icon have text label?**
   - What we know: Other nav items (Inbox/Archived) have text labels
   - What's unclear: Icon-only vs icon+text for settings
   - Recommendation: Use icon+text for consistency with other nav items, improve accessibility

## Sources

### Primary (HIGH confidence)
- Existing codebase patterns:
  - `/Users/esteban.torres/workspace/github/esttorhe/blogwatcher-ui/internal/server/handlers.go` - HTMX header detection pattern
  - `/Users/esteban.torres/workspace/github/esttorhe/blogwatcher-ui/internal/server/routes.go` - Route registration pattern
  - `/Users/esteban.torres/workspace/github/esttorhe/blogwatcher-ui/internal/storage/database.go` - SQL query patterns, window functions
  - `/Users/esteban.torres/workspace/github/esttorhe/blogwatcher-ui/templates/` - Template structure and hierarchy
  - `/Users/esteban.torres/workspace/github/esttorhe/blogwatcher-ui/static/styles.css` - CSS custom properties, component patterns

### Secondary (MEDIUM confidence)
- [HTMX Documentation - Navigation Patterns](https://htmx.org/docs/) - HTMX navigation and history management
- [Effortless Page Routing Using HTMX by Paul Allies](https://paulallies.medium.com/htmx-page-navigation-07b54742d251) - HTMX SPA-like patterns
- [Clean UI with Go's HTML Templates by Uygar Öztürk Ceylan](https://medium.com/@uygaroztcyln/clean-ui-with-gos-html-templates-base-partials-and-funcmaps-4915296c9097) - Go template best practices
- [Let's Go by Alex Edwards - HTML templating and inheritance](https://lets-go.alexedwards.net/sample/02.08-html-templating-and-inheritance.html) - Template organization patterns
- [SQLite GROUP BY documentation](https://www.sqlitetutorial.net/sqlite-group-by/) - GROUP BY with COUNT patterns
- [SQLite Query Optimizer Overview](https://www.sqlite.org/optoverview.html) - Index usage with GROUP BY
- [Feather Icons GitHub](https://github.com/feathericons/feather) - Settings icon SVG markup

### Tertiary (LOW confidence - general information only)
- WebSearch: HTMX patterns and Go template rendering (verified against primary sources)
- WebSearch: SQLite GROUP BY efficiency (verified with official SQLite docs)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All dependencies already in use, no new libraries needed
- Architecture: HIGH - Patterns directly observed in existing codebase, proven working
- Pitfalls: HIGH - Based on common SQLite/HTMX issues and codebase-specific patterns
- SQL queries: HIGH - Pattern matches existing codebase use of COUNT(*) OVER(), LEFT JOIN standard practice
- Template patterns: HIGH - Directly copied from existing working templates
- Icon SVG: MEDIUM - Feather Icons standard but specific SVG from documentation/knowledge, not verified in running app

**Research date:** 2026-02-08
**Valid until:** 30 days (stable technology stack, no fast-moving dependencies)
