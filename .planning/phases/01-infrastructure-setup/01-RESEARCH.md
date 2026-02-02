# Phase 1: Infrastructure Setup - Research

**Researched:** 2026-02-02
**Domain:** Go HTTP server with HTMX and SQLite integration
**Confidence:** HIGH

## Summary

Phase 1 requires building a Go HTTP server that serves web UI using Go's html/template package, connects to an existing SQLite database at `~/.blogwatcher/blogwatcher.db`, and integrates HTMX for dynamic updates without full page reloads.

The standard approach leverages Go's standard library (`net/http`, `html/template`, `database/sql`) paired with HTMX 2.0.8 and modernc.org/sqlite (a pure-Go SQLite driver already used by the reference CLI codebase). This stack requires zero external web frameworks and minimal JavaScript, aligning with the project's Go-native philosophy.

Key architectural patterns include: NewServer constructor pattern for dependency injection, dedicated routes.go file for API surface mapping, graceful shutdown via context cancellation, and HTMX request detection via HX-Request headers to return either full pages or HTML fragments.

**Primary recommendation:** Use Go 1.24's net/http with NewServer pattern, html/template with base/partial organization, HTMX 2.0.8 self-hosted (not CDN), and modernc.org/sqlite with single-connection pool configuration for SQLite's write constraints.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go net/http | 1.24+ | HTTP server | Standard library, battle-tested, handles 25k+ req/sec, goroutine-per-request model |
| html/template | stdlib | Server-side templating | Standard library, automatic XSS protection via contextual escaping |
| database/sql | stdlib | Database interface | Standard library, connection pooling, transaction support |
| modernc.org/sqlite | 1.38.2+ | SQLite driver | Pure-Go (no CGo), already used by reference CLI, active maintenance (Jan 2026) |
| HTMX | 2.0.8 | Dynamic UI updates | Industry standard for hypermedia approach, zero build step, declarative |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Air | latest | Hot reload (dev) | Development only - watches templates/Go files, restarts server |
| gernest/hot | latest | Template hot reload | Alternative to Air - reloads templates without server restart |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| net/http | gorilla/mux, chi, echo | No need - net/http routing sufficient for simple routes |
| html/template | templ, go-view | html/template preferred per project decisions |
| HTMX CDN | Self-hosted | Production reliability over CDN simplicity (CDN use discouraged in docs) |

**Installation:**
```bash
# Initialize Go module
go mod init github.com/esttorhe/blogwatcher-ui

# Add dependencies
go get modernc.org/sqlite@latest

# HTMX - download to static/htmx.min.js
curl -o static/htmx.min.js https://cdn.jsdelivr.net/npm/htmx.org@2.0.8/dist/htmx.min.js

# Development hot reload (optional)
go install github.com/air-verse/air@latest
```

## Architecture Patterns

### Recommended Project Structure
```
blogwatcher-ui/
├── cmd/
│   └── server/
│       └── main.go          # Entry point - minimal, calls run()
├── internal/
│   ├── server/
│   │   ├── server.go        # NewServer constructor, http.Handler
│   │   ├── routes.go        # All route registrations
│   │   └── handlers.go      # HTTP handlers
│   ├── storage/
│   │   └── database.go      # Database operations (can reuse from .reference)
│   └── model/
│       └── model.go         # Blog and Article structs (can reuse from .reference)
├── templates/
│   ├── base.gohtml          # Base layout (<!DOCTYPE>, <head>, <body>)
│   ├── partials/            # Reusable fragments for HTMX responses
│   │   ├── article-list.gohtml
│   │   └── blog-list.gohtml
│   └── pages/               # Full page templates
│       └── index.gohtml
├── static/
│   └── htmx.min.js          # Self-hosted HTMX
└── go.mod
```

### Pattern 1: NewServer Constructor (Dependency Injection)
**What:** Single constructor that takes all dependencies and returns http.Handler
**When to use:** Always - enables testability and explicit dependencies
**Example:**
```go
// Source: https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/
type Server struct {
    db     *storage.Database
    logger *log.Logger
    mux    *http.ServeMux
}

func NewServer(db *storage.Database, logger *log.Logger) http.Handler {
    s := &Server{
        db:     db,
        logger: logger,
        mux:    http.NewServeMux(),
    }
    s.registerRoutes()
    return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    s.mux.ServeHTTP(w, r)
}
```

### Pattern 2: Routes in Dedicated File
**What:** All route registrations in routes.go for API surface visibility
**When to use:** Always - single source of truth for all endpoints
**Example:**
```go
// Source: https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/
func (s *Server) registerRoutes() {
    // Static files
    s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    // Pages
    s.mux.HandleFunc("GET /", s.handleIndex)

    // HTMX endpoints (return HTML fragments)
    s.mux.HandleFunc("GET /articles", s.handleArticleList)
    s.mux.HandleFunc("GET /blogs", s.handleBlogList)
}
```

### Pattern 3: HTMX Request Detection
**What:** Check HX-Request header to return full page vs fragment
**When to use:** Handlers that serve both initial page load and HTMX updates
**Example:**
```go
// Source: https://htmx.org/docs/ + community patterns
func (s *Server) handleArticleList(w http.ResponseWriter, r *http.Request) {
    articles, err := s.db.ListArticles(false, nil)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    data := map[string]interface{}{
        "Articles": articles,
    }

    // If HTMX request, return partial; otherwise full page
    if r.Header.Get("HX-Request") == "true" {
        s.renderTemplate(w, "partials/article-list.gohtml", data)
    } else {
        s.renderTemplate(w, "pages/index.gohtml", data)
    }
}
```

### Pattern 4: Template Organization (Base + Partials)
**What:** Base template with shared layout, partials for HTMX responses
**When to use:** Always - prevents duplication, enables HTMX fragments
**Example:**
```go
// Source: https://medium.com/@uygaroztcyln/clean-ui-with-gos-html-templates-base-partials-and-funcmaps-4915296c9097
// Parse templates once at startup
templates := template.Must(template.ParseGlob("templates/**/*.gohtml"))

// Render with base layout
templates.ExecuteTemplate(w, "base.gohtml", data)

// Or render just a partial for HTMX
templates.ExecuteTemplate(w, "partials/article-list.gohtml", data)
```

### Pattern 5: Graceful Shutdown
**What:** Handle SIGTERM/SIGINT, shutdown with timeout for in-flight requests
**When to use:** Always - production requirement for Kubernetes, prevent data loss
**Example:**
```go
// Source: https://www.rudderstack.com/blog/implementing-graceful-shutdown-in-go/
func run(ctx context.Context) error {
    srv := &http.Server{
        Addr:    ":8080",
        Handler: NewServer(db, logger),
    }

    // Shutdown on context cancellation
    go func() {
        <-ctx.Done()
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        srv.Shutdown(shutdownCtx)
    }()

    return srv.ListenAndServe()
}

func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()

    if err := run(ctx); err != nil && err != http.ErrServerClosed {
        log.Fatal(err)
    }
}
```

### Pattern 6: Database Connection for SQLite
**What:** Single MaxOpenConn for SQLite's write constraints, WAL mode, busy_timeout
**When to use:** Always with SQLite - prevents SQLITE_BUSY errors
**Example:**
```go
// Source: https://turriate.com/articles/making-sqlite-faster-in-go
// Reference: .reference/blogwatcher/internal/storage/database.go
dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", path)
db, err := sql.Open("sqlite", dsn)
if err != nil {
    return nil, err
}

// SQLite single-writer constraint
db.SetMaxOpenConns(1)  // Critical for SQLite
db.SetMaxIdleConns(1)
db.SetConnMaxLifetime(0)  // No rotation needed for local file
```

### Anti-Patterns to Avoid
- **Global template variables:** Parse templates in NewServer, pass as dependency. Globals prevent testing and hot reload.
- **Template parsing on every request:** Parse once at startup (production). Only re-parse in development with hot reload.
- **Missing response.Body.Close():** Always defer resp.Body.Close() on HTTP client calls to prevent goroutine leaks.
- **No timeouts on http.Server:** Set ReadTimeout, WriteTimeout, IdleTimeout to prevent resource exhaustion.
- **Using typed template.HTML for user input:** Only use template.HTML for pre-sanitized content, never user-provided strings.

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Database connection pooling | Custom pool manager | database/sql built-in | Handles connection reuse, limits, lifetime automatically |
| XSS prevention | Manual escaping | html/template contextual escaping | Automatic, context-aware (HTML, JS, CSS, URL) |
| Request routing | String matching logic | http.ServeMux (Go 1.22+) | Method matching (`GET /path`), pattern matching built-in |
| Graceful shutdown | Manual goroutine tracking | http.Server.Shutdown | Handles in-flight requests, connection draining |
| Static file serving | Custom file reader | http.FileServer | Handles MIME types, ranges, caching headers |
| Signal handling | Raw os.Signal | signal.NotifyContext | Context integration, cleanup on cancel |
| Template hot reload (dev) | Custom file watcher | Air or gernest/hot | Battle-tested, watches multiple file types |

**Key insight:** Go's standard library is unusually comprehensive for web servers. Most problems already have stdlib solutions - only add dependencies when stdlib doesn't cover the use case.

## Common Pitfalls

### Pitfall 1: Goroutine Leaks from HTTP Clients
**What goes wrong:** Forgetting `resp.Body.Close()` causes goroutines to leak, holding file descriptors and connections indefinitely.
**Why it happens:** Go's HTTP client keeps connections alive for reuse, but unclosed bodies prevent cleanup.
**How to avoid:** Always `defer resp.Body.Close()` immediately after error check.
**Warning signs:** Increasing goroutine count (`runtime.NumGoroutine()`), open file descriptors, memory growth.

```go
// BAD
resp, err := http.Get(url)
if err != nil {
    return err
}
// Missing Close - goroutine leak!

// GOOD
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()
```

### Pitfall 2: SQLite SQLITE_BUSY Errors
**What goes wrong:** Multiple concurrent writes cause "database is locked" errors.
**Why it happens:** SQLite is single-writer, but database/sql connection pool creates multiple connections by default.
**How to avoid:** Set `db.SetMaxOpenConns(1)` and use `_pragma=busy_timeout(5000)` in DSN.
**Warning signs:** "database is locked", "SQLITE_BUSY", write errors under concurrent load.

### Pitfall 3: Template Type Confusion (template.HTML abuse)
**What goes wrong:** Using `template.HTML(userInput)` disables XSS protection, creating security hole.
**Why it happens:** Developers use template.HTML to bypass escaping without understanding consequences.
**How to avoid:** Never cast user input to template.HTML. Only use for pre-sanitized, trusted content.
**Warning signs:** `template.HTML()` calls with dynamic strings, user-provided data.

### Pitfall 4: Missing Server Timeouts
**What goes wrong:** Servers hang under load, holding goroutines and sockets indefinitely.
**Why it happens:** Default http.Server has no timeouts - connections stay open forever.
**How to avoid:** Set ReadTimeout, WriteTimeout, IdleTimeout on http.Server.
**Warning signs:** Increasing goroutine count under load, slow request handling, resource exhaustion.

```go
// BAD
srv := &http.Server{
    Addr:    ":8080",
    Handler: handler,
}

// GOOD
srv := &http.Server{
    Addr:         ":8080",
    Handler:      handler,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}
```

### Pitfall 5: Parsing Templates on Every Request
**What goes wrong:** High CPU usage, slow response times, template parse errors in production.
**Why it happens:** `template.ParseFiles()` called in handler instead of startup.
**How to avoid:** Parse templates once at startup with `template.Must()`, store in struct.
**Warning signs:** High CPU on template-heavy pages, slow response times, parse errors after deployment.

### Pitfall 6: Context Cancellation Not Respected
**What goes wrong:** Graceful shutdown hangs, requests don't cancel on client disconnect.
**Why it happens:** Handlers don't check `ctx.Done()` or pass context to dependencies.
**How to avoid:** Pass request context to database calls, check ctx.Err() in loops.
**Warning signs:** Long shutdown times, zombie requests after client disconnect.

## Code Examples

Verified patterns from official sources:

### Database Connection (SQLite-specific)
```go
// Source: .reference/blogwatcher/internal/storage/database.go
// Modified for production patterns from https://go.dev/doc/database/manage-connections
func OpenDatabase(path string) (*Database, error) {
    if path == "" {
        path = filepath.Join(os.Getenv("HOME"), ".blogwatcher", "blogwatcher.db")
    }

    // Check if database exists
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return nil, fmt.Errorf("database not found at %s - run blogwatcher CLI to initialize", path)
    }

    dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", path)
    conn, err := sql.Open("sqlite", dsn)
    if err != nil {
        return nil, err
    }

    // SQLite-specific connection pool settings
    conn.SetMaxOpenConns(1)           // Single writer
    conn.SetMaxIdleConns(1)           // Keep connection alive
    conn.SetConnMaxLifetime(0)        // No rotation for local file

    return &Database{path: path, conn: conn}, nil
}
```

### Template Setup and Rendering
```go
// Source: https://pkg.go.dev/html/template + patterns from search results
type Server struct {
    db        *storage.Database
    templates *template.Template
}

func NewServer(db *storage.Database) (http.Handler, error) {
    // Parse templates once at startup
    tmpl, err := template.ParseGlob("templates/**/*.gohtml")
    if err != nil {
        return nil, fmt.Errorf("failed to parse templates: %w", err)
    }

    s := &Server{
        db:        db,
        templates: tmpl,
    }

    mux := http.NewServeMux()
    s.registerRoutes(mux)
    return mux, nil
}

func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
    if err := s.templates.ExecuteTemplate(w, name, data); err != nil {
        log.Printf("Template error: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
    }
}
```

### HTMX Integration Handler
```go
// Source: HTMX docs + Go community patterns
func (s *Server) handleArticles(w http.ResponseWriter, r *http.Request) {
    // Fetch data from database
    articles, err := s.db.ListArticles(false, nil)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    data := map[string]interface{}{
        "Articles": articles,
        "Title":    "Articles",
    }

    // Check if HTMX request (fragment) or full page
    if r.Header.Get("HX-Request") == "true" {
        s.renderTemplate(w, "partials/article-list.gohtml", data)
    } else {
        s.renderTemplate(w, "pages/articles.gohtml", data)
    }
}
```

### Main Function with Graceful Shutdown
```go
// Source: https://www.rudderstack.com/blog/implementing-graceful-shutdown-in-go/
func run(ctx context.Context) error {
    // Setup database
    db, err := storage.OpenDatabase("")
    if err != nil {
        return err
    }
    defer db.Close()

    // Create server
    handler, err := NewServer(db)
    if err != nil {
        return err
    }

    srv := &http.Server{
        Addr:         ":8080",
        Handler:      handler,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    // Graceful shutdown
    go func() {
        <-ctx.Done()
        log.Println("Shutting down server...")
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        if err := srv.Shutdown(shutdownCtx); err != nil {
            log.Printf("Server shutdown error: %v", err)
        }
    }()

    log.Printf("Server starting on %s", srv.Addr)
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        return err
    }
    return nil
}

func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()

    if err := run(ctx); err != nil {
        log.Fatal(err)
    }
}
```

### Base Template Pattern
```html
<!-- Source: https://lets-go.alexedwards.net/sample/02.08-html-templating-and-inheritance.html -->
<!-- templates/base.gohtml -->
{{define "base"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{template "title" .}}</title>
    <script src="/static/htmx.min.js"></script>
</head>
<body>
    {{template "content" .}}
</body>
</html>
{{end}}

<!-- templates/pages/index.gohtml -->
{{define "title"}}BlogWatcher{{end}}
{{define "content"}}
<h1>Articles</h1>
<div id="article-list" hx-get="/articles" hx-trigger="load">
    {{template "partials/article-list.gohtml" .}}
</div>
{{end}}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| gorilla/mux for routing | net/http.ServeMux with method matching | Go 1.22 (Feb 2024) | Eliminate dependency, use `s.mux.HandleFunc("GET /path", handler)` |
| mattn/go-sqlite3 (CGo) | modernc.org/sqlite (pure Go) | 2020+ adoption | No CGo, easier cross-compilation, faster CI |
| Template parsing per request | Parse once with template.Must | Always best practice | 10x+ performance, catch errors at startup |
| Manual signal handling | signal.NotifyContext | Go 1.16+ | Context integration, cleaner shutdown |
| SPA frameworks (React) | HTMX hypermedia | 2023+ HTMX 2.0 | Simpler architecture, less JS, faster development |

**Deprecated/outdated:**
- **gorilla/mux:** Still works but unnecessary with Go 1.22+ net/http improvements
- **Server.ListenAndServe() without context:** Use signal.NotifyContext for graceful shutdown
- **Global http.DefaultServeMux:** Create own ServeMux for isolation and testing

## Open Questions

Things that couldn't be fully resolved:

1. **Hot Reload Tool Choice**
   - What we know: Air is popular, gernest/hot is lighter-weight
   - What's unclear: Which integrates better with html/template for this project
   - Recommendation: Start with Air (more features), can switch to gernest/hot if too heavy

2. **Template File Extension**
   - What we know: .tmpl, .gohtml, .tpl all used in community
   - What's unclear: Which has best IDE support for Esteban's environment
   - Recommendation: Use .gohtml (GoLand/VSCode support), document in project

3. **Port Configuration**
   - What we know: Default 8080 is standard, 3000 also common
   - What's unclear: User preference, conflicts with other services
   - Recommendation: Use 8080, make configurable via environment variable

## Sources

### Primary (HIGH confidence)
- [Go net/http package documentation](https://pkg.go.dev/net/http) - Official stdlib docs, January 2026
- [Go html/template package documentation](https://pkg.go.dev/html/template) - Official stdlib docs
- [Go database/sql connection management](https://go.dev/doc/database/manage-connections) - Official Go docs
- [HTMX 2.0.8 documentation](https://htmx.org/docs/) - Official HTMX docs, current version
- [HTMX reference](https://htmx.org/reference/) - Request/response headers
- [modernc.org/sqlite package](https://pkg.go.dev/modernc.org/sqlite) - Published January 20, 2026
- .reference/blogwatcher/ codebase - Existing implementation patterns

### Secondary (MEDIUM confidence)
- [How I write HTTP services in Go after 13 years](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/) - Mat Ryer, Grafana Labs, February 2024
- [Implementing Graceful Shutdown in Go](https://www.rudderstack.com/blog/implementing-graceful-shutdown-in-go/) - RudderStack
- [Making SQLite faster in Go](https://turriate.com/articles/making-sqlite-faster-in-go) - Connection pool patterns
- [Clean UI with Go's HTML Templates](https://medium.com/@uygaroztcyln/clean-ui-with-gos-html-templates-base-partials-and-funcmaps-4915296c9097) - Template organization
- [HTML templating and inheritance](https://lets-go.alexedwards.net/sample/02.08-html-templating-and-inheritance.html) - Alex Edwards

### Tertiary (LOW confidence)
- Various DEV.to articles on Go HTTP server structure (2024-2025) - Community patterns, cross-referenced
- Medium articles on HTMX integration with Go (2024) - Implementation examples, verified against official docs

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries well-documented in official sources, versions verified
- Architecture: HIGH - Patterns verified from Grafana engineering blog + official Go docs
- Pitfalls: HIGH - Documented in multiple authoritative sources, common knowledge in Go community

**Research date:** 2026-02-02
**Valid until:** 2026-03-04 (30 days - stable tech stack, Go 1.24 current)
