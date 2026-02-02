---
phase: 02-ui-layout-navigation
plan: 02
subsystem: ui
tags: [htmx, navigation, filtering, templates, go-templates]

# Dependency graph
requires:
  - phase: 02-01
    provides: Sidebar layout, dark theme CSS, hamburger menu
  - phase: 01-03
    provides: Database layer with ListArticles, handlers with HTMX detection
provides:
  - HTMX navigation for sidebar with filter and blog query params
  - ListArticlesByReadStatus database method for explicit read status filtering
  - Active state highlighting for current filter/blog in sidebar
  - Mobile sidebar auto-close on navigation
affects: [03-article-display, 04-article-management]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "HTMX hx-on::after-swap for mobile sidebar close"
    - "CurrentBlogID as int64 (0 = no filter) for template comparison"

key-files:
  created: []
  modified:
    - internal/server/handlers.go
    - internal/storage/database.go
    - templates/partials/sidebar.gohtml
    - templates/partials/blog-list.gohtml
    - templates/partials/article-list.gohtml
    - templates/pages/index.gohtml

key-decisions:
  - "Use ListArticlesByReadStatus for explicit read status filtering instead of modifying ListArticles"
  - "Pass CurrentBlogID as int64 (0 = no filter) instead of pointer for simpler template comparison"
  - "Include h1 title in article-list partial so HTMX swaps update the heading"

patterns-established:
  - "Filter params pattern: r.URL.Query().Get('filter') and r.URL.Query().Get('blog')"
  - "Active state in templates: {{if eq .CurrentFilter 'read'}} active{{end}}"

# Metrics
duration: 4min
completed: 2026-02-02
---

# Phase 2 Plan 02: HTMX Navigation Summary

**HTMX-powered sidebar navigation with filter (unread/read) and blog query params, active state highlighting, and mobile auto-close**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-02T22:01:44Z
- **Completed:** 2026-02-02T22:06:13Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments
- Clicking Inbox shows unread articles, Archived shows read articles
- Clicking a blog filters articles to that blog only
- URL updates on navigation (hx-push-url), browser back/forward works
- Active nav item is visually highlighted with accent color
- Mobile sidebar closes automatically after navigation

## Task Commits

Each task was committed atomically:

1. **Task 1: Update handlers to support filter and blog query params** - `4fc2fd6` (feat)
2. **Task 2: Add ListArticlesByReadStatus to database layer** - `95622da` (feat)
3. **Task 3: Add HTMX attributes to sidebar and blog list templates** - `ea858af` (feat)

## Files Created/Modified
- `internal/server/handlers.go` - Parse filter/blog query params, pass CurrentFilter/CurrentBlogID to templates
- `internal/storage/database.go` - New ListArticlesByReadStatus method for explicit read status filtering
- `templates/partials/sidebar.gohtml` - HTMX attributes (hx-get, hx-target, hx-push-url), active state classes
- `templates/partials/blog-list.gohtml` - HTMX navigation for blog filtering with active states
- `templates/partials/article-list.gohtml` - Include h1 title for HTMX swaps, ABOUTME comments
- `templates/pages/index.gohtml` - Remove duplicate h1 (now in article-list partial)

## Decisions Made
- **ListArticlesByReadStatus vs modifying ListArticles:** Created new method to avoid changing existing API. ListArticles(unreadOnly) had different semantics (unreadOnly=false shows ALL articles). New method is explicit: isRead=true returns read, isRead=false returns unread.
- **CurrentBlogID as int64:** Go templates can't easily dereference pointers. Using int64 with 0 meaning "no filter" works cleanly since database IDs start at 1.
- **h1 in article-list partial:** When HTMX swaps main-content, we need the title to update too. Moving h1 into the partial ensures consistent behavior.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Navigation fully functional with HTMX
- Ready for Phase 3 (Article Display) - article cards, metadata, date formatting
- Ready for Phase 4 (Article Management) - mark read/unread actions

---
*Phase: 02-ui-layout-navigation*
*Completed: 2026-02-02*
