# Phase 5: Theme Toggle - Context

**Gathered:** 2026-02-03
**Status:** Ready for planning

<domain>
## Phase Boundary

User can switch between dark and light themes with preference persisted. Three-way toggle: Light, Dark, System. Preference saved in browser storage.

</domain>

<decisions>
## Implementation Decisions

### Toggle placement
- Header area (top right corner of main content)
- Not in sidebar — keeps sidebar for navigation only

### Toggle style
- Segmented control with three options
- Visual representation: sun icon | computer/system icon | moon icon
- Clearly shows which mode is active

### Default behavior
- Default to "System" for new users
- Respects OS dark/light preference automatically
- User can override to force light or dark

### Light theme colors
- Warm cream feel, not pure white
- Slightly warm off-white backgrounds
- Easier on eyes than stark white
- Should complement the existing dark theme aesthetic

### Persistence
- localStorage for preference storage
- Restore preference on page load before render (avoid flash)

### Claude's Discretion
- Exact color values for light theme
- Transition animations between themes
- Icon choices for the segmented control
- How to detect system preference changes

</decisions>

<specifics>
## Specific Ideas

- Three-way toggle like many modern apps (Slack, Discord, VS Code)
- Warm cream aesthetic similar to Notion's light mode
- Should feel cohesive with the existing dark theme design language

</specifics>

<deferred>
## Deferred Ideas

The following UI improvements came up but belong in a separate phase (Phase 6 or backlog):

- **Masonry card layout option** — Alternative to current list layout
- **Full card clickable** — Entire card opens article, not just title
- **Mobile menu overlap fix** — Hamburger button covers title when menu open
- **Thumbnail support** — Phase 3 shows favicons only, user wants actual thumbnails

These should be addressed in a "UI Polish" or "Layout Options" phase.

</deferred>

---

*Phase: 05-theme-toggle*
*Context gathered: 2026-02-03*
