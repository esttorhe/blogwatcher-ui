# Phase 8: Masonry Layout - Research

**Researched:** 2026-02-03
**Domain:** CSS Grid responsive layouts, view toggle patterns, localStorage preferences
**Confidence:** HIGH

## Summary

Phase 8 implements a masonry-style grid layout as an alternative to the existing list view. Research reveals that native CSS Grid masonry (`grid-template-rows: masonry` or `display: grid-lanes`) remains experimental with limited browser support as of February 2026 (Firefox and Safari Technology Preview only). The recommended production approach is **CSS Grid with auto-fit and varied card heights**, which creates a responsive grid that adapts to viewport width without JavaScript libraries.

The existing codebase already has the infrastructure needed: localStorage pattern from theme toggle (Phase 5), CSS custom properties for theming, and stretched-link pattern from Phase 6. The implementation will reuse these patterns for consistency.

**Primary recommendation:** Use CSS Grid with `repeat(auto-fit, minmax(280px, 1fr))` for the grid layout, toggle via CSS class on the container, persist preference in localStorage using the same pattern as theme toggle, and create a segmented control with list/grid icons adjacent to the theme toggle.

## Standard Stack

The established approach for this phase uses existing technologies already in the codebase.

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| CSS Grid | Native | Responsive grid layout | Native browser support, no dependencies, works with auto-fit for true responsiveness |
| localStorage API | Native | Preference persistence | Same pattern as theme toggle (Phase 5), simple key-value storage |
| CSS Custom Properties | Native | Dynamic styling | Already used throughout codebase for theming |
| HTMX | 2.0.4 | Dynamic updates | Already integrated, preserves view state during updates |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| CSS `aspect-ratio` | Native | Preserve card proportions | For grid layout only, ensures consistent card sizing |
| `@supports` query | Native | Progressive enhancement | Feature detection for future masonry support |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| CSS Grid auto-fit | JavaScript libraries (Masonry.js, Isotope) | JS libraries add 24KB+ and require maintenance; CSS is native and faster |
| CSS Grid | CSS Columns | Columns break logical reading order (accessibility issue); Grid maintains document order |
| CSS Grid | Native CSS masonry | Native masonry has experimental support only (Firefox/Safari TP); not production-ready |
| CSS class toggle | Separate templates | Class toggle is simpler, preserves state, avoids duplicate markup |

**Installation:**
```bash
# No new dependencies required - all native CSS and existing patterns
```

## Architecture Patterns

### Recommended Project Structure
```
static/
├── styles.css           # Add grid layout styles and view toggle
└── htmx.min.js         # Already present

templates/
├── partials/
│   └── article-list.gohtml  # Add view toggle, modify article container
└── base.gohtml         # Add view preference script (mirrors theme script)
```

### Pattern 1: CSS Grid with Auto-Fit
**What:** Responsive grid that automatically adjusts columns based on viewport width without media queries
**When to use:** For masonry-style grid layouts that need to be truly responsive across all devices

**Example:**
```css
/* Source: https://css-tricks.com/auto-sizing-columns-css-grid-auto-fill-vs-auto-fit/ */
.articles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1rem;
  /* auto-fit collapses empty columns and stretches existing cards to fill space */
}

/* Mobile: 1 column naturally when viewport < 280px + gap */
/* Tablet: 2 columns when viewport >= 600px */
/* Desktop: 3-4 columns when viewport >= 900px+ */
```

**Why auto-fit over auto-fill:** Auto-fit collapses empty columns and expands cards to fill available space, preventing awkward gaps. Auto-fill would preserve empty columns, creating unnecessary whitespace.

### Pattern 2: View Toggle with localStorage Persistence
**What:** Reuse the theme toggle pattern from Phase 5 for view preference
**When to use:** Any user preference that needs to persist across sessions

**Example:**
```javascript
/* Source: Existing pattern from base.gohtml theme toggle */
(function() {
  var viewToggle = document.querySelectorAll('input[name="view"]');
  var stored = localStorage.getItem('view') || 'list';
  var articlesContainer = document.getElementById('articles-container');

  // Set initial checked state
  var initial = document.getElementById('view-' + stored);
  if (initial) initial.checked = true;

  // Apply initial view
  if (stored === 'grid') {
    articlesContainer.classList.add('articles-grid');
  }

  function updateView(value) {
    localStorage.setItem('view', value);
    articlesContainer.classList.toggle('articles-grid', value === 'grid');
  }

  viewToggle.forEach(function(r) {
    r.addEventListener('change', function(e) {
      updateView(e.target.value);
    });
  });
})();
```

### Pattern 3: Segmented Control for View Toggle
**What:** Button group with radio inputs and labels styled as a single control
**When to use:** Mutually exclusive options with 2-3 choices (list vs grid)

**Example:**
```html
<!-- Source: https://primer.style/components/segmented-control/ (accessibility guidance) -->
<div class="view-toggle" role="group" aria-label="View mode">
  <input type="radio" name="view" id="view-list" value="list">
  <label for="view-list" title="List view">
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
      <!-- List icon: horizontal lines -->
      <line x1="3" y1="6" x2="21" y2="6"/>
      <line x1="3" y1="12" x2="21" y2="12"/>
      <line x1="3" y1="18" x2="21" y2="18"/>
    </svg>
    <span class="visually-hidden">List</span>
  </label>
  <input type="radio" name="view" id="view-grid" value="grid">
  <label for="view-grid" title="Grid view">
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
      <!-- Grid icon: 2x2 squares -->
      <rect x="3" y="3" width="7" height="7"/>
      <rect x="14" y="3" width="7" height="7"/>
      <rect x="3" y="14" width="7" height="7"/>
      <rect x="14" y="14" width="7" height="7"/>
    </svg>
    <span class="visually-hidden">Grid</span>
  </label>
</div>
```

**CSS:**
```css
/* Source: Existing theme-toggle pattern from styles.css */
.view-toggle {
  display: inline-flex;
  background-color: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 4px;
  gap: 4px;
}

.view-toggle input[type="radio"] {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
  pointer-events: none;
}

.view-toggle label {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 6px 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s ease, color 0.2s ease;
  color: var(--text-secondary);
}

.view-toggle input[type="radio"]:checked + label {
  background-color: var(--bg-elevated);
  color: var(--text-primary);
}
```

### Pattern 4: Grid Layout Card Adjustments
**What:** Modify existing article cards to work in both list and grid layouts
**When to use:** When cards need to adapt between vertical list and grid arrangements

**Key changes needed:**
```css
/* List view (default) - existing styles */
.article-card {
  display: flex;
  flex-direction: row;
  align-items: flex-start;
  /* ... existing styles ... */
}

/* Grid view - vertical card layout */
.articles-grid .article-card {
  display: flex;
  flex-direction: column;
  align-items: stretch;
}

.articles-grid .article-thumbnail {
  width: 100%;
  height: auto;
  aspect-ratio: 16 / 9; /* Preserve aspect ratio */
}

.articles-grid .article-content {
  flex: 1; /* Allow varied heights */
  display: flex;
  flex-direction: column;
}

.articles-grid .article-title {
  /* Allow multi-line in grid view */
  white-space: normal;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
```

### Anti-Patterns to Avoid
- **Don't use CSS columns for masonry:** Items flow vertically (down column 1, then down column 2), breaking logical reading order and tab navigation. Accessibility issue for keyboard and screen reader users.
- **Don't create separate templates for each view:** Duplicates markup, harder to maintain. Use CSS class toggle instead.
- **Don't use JavaScript libraries for layout:** Masonry.js (24KB, last updated 2017) and similar are unnecessary when CSS Grid provides native responsive behavior.
- **Don't use grid-template-rows: masonry in production:** Experimental feature with limited browser support (Firefox/Safari TP only as of Feb 2026).
- **Don't break stretched-link pattern:** Grid cards need same z-index layering as list cards to keep title links and action buttons both functional.

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Responsive column count | Manual media queries for 1/2/3/4 columns | CSS Grid `repeat(auto-fit, minmax())` | Auto-fit handles all viewport sizes without breakpoints; content-driven not device-driven |
| View preference persistence | Custom storage wrapper or cookies | localStorage.setItem/getItem | Native API, same pattern as theme toggle, 5MB storage, synchronous |
| Toggle button accessibility | Custom button group with divs | Radio inputs + labels with role="group" | Native keyboard navigation, screen reader support, no ARIA required |
| Card aspect ratios | Padding-top percentage hack | CSS `aspect-ratio` property | Native since 2021, well-supported, clearer intent |
| Icon resources | Hand-draw SVG icons | Use established icon libraries or free resources | Icons8, Bootstrap Icons, Flaticon provide tested, accessible SVGs |

**Key insight:** CSS Grid's auto-fit/auto-fill with minmax() eliminates the need for media query breakpoints entirely. The layout responds to container width, not viewport width, making it more resilient to unknown devices and truly responsive.

## Common Pitfalls

### Pitfall 1: Reading Order and Accessibility with CSS Masonry
**What goes wrong:** CSS columns and some grid implementations reorder content visually, breaking tab navigation and screen reader order. Users must navigate down entire first column before moving to second column, then scroll back to top.

**Why it happens:** CSS columns fill vertically first (column 1 top-to-bottom, then column 2 top-to-bottom). Native CSS masonry can also reorder items to fill gaps, disconnecting DOM order from visual order.

**How to avoid:**
- Use CSS Grid with auto-fit (preserves DOM order horizontally)
- Don't use CSS columns for interactive content
- Native masonry proposal includes `reading-flow` property to address this, but not yet implemented

**Warning signs:**
- Tab navigation jumps unexpectedly between cards
- Screen reader announces cards in non-visual order
- Cards appear in different order than in HTML

**Source:** [CSS Grid Masonry Accessibility Discussion](https://github.com/w3c/csswg-drafts/issues/5675)

### Pitfall 2: Auto-Fill vs Auto-Fit Confusion
**What goes wrong:** Using `auto-fill` instead of `auto-fit` creates empty ghost columns that take up space, leaving awkward gaps when viewport is wide.

**Why it happens:** Auto-fill creates maximum possible columns even if empty; auto-fit collapses empty columns and redistributes space.

**How to avoid:**
- Use `auto-fit` for card layouts where you want items to expand
- Use `auto-fill` only when you need consistent grid structure (rare)
- Formula: `repeat(auto-fit, minmax(MIN, 1fr))` where MIN is minimum card width

**Warning signs:**
- Large gaps appear on wide viewports
- Cards don't expand to fill available space
- Layout looks "unfinished" on desktop

**Source:** [CSS-Tricks: Auto-Fill vs Auto-Fit](https://css-tricks.com/auto-sizing-columns-css-grid-auto-fill-vs-auto-fit/)

### Pitfall 3: Stretched-Link Breaking in Grid Layout
**What goes wrong:** When converting cards to grid layout, z-index stacking context can break, making action buttons unclickable or entire card unclickable.

**Why it happens:** CSS Grid can create new stacking contexts. Stretched-link requires parent to be containing block (position: relative) and buttons to have z-index higher than stretched link's ::after pseudo-element.

**How to avoid:**
- Keep `position: relative` on `.article-card`
- Keep `.stretched-link::after { z-index: 1 }`
- Keep `.action-btn { position: relative; z-index: 2 }`
- Don't add `transform`, `perspective`, or `filter` to cards (creates new containing block)

**Warning signs:**
- Clicking card doesn't open article
- Clicking action button doesn't work
- Hover states don't match clickable areas

**Source:** [Bootstrap Stretched Link Documentation](https://getbootstrap.com/docs/5.3/helpers/stretched-link/)

### Pitfall 4: localStorage Without Error Handling
**What goes wrong:** localStorage can throw exceptions (quota exceeded, disabled by user, private browsing), breaking JavaScript and view toggle.

**Why it happens:** Browsers allow disabling localStorage, private mode may not persist it, or quota can be exceeded (5MB limit).

**How to avoid:**
```javascript
function safeSetItem(key, value) {
  try {
    localStorage.setItem(key, value);
  } catch (e) {
    console.warn('localStorage not available:', e);
    // Fallback to in-memory or accept default
  }
}

function safeGetItem(key, defaultValue) {
  try {
    return localStorage.getItem(key) || defaultValue;
  } catch (e) {
    console.warn('localStorage not available:', e);
    return defaultValue;
  }
}
```

**Warning signs:**
- Console errors about localStorage
- View toggle works but doesn't persist
- JavaScript breaks in private browsing mode

**Source:** [localStorage Best Practices](https://blog.logrocket.com/localstorage-javascript-complete-guide/)

### Pitfall 5: Grid Cards with Fixed Heights
**What goes wrong:** Setting fixed height on grid cards creates awkward whitespace when content is shorter or cuts off content when longer.

**Why it happens:** Developers try to make grid look uniform but content varies (title length, metadata presence).

**How to avoid:**
- Let cards grow naturally based on content (`flex: 1` on content area)
- Use `aspect-ratio` on images only, not cards
- Use `-webkit-line-clamp` to limit title lines if needed, but let card height be flexible
- Accept that masonry-style means varied heights (that's the point)

**Warning signs:**
- Text overflow hidden or ellipsis on multi-line content
- Large gaps inside cards
- Cards all same height despite varied content

## Code Examples

Verified patterns from official sources:

### Responsive Grid Container
```css
/* Source: https://css-tricks.com/auto-sizing-columns-css-grid-auto-fill-vs-auto-fit/ */
/* Source: https://web.dev/patterns/layout/aspect-ratio-image-card */

.articles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1rem;
  /* Automatically creates:
     - 1 column on mobile (< 280px + gap)
     - 2 columns on tablet (~600px)
     - 3 columns on small desktop (~900px)
     - 4 columns on large desktop (~1200px+)
     Without any media queries!
  */
}
```

### Grid Card Layout
```css
/* Source: https://web.dev/patterns/layout/aspect-ratio-image-card */

/* List view: horizontal layout (default) */
.article-card {
  display: flex;
  flex-direction: row;
  align-items: flex-start;
  gap: 0.75rem;
  /* ... existing styles ... */
}

/* Grid view: vertical layout */
.articles-grid .article-card {
  display: flex;
  flex-direction: column;
  align-items: stretch;
}

/* Grid view: full-width thumbnail with aspect ratio */
.articles-grid .article-thumbnail {
  width: 100%;
  height: auto;
  aspect-ratio: 16 / 9;
  object-fit: cover;
}

/* Grid view: multi-line title with clamp */
.articles-grid .article-title {
  white-space: normal;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
}

/* Grid view: let content area grow */
.articles-grid .article-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}
```

### View Toggle with Persistence
```javascript
/* Source: Existing pattern from templates/base.gohtml theme toggle */

(function() {
  var viewRadios = document.querySelectorAll('input[name="view"]');
  var stored = localStorage.getItem('view') || 'list';
  var container = document.getElementById('articles-container');

  // Set initial checked state
  var initial = document.getElementById('view-' + stored);
  if (initial) initial.checked = true;

  // Apply initial view class
  if (stored === 'grid') {
    container.classList.add('articles-grid');
  }

  function updateView(value) {
    localStorage.setItem('view', value);
    container.classList.toggle('articles-grid', value === 'grid');
  }

  viewRadios.forEach(function(r) {
    r.addEventListener('change', function(e) {
      updateView(e.target.value);
    });
  });
})();
```

### View Toggle HTML (Segmented Control)
```html
<!-- Source: https://primer.style/components/segmented-control/ (accessibility pattern) -->
<!-- Source: Existing theme-toggle pattern in article-list.gohtml -->

<div class="view-toggle" role="group" aria-label="View mode">
  <input type="radio" name="view" id="view-list" value="list">
  <label for="view-list" title="List view">
    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"
         fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <line x1="8" y1="6" x2="21" y2="6"/>
      <line x1="8" y1="12" x2="21" y2="12"/>
      <line x1="8" y1="18" x2="21" y2="18"/>
      <line x1="3" y1="6" x2="3.01" y2="6"/>
      <line x1="3" y1="12" x2="3.01" y2="12"/>
      <line x1="3" y1="18" x2="3.01" y2="18"/>
    </svg>
    <span class="visually-hidden">List</span>
  </label>
  <input type="radio" name="view" id="view-grid" value="grid">
  <label for="view-grid" title="Grid view">
    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"
         fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <rect x="3" y="3" width="7" height="7"/>
      <rect x="14" y="3" width="7" height="7"/>
      <rect x="3" y="14" width="7" height="7"/>
      <rect x="14" y="14" width="7" height="7"/>
    </svg>
    <span class="visually-hidden">Grid</span>
  </label>
</div>
```

### Progressive Enhancement for Future Native Masonry
```css
/* Source: https://developer.mozilla.org/en-US/docs/Web/CSS/Guides/Grid_layout/Masonry_layout */

/* Default: CSS Grid with auto-fit (works everywhere) */
.articles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1rem;
}

/* Future: Native masonry when supported */
@supports (grid-template-rows: masonry) {
  .articles-grid {
    grid-template-rows: masonry;
    /* This will pack items more tightly when browsers support it */
    /* As of Feb 2026: Firefox and Safari TP only */
  }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Masonry.js library (24KB) | CSS Grid auto-fit | 2021+ (Grid support mature) | No JavaScript required, faster, more maintainable |
| Media query breakpoints | Container-based Grid | 2016+ (Grid adoption) | Truly responsive to container not viewport |
| Padding-top aspect ratio hack | CSS `aspect-ratio` property | 2021 (widespread support) | Cleaner code, better semantics |
| `grid-template-rows: masonry` | `display: grid-lanes` | Jan 2025 (CSSWG consensus) | Final syntax changed, but still experimental |
| Fixed breakpoints (320/768/1024) | Content-driven breakpoints | 2023+ (foldables, varied devices) | More flexible, works on unknown devices |

**Deprecated/outdated:**
- **Masonry.js:** Last commit 2017, 24KB, unnecessary with CSS Grid
- **Isotope:** jQuery-dependent (jQuery itself falling out of favor), performance overhead
- **CSS Columns for clickable content:** Accessibility issues with reading order and tab navigation
- **Padding-top percentage hack:** Replaced by native `aspect-ratio` property (supported since 2021)
- **`grid-template-rows: masonry` syntax:** Replaced by `display: grid-lanes` in Jan 2025 CSSWG decision

## Open Questions

Things that couldn't be fully resolved:

1. **Native CSS Masonry Timeline**
   - What we know: CSS Working Group decided on `display: grid-lanes` syntax in Jan 2025. Firefox and Safari Technology Preview have experimental support. Chrome has experimental support in Chrome 140+ behind flags.
   - What's unclear: When will this reach stable browsers across the board? Mid-2026 is realistic estimate but no confirmed dates.
   - Recommendation: Implement CSS Grid auto-fit now with `@supports` query for progressive enhancement. Grid layout provides 80% of masonry benefits without experimental features.

2. **Optimal Minimum Card Width**
   - What we know: `minmax(280px, 1fr)` is a reasonable starting point for content-rich cards with thumbnails, titles, and metadata.
   - What's unclear: Actual optimal value depends on real content—title length distribution, thumbnail presence, metadata density.
   - Recommendation: Start with 280px, test with production data, adjust based on how cards look when content varies. Could be anywhere from 240px (compact) to 320px (spacious).

3. **Grid vs List Default**
   - What we know: Current UI is list view. Most RSS readers default to list view (Feedly, Inoreader, NetNewsWire).
   - What's unclear: User preference data—would grid view be more popular if offered?
   - Recommendation: Default to list view (current behavior), let users opt into grid. Consider analytics post-launch to see adoption.

4. **HTMX Interaction with View Toggle**
   - What we know: HTMX replaces `#main-content` during sync/filter operations. View toggle targets `#articles-container` (child of main-content).
   - What's unclear: Will HTMX swap preserve the articles-grid class, or will it be lost on update?
   - Recommendation: Test thoroughly. May need to apply class to a wrapper outside HTMX swap target, or reapply class after htmx:afterSwap event. Alternatively, move class toggle to parent element that persists across swaps.

## Sources

### Primary (HIGH confidence)
- [MDN: CSS Grid Masonry Layout](https://developer.mozilla.org/en-US/docs/Web/CSS/Guides/Grid_layout/Masonry_layout) - Native masonry syntax and browser support
- [CSS-Tricks: Auto-Fill vs Auto-Fit](https://css-tricks.com/auto-sizing-columns-css-grid-auto-fill-vs-auto-fit/) - Grid column sizing behavior
- [Bootstrap Stretched Link](https://getbootstrap.com/docs/5.3/helpers/stretched-link/) - Z-index and positioning patterns
- [web.dev: Aspect Ratio Image Card](https://web.dev/patterns/layout/aspect-ratio-image-card) - Modern aspect ratio patterns
- [Primer: SegmentedControl Accessibility](https://primer.style/product/components/segmented-control/accessibility/) - Accessibility patterns for toggle controls

### Secondary (MEDIUM confidence)
- [Piccalilli: Simple Masonry-Like Layout](https://piccalil.li/blog/a-simple-masonry-like-composable-layout/) - CSS-only alternatives to native masonry
- [Chrome Developers: CSS Masonry Update](https://developer.chrome.com/blog/masonry-update) - Browser vendor perspectives and timeline
- [CSS-Tricks: Masonry Layout is Now grid-lanes](https://css-tricks.com/masonry-layout-is-now-grid-lanes/) - Syntax evolution and CSSWG decision
- [LogRocket: localStorage Complete Guide](https://blog.logrocket.com/localstorage-javascript-complete-guide/) - localStorage best practices
- [BrowserStack: Responsive Design Breakpoints 2026](https://www.browserstack.com/guide/responsive-design-breakpoints) - Modern breakpoint strategies

### Secondary (MEDIUM confidence - WebSearch verified)
- [W3C CSSWG Issue #5675](https://github.com/w3c/csswg-drafts/issues/5675) - Masonry accessibility concerns with reading order
- [Smashing Magazine: Native CSS Masonry](https://www.smashingmagazine.com/native-css-masonry-layout-css-grid/) - Implementation patterns and browser status
- Multiple icon sources (Noun Project, Icons8, Flaticon, Bootstrap Icons) - SVG icon resources for list/grid toggle

### Tertiary (LOW confidence - patterns from WebSearch)
- Various CodePen examples for view toggle animations - reference only, not used for technical decisions
- Medium articles on localStorage patterns - cross-referenced with official docs

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - CSS Grid auto-fit is well-established since 2017, widely supported, documented by MDN and CSS-Tricks
- Architecture: HIGH - Patterns reuse existing codebase patterns (theme toggle, stretched-link), verified with official Bootstrap and MDN docs
- Pitfalls: MEDIUM-HIGH - Accessibility concerns documented in W3C issues, stretched-link documented by Bootstrap, auto-fit/auto-fill clarified by CSS-Tricks
- Native masonry timeline: LOW - Experimental feature with no firm production timeline, based on blog posts and browser previews

**Research date:** 2026-02-03
**Valid until:** 2026-04-03 (60 days - CSS Grid is stable, but native masonry is evolving)
