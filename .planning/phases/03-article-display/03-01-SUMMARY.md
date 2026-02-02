---
phase: 03-article-display
plan: 01
completed: 2026-02-02
duration: 11m
subsystem: backend-templates
tags: [go, templates, database, join, time-formatting]

dependency_graph:
  requires:
    - 01-01 (database layer)
    - 01-02 (server structure)
  provides:
    - ArticleWithBlog model for rich article display
    - timeAgo template function for relative timestamps
    - faviconURL template function for blog favicons
    - ListArticlesWithBlog database method with JOIN
  affects:
    - 03-02 (article card templates will use these functions)

tech_stack:
  added: []
  patterns:
    - template.FuncMap registration before parsing
    - SQL INNER JOIN for denormalized display data

files:
  created:
    - internal/server/template_funcs.go
  modified:
    - internal/model/model.go
    - internal/storage/database.go
    - internal/server/server.go
    - .gitignore

decisions:
  - id: funcmap-before-parse
    choice: Register FuncMap before ParseGlob calls
    rationale: Go templates require functions registered before parsing

metrics:
  tasks_completed: 3/3
  commits: 3
---

# Phase 3 Plan 01: Template Functions and Database Infrastructure Summary

**One-liner:** Added timeAgo/faviconURL template functions and ArticleWithBlog model with JOIN query for rich article card display.

## What Was Built

### 1. ArticleWithBlog Model (`internal/model/model.go`)

New struct extending Article with blog metadata:

```go
type ArticleWithBlog struct {
    ID, BlogID         int64
    Title, URL         string
    PublishedDate      *time.Time
    DiscoveredDate     *time.Time
    IsRead             bool
    BlogName           string  // From JOIN
    BlogURL            string  // For favicon
}
```

### 2. ListArticlesWithBlog Database Method (`internal/storage/database.go`)

New method using INNER JOIN to fetch articles with blog info:

```go
func (db *Database) ListArticlesWithBlog(isRead bool, blogID *int64) ([]model.ArticleWithBlog, error)
```

- Uses `INNER JOIN blogs b ON a.blog_id = b.id`
- Returns articles with BlogName and BlogURL populated
- Supports filtering by read status and optional blog ID

### 3. Template Functions (`internal/server/template_funcs.go`)

Two functions for article card rendering:

**timeAgo(t *time.Time) string**
- Converts timestamps to human-readable relative strings
- "just now", "7 minutes ago", "3 hours ago", "yesterday", "2 weeks ago", etc.
- Handles nil input (returns "") and future times

**faviconURL(blogURL string) string**
- Extracts domain from blog URL
- Returns Google S2 favicon API URL: `https://www.google.com/s2/favicons?domain={host}&sz=32`
- Handles invalid URLs gracefully

### 4. FuncMap Registration (`internal/server/server.go`)

Critical change: FuncMap registered BEFORE template parsing:

```go
funcMap := template.FuncMap{
    "timeAgo":    timeAgo,
    "faviconURL": faviconURL,
}
tmpl := template.New("").Funcs(funcMap)
// ParseGlob calls follow...
```

## Commits

| Hash | Description |
|------|-------------|
| a8e4015 | feat(03-01): add ArticleWithBlog model and ListArticlesWithBlog method |
| e2f7b52 | feat(03-01): add timeAgo and faviconURL template functions |
| c56c7e9 | feat(03-01): register FuncMap in server before template parsing |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed gitignore excluding internal/server**

- **Found during:** Task 2 (git add failed)
- **Issue:** `.gitignore` had bare `server` pattern matching `internal/server` directory
- **Fix:** Changed to `/server` to only match root-level binary
- **Files modified:** .gitignore
- **Commit:** e2f7b52

## Verification Results

- `go build ./...` - Passes
- `go vet ./...` - Passes
- Server starts without template parsing errors
- FuncMap registered before ParseGlob (line 30 vs lines 33/39/45)
- INNER JOIN query present in ListArticlesWithBlog

## Next Phase Readiness

**Ready for 03-02:** Article card templates can now use:
- `{{ timeAgo .DiscoveredDate }}` for relative timestamps
- `{{ faviconURL .BlogURL }}` for favicon images
- Handler updates to use ListArticlesWithBlog and ArticleWithBlog data
