# BlogWatcher UI

## What This Is

A web-based reader UI for the existing blogwatcher CLI tool. It provides an Omnivore-style dark/light interface to browse, read, and manage blog articles tracked by blogwatcher. Single-user, self-hosted, accessed via browser on desktop or mobile.

## Core Value

Read and manage blog articles through a clean, responsive web interface without touching the CLI.

## Current Milestone: v1.2 Blog Management

**Goal:** Manage blogs entirely from the web UI — add, edit, remove — without touching the CLI.

**Target features:**
- Add new blogs by URL (auto-discover RSS/Atom feed via CLI)
- Auto-sync new blog after adding (using UI's sync with thumbnails)
- Edit blog settings (name, URL, feed URL)
- Remove blogs with confirmation (choice to keep or delete articles)
- Settings page for blog management
- RSS/Atom feeds only (no scrape support this milestone)

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

Shipped in v1.1:

- ✓ **POLISH-01**: Entire article card clickable (opens URL in new tab) — v1.1
- ✓ **POLISH-02**: Masonry grid layout as alternative to list view — v1.1
- ✓ **POLISH-03**: View toggle to switch between list and grid layouts — v1.1
- ✓ **POLISH-04**: View preference persists across sessions — v1.1
- ✓ **THUMB-01**: Extract thumbnail URL from RSS media/enclosures during sync — v1.1
- ✓ **THUMB-02**: Extract thumbnail from Open Graph meta tags as fallback — v1.1
- ✓ **THUMB-03**: Fall back to favicon when no thumbnail available — v1.1
- ✓ **THUMB-04**: Display thumbnail in article card (both list and grid views) — v1.1
- ✓ **SRCH-01**: Search articles by title text — v1.1
- ✓ **SRCH-02**: Search input with 300ms debounce (HTMX active search) — v1.1
- ✓ **SRCH-03**: Date filter: Last Week shortcut — v1.1
- ✓ **SRCH-04**: Date filter: Last Month shortcut — v1.1
- ✓ **SRCH-05**: Date filter: Custom date range picker — v1.1
- ✓ **SRCH-06**: Combined filters (blog + status + search + date together) — v1.1
- ✓ **SRCH-07**: Display results count showing how many articles match — v1.1

### Active

v1.2 scope:

- [ ] Add new blogs by URL with auto feed discovery
- [ ] Auto-sync new blog after adding
- [ ] Edit blog settings (name, URL, feed URL)
- [ ] Remove blogs with confirmation dialog
- [ ] Settings page for blog management

### Out of Scope

- User authentication/multi-user — single user, local access
- Labels/tags — not needed yet
- In-app reader view — just link to originals
- Auto-sync/background refresh — manual only
- Full-text search — would require fetching/storing article content
- Read time estimates — not in current database
- OPML import/export — deferred to future milestone
- Scrape-based blogs — RSS/Atom only for v1.2

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

| Shell exec for CLI integration | Leverage existing blogwatcher CLI for feed discovery, keep logic centralized | — Pending |

---
*Last updated: 2026-02-08 after v1.2 milestone start*
