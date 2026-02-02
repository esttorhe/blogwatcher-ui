# Phase 2: UI Layout & Navigation - Research

**Researched:** 2026-02-02
**Domain:** CSS Layout, HTMX Navigation, Dark Theme
**Confidence:** HIGH

## Summary

This phase implements a responsive sidebar layout with dark theme styling for the BlogWatcher UI. The existing Phase 1 infrastructure provides Go templates, HTMX 2.0.8, and handlers that already support partial/full page rendering based on HX-Request header.

The recommended approach combines:
1. CSS Grid for the macro page layout (sidebar + main content)
2. CSS Flexbox for micro layouts within components
3. Pure CSS checkbox hack for mobile hamburger menu toggle
4. CSS custom properties for dark theme implementation
5. HTMX hx-get/hx-target for sidebar navigation with hx-swap-oob for active states

**Primary recommendation:** Use CSS Grid with named template areas for responsive layout, implement dark theme using CSS variables with the #121212 standard background, and leverage existing HTMX partial rendering infrastructure for navigation.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| CSS Grid | Native | Page-level layout | Universal browser support, named areas simplify responsive design |
| CSS Flexbox | Native | Component-level layout | 98%+ browser support, perfect for sidebar internals |
| CSS Custom Properties | Native | Theme variables | Enables dark theme without preprocessors |
| HTMX 2.0.8 | Existing | Dynamic updates | Already installed, handles partial swaps |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| prefers-color-scheme | Native | System preference detection | Respect user's OS dark mode setting |
| CSS :checked selector | Native | State management | Mobile menu toggle without JS |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| CSS Grid | Pure Flexbox | Grid is better for 2D layouts; Flexbox would require nested containers |
| Checkbox hack | JavaScript toggle | JS provides more control but violates minimal JS approach |
| CSS variables | Hardcoded colors | Variables enable future light theme without refactoring |

**Installation:**
No additional libraries needed. All CSS is native. HTMX 2.0.8 already present from Phase 1.

## Architecture Patterns

### Recommended Project Structure
```
static/
├── htmx.min.js          # Already exists
└── styles.css           # NEW: All CSS in one file
templates/
├── base.gohtml          # UPDATE: Add CSS link, dark theme base
├── pages/
│   └── index.gohtml     # UPDATE: Grid layout structure
└── partials/
    ├── sidebar.gohtml   # NEW: Collapsible sidebar component
    ├── article-list.gohtml  # UPDATE: Add styling classes
    └── blog-list.gohtml     # UPDATE: Add click handlers
```

### Pattern 1: CSS Grid Named Areas Layout
**What:** Use grid-template-areas for semantic layout definition
**When to use:** Page-level layout with header/sidebar/main/footer regions
**Example:**
```css
/* Source: https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_grid_layout/Realizing_common_layouts_using_grids */
.layout {
  display: grid;
  min-height: 100vh;
  grid-template-areas:
    "sidebar main";
  grid-template-columns: 250px 1fr;
  grid-template-rows: 1fr;
}

/* Mobile: stack vertically */
@media (max-width: 768px) {
  .layout {
    grid-template-areas:
      "main";
    grid-template-columns: 1fr;
  }
}
```

### Pattern 2: CSS Checkbox Hack for Mobile Menu
**What:** Hidden checkbox + label to toggle sidebar visibility without JavaScript
**When to use:** Mobile hamburger menu that slides sidebar in/out
**Example:**
```html
<!-- Source: https://janessagarrow.com/blog/pure-css-hamburger-menu/ -->
<input type="checkbox" id="sidebar-toggle" class="sidebar-toggle" />
<label for="sidebar-toggle" class="hamburger" aria-label="Toggle menu">
  <span class="hamburger-line"></span>
</label>
```
```css
.sidebar-toggle {
  display: none; /* Hidden but functional */
}

.sidebar-toggle:checked ~ .sidebar {
  transform: translateX(0); /* Slide in */
}

@media (min-width: 769px) {
  .hamburger { display: none; }
  .sidebar { transform: none; }
}
```

### Pattern 3: HTMX Navigation with hx-target
**What:** Sidebar links update main content area via HTMX
**When to use:** All navigation items in sidebar
**Example:**
```html
<!-- Source: https://htmx.org/docs/ -->
<nav class="sidebar-nav">
  <a href="/articles?filter=unread"
     hx-get="/articles?filter=unread"
     hx-target="#main-content"
     hx-push-url="true"
     class="nav-link">
    Inbox
  </a>
</nav>
<main id="main-content">
  <!-- Content swapped here -->
</main>
```

### Pattern 4: Active State via hx-swap-oob
**What:** Server returns main content + updated sidebar with active class
**When to use:** Highlight currently selected navigation item
**Example:**
```html
<!-- Server response includes both main content and OOB sidebar update -->
<!-- Source: https://htmx.org/attributes/hx-swap-oob/ -->
<div id="main-content">
  <!-- Article list content -->
</div>
<nav id="sidebar-nav" hx-swap-oob="true">
  <a href="/articles" class="nav-link active">Inbox</a>
  <a href="/articles?filter=read" class="nav-link">Archived</a>
</nav>
```

### Anti-Patterns to Avoid
- **JavaScript-heavy state management:** Don't reach for JS when CSS :checked selector works
- **Inline styles in templates:** Keep all styling in styles.css for maintainability
- **Multiple CSS files:** One file is sufficient at this scale, avoid over-engineering
- **Pure black (#000) backgrounds:** Causes eye strain; use #121212 or similar dark gray
- **Fixed viewport units for sidebar:** Use CSS Grid with flexible columns instead

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Responsive layout | Custom media queries per element | CSS Grid template areas | Grid handles reflow automatically |
| Mobile menu toggle | Custom JavaScript | CSS checkbox hack | Zero JS, better performance |
| Theme switching | Inline style toggling | CSS custom properties | Variables cascade automatically |
| Active nav highlighting | Client-side class toggling | HTMX hx-swap-oob | Server controls state, simpler logic |
| Smooth transitions | Custom animation library | CSS transitions | Native, hardware accelerated |

**Key insight:** Modern CSS provides intrinsic responsive behavior. Flexbox and Grid adapt to container size without explicit breakpoints in many cases.

## Common Pitfalls

### Pitfall 1: Flash of Incorrect Theme (FOIT)
**What goes wrong:** Page loads with light colors then switches to dark
**Why it happens:** CSS loads after HTML renders
**How to avoid:** Set dark theme class on `<html>` element, or use `prefers-color-scheme` media query in CSS (no JS needed for default state)
**Warning signs:** Visible color flash on page load

### Pitfall 2: Sidebar Not Closing on Mobile Navigation
**What goes wrong:** User clicks link, content updates, but sidebar stays open
**Why it happens:** Checkbox state persists across HTMX swaps
**How to avoid:** Either use hx-swap-oob to return sidebar with unchecked state, or use hx-on:htmx:after-swap to reset checkbox
**Warning signs:** Testing mobile navigation shows sidebar remains visible

### Pitfall 3: Keyboard Accessibility of Hidden Sidebar
**What goes wrong:** Tab navigation reaches hidden off-screen sidebar items
**Why it happens:** `transform: translateX(-100%)` hides visually but not from keyboard
**How to avoid:** Add `visibility: hidden` when sidebar is closed, `visibility: visible` when open
**Warning signs:** Tabbing through page goes "nowhere" for several presses

### Pitfall 4: Grid Collapse on Empty Main Content
**What goes wrong:** Layout breaks when article list is empty
**Why it happens:** Grid columns can collapse without content
**How to avoid:** Set `min-height` on main area or use `minmax()` in grid-template-columns
**Warning signs:** Layout shifts when navigating to empty views

### Pitfall 5: HTMX Partial vs Full Page Confusion
**What goes wrong:** Direct URL navigation returns partial HTML without layout
**Why it happens:** Not checking HX-Request header properly
**How to avoid:** Existing handlers already check this - maintain pattern. Always return full page for non-HTMX requests
**Warning signs:** Bookmarked URLs show unstyled fragments

## Code Examples

Verified patterns from official sources:

### Dark Theme CSS Variables
```css
/* Source: https://css-tricks.com/a-complete-guide-to-dark-mode-on-the-web/ */
:root {
  /* Dark theme as default */
  --bg-primary: #121212;
  --bg-surface: #1e1e1e;
  --bg-elevated: #2d2d2d;
  --text-primary: #e0e0e0;
  --text-secondary: #a0a0a0;
  --accent: #64b5f6;
  --border: #333333;
}

body {
  background-color: var(--bg-primary);
  color: var(--text-primary);
}
```

### Responsive Grid Layout
```css
/* Source: https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_grid_layout/Realizing_common_layouts_using_grids */
.app-layout {
  display: grid;
  min-height: 100vh;
  grid-template-columns: 250px 1fr;
  grid-template-areas: "sidebar main";
}

.sidebar {
  grid-area: sidebar;
  background: var(--bg-surface);
}

.main-content {
  grid-area: main;
  background: var(--bg-primary);
  padding: 1rem;
}

@media (max-width: 768px) {
  .app-layout {
    grid-template-columns: 1fr;
    grid-template-areas: "main";
  }

  .sidebar {
    position: fixed;
    left: 0;
    top: 0;
    height: 100vh;
    width: 250px;
    transform: translateX(-100%);
    transition: transform 0.3s ease;
    z-index: 100;
  }
}
```

### Hamburger Icon with CSS
```css
/* Source: https://janessagarrow.com/blog/pure-css-hamburger-menu/ */
.hamburger {
  display: none;
  cursor: pointer;
  padding: 10px;
}

.hamburger-line {
  display: block;
  width: 25px;
  height: 3px;
  background: var(--text-primary);
  position: relative;
}

.hamburger-line::before,
.hamburger-line::after {
  content: '';
  position: absolute;
  width: 100%;
  height: 100%;
  background: inherit;
  transition: transform 0.3s ease;
}

.hamburger-line::before { top: -8px; }
.hamburger-line::after { top: 8px; }

/* Transform to X when checked */
.sidebar-toggle:checked ~ .hamburger .hamburger-line {
  background: transparent;
}
.sidebar-toggle:checked ~ .hamburger .hamburger-line::before {
  transform: rotate(45deg) translate(5px, 6px);
}
.sidebar-toggle:checked ~ .hamburger .hamburger-line::after {
  transform: rotate(-45deg) translate(5px, -6px);
}

@media (max-width: 768px) {
  .hamburger { display: block; }
}
```

### HTMX Navigation Links
```html
<!-- Source: https://htmx.org/docs/ -->
<nav class="sidebar-nav">
  <a href="/articles?filter=unread"
     hx-get="/articles?filter=unread"
     hx-target="#main-content"
     hx-push-url="true"
     class="nav-link {{if eq .CurrentFilter "unread"}}active{{end}}">
    Inbox
  </a>
  <a href="/articles?filter=read"
     hx-get="/articles?filter=read"
     hx-target="#main-content"
     hx-push-url="true"
     class="nav-link {{if eq .CurrentFilter "read"}}active{{end}}">
    Archived
  </a>
</nav>
```

### Sidebar Blog List with Click Filtering
```html
<!-- Each blog filters the article list when clicked -->
<div class="blog-list">
  {{range .Blogs}}
  <a href="/articles?blog={{.ID}}"
     hx-get="/articles?blog={{.ID}}"
     hx-target="#main-content"
     hx-push-url="true"
     class="blog-item">
    {{.Name}}
  </a>
  {{else}}
  <p class="empty-state">No blogs tracked yet.</p>
  {{end}}
</div>
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Media queries on every element | CSS Grid auto-flow + named areas | CSS Grid matured ~2020 | Fewer breakpoints, intrinsic sizing |
| JavaScript menu toggles | CSS :checked hack | Always available | Zero JS for basic interactions |
| SASS/LESS variables | CSS Custom Properties | Wide support ~2019 | No build step needed |
| Framework grids (Bootstrap) | Native CSS Grid | CSS Grid adoption ~2018 | No dependencies |
| JSON APIs + client rendering | HTMX HTML fragments | HTMX 2.0 (2024) | Server controls state |

**Deprecated/outdated:**
- Float-based layouts: Grid/Flexbox replaced these entirely
- jQuery for DOM manipulation: HTMX declarative approach is cleaner
- Separate mobile stylesheets: Responsive CSS handles all breakpoints

## Open Questions

Things that couldn't be fully resolved:

1. **Omnivore exact color values**
   - What we know: Uses dark theme similar to Material Design guidelines
   - What's unclear: Exact hex values from Omnivore's implementation
   - Recommendation: Use Material Design dark theme standard (#121212 background) as base, refine based on user feedback

2. **Sidebar width on tablets**
   - What we know: 250px works well for desktop, mobile hides completely
   - What's unclear: Whether tablets need intermediate width
   - Recommendation: Start with 250px collapsible, test on iPad sizes, adjust if needed

3. **Animation performance on low-end devices**
   - What we know: CSS transforms are GPU-accelerated
   - What's unclear: Performance on older mobile devices
   - Recommendation: Use `will-change: transform` and keep animations simple (transform only)

## Sources

### Primary (HIGH confidence)
- [HTMX Official Documentation](https://htmx.org/docs/) - Core attributes, navigation patterns, partial rendering
- [MDN CSS Grid Layout](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_grid_layout/Realizing_common_layouts_using_grids) - Grid template areas, responsive patterns
- [HTMX hx-swap-oob Attribute](https://htmx.org/attributes/hx-swap-oob/) - Out-of-band updates for active states
- [CSS-Tricks Dark Mode Guide](https://css-tricks.com/a-complete-guide-to-dark-mode-on-the-web/) - CSS variables, prefers-color-scheme, best practices

### Secondary (MEDIUM confidence)
- [Pure CSS Hamburger Menu](https://janessagarrow.com/blog/pure-css-hamburger-menu/) - Checkbox hack implementation
- [Every Layout: The Sidebar](https://every-layout.dev/layouts/sidebar/) - Intrinsic responsive sidebar pattern
- [ColorsWall Dark Mode Palette](https://colorswall.com/palette/552154) - #121212 based color scheme

### Tertiary (LOW confidence)
- Various blog posts on dark theme color recommendations - Cross-referenced with Material Design guidelines

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All native CSS, well-documented patterns
- Architecture: HIGH - Grid/Flexbox patterns are mature and stable
- Pitfalls: MEDIUM - Based on common issues documented across multiple sources
- Dark theme colors: MEDIUM - Using established Material Design guidelines

**Research date:** 2026-02-02
**Valid until:** Stable patterns, valid indefinitely for CSS. HTMX patterns valid until major version change.
