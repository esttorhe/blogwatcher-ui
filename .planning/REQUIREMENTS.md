# Requirements: BlogWatcher UI

**Defined:** 2026-02-02
**Core Value:** Read and manage blog articles through a clean, responsive web interface without touching the CLI.

## v1.0 Requirements (Complete)

Shipped in v1.0 milestone:

### Infrastructure

- [x] **INFRA-01**: Go HTTP server serving web UI
- [x] **INFRA-02**: Connect to existing blogwatcher SQLite database
- [x] **INFRA-03**: HTMX for dynamic updates without full page reloads

### UI/Layout

- [x] **UI-01**: Responsive layout with collapsible sidebar on mobile
- [x] **UI-02**: Filter views in sidebar (Inbox/unread, Archived/read)
- [x] **UI-03**: Subscriptions list in sidebar showing tracked blogs
- [x] **UI-04**: Dark/light theme toggle

### Article Display

- [x] **DISP-01**: Article cards show thumbnail or site favicon
- [x] **DISP-02**: Article cards show time ago ("7 hours ago")
- [x] **DISP-03**: Article cards show title and source blog name
- [x] **DISP-04**: Clicking article opens original URL in new tab

### Article Management

- [x] **MGMT-01**: Button to mark individual article as read
- [x] **MGMT-02**: Button to mark article as unread
- [x] **MGMT-03**: "Mark all read" button for bulk action
- [x] **MGMT-04**: Manual sync button to scan blogs for new articles

## v1.1 Requirements

Requirements for v1.1 milestone (UI Polish & Search).

### UI Polish

- [ ] **POLISH-01**: Entire article card is clickable (opens URL in new tab)
- [ ] **POLISH-02**: Masonry grid layout as alternative to list view
- [ ] **POLISH-03**: View toggle to switch between list and grid layouts
- [ ] **POLISH-04**: View preference persists across sessions

### Thumbnails

- [ ] **THUMB-01**: Extract thumbnail URL from RSS media/enclosures during sync
- [ ] **THUMB-02**: Extract thumbnail from Open Graph meta tags as fallback
- [ ] **THUMB-03**: Fall back to favicon when no thumbnail available
- [ ] **THUMB-04**: Display thumbnail in article card (both list and grid views)

### Search & Filtering

- [ ] **SRCH-01**: Search articles by title text
- [ ] **SRCH-02**: Search input with 300ms debounce (HTMX active search)
- [ ] **SRCH-03**: Date filter: Last Week shortcut
- [ ] **SRCH-04**: Date filter: Last Month shortcut
- [ ] **SRCH-05**: Date filter: Custom date range picker
- [ ] **SRCH-06**: Combined filters (blog + status + search + date together)
- [ ] **SRCH-07**: Display results count showing how many articles match

## Future Requirements

Deferred to v1.2 or later. Tracked but not in current roadmap.

### Search Enhancements

- **SRCH-F01**: Full-text search of article content (requires content storage)
- **SRCH-F02**: Saved searches with quick access
- **SRCH-F03**: Search suggestions/autocomplete

### Blog Management

- **BLOG-01**: Add new blog from UI
- **BLOG-02**: Remove blog from UI
- **BLOG-03**: Edit blog settings from UI

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| User authentication | Single user, local access only |
| Labels/tags | Not needed, adds complexity |
| In-app reader view | Just link to originals, avoids content fetching |
| Auto-sync/background refresh | Manual only keeps it simple |
| Full-text search | Would require fetching/storing article content |
| Read time estimates | Not in current database schema |
| Keyboard shortcuts | Nice to have but not essential |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

### v1.0 Requirements (Complete)

| Requirement | Phase | Status |
|-------------|-------|--------|
| INFRA-01 | Phase 1 | Complete |
| INFRA-02 | Phase 1 | Complete |
| INFRA-03 | Phase 1 | Complete |
| UI-01 | Phase 2 | Complete |
| UI-02 | Phase 2 | Complete |
| UI-03 | Phase 2 | Complete |
| DISP-01 | Phase 3 | Complete |
| DISP-02 | Phase 3 | Complete |
| DISP-03 | Phase 3 | Complete |
| DISP-04 | Phase 3 | Complete |
| MGMT-01 | Phase 4 | Complete |
| MGMT-02 | Phase 4 | Complete |
| MGMT-03 | Phase 4 | Complete |
| MGMT-04 | Phase 4 | Complete |
| UI-04 | Phase 5 | Complete |

### v1.1 Requirements (In Progress)

| Requirement | Phase | Status |
|-------------|-------|--------|
| POLISH-01 | Phase 6 | Complete |
| THUMB-01 | Phase 6 | Complete |
| THUMB-02 | Phase 6 | Complete |
| THUMB-03 | Phase 6 | Complete |
| THUMB-04 | Phase 6 | Complete |
| SRCH-01 | Phase 7 | Pending |
| SRCH-02 | Phase 7 | Pending |
| SRCH-03 | Phase 7 | Pending |
| SRCH-04 | Phase 7 | Pending |
| SRCH-05 | Phase 7 | Pending |
| SRCH-06 | Phase 7 | Pending |
| SRCH-07 | Phase 7 | Pending |
| POLISH-02 | Phase 8 | Pending |
| POLISH-03 | Phase 8 | Pending |
| POLISH-04 | Phase 8 | Pending |

**Coverage:**
- v1.0 requirements: 15 mapped (100%)
- v1.1 requirements: 15 mapped (100%)
- Total: 30/30 requirements mapped
- Unmapped: 0

---
*Requirements defined: 2026-02-02*
*v1.1 roadmap added: 2026-02-03*
*Last updated: 2026-02-03*
