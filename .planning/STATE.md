# Project State: BlogWatcher UI

**Last updated:** 2026-02-02
**Current phase:** Phase 3 - Article Display (Complete)
**Overall progress:** 60% (3/5 phases complete, 8 plans executed)

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-02)

**Core value:** Read and manage blog articles through a clean, responsive web interface without touching the CLI.

**Current focus:** Phase 3 complete - Article cards with rich metadata displaying. Ready for Phase 4 article management.

## Phase Status

| Phase | Name | Status | Progress |
|-------|------|--------|----------|
| 1 | Infrastructure Setup | Complete | 100% |
| 2 | UI Layout & Navigation | Complete | 100% |
| 3 | Article Display | Complete | 100% |
| 4 | Article Management | Pending | 0% |
| 5 | Theme Toggle | Pending | 0% |

## Current Phase

**Phase:** 3 of 5 (Article Display) - COMPLETE
**Plan:** 2 of 2 in current phase
**Status:** Phase complete - Ready for Phase 4
**Last activity:** 2026-02-02 - Completed 03-02-PLAN.md (article card templates)

**Progress bar:** `[██████----] 60%` (8/14 plans complete estimate)

## Performance Metrics

**Phases completed:** 3/5
**Plans executed:** 8
**Requirements delivered:** 10/15 (INFRA-01, INFRA-02, INFRA-03, UI-01, UI-02, UI-03, DISP-01, DISP-02, DISP-03, DISP-04)

**Velocity:** ~4 min per plan (Phase 1: 3 plans in ~11 min, Phase 2: 2 plans in ~7 min, Phase 3: 2 plans in ~13 min)

## Accumulated Context

### Key Decisions

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-02-02 | Use Go templates + HTMX | Go-native, minimal JS, matches reference codebase style |
| 2026-02-02 | Share CLI database | Single source of truth, CLI and UI coexist |
| 2026-02-02 | No in-app reader | Simpler, just link to originals |
| 2026-02-02 | Manual sync only | Keeps it simple, user controls refresh |
| 2026-02-02 | Use modernc.org/sqlite | Same pure-Go driver as CLI, no CGO dependency |
| 2026-02-02 | Read-only database access | UI doesn't create schema - that's CLI's job |
| 2026-02-02 | Self-host HTMX | Downloaded 2.0.4 for production reliability |
| 2026-02-02 | Go 1.22+ method routing | Cleaner route definitions with explicit HTTP methods |
| 2026-02-02 | Template composition pattern | Base wraps content blocks for layout reuse |
| 2026-02-02 | HTMX request detection | HX-Request header determines fragment vs full page |
| 2026-02-02 | Index fetches all data | Avoids extra HTMX request on initial page load |
| 2026-02-02 | Empty state guidance | Direct users to CLI for blog management |
| 2026-02-02 | Dark theme as default | Applied class="dark" to html element for immediate dark rendering |
| 2026-02-02 | Pure CSS hamburger toggle | Checkbox hack instead of JavaScript for mobile menu |
| 2026-02-02 | 768px mobile breakpoint | Standard tablet/phone breakpoint for responsive layout |
| 2026-02-02 | CSS custom properties | Established naming convention for future theme toggle |
| 2026-02-02 | ListArticlesByReadStatus for filtering | New method for explicit read status (vs ListArticles which has different semantics) |
| 2026-02-02 | CurrentBlogID as int64 | 0 means no filter, simpler than pointer for template comparison |
| 2026-02-02 | h1 in article-list partial | Ensures title updates on HTMX swaps |
| 2026-02-02 | FuncMap before ParseGlob | Go templates require functions registered before parsing |
| 2026-02-02 | Google S2 for favicons | faviconURL uses google.com/s2/favicons for reliable icons |
| 2026-02-02 | ListArticlesWithBlog for display | JOIN query provides BlogName and BlogURL for card display |

### Active TODOs

- Execute Phase 4: Article management (mark read/unread)

### Known Blockers

None currently

### Technical Notes

**Database location:** `~/.blogwatcher/blogwatcher.db`

**Schema (from reference):**
- `blogs` table: id, name, url, feed_url, scrape_selector, last_scanned
- `articles` table: id, blog_id, title, url, published_date, discovered_date, is_read

**Tech stack:** Go server, Go templates, HTMX, SQLite (modernc.org/sqlite), CSS custom properties

**CSS Variables (established in 02-01):**
- --bg-primary: #121212
- --bg-surface: #1e1e1e
- --bg-elevated: #2d2d2d
- --text-primary: #e0e0e0
- --text-secondary: #a0a0a0
- --accent: #64b5f6
- --border: #333333

**HTMX Navigation Patterns (established in 02-02):**
- Filter params: r.URL.Query().Get("filter") and r.URL.Query().Get("blog")
- Active state: {{if eq .CurrentFilter "read"}} active{{end}}
- Mobile close: hx-on::after-swap="document.getElementById('sidebar-toggle').checked = false"

**Template Functions (established in 03-01):**
- `{{ timeAgo .PublishedDate }}` - Relative timestamps ("7 hours ago")
- `{{ faviconURL .BlogURL }}` - Google S2 favicon URL
- ArticleWithBlog model has BlogName and BlogURL from JOIN

**Article Card Pattern (established in 03-02):**
- Flexbox layout: favicon (32px) + content area
- Title truncated with ellipsis for long titles
- Meta line: blog name + dot separator + relative time
- External links: target="_blank" rel="noopener noreferrer"

## Session History

| Date | Action | Notes |
|------|--------|-------|
| 2026-02-02 | Project initialized | Created PROJECT.md, REQUIREMENTS.md, ROADMAP.md with 5 phases |
| 2026-02-02 | Completed 01-01 | Database layer and HTMX setup (2 min) |
| 2026-02-02 | Completed 01-02 | HTTP server with HTMX integration (6 min) |
| 2026-02-02 | Completed 01-03 | Wire handlers to database, full integration (3 min) |
| 2026-02-02 | Phase 1 complete | Infrastructure foundation ready, real data flowing |
| 2026-02-02 | Completed 02-01 | Dark theme CSS, grid layout, hamburger menu (3 min) |
| 2026-02-02 | Completed 02-02 | HTMX navigation, filter/blog params, active states (4 min) |
| 2026-02-02 | Phase 2 complete | Navigation fully functional, ready for article display |
| 2026-02-02 | Completed 03-01 | Template functions + ArticleWithBlog database method (11 min) |
| 2026-02-02 | Completed 03-02 | Article card templates with rich metadata (2 min) |
| 2026-02-02 | Phase 3 complete | Article display with favicons, titles, blog names, timestamps |

## Session Continuity

Last session: 2026-02-02
Stopped at: Phase 3 complete, ready for Phase 4 planning
Resume file: None
Next action: /gsd:plan-phase 4 (Article Management)

---

*State initialized: 2026-02-02*
