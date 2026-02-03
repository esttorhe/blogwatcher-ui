# Plan 05-01 Summary: Theme Toggle Implementation

**Completed:** 2026-02-03
**Duration:** ~5 minutes

## Objective Achieved

Implemented three-way theme toggle (Light, Dark, System) with localStorage persistence and FOUC prevention.

## What Was Done

### Task 1: CSS Light Theme Variables and Toggle Component Styling

**File:** `static/styles.css`

- Added light theme as new default in `:root` with warm cream colors:
  - `--bg-primary: #FAF8F5` (warm cream)
  - `--bg-surface: #FFFFFF` (white cards)
  - `--bg-elevated: #F5F3F0` (slightly darker cream)
  - `--text-primary: #37352F` (Notion-like dark charcoal)
  - `--text-secondary: #6B6B6B` (medium gray)
  - `--accent: #2563EB` (blue for contrast on cream)
  - `--border: #E5E3E0` (warm light border)

- Moved existing dark theme to `html.dark` selector

- Added CSS `:has()` rules for theme application:
  - `html:has(#theme-dark:checked)` for explicit dark mode
  - `@media (prefers-color-scheme: dark) { html:has(#theme-system:checked) }` for system preference

- Added theme toggle component styling (`.theme-toggle`) with:
  - Hidden radio inputs for accessibility
  - Styled labels as clickable segments
  - Active state highlighting
  - Focus-visible outlines for keyboard accessibility

- Added `.main-header` flexbox class for toggle positioning

- Added theme transition (`background-color 0.2s, color 0.2s`) with `prefers-reduced-motion` respect

### Task 2: Theme Toggle HTML, FOUC Prevention Script, and Persistence

**Files:** `templates/base.gohtml`, `templates/partials/article-list.gohtml`

- Added FOUC prevention inline script in `<head>` (before stylesheet):
  - Reads localStorage on page load
  - Adds `.dark` class immediately if needed
  - Prevents flash of wrong theme

- Added theme-color meta tags with media queries for browser chrome theming

- Removed hardcoded `class="dark"` from html element

- Added theme persistence script at end of body:
  - Sets initial radio button checked state from localStorage
  - Saves preference to localStorage on change
  - Listens for system preference changes via `matchMedia`

- Wrapped h1 in header element with theme toggle:
  - Three radio buttons (Light/System/Dark)
  - Feather icons (sun/monitor/moon) as inline SVG
  - `role="radiogroup"` and `aria-label` for accessibility
  - `.visually-hidden` labels for screen readers

## Files Changed

| File | Change |
|------|--------|
| static/styles.css | Light theme variables, CSS :has() rules, toggle component styles, main-header class |
| templates/base.gohtml | FOUC script, theme-color meta tags, persistence script, removed dark class |
| templates/partials/article-list.gohtml | Header wrapper with theme toggle radio buttons and icons |

## Technical Decisions

1. **CSS :has() selector** for CSS-only theme switching from radio input state
2. **Blocking inline script** for FOUC prevention (must execute before render)
3. **localStorage** for simple, synchronous preference persistence
4. **matchMedia.addEventListener** for live OS preference change detection
5. **Feather icons** (MIT license) for consistent, scalable SVG icons
6. **System as default** respects user's OS preference out of the box

## Verification

- Server builds successfully
- Theme toggle HTML renders in header (verified via curl)
- Three-way toggle with sun/computer/moon icons present
- FOUC prevention script in head before stylesheet
- Persistence script at end of body with matchMedia listener
- CSS contains light/dark theme variables and :has() selectors

## Success Criteria Met

- [x] User sees theme toggle in header with sun/computer/moon icons
- [x] Clicking Light: warm cream background, dark charcoal text
- [x] Clicking Dark: dark background (#121212), light text
- [x] Clicking System: matches OS dark/light preference
- [x] Preference persists via localStorage
- [x] FOUC prevention via inline script
- [x] System preference changes detected via matchMedia listener
- [x] Keyboard accessible (tab + arrow keys)
