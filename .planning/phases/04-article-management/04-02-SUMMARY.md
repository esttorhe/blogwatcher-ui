---
phase: 04-article-management
plan: 02
subsystem: article-management
tags: [htmx, post-handlers, mark-read, sync, bulk-actions, ui-buttons]
dependency-graph:
  requires: [04-01-scanner-infrastructure]
  provides: [mark-read-api, mark-unread-api, mark-all-read-api, sync-api, article-action-ui]
  affects: [05-theme-toggle]
tech-stack:
  added: []
  patterns: [htmx-post-actions, outerhtml-swap, fade-animation, loading-indicators]
key-files:
  created: []
  modified:
    - internal/storage/database.go
    - internal/server/handlers.go
    - internal/server/routes.go
    - templates/partials/article-list.gohtml
    - static/styles.css
decisions:
  - Return 200 OK with empty body for mark read/unread (HTMX removes card via outerHTML swap)
  - Return refreshed article list for mark all read and sync actions
  - Use single worker for sync to avoid SQLite write conflicts
  - Add hx-confirm for mark all read to prevent accidental bulk actions
patterns-established:
  - "HTMX outerHTML swap with empty response removes element"
  - "Fade animation via htmx-swapping class"
  - "Loading state via htmx-request class"
  - "Toolbar pattern for bulk actions above article list"
metrics:
  duration: ~4 minutes
  completed: 2026-02-02
---

# Phase 04 Plan 02: Article Management UI Summary

**Mark read/unread buttons on article cards, bulk mark all read, and sync button - all wired to POST handlers with HTMX**

## Performance

- **Duration:** ~4 minutes
- **Started:** 2026-02-02T23:00:00Z
- **Completed:** 2026-02-02T23:16:31Z
- **Tasks:** 3 (2 auto + 1 checkpoint)
- **Files modified:** 5

## Accomplishments

- Individual mark read/unread buttons on each article card with fade-out animation
- Bulk "Mark All Read" button in toolbar with confirmation dialog
- Sync button triggers scanner and refreshes article list
- All state changes persist to database and visible in CLI tool

## Task Commits

Each task was committed atomically:

1. **Task 1: Add bulk mark read database method and POST handlers** - `974670e` (feat)
2. **Task 2: Register POST routes and update templates** - `1a30d50` (feat)
3. **Task 3: Human verification checkpoint** - User approved (no commit needed)

## Files Created/Modified

- `internal/storage/database.go` - Added MarkAllUnreadArticlesRead method with optional blog filter
- `internal/server/handlers.go` - Added handleMarkRead, handleMarkUnread, handleMarkAllRead, handleSync handlers
- `internal/server/routes.go` - Registered POST routes for all article management actions
- `templates/partials/article-list.gohtml` - Added toolbar with bulk actions, action buttons on cards
- `static/styles.css` - Added toolbar styles, action button styles, loading states, fade animation

## Decisions Made

- **Empty response for single actions**: Mark read/unread returns 200 OK with empty body - HTMX removes the card via outerHTML swap
- **Full refresh for bulk actions**: Mark all read and sync return re-rendered article list partial
- **Single worker for sync**: Prevents SQLite write conflicts when scanning blogs
- **Confirmation for bulk actions**: Added hx-confirm to prevent accidental mark all read

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Article management complete (MGMT-01, MGMT-02, MGMT-03, MGMT-04 requirements satisfied)
- Phase 4 complete after this plan
- Ready for Phase 5: Theme Toggle

---
*Phase: 04-article-management*
*Completed: 2026-02-02*
