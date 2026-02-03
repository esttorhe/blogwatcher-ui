# Roadmap: BlogWatcher UI

**Created:** 2026-02-02
**Depth:** standard
**Total Phases:** 5

## Milestone: v1.0

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
- [ ] 05-01-PLAN.md — CSS light theme variables, toggle component, FOUC prevention, localStorage persistence

---

## Coverage Validation

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

**Coverage:** 15/15 requirements mapped (100%)

---

## Phase Progress

| Phase | Status | Progress |
|-------|--------|----------|
| 1 - Infrastructure Setup | Complete | 100% |
| 2 - UI Layout & Navigation | Complete | 100% |
| 3 - Article Display | Complete | 100% |
| 4 - Article Management | Complete | 100% |
| 5 - Theme Toggle | Planned | 0% |

**Overall Progress:** 4/5 phases complete (80%)

---

*Roadmap created: 2026-02-02*
*Last updated: 2026-02-03 (Phase 5 planned)*
