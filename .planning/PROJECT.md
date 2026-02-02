# BlogWatcher UI

## What This Is

A web-based reader UI for the existing blogwatcher CLI tool. It provides an Omnivore-style dark interface to browse, read, and manage blog articles tracked by blogwatcher. Single-user, self-hosted, accessed via browser on desktop or mobile.

## Core Value

Read and manage blog articles through a clean, responsive web interface without touching the CLI.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Dark theme UI matching Omnivore aesthetic
- [ ] Sidebar with filter views (Inbox/unread, Archived/read)
- [ ] Sidebar showing subscriptions (tracked blogs) for filtering
- [ ] Article list with cards showing thumbnail/favicon, time ago, title, source blog
- [ ] Click article opens original URL in new tab
- [ ] Checkbox/button to mark individual articles as read
- [ ] "Mark all read" button for bulk action
- [ ] Manual sync button to trigger blog scanning
- [ ] Responsive layout (collapsible sidebar on mobile)
- [ ] Connects to existing blogwatcher SQLite database

### Out of Scope

- User authentication/multi-user — single user, local access
- Labels/tags — not needed for v1
- In-app reader view — just link to originals
- Auto-sync/background refresh — manual only
- Article search — not needed for v1
- Read time estimates — not in current database
- Blog management (add/remove) — use CLI for that

## Context

**Reference codebase:** `.reference/blogwatcher/` contains the Go CLI tool that:
- Tracks blogs via RSS/Atom feeds or HTML scraping
- Stores data in SQLite at `~/.blogwatcher/blogwatcher.db`
- Has `blogs` table (id, name, url, feed_url, scrape_selector, last_scanned)
- Has `articles` table (id, blog_id, title, url, published_date, discovered_date, is_read)
- Provides scanning, read/unread management via CLI

**Database location:** `~/.blogwatcher/blogwatcher.db` (shared with CLI)

**Existing patterns:** The reference code uses modernc.org/sqlite, clean Go patterns, tested storage layer.

## Constraints

- **Tech stack:** Go server with templates + HTMX — server-rendered, minimal JS
- **Database:** Must use existing SQLite database and schema (no modifications)
- **Deployment:** Single binary that serves the web UI
- **Compatibility:** Share database with CLI tool — both can coexist

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go templates + HTMX | Go-native, minimal JS, matches reference codebase style | — Pending |
| Share CLI database | Single source of truth, CLI and UI coexist | — Pending |
| No in-app reader | Simpler, just link to originals, avoids content fetching complexity | — Pending |
| Manual sync only | Keeps it simple, user controls when to refresh | — Pending |

---
*Last updated: 2026-02-02 after initialization*
