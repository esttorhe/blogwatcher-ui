# Requirements: BlogWatcher UI

**Defined:** 2026-02-02
**Core Value:** Read and manage blog articles through a clean, responsive web interface without touching the CLI.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### UI/Layout

- [ ] **UI-01**: Responsive layout with collapsible sidebar on mobile
- [ ] **UI-02**: Filter views in sidebar (Inbox/unread, Archived/read)
- [ ] **UI-03**: Subscriptions list in sidebar showing tracked blogs
- [ ] **UI-04**: Dark/light theme toggle

### Article Display

- [ ] **DISP-01**: Article cards show thumbnail or site favicon
- [ ] **DISP-02**: Article cards show time ago ("7 hours ago")
- [ ] **DISP-03**: Article cards show title and source blog name
- [ ] **DISP-04**: Clicking article opens original URL in new tab

### Article Management

- [ ] **MGMT-01**: Button to mark individual article as read
- [ ] **MGMT-02**: Button to mark article as unread
- [ ] **MGMT-03**: "Mark all read" button for bulk action
- [ ] **MGMT-04**: Manual sync button to scan blogs for new articles

### Infrastructure

- [ ] **INFRA-01**: Go HTTP server serving web UI
- [ ] **INFRA-02**: Connect to existing blogwatcher SQLite database
- [ ] **INFRA-03**: HTMX for dynamic updates without full page reloads

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Search

- **SRCH-01**: Search articles by title
- **SRCH-02**: Search articles by blog name

### Blog Management

- **BLOG-01**: Add new blog from UI
- **BLOG-02**: Remove blog from UI
- **BLOG-03**: Edit blog settings from UI

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
| ------- | ------ |
| User authentication | Single user, local access only |
| Labels/tags | Not needed for v1, adds complexity |
| In-app reader view | Just link to originals, avoids content fetching |
| Auto-sync/background refresh | Manual only keeps it simple |
| Read time estimates | Not in current database schema |
| Keyboard shortcuts | Nice to have but not essential for v1 |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
| ----------- | ----- | ------ |
| INFRA-01 | Phase 1 | Pending |
| INFRA-02 | Phase 1 | Pending |
| INFRA-03 | Phase 1 | Pending |
| UI-01 | Phase 2 | Pending |
| UI-02 | Phase 2 | Pending |
| UI-03 | Phase 2 | Pending |
| DISP-01 | Phase 3 | Pending |
| DISP-02 | Phase 3 | Pending |
| DISP-03 | Phase 3 | Pending |
| DISP-04 | Phase 3 | Pending |
| MGMT-01 | Phase 4 | Pending |
| MGMT-02 | Phase 4 | Pending |
| MGMT-03 | Phase 4 | Pending |
| MGMT-04 | Phase 4 | Pending |
| UI-04 | Phase 5 | Pending |

**Coverage:**

- v1 requirements: 15 total
- Mapped to phases: 15
- Unmapped: 0

---

*Requirements defined: 2026-02-02*
*Last updated: 2026-02-02 after roadmap creation*
