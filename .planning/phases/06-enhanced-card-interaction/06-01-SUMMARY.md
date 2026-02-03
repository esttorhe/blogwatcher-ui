---
phase: 06-enhanced-card-interaction
plan: 01
subsystem: database
tags: [sqlite, thumbnail, opengraph, gofeed, migration]

# Dependency graph
requires:
  - phase: 01-infrastructure-setup
    provides: "SQLite database schema with articles table"
provides:
  - "thumbnail_url column in articles table (nullable TEXT)"
  - "ThumbnailURL field on Article and ArticleWithBlog models"
  - "Thumbnail extraction package with RSS and Open Graph support"
affects: [06-02-scanner-integration, 06-03-ui-cards]

# Tech tracking
tech-stack:
  added: [github.com/otiai10/opengraph/v2]
  patterns: [idempotent-migrations, nullable-scan-pattern]

key-files:
  created:
    - internal/thumbnail/thumbnail.go
  modified:
    - internal/model/model.go
    - internal/storage/database.go
    - go.mod
    - go.sum

key-decisions:
  - "Use ADD COLUMN IF NOT EXISTS for idempotent migration"
  - "Run migrations on database open (ensureMigrations) not CLI"
  - "Store thumbnail_url as nullable TEXT (empty string if no thumbnail)"

patterns-established:
  - "ensureMigrations: Run idempotent schema migrations on database open"
  - "sql.NullString scan: Convert to string field (empty string for NULL)"

# Metrics
duration: 3min
completed: 2026-02-03
---

# Phase 6 Plan 1: Thumbnail Infrastructure Summary

**Database schema migration for thumbnail_url column with model updates and thumbnail extraction package using opengraph/v2**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-03T12:51:42Z
- **Completed:** 2026-02-03T12:54:38Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments
- Added thumbnail_url column to articles table via idempotent migration
- Updated Article and ArticleWithBlog models with ThumbnailURL field
- Created internal/thumbnail package with RSS and Open Graph extraction functions
- All database queries now include and properly scan thumbnail_url

## Task Commits

Each task was committed atomically:

1. **Task 1: Schema migration and model updates** - `a20130c` (feat)
2. **Task 2: Create thumbnail extraction package** - `192b2e3` (feat)
3. **Task 3: Verify database queries return thumbnail** - verification only, no changes needed

## Files Created/Modified
- `internal/model/model.go` - Added ThumbnailURL field to Article and ArticleWithBlog structs
- `internal/storage/database.go` - Added ensureMigrations(), updated all queries for thumbnail_url
- `internal/thumbnail/thumbnail.go` - New package for thumbnail URL extraction
- `go.mod` / `go.sum` - Added opengraph/v2 dependency

## Decisions Made
- **Idempotent migration approach:** Using `ADD COLUMN IF NOT EXISTS` allows migration to run safely on every database open without tracking migration state
- **ensureMigrations on open:** Running migrations on database open keeps migration logic centralized and automatic
- **Nullable to empty string:** Converting sql.NullString to empty string in model fields simplifies template logic (just check if empty)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - all implementations followed the plan specification.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Schema ready for scanner to populate thumbnail_url during sync
- Model fields ready for template rendering
- Thumbnail extraction functions ready for scanner integration in Plan 06-02
- All existing functionality continues to work (thumbnail defaults to empty string)

---
*Phase: 06-enhanced-card-interaction*
*Completed: 2026-02-03*
