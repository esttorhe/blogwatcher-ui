# Phase 4: Article Management - Research

**Researched:** 2026-02-02
**Domain:** HTMX mutations, database write operations, blog synchronization, Go HTTP POST handlers
**Confidence:** HIGH

## Summary

Phase 4 introduces write operations to the database and integrates blog synchronization into the UI. This phase transitions from read-only to read-write database access, adding handlers for marking articles read/unread and triggering the existing scanner functionality from the reference blogwatcher CLI.

The recommended approach involves:
1. Adding POST endpoints for mark read/unread that return updated article cards or empty responses
2. Implementing a "Mark all read" bulk action using HTMX form submission
3. Copying and integrating the scanner package from the reference codebase for sync functionality
4. Using HTMX patterns for optimistic UI updates with proper loading indicators
5. Ensuring database methods already exist (MarkArticleRead/MarkArticleUnread are implemented)

Key architectural insight: The database already has `MarkArticleRead` and `MarkArticleUnread` methods. The scanner package from the reference codebase provides complete RSS/scraper functionality that can be imported. The main work is HTTP handlers and HTMX UI integration.

**Primary recommendation:** Use HTMX `hx-post` with `hx-swap="outerHTML"` for individual article actions that remove the card from view, use `hx-swap-oob` for updating counts/lists after bulk operations, and integrate the reference scanner package with gofeed/goquery dependencies for sync.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| HTMX hx-post | 2.0.4+ | Issue POST requests for mutations | Already using HTMX, standard pattern for form submissions |
| HTMX hx-swap | 2.0.4+ | Replace/remove elements after mutation | Delete row pattern well-documented |
| HTMX hx-indicator | 2.0.4+ | Show loading state during requests | Built-in, CSS-based feedback |
| gofeed | 1.3.0 | Parse RSS/Atom feeds | Same as reference CLI, proven library |
| goquery | 1.10.3 | HTML scraping fallback | Same as reference CLI, jQuery-like API |
| Go http.ServeMux | stdlib | POST route handling | Go 1.22+ method routing already in use |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| HTMX hx-swap-oob | 2.0.4+ | Update multiple elements from one response | Updating sidebar counts after bulk action |
| HTMX hx-confirm | 2.0.4+ | Confirmation dialogs | "Mark all read" confirmation |
| CSS transitions | Native | Fade out removed articles | Visual feedback on deletion |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| hx-swap="outerHTML" with empty | hx-swap="delete" | delete requires no response content but less flexible |
| Copy scanner package | Import as dependency | Copying simpler, avoids module path issues |
| Inline loading spinner | Global indicator | Inline shows action-specific feedback |

**Installation:**
```bash
# Add dependencies for scanner functionality
go get github.com/mmcdole/gofeed@v1.3.0
go get github.com/PuerkitoBio/goquery@v1.10.3
```

## Architecture Patterns

### Recommended Project Structure
```
internal/
  server/
    handlers.go      # UPDATE: Add POST handlers for mark read/unread/sync
    routes.go        # UPDATE: Register new POST routes
  storage/
    database.go      # Already has MarkArticleRead/MarkArticleUnread + add bulk + sync helpers
  scanner/           # NEW: Copy from reference codebase
    scanner.go       # ScanAllBlogs, ScanBlog functions
  rss/               # NEW: Copy from reference codebase
    rss.go           # ParseFeed, DiscoverFeedURL
  scraper/           # NEW: Copy from reference codebase
    scraper.go       # ScrapeBlog fallback
templates/
  partials/
    article-list.gohtml     # UPDATE: Add mark read/unread buttons
    article-card.gohtml     # NEW: Individual card partial for swap responses
static/
  styles.css                # UPDATE: Button styles, loading indicator CSS
```

### Pattern 1: HTMX POST for Single Article Action
**What:** Button triggers POST, server returns replacement HTML or empty
**When to use:** Mark individual article read/unread
**Example:**
```html
{{/* Source: https://htmx.org/examples/delete-row/ */}}
<article class="article-card" id="article-{{.ID}}">
    <img class="article-favicon" src="{{faviconURL .BlogURL}}" alt="" width="32" height="32">
    <div class="article-content">
        <a href="{{.URL}}" target="_blank" rel="noopener noreferrer" class="article-title">
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
            hx-indicator="#article-{{.ID}} .action-indicator">
        {{if .IsRead}}Unread{{else}}Read{{end}}
        <span class="action-indicator htmx-indicator">...</span>
    </button>
</article>
```

### Pattern 2: Server Response for Mark Read (Remove from Inbox)
**What:** Return empty response when article should disappear from current view
**When to use:** Mark read in Inbox view, mark unread in Archived view
**Example:**
```go
// Source: https://htmx.org/examples/delete-row/
func (s *Server) handleMarkRead(w http.ResponseWriter, r *http.Request) {
    // Parse article ID from path
    id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        http.Error(w, "Invalid article ID", http.StatusBadRequest)
        return
    }

    // Mark as read
    found, err := s.db.MarkArticleRead(id)
    if err != nil {
        log.Printf("Error marking article read: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    if !found {
        http.NotFound(w, r)
        return
    }

    // Return empty 200 - HTMX will swap with nothing, removing the card
    w.WriteHeader(http.StatusOK)
}
```

### Pattern 3: Bulk Mark All Read with Form
**What:** Form wraps article list, button submits all visible article IDs
**When to use:** "Mark all read" button
**Example:**
```html
{{/* Source: https://htmx.org/examples/bulk-update/ */}}
<form id="article-form" hx-post="/articles/mark-all-read"
      hx-target="#main-content"
      hx-swap="innerHTML"
      hx-confirm="Mark all articles as read?">
    <div class="toolbar">
        <button type="submit" class="btn-action">
            Mark All Read
            <span class="htmx-indicator">...</span>
        </button>
    </div>
    <div id="article-list">
        {{range .Articles}}
        <input type="hidden" name="ids" value="{{.ID}}">
        <article class="article-card">...</article>
        {{end}}
    </div>
</form>
```

### Pattern 4: Sync Button with Loading State
**What:** Button triggers sync, shows progress, refreshes article list
**When to use:** Manual sync button
**Example:**
```html
{{/* Sync button in toolbar */}}
<button class="btn-action sync-btn"
        hx-post="/sync"
        hx-target="#main-content"
        hx-swap="innerHTML"
        hx-indicator=".sync-indicator">
    <span class="sync-icon">Sync</span>
    <span class="sync-indicator htmx-indicator">Syncing...</span>
</button>
```

### Pattern 5: Scanner Integration
**What:** Copy scanner, rss, scraper packages from reference
**When to use:** Sync functionality
**Example:**
```go
// Source: .reference/blogwatcher/internal/scanner/scanner.go
// Copy entire scanner package, update import paths
// internal/scanner/scanner.go
package scanner

import (
    "github.com/esttorhe/blogwatcher-ui/internal/model"
    "github.com/esttorhe/blogwatcher-ui/internal/rss"
    "github.com/esttorhe/blogwatcher-ui/internal/scraper"
    "github.com/esttorhe/blogwatcher-ui/internal/storage"
)

// ScanResult, ScanBlog, ScanAllBlogs functions unchanged
```

### Pattern 6: CSS Fade Out Animation for Removed Cards
**What:** Animate card removal for visual feedback
**When to use:** When article is marked read and removed from view
**Example:**
```css
/* Source: https://htmx.org/examples/delete-row/ */
.article-card.htmx-swapping {
    opacity: 0;
    transition: opacity 300ms ease-out;
}
```

### Anti-Patterns to Avoid
- **POST without CSRF protection:** Not needed for single-user local app, but consider if deploying remotely
- **Returning 204 No Content:** HTMX won't swap on 204, return 200 with empty body instead
- **Full page reload after action:** Use HTMX partial swaps for SPA-like experience
- **Sync blocking UI:** Scanner can take time, show loading indicator prominently
- **N+1 mark read calls:** Use bulk endpoint for mark all, not individual calls per article

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| RSS parsing | Custom XML parsing | gofeed library | Handles Atom, RSS 1.0/2.0, JSON Feed, edge cases |
| HTML scraping | regex or manual DOM | goquery library | jQuery-like API, handles malformed HTML |
| Feed URL discovery | Manual link parsing | rss.DiscoverFeedURL from reference | Handles autodiscovery, common paths |
| Loading indicator | JavaScript spinner | HTMX htmx-indicator class | Pure CSS, no JS needed |
| Confirmation dialog | Custom modal | hx-confirm attribute | Browser native, accessible |
| Delete animation | JavaScript animation | CSS transition on htmx-swapping class | Declarative, no JS |

**Key insight:** The reference blogwatcher CLI already has battle-tested scanner, rss, and scraper packages. Copy them rather than reimplementing. The only adaptation needed is updating import paths.

## Common Pitfalls

### Pitfall 1: Returning 204 for HTMX Swap
**What goes wrong:** Article card not removed from DOM
**Why it happens:** HTMX ignores 204 No Content responses by default
**How to avoid:** Return 200 OK with empty body, or configure htmx to handle 204
```go
// WRONG - HTMX ignores 204
w.WriteHeader(http.StatusNoContent)

// CORRECT - Return 200 with empty body
w.WriteHeader(http.StatusOK)
```
**Warning signs:** Card remains visible after successful POST

### Pitfall 2: Missing hx-swap Timing for Animations
**What goes wrong:** Card disappears instantly without fade
**Why it happens:** Default swap happens immediately
**How to avoid:** Use `hx-swap="outerHTML swap:300ms"` to delay swap
**Warning signs:** Jarring instant removal instead of smooth fade

### Pitfall 3: Scanner Timeout on Slow Feeds
**What goes wrong:** Sync hangs or times out
**Why it happens:** Some feeds are slow, HTTP default timeout too long
**How to avoid:** Set explicit timeout in HTTP client (30s like reference)
```go
client := &http.Client{Timeout: 30 * time.Second}
```
**Warning signs:** UI appears frozen during sync

### Pitfall 4: Race Condition with Multiple Syncs
**What goes wrong:** Database locked errors, duplicate articles
**Why it happens:** User clicks sync multiple times
**How to avoid:** Disable button during request using htmx-request class
```css
.sync-btn.htmx-request {
    pointer-events: none;
    opacity: 0.5;
}
```
**Warning signs:** SQLite BUSY errors, duplicate article entries

### Pitfall 5: Bulk Operation Without Transaction
**What goes wrong:** Partial updates if error occurs
**Why it happens:** Individual UPDATE calls without transaction
**How to avoid:** Use transaction for bulk mark all read
```go
func (db *Database) MarkAllArticlesRead(ids []int64) error {
    tx, err := db.conn.Begin()
    // ... mark each, rollback on error
}
```
**Warning signs:** Some articles marked, others not, on error

### Pitfall 6: Path Variable Parsing in Go 1.22+
**What goes wrong:** Article ID not parsed from URL
**Why it happens:** Using wrong method to get path variable
**How to avoid:** Use `r.PathValue("id")` not `r.URL.Query().Get("id")`
```go
// WRONG for /articles/{id}/read
id := r.URL.Query().Get("id")

// CORRECT for Go 1.22+ method routing
id := r.PathValue("id")
```
**Warning signs:** Empty or nil ID, 404 errors

### Pitfall 7: Scanner Package Import Path Issues
**What goes wrong:** Circular imports or module not found
**Why it happens:** Copying reference code without updating imports
**How to avoid:** Update all import paths to new module
```go
// Change from
import "github.com/Hyaxia/blogwatcher/internal/model"

// To
import "github.com/esttorhe/blogwatcher-ui/internal/model"
```
**Warning signs:** Build errors about module paths

## Code Examples

Verified patterns from official sources:

### Mark Read Handler
```go
// Source: Go 1.22+ routing + HTMX delete row pattern
func (s *Server) handleMarkRead(w http.ResponseWriter, r *http.Request) {
    idStr := r.PathValue("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid article ID", http.StatusBadRequest)
        return
    }

    found, err := s.db.MarkArticleRead(id)
    if err != nil {
        log.Printf("Error marking article %d read: %v", id, err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    if !found {
        http.NotFound(w, r)
        return
    }

    // Return empty 200 - article card will be removed from DOM
    w.WriteHeader(http.StatusOK)
}
```

### Mark Unread Handler (Returns Updated Card)
```go
// When in Archived view, mark unread removes from view
// When in Inbox view, mark unread does nothing (shouldn't be visible)
func (s *Server) handleMarkUnread(w http.ResponseWriter, r *http.Request) {
    idStr := r.PathValue("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid article ID", http.StatusBadRequest)
        return
    }

    found, err := s.db.MarkArticleUnread(id)
    if err != nil {
        log.Printf("Error marking article %d unread: %v", id, err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    if !found {
        http.NotFound(w, r)
        return
    }

    // Return empty 200 - article card will be removed from Archived view
    w.WriteHeader(http.StatusOK)
}
```

### Mark All Read Handler
```go
// Source: HTMX bulk update pattern
func (s *Server) handleMarkAllRead(w http.ResponseWriter, r *http.Request) {
    // Parse form to get current filter context
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Invalid form", http.StatusBadRequest)
        return
    }

    // Get blogID filter if present
    var blogID *int64
    if blogParam := r.FormValue("blog"); blogParam != "" {
        if id, err := strconv.ParseInt(blogParam, 10, 64); err == nil {
            blogID = &id
        }
    }

    // Mark all unread articles as read
    err := s.db.MarkAllUnreadArticlesRead(blogID)
    if err != nil {
        log.Printf("Error marking all read: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    // Return refreshed article list (now empty for inbox)
    articles, err := s.db.ListArticlesWithBlog(false, blogID)
    if err != nil {
        log.Printf("Error fetching articles: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    data := map[string]interface{}{
        "Articles":      articles,
        "CurrentFilter": "unread",
        "CurrentBlogID": blogID,
    }
    s.renderTemplate(w, "article-list.gohtml", data)
}
```

### Sync Handler
```go
// Source: Reference blogwatcher scanner integration
func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
    // Run scanner synchronously (could be async with goroutine for large feeds)
    results, err := scanner.ScanAllBlogs(s.db, 1) // Single worker for simplicity
    if err != nil {
        log.Printf("Error during sync: %v", err)
        http.Error(w, "Sync failed", http.StatusInternalServerError)
        return
    }

    // Log results
    totalNew := 0
    for _, result := range results {
        if result.Error != "" {
            log.Printf("Sync error for %s: %s", result.BlogName, result.Error)
        } else {
            log.Printf("Synced %s: %d new articles", result.BlogName, result.NewArticles)
            totalNew += result.NewArticles
        }
    }
    log.Printf("Sync complete: %d new articles total", totalNew)

    // Return refreshed article list
    filter := r.URL.Query().Get("filter")
    if filter == "" {
        filter = "unread"
    }
    isRead := filter == "read"

    articles, err := s.db.ListArticlesWithBlog(isRead, nil)
    if err != nil {
        log.Printf("Error fetching articles: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    data := map[string]interface{}{
        "Articles":      articles,
        "CurrentFilter": filter,
        "CurrentBlogID": int64(0),
    }
    s.renderTemplate(w, "article-list.gohtml", data)
}
```

### Route Registration
```go
// Source: Go 1.22+ method routing
func (s *Server) registerRoutes() {
    // Static files
    fs := http.FileServer(http.Dir("static"))
    s.mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

    // Pages (existing)
    s.mux.HandleFunc("GET /", s.handleIndex)
    s.mux.HandleFunc("GET /articles", s.handleArticleList)
    s.mux.HandleFunc("GET /blogs", s.handleBlogList)

    // Article actions (NEW)
    s.mux.HandleFunc("POST /articles/{id}/read", s.handleMarkRead)
    s.mux.HandleFunc("POST /articles/{id}/unread", s.handleMarkUnread)
    s.mux.HandleFunc("POST /articles/mark-all-read", s.handleMarkAllRead)

    // Sync (NEW)
    s.mux.HandleFunc("POST /sync", s.handleSync)
}
```

### Database Bulk Operation
```go
// Add to storage/database.go
func (db *Database) MarkAllUnreadArticlesRead(blogID *int64) error {
    query := `UPDATE articles SET is_read = 1 WHERE is_read = 0`
    var args []interface{}

    if blogID != nil {
        query += " AND blog_id = ?"
        args = append(args, *blogID)
    }

    _, err := db.conn.Exec(query, args...)
    return err
}
```

### Updated Article Card Template
```html
{{define "article-card.gohtml"}}
{{/* ABOUTME: Individual article card with action button for mark read/unread.
     ABOUTME: Used both in list and as HTMX swap response. */}}
<article class="article-card" id="article-{{.ID}}">
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
    {{if .IsRead}}
    <button class="action-btn"
            hx-post="/articles/{{.ID}}/unread"
            hx-target="#article-{{.ID}}"
            hx-swap="outerHTML swap:300ms"
            title="Mark as unread">
        <span class="action-icon">Unread</span>
    </button>
    {{else}}
    <button class="action-btn"
            hx-post="/articles/{{.ID}}/read"
            hx-target="#article-{{.ID}}"
            hx-swap="outerHTML swap:300ms"
            title="Mark as read">
        <span class="action-icon">Read</span>
    </button>
    {{end}}
</article>
{{end}}
```

### CSS Additions
```css
/* Action button styling */
.action-btn {
    flex-shrink: 0;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--border);
    border-radius: 4px;
    background-color: var(--bg-elevated);
    color: var(--text-primary);
    cursor: pointer;
    font-size: 0.75rem;
    transition: background-color var(--transition-speed) ease,
                opacity var(--transition-speed) ease;
}

.action-btn:hover {
    background-color: var(--accent);
    color: var(--bg-primary);
}

/* Disable during request */
.action-btn.htmx-request {
    pointer-events: none;
    opacity: 0.5;
}

/* Fade out animation for removed cards */
.article-card.htmx-swapping {
    opacity: 0;
    transition: opacity 300ms ease-out;
}

/* Toolbar styling */
.toolbar {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1rem;
    padding: 0.5rem 0;
    border-bottom: 1px solid var(--border);
}

.btn-action {
    padding: 0.5rem 1rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background-color: var(--bg-surface);
    color: var(--text-primary);
    cursor: pointer;
    font-size: 0.875rem;
    transition: background-color var(--transition-speed) ease;
}

.btn-action:hover {
    background-color: var(--bg-elevated);
}

.btn-action.htmx-request {
    pointer-events: none;
    opacity: 0.5;
}

/* Loading indicator */
.htmx-indicator {
    display: none;
}

.htmx-request .htmx-indicator,
.htmx-request.htmx-indicator {
    display: inline;
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Form POST with full page reload | HTMX partial swap | HTMX 2.0+ | SPA-like UX without JavaScript |
| JavaScript delete animation | CSS transition on htmx-swapping | HTMX 1.0+ | Declarative, no JS needed |
| Background sync with WebSocket | Manual sync button with loading state | Design decision | Simpler, user controls refresh |
| Multiple AJAX calls for bulk | Single POST with form data | Always best practice | Single round-trip |

**Deprecated/outdated:**
- **JavaScript confirm() calls:** Use hx-confirm for native browser dialog
- **XMLHttpRequest manual handling:** HTMX abstracts this
- **Full page reload patterns:** HTMX partial swaps are standard now

## Open Questions

Things that couldn't be fully resolved:

1. **Sync progress reporting**
   - What we know: Sync can take 30+ seconds for many blogs
   - What's unclear: Should we show per-blog progress or just a spinner?
   - Recommendation: Start with simple spinner, add progress if users complain about uncertainty

2. **Confirmation for mark all read**
   - What we know: hx-confirm provides native browser dialog
   - What's unclear: Is this disruptive or helpful?
   - Recommendation: Add hx-confirm since action is destructive and not easily reversible

3. **Error handling UI**
   - What we know: HTMX has error events (htmx:responseError)
   - What's unclear: How to show errors in this UI without JavaScript
   - Recommendation: Server can return error partial that replaces content with error message

4. **Scanner concurrency**
   - What we know: Reference CLI supports multiple workers
   - What's unclear: Whether concurrent DB writes cause issues with single SQLite connection
   - Recommendation: Use single worker (workers=1) for simplicity, avoiding concurrent write issues

## Sources

### Primary (HIGH confidence)
- [htmx hx-post documentation](https://htmx.org/attributes/hx-post/) - Official POST attribute docs
- [htmx Delete Row example](https://htmx.org/examples/delete-row/) - Official delete row pattern
- [htmx Bulk Update example](https://htmx.org/examples/bulk-update/) - Official bulk update pattern
- [htmx hx-indicator documentation](https://htmx.org/attributes/hx-indicator/) - Official loading indicator docs
- [htmx hx-swap-oob documentation](https://htmx.org/attributes/hx-swap-oob/) - Out-of-band swaps
- [Go 1.22 http.ServeMux routing](https://go.dev/blog/routing-enhancements) - PathValue method

### Secondary (MEDIUM confidence)
- [gofeed library](https://github.com/mmcdole/gofeed) - RSS/Atom parsing
- [goquery library](https://github.com/PuerkitoBio/goquery) - HTML scraping
- Reference blogwatcher CLI scanner package - Proven implementation

### Tertiary (LOW confidence)
- Community HTMX patterns from search results - Various tutorials

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - HTMX patterns well-documented, reference code exists
- Architecture: HIGH - Patterns from official docs, database methods already exist
- Pitfalls: HIGH - Common HTMX issues documented, tested patterns
- Scanner integration: MEDIUM - Copying code, may need minor adjustments

**Research date:** 2026-02-02
**Valid until:** 2026-03-04 (30 days - HTMX stable, scanner code proven)
