# Phase 5: Theme Toggle - Research

**Researched:** 2026-02-03
**Domain:** CSS theming, user preference persistence, prefers-color-scheme media query
**Confidence:** HIGH

## Summary

This phase implements a three-way theme toggle (Light, Dark, System) that allows users to switch between themes with preference persisted in localStorage. The implementation builds on the existing CSS custom properties architecture established in Phase 2 and follows the project's CSS-first, minimal JavaScript approach.

The standard approach uses:
1. CSS custom properties (already in place) for theme colors
2. `prefers-color-scheme` media query for system preference detection
3. CSS `:has()` selector with radio buttons for CSS-only toggle state management
4. Inline `<script>` in `<head>` to prevent FOUC (Flash of Unstyled Content)
5. Minimal JavaScript for localStorage persistence and system preference change listeners

The existing dark theme CSS variables provide a solid foundation. The light theme will use warm cream tones (not pure white) as specified in CONTEXT.md, following similar design language to Notion's light mode.

**Primary recommendation:** Implement a segmented control using hidden radio buttons with CSS `:has()` for styling, minimal JS for persistence, and inline head script for FOUC prevention.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Technology | Version/Status | Purpose | Why Standard |
|------------|----------------|---------|--------------|
| CSS Custom Properties | Native | Theme color definitions | Already established in Phase 2, standard pattern |
| prefers-color-scheme | Baseline since 2020 | System preference detection | W3C standard, universal browser support |
| CSS :has() | Baseline since Dec 2023 | Toggle state to style mapping | Enables CSS-only styling from input state |
| localStorage | Native | Preference persistence | Simple, synchronous, no dependencies |
| matchMedia() | Native | JS system preference detection | Standard API for media query listening |

### Supporting
| Technology | Purpose | When to Use |
|------------|---------|-------------|
| color-scheme CSS property | Browser form styling | Set on :root for native form element theming |
| meta theme-color | Browser chrome theming | Update URL bar color based on theme |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| CSS :has() | JavaScript class toggling | :has() is CSS-only but needs Dec 2023+ browsers |
| Radio buttons | `<select>` element | Radio buttons allow segmented control styling |
| localStorage | Cookies | localStorage is simpler, synchronous, client-only |

**No additional installation required** - all native browser APIs.

## Architecture Patterns

### Theme Variable Structure
```css
/* Light theme (applied when html.light or system prefers light) */
:root {
  --bg-primary: #FAF8F5;      /* Warm cream */
  --bg-surface: #FFFFFF;       /* White cards */
  --bg-elevated: #F5F3F0;     /* Slightly darker cream */
  --text-primary: #37352F;     /* Dark charcoal (Notion-like) */
  --text-secondary: #6B6B6B;   /* Medium gray */
  --accent: #2563EB;           /* Blue for contrast */
  --border: #E5E3E0;           /* Warm light border */
}

/* Dark theme (default, already exists) */
html.dark {
  --bg-primary: #121212;
  --bg-surface: #1e1e1e;
  --bg-elevated: #2d2d2d;
  --text-primary: #e0e0e0;
  --text-secondary: #a0a0a0;
  --accent: #64b5f6;
  --border: #333333;
}
```

### Pattern 1: Three-Way Toggle with CSS :has()
**What:** Use CSS `:has()` to style based on which radio button is checked
**When to use:** Three-way theme toggle without JavaScript for styling

```html
<!-- Source: MDN, Smashing Magazine pattern -->
<div class="theme-toggle" role="radiogroup" aria-label="Color theme">
  <input type="radio" name="theme" id="theme-light" value="light">
  <label for="theme-light" title="Light theme">
    <svg><!-- sun icon --></svg>
    <span class="visually-hidden">Light</span>
  </label>

  <input type="radio" name="theme" id="theme-system" value="system" checked>
  <label for="theme-system" title="System theme">
    <svg><!-- computer icon --></svg>
    <span class="visually-hidden">System</span>
  </label>

  <input type="radio" name="theme" id="theme-dark" value="dark">
  <label for="theme-dark" title="Dark theme">
    <svg><!-- moon icon --></svg>
    <span class="visually-hidden">Dark</span>
  </label>
</div>
```

```css
/* Source: MDN :has() documentation, Smashing Magazine */
/* Base light theme variables in :root */
:root {
  color-scheme: light dark;
  --bg-primary: #FAF8F5;
  /* ... light theme values ... */
}

/* Dark mode when explicitly selected */
html:has(#theme-dark:checked) {
  --bg-primary: #121212;
  /* ... dark theme values ... */
}

/* Dark mode when system preference and system selected */
@media (prefers-color-scheme: dark) {
  html:has(#theme-system:checked) {
    --bg-primary: #121212;
    /* ... dark theme values ... */
  }
}
```

### Pattern 2: FOUC Prevention with Inline Script
**What:** Restore theme preference before page renders
**When to use:** Always - prevents flash of wrong theme

```html
<!-- Source: CSS-Tricks FART article, web.dev -->
<!-- Place immediately after <head> opening tag -->
<script>
  (function() {
    var theme = localStorage.getItem('theme') || 'system';
    var radio = document.getElementById('theme-' + theme);
    if (radio) radio.checked = true;

    // Apply dark class immediately if needed
    if (theme === 'dark' ||
        (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  })();
</script>
```

### Pattern 3: localStorage Persistence
**What:** Save and restore user preference
**When to use:** On toggle change events

```javascript
// Source: Smashing Magazine
const themeRadios = document.querySelectorAll('input[name="theme"]');

function saveThemePreference(event) {
  localStorage.setItem('theme', event.target.value);
  updateThemeClass(event.target.value);
}

function updateThemeClass(theme) {
  if (theme === 'dark' ||
      (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    document.documentElement.classList.add('dark');
  } else {
    document.documentElement.classList.remove('dark');
  }
}

themeRadios.forEach(radio => {
  radio.addEventListener('change', saveThemePreference);
});

// Listen for system preference changes
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function(e) {
  const currentTheme = localStorage.getItem('theme') || 'system';
  if (currentTheme === 'system') {
    updateThemeClass('system');
  }
});
```

### Anti-Patterns to Avoid
- **Storing colors in localStorage:** Store theme NAME only ('light', 'dark', 'system'), not color values
- **Using setTimeout for FOUC:** Unreliable, still causes flash - use inline blocking script instead
- **Pure white backgrounds:** Use warm cream (#FAF8F5) - easier on eyes, more inviting
- **Forgetting system preference listener:** User may change OS setting while page is open

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| System preference detection | Custom polling | `matchMedia('(prefers-color-scheme: dark)')` | Standard API, includes change events |
| Theme persistence | Cookies or custom storage | `localStorage` | Synchronous, simple, no expiry |
| Segmented control styling | Complex div structures | Hidden radio + label pattern | Native accessibility, form integration |
| Icon toggle animations | Custom JS animation | CSS transitions on transform | Smoother, GPU-accelerated |

**Key insight:** The browser provides everything needed for theme toggling. CSS custom properties + matchMedia + localStorage = complete solution with no dependencies.

## Common Pitfalls

### Pitfall 1: Flash of Wrong Theme (FOUC/FART)
**What goes wrong:** Page loads with light theme, then flashes to dark after JS runs
**Why it happens:** JavaScript runs after initial render; CSS loads before JS
**How to avoid:** Place theme-restoration script in `<head>` before stylesheets, make it blocking (no async/defer)
**Warning signs:** Brief white flash when loading a dark-themed page

### Pitfall 2: Forgetting System Preference Changes
**What goes wrong:** User changes OS theme while page is open, nothing happens
**Why it happens:** Only checked preference on page load, not listening for changes
**How to avoid:** Add `matchMedia.addEventListener('change', callback)` listener
**Warning signs:** Theme doesn't update when toggling OS dark mode

### Pitfall 3: :has() Specificity Conflicts
**What goes wrong:** Theme variables don't apply correctly, wrong colors show
**Why it happens:** :has() selectors may conflict with existing specificity
**How to avoid:** Keep theme CSS in one place, use same specificity pattern throughout
**Warning signs:** Some elements themed correctly, others not

### Pitfall 4: Inaccessible Toggle Control
**What goes wrong:** Screen readers announce nothing or incorrect state
**Why it happens:** Using div/spans instead of semantic inputs, missing labels
**How to avoid:** Use native radio inputs with proper labels, add role="radiogroup"
**Warning signs:** VoiceOver/NVDA don't announce theme changes

### Pitfall 5: Pure White Light Theme
**What goes wrong:** Light theme feels harsh, users prefer dark mode
**Why it happens:** Using #FFFFFF for backgrounds
**How to avoid:** Use warm cream tones (#FAF8F5, #FBF7F5) for backgrounds
**Warning signs:** User feedback about "harsh" or "too bright" light mode

## Code Examples

Verified patterns from official sources:

### Complete Segmented Control CSS
```css
/* Source: segmented-control-css pattern, adapted for theme toggle */
.theme-toggle {
  display: inline-flex;
  background-color: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 4px;
  gap: 4px;
}

.theme-toggle input[type="radio"] {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
  pointer-events: none;
}

.theme-toggle label {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 6px 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s ease, color 0.2s ease;
  color: var(--text-secondary);
}

.theme-toggle label:hover {
  background-color: var(--bg-elevated);
}

.theme-toggle input[type="radio"]:checked + label {
  background-color: var(--bg-elevated);
  color: var(--text-primary);
}

.theme-toggle input[type="radio"]:focus-visible + label {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.theme-toggle svg {
  width: 18px;
  height: 18px;
}
```

### System Preference Detection (JavaScript)
```javascript
// Source: MDN matchMedia documentation
const darkModeQuery = window.matchMedia('(prefers-color-scheme: dark)');

// Check current preference
if (darkModeQuery.matches) {
  console.log('User prefers dark mode');
}

// Listen for changes
darkModeQuery.addEventListener('change', (e) => {
  if (e.matches) {
    // Switched to dark mode
  } else {
    // Switched to light mode
  }
});
```

### Warm Cream Light Theme Colors
```css
/* Source: Notion color palette, web.dev best practices */
:root {
  /* Warm cream backgrounds - not pure white */
  --bg-primary: #FAF8F5;       /* Main background - warm off-white */
  --bg-surface: #FFFFFF;        /* Card surfaces - white */
  --bg-elevated: #F5F3F0;       /* Hover states - slightly darker cream */

  /* Text colors - not pure black */
  --text-primary: #37352F;      /* Main text - Notion dark gray */
  --text-secondary: #6B6B6B;    /* Secondary text - medium gray */

  /* Accent and borders */
  --accent: #2563EB;            /* Blue accent - good contrast on cream */
  --border: #E5E3E0;            /* Warm light border */
}
```

### Meta Tag for Theme Color
```html
<!-- Source: web.dev prefers-color-scheme article -->
<meta name="theme-color" content="#121212" media="(prefers-color-scheme: dark)">
<meta name="theme-color" content="#FAF8F5" media="(prefers-color-scheme: light)">
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| JS class toggling | CSS :has() selector | Dec 2023 | CSS-only styling possible |
| Single theme meta | Media query meta tags | Chromium 93, Safari 15 | Dynamic browser chrome theming |
| @media queries only | light-dark() function | May 2024 | Cleaner inline color declaration |
| Custom event handling | matchMedia.addEventListener | Well established | Standard event pattern |

**Available but not used:**
- `light-dark()` CSS function: Available since May 2024, but using CSS :has() pattern is more explicit and compatible with the three-way toggle requirement

**Deprecated/outdated:**
- `matchMedia.addListener()`: Deprecated, use `addEventListener('change', callback)` instead

## Open Questions

Things that couldn't be fully resolved:

1. **Icon choices for toggle**
   - What we know: Need sun, computer/monitor, moon icons
   - What's unclear: Whether to use emoji, inline SVG, or icon font
   - Recommendation: Use inline SVG for consistency and color control

2. **Transition animation between themes**
   - What we know: CSS transitions on color properties work well
   - What's unclear: Optimal duration and which properties to animate
   - Recommendation: 200ms transition on background-color, color, border-color; respect prefers-reduced-motion

## Sources

### Primary (HIGH confidence)
- [MDN prefers-color-scheme](https://developer.mozilla.org/en-US/docs/Web/CSS/Reference/At-rules/@media/prefers-color-scheme) - Media query syntax, matchMedia API
- [MDN light-dark()](https://developer.mozilla.org/en-US/docs/Web/CSS/Reference/Values/color_value/light-dark) - CSS function availability
- [MDN :has()](https://developer.mozilla.org/en-US/docs/Web/CSS/:has) - Browser support (Dec 2023 baseline)
- [web.dev prefers-color-scheme](https://web.dev/articles/prefers-color-scheme) - color-scheme property, best practices

### Secondary (MEDIUM confidence)
- [Smashing Magazine - Color Scheme Preferences](https://www.smashingmagazine.com/2024/03/setting-persisting-color-scheme-preferences-css-javascript/) - :has() + select pattern, localStorage persistence
- [CSS-Tricks FART article](https://css-tricks.com/flash-of-inaccurate-color-theme-fart/) - FOUC prevention inline script technique
- [Notion color codes](https://www.notionavenue.co/post/notion-color-code-hex-palette) - Warm color palette reference
- [Super.so Notion colors](https://docs.super.so/notion-colors) - Additional Notion palette data
- [segmented-control-css](https://github.com/basilebong/segmented-control-css) - CSS-only segmented control pattern
- [Sara Soueidan - Inclusive checkboxes](https://www.sarasoueidan.com/blog/inclusively-hiding-and-styling-checkboxes-and-radio-buttons/) - Accessible hidden input pattern

### Tertiary (LOW confidence)
- Various CodePen examples for segmented control patterns

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Native browser APIs, well-documented standards
- Architecture: HIGH - Multiple official sources confirm patterns
- Pitfalls: HIGH - Well-documented in multiple authoritative articles
- Light theme colors: MEDIUM - Based on Notion palette extrapolation

**Research date:** 2026-02-03
**Valid until:** 2026-03-03 (30 days - stable CSS/JS standards)
