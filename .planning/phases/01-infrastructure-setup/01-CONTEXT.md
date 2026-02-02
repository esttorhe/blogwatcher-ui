# Phase 1: Infrastructure Setup - Context

**Gathered:** 2026-02-02
**Status:** Ready for planning

<domain>
## Phase Boundary

Go HTTP server that serves the web UI, connects to the existing blogwatcher SQLite database, and integrates HTMX for dynamic updates. This is foundation infrastructure — no actual UI or features yet.

</domain>

<decisions>
## Implementation Decisions

### Database path
- Use CLI's default path: `~/.blogwatcher/blogwatcher.db`
- No configuration options needed — single, fixed location
- Share database with CLI tool (read/write same file)

### Missing database handling
- If database doesn't exist at startup: show setup message
- Message: "Run blogwatcher to set up your database first"
- Don't error out harshly — friendly guidance to user
- Don't create empty database — let CLI handle initialization

### Claude's Discretion
- Server port (default 8080 is fine, or pick another)
- Host binding (localhost vs 0.0.0.0)
- Startup message format
- Hot reload in development
- HTMX integration approach (CDN vs embedded)
- Template organization

</decisions>

<specifics>
## Specific Ideas

- Reference codebase at `.reference/blogwatcher/` uses `modernc.org/sqlite` — reuse same driver for compatibility
- Database schema is already defined in reference code's `internal/storage/database.go`

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 01-infrastructure-setup*
*Context gathered: 2026-02-02*
