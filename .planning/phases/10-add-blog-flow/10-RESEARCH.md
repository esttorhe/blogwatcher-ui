# Phase 10: Add Blog Flow - Research

**Researched:** 2026-02-09
**Domain:** Go os/exec, HTMX forms, FAB UI patterns
**Confidence:** HIGH

## Summary

Phase 10 adds the ability to add new blogs via the web UI by calling the blogwatcher CLI tool using Go's `os/exec` package. The standard approach involves:

1. **CLI Integration**: Use `exec.CommandContext` with timeout to call `blogwatcher add <name> <url>`
2. **HTMX Form Pattern**: POST form that returns either success confirmation or error message HTML
3. **FAB Component**: Fixed-position button using `position: fixed` with existing `.btn-action` styling
4. **Auto-sync After Add**: Reuse existing `handleSync` to fetch articles immediately after adding blog

The blogwatcher CLI already provides all feed discovery logic - the UI simply needs to invoke it and handle success/error responses. The CLI returns exit code 0 on success with "Added blog 'Name'" message, and exit code 1 on error with "Error: <message>" output to stderr.

**Primary recommendation:** Use `exec.CommandContext` with 30-second timeout, capture both stdout and stderr, return HTMX-friendly HTML fragments showing success/error state, then trigger sync on success.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| os/exec | stdlib (Go 1.22+) | Execute external commands | Standard library package for process execution, no dependencies needed |
| HTMX | 1.x (existing) | Form submission and swapping | Already used throughout codebase for all dynamic interactions |
| context | stdlib | Timeout management | Standard way to cancel long-running operations in Go |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| strings | stdlib | Parse CLI output | Extract success/error messages from command output |
| bytes | stdlib | Capture command output | Use Buffer to collect stdout/stderr |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| exec.Command | exec.CommandContext | CommandContext provides timeout/cancellation, always prefer it |
| Direct database writes | CLI invocation | CLI already has feed discovery logic - don't duplicate |
| JavaScript validation | Server-side only | Keep logic server-side, HTMX handles UI updates |

**Installation:**
No new dependencies required - all stdlib packages.

## Architecture Patterns

### Recommended Project Structure
```
internal/server/
├── handlers.go        # Add handleAddBlog handler
├── routes.go          # Register POST /blogs/add route
assets/templates/partials/
├── settings-page.gohtml  # Add form to this existing template
assets/static/
└── styles.css         # Add FAB and form styles
```

### Pattern 1: CLI Invocation with Timeout
**What:** Execute blogwatcher CLI with context-based timeout
**When to use:** Any time you need to call external command from Go
**Example:**
```go
// Source: https://pkg.go.dev/os/exec
func (s *Server) handleAddBlog(w http.ResponseWriter, r *http.Request) {
    name := r.FormValue("name")
    url := r.FormValue("url")

    // Create command with 30-second timeout
    ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, "blogwatcher", "add", name, url)

    // Capture both stdout and stderr
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()
    if err != nil {
        // Parse error from stderr: "Error: Blog with name 'X' already exists"
        errorMsg := strings.TrimSpace(stderr.String())
        // Return error HTML fragment
        renderAddBlogError(w, errorMsg)
        return
    }

    // Success - trigger sync for new blog and return success message
    renderAddBlogSuccess(w, name)
}
```

### Pattern 2: HTMX Form Submission with Error Feedback
**What:** Form posts to server, returns HTML fragment showing success/error state
**When to use:** Any form that needs inline validation/feedback without page reload
**Example:**
```html
<!-- Source: https://htmx.org/examples/inline-validation/ -->
<div id="add-blog-form-container">
    <form hx-post="/blogs/add"
          hx-target="#add-blog-form-container"
          hx-swap="innerHTML">
        <label>
            Blog Name:
            <input type="text" name="name" required>
        </label>
        <label>
            Blog URL:
            <input type="url" name="url" required>
        </label>
        <button type="submit" class="btn-action">Add Blog</button>
    </form>
</div>
```

**Server returns on error:**
```html
<div id="add-blog-form-container" class="form-error">
    <p class="error-message">Error: Blog with name 'X' already exists</p>
    <form hx-post="/blogs/add" hx-target="#add-blog-form-container" hx-swap="innerHTML">
        <!-- Same form fields, pre-populated with submitted values -->
    </form>
</div>
```

**Server returns on success:**
```html
<div id="add-blog-form-container" class="form-success">
    <p class="success-message">Successfully added 'Blog Name'. Syncing articles...</p>
    <button hx-get="/blogs/add/form"
            hx-target="#add-blog-form-container"
            hx-swap="innerHTML"
            class="btn-action">Add Another Blog</button>
</div>
```

### Pattern 3: Floating Action Button (FAB)
**What:** Fixed-position button that floats above content, typically bottom-right corner
**When to use:** Quick access to primary action from any scroll position
**Example:**
```css
/* Source: https://web.dev/articles/building/a-fab-component */
.fab {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  z-index: 1000; /* Above other content */

  /* Reuse existing button styles */
  padding: 1rem;
  border: 1px solid var(--border);
  border-radius: 50%; /* Circular */
  background-color: var(--accent);
  color: var(--bg-primary);
  cursor: pointer;

  /* Drop shadow for floating effect */
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.fab:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 12px rgba(0, 0, 0, 0.3);
}

/* Mobile: adjust position for smaller screens */
@media (max-width: 768px) {
  .fab {
    bottom: 1rem;
    right: 1rem;
  }
}
```

```html
<!-- FAB triggers navigation to settings page with add form -->
<a href="/settings#add-blog"
   class="fab"
   title="Add Blog"
   aria-label="Add new blog">
  <svg><!-- Plus icon --></svg>
</a>
```

### Pattern 4: Auto-Sync After Add
**What:** Reuse existing sync handler to fetch articles immediately after adding blog
**When to use:** After any operation that adds new data sources
**Example:**
```go
// In handleAddBlog after successful CLI execution:
func (s *Server) renderAddBlogSuccess(w http.ResponseWriter, name string) {
    // Trigger sync for the newly added blog
    results, err := scanner.ScanAllBlogs(s.db, 1)
    if err != nil {
        log.Printf("Auto-sync failed after adding %s: %v", name, err)
    } else {
        // Log sync results
        for _, result := range results {
            if result.BlogName == name {
                log.Printf("Auto-synced %s: %d new articles", name, result.NewArticles)
            }
        }
    }

    // Return success message with article count if available
    data := map[string]interface{}{
        "BlogName": name,
        "Success":  true,
    }
    s.renderTemplate(w, "add-blog-success.gohtml", data)
}
```

### Anti-Patterns to Avoid
- **Don't duplicate feed discovery logic**: The CLI already handles RSS/Atom auto-discovery via `--feed-url` flag. Don't try to replicate this in the UI.
- **Don't use exec.Command without context**: Always use `CommandContext` with timeout to prevent hung requests.
- **Don't ignore stderr**: Error messages come through stderr, not error type alone.
- **Don't shell out with bash -c**: Use `exec.Command` directly - it's safer and doesn't invoke shell.
- **Don't store command output in error messages directly**: Parse and clean CLI output before showing to users.

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| RSS/Atom feed discovery | Custom HTTP fetcher + parser | blogwatcher CLI `add` command | CLI already handles auto-discovery, multiple feed formats, error cases |
| Command timeouts | Manual goroutine + timer | `context.WithTimeout` + `CommandContext` | Context cancellation is stdlib pattern, handles cleanup properly |
| Form validation | JavaScript client-side checks | HTML5 `required` + server validation | HTMX pattern is server-driven, ensures validation even with JS disabled |
| Success/error UI state | JavaScript DOM manipulation | HTMX `hx-target` + `hx-swap` | HTMX handles swapping HTML fragments, no custom JS needed |

**Key insight:** The blogwatcher CLI is the source of truth for blog operations. The UI is a thin wrapper that invokes CLI commands and presents results. Don't reimplement CLI logic in the UI.

## Common Pitfalls

### Pitfall 1: Command Path Assumptions
**What goes wrong:** Assuming `blogwatcher` is in PATH, command fails in production
**Why it happens:** Development environment has different PATH than production/systemd service
**How to avoid:**
- Option 1: Use full path `/Users/esteban.torres/workspace/bin/blogwatcher` (check where installed)
- Option 2: Use `exec.LookPath("blogwatcher")` to find it dynamically
- Option 3: Make blogwatcher path configurable via environment variable
**Warning signs:** Command works in dev terminal but fails when server runs as service

### Pitfall 2: Not Capturing stderr for Errors
**What goes wrong:** User sees generic "command failed" instead of actual error like "Blog already exists"
**Why it happens:** Only checking `err` from `cmd.Run()`, not reading stderr buffer
**How to avoid:** Always set `cmd.Stderr = &stderr` and include `stderr.String()` in error messages
**Warning signs:** Error messages are vague, actual CLI errors not visible to users

### Pitfall 3: Blocking on Long-Running Commands
**What goes wrong:** HTTP request times out, user sees blank page or loading forever
**Why it happens:** Using `exec.Command` without timeout, feed discovery can take 30+ seconds
**How to avoid:** Always use `CommandContext` with reasonable timeout (30s for add command)
**Warning signs:** Some URLs work fine, others hang indefinitely

### Pitfall 4: Not Escaping User Input
**What goes wrong:** Blog name with quotes or special chars breaks command, or worse - command injection
**Why it happens:** Passing user input directly to command args without validation
**How to avoid:** `exec.Command` args are already safe (no shell parsing), but validate name/URL format first
**Warning signs:** Blog names with spaces, quotes, or special characters fail mysteriously

### Pitfall 5: Form State Lost on Error
**What goes wrong:** User types name + URL, hits submit, gets error, form is empty - has to retype
**Why it happens:** Error response returns fresh empty form instead of pre-populating with submitted values
**How to avoid:** Pass `r.FormValue("name")` and `r.FormValue("url")` back in error template data
**Warning signs:** Users complain about having to retype after validation errors

### Pitfall 6: No Loading Indicator
**What goes wrong:** User clicks "Add Blog", nothing happens for 30 seconds, clicks again (duplicate submit)
**Why it happens:** Not using HTMX `hx-indicator` to show progress during CLI execution
**How to avoid:** Add `hx-indicator=".add-blog-spinner"` and loading message element
**Warning signs:** Users report "nothing happening" or accidentally submitting twice

## Code Examples

Verified patterns from official sources:

### Complete Handler with Error Handling
```go
// Based on patterns from internal/server/handlers.go
func (s *Server) handleAddBlog(w http.ResponseWriter, r *http.Request) {
    // Parse form values
    name := strings.TrimSpace(r.FormValue("name"))
    url := strings.TrimSpace(r.FormValue("url"))

    // Basic validation
    if name == "" || url == "" {
        s.renderAddBlogError(w, "Blog name and URL are required", name, url)
        return
    }

    // Create command with timeout
    ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
    defer cancel()

    // Find blogwatcher command (prefer explicit path)
    blogwatcherPath, err := exec.LookPath("blogwatcher")
    if err != nil {
        s.renderAddBlogError(w, "blogwatcher CLI not found", name, url)
        return
    }

    cmd := exec.CommandContext(ctx, blogwatcherPath, "add", name, url)

    // Capture both stdout and stderr
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    // Execute command
    err = cmd.Run()
    if err != nil {
        // Extract error message from stderr
        // Format: "Error: Blog with name 'X' already exists"
        errorMsg := strings.TrimPrefix(strings.TrimSpace(stderr.String()), "Error: ")
        if errorMsg == "" {
            errorMsg = "Failed to add blog: " + err.Error()
        }
        s.renderAddBlogError(w, errorMsg, name, url)
        return
    }

    // Success - auto-sync the new blog
    log.Printf("Added blog '%s' at %s", name, url)
    go s.autoSyncNewBlog(name) // Don't block response on sync

    // Return success message
    s.renderAddBlogSuccess(w, name)
}

func (s *Server) autoSyncNewBlog(blogName string) {
    result, err := scanner.ScanBlogByName(s.db, blogName)
    if err != nil {
        log.Printf("Auto-sync failed for %s: %v", blogName, err)
        return
    }
    if result != nil {
        log.Printf("Auto-synced %s: %d new articles", blogName, result.NewArticles)
    }
}

func (s *Server) renderAddBlogError(w http.ResponseWriter, message, name, url string) {
    data := map[string]interface{}{
        "Error":   message,
        "Name":    name, // Pre-populate form
        "URL":     url,  // Pre-populate form
    }
    s.renderTemplate(w, "add-blog-form.gohtml", data)
}

func (s *Server) renderAddBlogSuccess(w http.ResponseWriter, name string) {
    data := map[string]interface{}{
        "Success":  true,
        "BlogName": name,
    }
    s.renderTemplate(w, "add-blog-form.gohtml", data)
}
```

### Add Blog Form Template
```html
{{define "add-blog-form.gohtml"}}
<div id="add-blog-form-container" class="add-blog-section">
    {{if .Success}}
    <div class="success-message">
        <p>✓ Successfully added '{{.BlogName}}'. Fetching articles...</p>
        <button hx-get="/settings"
                hx-target="#main-content"
                hx-swap="innerHTML"
                class="btn-action">Back to Settings</button>
    </div>
    {{else}}
    <h2>Add New Blog</h2>
    {{if .Error}}
    <div class="error-message">
        <p>{{.Error}}</p>
    </div>
    {{end}}
    <form hx-post="/blogs/add"
          hx-target="#add-blog-form-container"
          hx-swap="innerHTML"
          hx-indicator=".add-blog-spinner">
        <div class="form-group">
            <label for="blog-name">Blog Name</label>
            <input type="text"
                   id="blog-name"
                   name="name"
                   value="{{.Name}}"
                   placeholder="My Favorite Blog"
                   required>
        </div>
        <div class="form-group">
            <label for="blog-url">Blog URL</label>
            <input type="url"
                   id="blog-url"
                   name="url"
                   value="{{.URL}}"
                   placeholder="https://example.com"
                   required>
            <small>RSS/Atom feed will be auto-discovered</small>
        </div>
        <div class="form-actions">
            <button type="submit" class="btn-action">
                <span class="btn-text">Add Blog</span>
                <span class="add-blog-spinner htmx-indicator">Adding...</span>
            </button>
        </div>
    </form>
    {{end}}
</div>
{{end}}
```

### FAB Button (Added to Settings Page)
```html
<!-- In settings-page.gohtml or index.gohtml -->
<a href="/settings#add-blog"
   class="fab"
   title="Add New Blog"
   aria-label="Add new blog"
   hx-boost="true">
    <svg xmlns="http://www.w3.org/2000/svg"
         width="24"
         height="24"
         viewBox="0 0 24 24"
         fill="none"
         stroke="currentColor"
         stroke-width="2">
        <line x1="12" y1="5" x2="12" y2="19"/>
        <line x1="5" y1="12" x2="19" y2="12"/>
    </svg>
</a>
```

### CSS Additions
```css
/* Add Blog Form Styles */
.add-blog-section {
    padding: 2rem;
    max-width: 600px;
}

.add-blog-section h2 {
    font-size: 1.25rem;
    color: var(--text-secondary);
    font-weight: 500;
    margin-bottom: 1.5rem;
}

.form-group {
    margin-bottom: 1.5rem;
}

.form-group label {
    display: block;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text-primary);
    margin-bottom: 0.5rem;
}

.form-group input {
    width: 100%;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background-color: var(--bg-surface);
    color: var(--text-primary);
    font-size: 0.875rem;
}

.form-group input:focus {
    outline: none;
    border-color: var(--accent);
    box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

.form-group small {
    display: block;
    font-size: 0.75rem;
    color: var(--text-secondary);
    margin-top: 0.25rem;
}

.error-message {
    padding: 0.75rem 1rem;
    background-color: rgba(220, 38, 38, 0.1);
    border: 1px solid rgba(220, 38, 38, 0.3);
    border-radius: 6px;
    color: #dc2626;
    margin-bottom: 1rem;
}

.success-message {
    padding: 0.75rem 1rem;
    background-color: rgba(34, 197, 94, 0.1);
    border: 1px solid rgba(34, 197, 94, 0.3);
    border-radius: 6px;
    color: #22c55e;
    margin-bottom: 1rem;
}

.form-actions {
    display: flex;
    gap: 0.5rem;
}

/* Floating Action Button */
.fab {
    position: fixed;
    bottom: 2rem;
    right: 2rem;
    z-index: 1000;

    width: 56px;
    height: 56px;
    display: flex;
    align-items: center;
    justify-content: center;

    border: none;
    border-radius: 50%;
    background-color: var(--accent);
    color: white;
    cursor: pointer;

    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
    transition: transform 0.2s ease, box-shadow 0.2s ease;
    text-decoration: none;
}

.fab:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 12px rgba(0, 0, 0, 0.3);
    text-decoration: none;
}

.fab svg {
    width: 24px;
    height: 24px;
}

@media (max-width: 768px) {
    .fab {
        bottom: 1rem;
        right: 1rem;
        width: 48px;
        height: 48px;
    }

    .fab svg {
        width: 20px;
        height: 20px;
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Shell scripts with bash -c | exec.Command with direct args | Go 1.0+ | Safer - no shell injection, clearer argument passing |
| Manual timeout goroutines | context.WithTimeout + CommandContext | Go 1.7 (2016) | Cleaner cancellation, stdlib pattern |
| JavaScript-heavy forms | HTMX declarative attributes | HTMX 1.0 (2020) | Less JavaScript, server-driven UI |
| Fixed #hash anchors for FAB | JavaScript scroll listeners | Modern CSS (2020+) | position: fixed is widely supported, simpler |

**Deprecated/outdated:**
- `exec.Command` without context: Use `CommandContext` for timeout support (since Go 1.7)
- jQuery form plugins: HTMX provides better server-integration patterns
- Material Design Lite FAB: Pure CSS with custom properties matches existing theme system

## Open Questions

Things that couldn't be fully resolved:

1. **Where is blogwatcher installed in production?**
   - What we know: It's at `/Users/esteban.torres/workspace/bin/blogwatcher` in dev
   - What's unclear: Production path might differ, systemd service PATH might not include it
   - Recommendation: Make path configurable via env var `BLOGWATCHER_CLI_PATH`, fall back to `exec.LookPath`

2. **Should FAB be on all pages or just settings?**
   - What we know: Phase requirements say "User can access quick add via floating action button (FAB)"
   - What's unclear: Whether FAB should appear on article list page or only settings page
   - Recommendation: Add to settings page only for v1.2, can expand to all pages later if needed

3. **How to handle feed discovery timeout?**
   - What we know: CLI auto-discovers RSS/Atom feeds, can take 10-30 seconds for slow sites
   - What's unclear: What's an acceptable timeout before user sees "taking too long" message
   - Recommendation: 30-second timeout with clear loading indicator, matches existing sync pattern

## Sources

### Primary (HIGH confidence)
- [Go os/exec Package Documentation](https://pkg.go.dev/os/exec) - Official stdlib docs
- [HTMX Inline Validation Example](https://htmx.org/examples/inline-validation/) - Official pattern
- Internal codebase: `internal/server/handlers.go` - Existing HTMX handler patterns
- Internal codebase: `internal/scanner/scanner.go` - Scanner integration patterns
- Internal codebase: `assets/static/styles.css` - Existing button and form styling
- blogwatcher CLI testing - Live command behavior and output format

### Secondary (MEDIUM confidence)
- [Advanced command execution in Go with os/exec](https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html)
- [DoltHub Go os/exec Patterns](https://www.dolthub.com/blog/2022-11-28-go-os-exec-patterns/)
- [web.dev FAB Component Guide](https://web.dev/articles/building/a-fab-component)
- [DEV Community: Handling form errors in htmx](https://dev.to/yawaramin/handling-form-errors-in-htmx-3ncg)

### Tertiary (LOW confidence)
- [FreeFrontend CSS FAB Examples](https://freefrontend.com/css-floating-action-buttons/) - Visual inspiration only
- [Materialize CSS FAB](https://materializecss.com/floating-action-button.html) - Reference for patterns, not using framework

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - stdlib only, no version compatibility issues
- Architecture: HIGH - patterns verified in existing codebase and official docs
- Pitfalls: HIGH - identified through CLI testing and Go stdlib documentation

**Research date:** 2026-02-09
**Valid until:** 60 days (stable domain - Go stdlib and HTMX patterns don't change frequently)
