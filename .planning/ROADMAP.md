# Roadmap: BlogWatcher UI

**Created:** 2026-02-02
**Depth:** standard
**Total Phases:** 8

## Milestones

- ✅ **v1.0 MVP** - Phases 1-5 (shipped 2026-02-03)
- 🚧 **v1.1 UI Polish & Search** - Phases 6-8 (in progress)

## Phases

<details>
<summary>✅ v1.0 MVP (Phases 1-5) - SHIPPED 2026-02-03</summary>

### Phase 1: Infrastructure Setup

**Goal:** Foundation server and database connection are functional and ready to serve the UI.

**Requirements:**
- INFRA-01: Go HTTP server serving web UI
- INFRA-02: Connect to existing blogwatcher SQLite database
- INFRA-03: HTMX for dynamic updates without full page reloads

**Success Criteria:**
1. User can navigate to localhost URL and see a basic page served by Go server
2. Server successfully reads articles and blogs from existing SQLite database at ~/.blogwatcher/blogwatcher.db
3. HTMX requests can fetch data from server endpoints and update page sections without full reload

**Depends on:** None

**Plans:** 3 plans

Plans:
- [x] 01-01-PLAN.md — Project setup, database layer, HTMX static file
- [x] 01-02-PLAN.md — HTTP server with NewServer pattern and templates
- [x] 01-03-PLAN.md — Wire handlers to database, full integration verification

---

### Phase 2: UI Layout & Navigation

**Goal:** User can navigate between different views of their articles using a responsive sidebar.

**Requirements:**
- UI-01: Responsive layout with collapsible sidebar on mobile
- UI-02: Filter views in sidebar (Inbox/unread, Archived/read)
- UI-03: Subscriptions list in sidebar showing tracked blogs

**Success Criteria:**
1. User sees a sidebar with "Inbox" and "Archived" filter options
2. User sees list of subscribed blogs in sidebar matching their blogwatcher database
3. User can collapse/expand sidebar on mobile screen sizes
4. Clicking a filter or blog in sidebar changes the main content area (even if just showing placeholder)

**Depends on:** Phase 1 (Infrastructure Setup)

**Plans:** 2 plans

Plans:
- [x] 02-01-PLAN.md — CSS foundation with dark theme, responsive grid layout, collapsible sidebar structure
- [x] 02-02-PLAN.md — HTMX navigation wiring, filter query params, active state highlighting

---

### Phase 3: Article Display

**Goal:** User can see their articles with rich metadata and open them to read.

**Requirements:**
- DISP-01: Article cards show thumbnail or site favicon
- DISP-02: Article cards show time ago ("7 hours ago")
- DISP-03: Article cards show title and source blog name
- DISP-04: Clicking article opens original URL in new tab

**Success Criteria:**
1. User sees article cards displaying title, source blog name, and relative time ("2 hours ago")
2. Each article card shows either a thumbnail image or favicon for the source blog
3. User can click an article card and original blog post opens in new browser tab
4. Articles from database appear in correct filtered view (unread in Inbox, read in Archived)
5. Clicking a blog in sidebar filters articles to only show that blog's content

**Depends on:** Phase 2 (UI Layout & Navigation)

**Plans:** 2 plans

Plans:
- [x] 03-01-PLAN.md — Template functions (timeAgo, faviconURL), ArticleWithBlog model, database JOIN query
- [x] 03-02-PLAN.md — Article card template and CSS styling with rich metadata display

---

### Phase 4: Article Management

**Goal:** User can mark articles as read/unread and trigger blog syncing from the UI.

**Requirements:**
- MGMT-01: Button to mark individual article as read
- MGMT-02: Button to mark article as unread
- MGMT-03: "Mark all read" button for bulk action
- MGMT-04: Manual sync button to scan blogs for new articles

**Success Criteria:**
1. User can mark individual article as read and see it move from Inbox to Archived
2. User can mark individual article as unread and see it move from Archived to Inbox
3. User can click "Mark all read" and see all visible articles move to Archived
4. User can click sync button and see new articles appear from blogs after scan completes
5. Read/unread state persists to database and matches CLI tool's view

**Depends on:** Phase 3 (Article Display)

**Plans:** 2 plans

Plans:
- [x] 04-01-PLAN.md — Scanner package setup (copy RSS, scraper, scanner from reference + database extensions)
- [x] 04-02-PLAN.md — Article management handlers, routes, templates with action buttons and toolbar

---

### Phase 5: Theme Toggle

**Goal:** User can switch between dark and light themes with preference persisted.

**Requirements:**
- UI-04: Dark/light theme toggle

**Success Criteria:**
1. User sees theme toggle control in UI
2. User can click toggle and interface switches between dark and light themes
3. Theme preference persists across browser sessions

**Depends on:** Phase 2 (UI Layout & Navigation)

**Plans:** 1 plan

Plans:
- [x] 05-01-PLAN.md — CSS light theme variables, toggle component, FOUC prevention, localStorage persistence

</details>

---

## 🚧 v1.1 UI Polish & Search (In Progress)

**Milestone Goal:** Improve visual presentation with masonry layout and thumbnails, add search and filtering capabilities.

### Phase 6: Enhanced Card Interaction

**Goal:** User can click entire article card to open article, and cards display rich thumbnails with fallback chain.

**Requirements:**
- POLISH-01: Entire article card is clickable (opens URL in new tab)
- THUMB-01: Extract thumbnail URL from RSS media/enclosures during sync
- THUMB-02: Extract thumbnail from Open Graph meta tags as fallback
- THUMB-03: Fall back to favicon when no thumbnail available
- THUMB-04: Display thumbnail in article card (both list and grid views)

**Success Criteria:**
1. User can click anywhere on article card and original article opens in new tab
2. Article cards display thumbnails extracted from RSS media/enclosures when available
3. When RSS has no thumbnail, article cards display Open Graph image from article URL
4. When neither RSS nor Open Graph provide thumbnail, article cards display blog favicon
5. Thumbnail images render with proper aspect ratio and no cumulative layout shift

**Depends on:** Phase 5 (Theme Toggle)

**Plans:** TBD

Plans:
- [ ] 06-01: TBD
- [ ] 06-02: TBD

---

### Phase 7: Search & Date Filtering

**Goal:** User can find articles by title search and filter by date ranges.

**Requirements:**
- SRCH-01: Search articles by title text
- SRCH-02: Search input with 300ms debounce (HTMX active search)
- SRCH-03: Date filter: Last Week shortcut
- SRCH-04: Date filter: Last Month shortcut
- SRCH-05: Date filter: Custom date range picker
- SRCH-06: Combined filters (blog + status + search + date together)
- SRCH-07: Display results count showing how many articles match

**Success Criteria:**
1. User can type in search box and see results filter to articles matching title text
2. Search input debounces at 300ms and does not trigger on every keystroke
3. User can click "Last Week" filter and see only articles from past 7 days
4. User can click "Last Month" filter and see only articles from past 30 days
5. User can select custom date range and see articles within that range
6. User can combine multiple filters (blog + status + search + date) and see articles matching all conditions
7. Results count displays "Showing X articles" or "No articles found" based on active filters

**Depends on:** Phase 5 (Theme Toggle)

**Plans:** TBD

Plans:
- [ ] 07-01: TBD
- [ ] 07-02: TBD

---

### Phase 8: Masonry Layout

**Goal:** User can toggle between list and masonry grid layouts with preference persisted.

**Requirements:**
- POLISH-02: Masonry grid layout as alternative to list view
- POLISH-03: View toggle to switch between list and grid layouts
- POLISH-04: View preference persists across sessions

**Success Criteria:**
1. User sees view toggle button to switch between list and grid layouts
2. User can click grid view and see articles arranged in masonry layout with varied card heights
3. Masonry layout responds to viewport width (1 col mobile, 2 col tablet, 3-4 col desktop)
4. User can switch back to list view and see traditional vertical layout
5. View preference persists across browser sessions (remembered on next visit)

**Depends on:** Phase 6 (Enhanced Card Interaction)

**Plans:** TBD

Plans:
- [ ] 08-01: TBD

---

## Coverage Validation

### v1.0 Requirements

| Requirement | Phase | Covered |
|-------------|-------|---------|
| INFRA-01 | 1 | ✓ |
| INFRA-02 | 1 | ✓ |
| INFRA-03 | 1 | ✓ |
| UI-01 | 2 | ✓ |
| UI-02 | 2 | ✓ |
| UI-03 | 2 | ✓ |
| DISP-01 | 3 | ✓ |
| DISP-02 | 3 | ✓ |
| DISP-03 | 3 | ✓ |
| DISP-04 | 3 | ✓ |
| MGMT-01 | 4 | ✓ |
| MGMT-02 | 4 | ✓ |
| MGMT-03 | 4 | ✓ |
| MGMT-04 | 4 | ✓ |
| UI-04 | 5 | ✓ |

**v1.0 Coverage:** 15/15 requirements mapped (100%)

### v1.1 Requirements

| Requirement | Phase | Covered |
|-------------|-------|---------|
| POLISH-01 | 6 | ✓ |
| POLISH-02 | 8 | ✓ |
| POLISH-03 | 8 | ✓ |
| POLISH-04 | 8 | ✓ |
| THUMB-01 | 6 | ✓ |
| THUMB-02 | 6 | ✓ |
| THUMB-03 | 6 | ✓ |
| THUMB-04 | 6 | ✓ |
| SRCH-01 | 7 | ✓ |
| SRCH-02 | 7 | ✓ |
| SRCH-03 | 7 | ✓ |
| SRCH-04 | 7 | ✓ |
| SRCH-05 | 7 | ✓ |
| SRCH-06 | 7 | ✓ |
| SRCH-07 | 7 | ✓ |

**v1.1 Coverage:** 15/15 requirements mapped (100%)

---

## Phase Progress

### v1.0 (Complete)

| Phase | Status | Progress |
|-------|--------|----------|
| 1 - Infrastructure Setup | Complete | 100% |
| 2 - UI Layout & Navigation | Complete | 100% |
| 3 - Article Display | Complete | 100% |
| 4 - Article Management | Complete | 100% |
| 5 - Theme Toggle | Complete | 100% |

### v1.1 (In Progress)

| Phase | Status | Progress |
|-------|--------|----------|
| 6 - Enhanced Card Interaction | Not started | 0% |
| 7 - Search & Date Filtering | Not started | 0% |
| 8 - Masonry Layout | Not started | 0% |

**v1.0 Progress:** 5/5 phases complete (100%)
**v1.1 Progress:** 0/3 phases complete (0%)
**Overall Progress:** 5/8 phases complete (62.5%)

---

*Roadmap created: 2026-02-02*
*v1.1 roadmap added: 2026-02-03*
*Last updated: 2026-02-03*
