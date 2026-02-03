# BlogWatcher UI

## What This Is

A web-based reader UI for the existing blogwatcher CLI tool. It provides an Omnivore-style dark/light interface to browse, read, and manage blog articles tracked by blogwatcher. Single-user, self-hosted, accessed via browser on desktop or mobile.

## Core Value

Read and manage blog articles through a clean, responsive web interface without touching the CLI.

## Current Milestone: v1.1 UI Polish & Search

**Goal:** Improve visual presentation with masonry layout and thumbnails, add search and filtering capabilities.

**Target features:**
- Masonry layout (grid alternative to current list view)
- Clickable cards (entire card opens article, not just title)
- Thumbnail support (RSS → Open Graph → favicon fallback)
- Title search
- Date filtering (last week, month, custom range)
- Combined filters (blog + status + search together)

## Requirements

### Validated

Shipped in v1.0:

- ✓ **INFRA-01**: Go HTTP server serving web UI — v1.0
- ✓ **INFRA-02**: Connect to existing blogwatcher SQLite database — v1.0
- ✓ **INFRA-03**: HTMX for dynamic updates without full page reloads — v1.0
- ✓ **UI-01**: Responsive layout with collapsible sidebar on mobile — v1.0
- ✓ **UI-02**: Filter views in sidebar (Inbox/unread, Archived/read) — v1.0
- ✓ **UI-03**: Subscriptions list in sidebar showing tracked blogs — v1.0
- ✓ **UI-04**: Dark/light theme toggle — v1.0
- ✓ **DISP-01**: Article cards show thumbnail or site favicon — v1.0
- ✓ **DISP-02**: Article cards show time ago ("7 hours ago") — v1.0
- ✓ **DISP-03**: Article cards show title and source blog name — v1.0
- ✓ **DISP-04**: Clicking article opens original URL in new tab — v1.0
- ✓ **MGMT-01**: Button to mark individual article as read — v1.0
- ✓ **MGMT-02**: Button to mark article as unread — v1.0
- ✓ **MGMT-03**: "Mark all read" button for bulk action — v1.0
- ✓ **MGMT-04**: Manual sync button to scan blogs for new articles — v1.0

### Active

v1.1 scope:

- [ ] Masonry grid layout as alternative to list view
- [ ] Entire article card clickable (not just title)
- [ ] Article thumbnails with fallback chain (RSS → OG → favicon)
- [ ] Search articles by title
- [ ] Date filtering (last week, month, custom)
- [ ] Combined filters (blog + status + search + date)

### Out of Scope

- User authentication/multi-user — single user, local access
- Labels/tags — not needed yet
- In-app reader view — just link to originals
- Auto-sync/background refresh — manual only
- Full-text search — would require fetching/storing article content
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

**v1.0 architecture:** Go server with HTMX, CSS custom properties for theming, scanner packages (RSS, scraper).

## Constraints

- **Tech stack:** Go server with templates + HTMX — server-rendered, minimal JS
- **Database:** Must use existing SQLite database and schema (may add thumbnail_url column if needed)
- **Deployment:** Single binary that serves the web UI
- **Compatibility:** Share database with CLI tool — both can coexist

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go templates + HTMX | Go-native, minimal JS, matches reference codebase style | ✓ Good |
| Share CLI database | Single source of truth, CLI and UI coexist | ✓ Good |
| No in-app reader | Simpler, just link to originals, avoids content fetching complexity | ✓ Good |
| Manual sync only | Keeps it simple, user controls when to refresh | ✓ Good |
| Three-way theme toggle | Light/Dark/System with CSS :has() and localStorage | ✓ Good |

---
*Last updated: 2026-02-03 after v1.1 milestone start*
