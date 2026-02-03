# Feature Landscape: RSS Reader UI Polish & Search

**Domain:** RSS reader web application
**Milestone:** v1.1 - UI Polish & Search
**Researched:** 2026-02-03
**Confidence:** HIGH

## Context

BlogWatcher UI v1.0 shipped with:
- List view article cards with favicons, titles, time ago
- Read/unread management
- Blog and status filtering (sidebar)
- Three-way theme toggle

v1.1 adds:
- Masonry layout option
- Clickable card interactions
- Article thumbnails
- Title search
- Date filtering
- Combined filters

## Table Stakes

Features users expect in modern RSS readers. Missing = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **Entire card clickable** | Users expect the full card area to be a click target, not just title text | LOW | Research shows removing "read more" links in favor of full card clickability improves UX. Make entire card a single link with hover states. |
| **Hover/focus states on cards** | Visual feedback confirms clickability | LOW | Border/shadow change on hover. Keyboard focus ring for accessibility. |
| **Thumbnail fallback chain** | Articles without featured images should gracefully degrade to site favicon | MEDIUM | RSS media:thumbnail → Open Graph image → Favicon. Already have favicon from v1.0. |
| **Search debouncing** | Real-time search that doesn't hammer the server on every keystroke | LOW | 300-350ms debounce is industry standard. Prevents lag and unnecessary requests. |
| **Clear search results count** | "Showing 47 articles" gives context | LOW | Display result count. Show "No results" state when empty. |
| **Persistent filter state** | Selected filters stay when navigating away and back | LOW | Already partially handled by HTMX - ensure search/date persist on partial updates. |
| **Mobile-responsive masonry** | Masonry must adapt column count to viewport | MEDIUM | Standard: 1 col mobile, 2 col tablet, 3-4 col desktop. Use CSS auto-fit minmax(). |
| **View toggle persistence** | User's choice (list vs masonry) remembered across sessions | LOW | Save in localStorage or cookie. Server should respect preference. |
| **Date filter shortcuts** | Predefined ranges (Last Week, Month) faster than custom range | LOW | Common pattern: Quick options + custom date picker for power users. |
| **Combined filter AND logic** | All active filters apply together (blog + unread + date + search) | MEDIUM | Query must combine: WHERE blog_id = X AND is_read = 0 AND title LIKE '%search%' AND published_date > Y |

## Differentiators

Features that set products apart. Not expected, but highly valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Saved searches** | Power users can bookmark complex filter combinations | MEDIUM | FreshRSS pattern: Save queries as URLs. Could add "Save current filters" button → named bookmark. Out of scope for v1.1 but low-hanging fruit for v1.2. |
| **Masonry with mixed widths** | Some cards wider (2 col span) for featured articles | HIGH | Requires grid-column: span 2 logic and detecting "featured" articles. V1.1 should do uniform masonry first. |
| **Thumbnail position options** | Let user choose: left, top, or no thumbnails | MEDIUM | Feedly has this. Could use CSS classes: card--thumb-left, card--thumb-top. V1.2 material. |
| **Regex search** | Power users love /pattern/ syntax for complex queries | MEDIUM | FreshRSS supports this. SQLite REGEXP needs custom function in Go. Defer to v1.2+. |
| **Keyboard shortcuts** | j/k navigation, x to mark read | MEDIUM | Requires minimal JS for keypress handling. Good v1.2 feature. |
| **Estimated read time** | "3 min read" helps prioritize | MEDIUM | Requires fetching article content or using word count heuristics. Out of scope (would need content storage). |
| **Search highlighting** | Highlight search terms in results | LOW | CSS mark tag or background on matched text. Nice polish for v1.1 if time allows. |
| **Filter animation** | Smooth transitions when filters change results | LOW | HTMX swap with CSS transitions. Could enhance perceived performance. |

## Anti-Features

Features to explicitly NOT build. Common mistakes in this domain.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| **Infinite scroll for masonry** | Breaks back button, loses position, bad accessibility | Use pagination or "Load More" button. Preserve scroll position. |
| **Auto-refresh search on every filter change** | Overwhelming when adjusting multiple filters | Hybrid approach: Real-time for search text (debounced), "Apply" button for date/combined filters OR instant updates if fast enough. |
| **CSS-only masonry layout (column-count)** | Breaks card reading order (flows vertically down columns, not left-to-right) | Use CSS Grid with masonry (progressive enhancement) or JS library. Column-count is wrong flow. |
| **Image-heavy masonry without lazy loading** | Slow page loads, especially on mobile | Defer images below fold. Use loading="lazy" attribute. |
| **Complex filter UI** | Dropdowns, multi-selects, accordions overwhelm users | Keep filters simple: Sidebar for blogs, inline search + date + view toggle. Maximum 2 clicks to any filter. |
| **Search that requires Enter key** | Friction. Users expect instant results (with debouncing) | Real-time debounced search. Enter key optional (can trigger early). |
| **Masonry without aspect ratio hints** | Layout shift as images load (CLS) | Reserve space with aspect-ratio CSS or explicit height/width. Use placeholders. |
| **Full article preview on hover** | Requires fetching content. Slow. Breaks "open in new tab" model. | Stick to metadata (title, excerpt if available, thumbnail). External links only. |

## Feature Dependencies

```
Existing v1.0 Features
├── Article cards with metadata
│   ├── → Clickable cards (wraps existing card in <a>)
│   └── → Thumbnail support (add image, keep favicon fallback)
├── List view layout
│   └── → Masonry layout (alternative CSS Grid rendering)
├── Blog filter (sidebar)
│   └── → Combined filters (AND logic with search + date)
└── HTMX partial updates
    └── → Search with debouncing (HTMX with hx-trigger="keyup changed delay:350ms")

New v1.1 Features
├── Masonry layout
│   ├── Requires: Viewport breakpoints
│   └── Requires: View toggle (list vs masonry)
├── Clickable cards
│   ├── Requires: Hover states
│   ├── Requires: Keyboard focus styles
│   └── Requires: Accessibility (ARIA labels)
├── Thumbnail support
│   ├── Requires: Database column (thumbnail_url)
│   ├── Requires: Fallback chain logic
│   └── Requires: Lazy loading
├── Title search
│   ├── Requires: Debouncing
│   ├── Requires: Results count
│   └── Requires: No results state
├── Date filtering
│   ├── Requires: Quick shortcuts (Last Week, Month)
│   ├── Requires: Custom date range picker
│   └── Requires: Clear date filter button
└── Combined filters
    ├── Requires: AND logic in SQL queries
    └── Requires: Filter state management
```

## MVP Recommendation for v1.1

Prioritize in this order:

### Phase 1: Enhanced Card Interaction (HIGH VALUE, LOW EFFORT)
1. **Entire card clickable** - Wrap card in <a>, add hover states
2. **Keyboard accessibility** - Focus rings, proper ARIA labels
3. **Thumbnail support** - Add thumbnail_url column, render images with fallback

**Rationale:** Biggest UX improvement for least effort. Makes existing list view feel modern.

### Phase 2: Search & Date Filtering (HIGH VALUE, MEDIUM EFFORT)
1. **Title search with debouncing** - HTMX hx-trigger with 350ms delay
2. **Results count** - Display "X articles" feedback
3. **Date filter shortcuts** - Last Week, Last Month, custom range
4. **Combined filters** - AND logic: blog + status + search + date

**Rationale:** Core functionality for finding articles. Search is table stakes for any modern reader.

### Phase 3: Masonry Layout (MEDIUM VALUE, MEDIUM EFFORT)
1. **View toggle** - Button to switch list ↔ masonry
2. **CSS Grid masonry** - Progressive enhancement with @supports
3. **Responsive breakpoints** - 1/2/3 columns based on viewport
4. **Persistence** - Remember user's view preference

**Rationale:** Nice visual improvement but less critical than search. Can be v1.1 stretch goal or v1.2.

### Defer to Post-v1.1

- **Saved searches** - Nice-to-have, not blocking
- **Keyboard shortcuts** - Power user feature, needs JS
- **Search highlighting** - Visual polish, low priority
- **Regex search** - Power user feature, niche use case
- **Mixed width masonry** - Complex, diminishing returns

## Implementation Notes

### Masonry Layout: Browser Support Reality Check

**CSS Grid Level 3 masonry is EXPERIMENTAL (2026)**
- Only Firefox Nightly with flag
- Not production-ready
- Spec still being debated (masonry vs grid-lanes syntax)

**Recommendation for v1.1:**
```css
/* Use CSS Grid auto-fit as "masonry-like" layout */
.masonry-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1rem;
  grid-auto-rows: auto; /* NOT true masonry, but good enough */
}

/* Progressive enhancement if native masonry ships */
@supports (grid-template-rows: masonry) {
  .masonry-grid {
    grid-template-rows: masonry;
  }
}
```

**Alternative:** Use lightweight JS library (Masonry.js) if true packed layout required. But CSS Grid auto-fit is 90% there without the complexity.

### Thumbnail Fallback Strategy

```go
// Pseudo-code for thumbnail resolution
func GetArticleThumbnail(article Article) string {
    // 1. RSS media:thumbnail or enclosure
    if article.ThumbnailURL != "" {
        return article.ThumbnailURL
    }

    // 2. Open Graph image (requires fetching article HTML)
    // Only fetch on-demand, cache result
    if ogImage := FetchOGImage(article.URL); ogImage != "" {
        return ogImage
    }

    // 3. Favicon (already have from v1.0)
    return article.Blog.FaviconURL
}
```

**Complexity consideration:** Fetching OG images adds HTTP requests. Options:
- **Eager:** Fetch during sync (slower sync, but thumbnails ready)
- **Lazy:** Fetch on first view (faster sync, slower first render)
- **Skip OG:** RSS thumbnail → Favicon only (simplest, 80% coverage)

**Recommendation:** Start with RSS → Favicon for v1.1. Add OG fetching in v1.2 if needed.

### Search Debouncing with HTMX

```html
<input
  type="search"
  name="q"
  hx-get="/articles"
  hx-trigger="keyup changed delay:350ms, search"
  hx-target="#article-list"
  hx-indicator="#search-spinner"
  placeholder="Search articles..."
/>
```

**Key points:**
- `delay:350ms` = debounce timer
- `search` event = triggered by clear button or Enter key
- `changed` = only fires if value actually changed (avoids duplicate requests)

### Accessibility for Clickable Cards

**Problem:** Entire card is a link, but screenreaders need meaningful text.

**Solution:**
```html
<a href="/article/123" class="article-card" aria-labelledby="article-title-123">
  <img src="thumbnail.jpg" alt="" loading="lazy" /> <!-- decorative, empty alt -->
  <div>
    <h2 id="article-title-123">Article Title Here</h2>
    <p>Blog Name · 2 hours ago</p>
  </div>
</a>
```

**Do NOT:**
- Nest links inside cards (link within link is invalid HTML)
- Use tabindex=-1 on card children (breaks keyboard nav)
- Add redundant "Read more" text

### Combined Filters: SQL Query Building

```go
func BuildArticleQuery(filters ArticleFilters) (query string, args []interface{}) {
    query = "SELECT * FROM articles WHERE 1=1"

    if filters.BlogID != 0 {
        query += " AND blog_id = ?"
        args = append(args, filters.BlogID)
    }

    if filters.Status == "unread" {
        query += " AND is_read = 0"
    } else if filters.Status == "read" {
        query += " AND is_read = 1"
    }

    if filters.Search != "" {
        query += " AND title LIKE ?"
        args = append(args, "%"+filters.Search+"%")
    }

    if !filters.DateFrom.IsZero() {
        query += " AND published_date >= ?"
        args = append(args, filters.DateFrom)
    }

    query += " ORDER BY published_date DESC"
    return query, args
}
```

**Test edge cases:**
- All filters active (blog + unread + search + date)
- No filters (show all)
- Search with no results
- Invalid date ranges

## Complexity Assessment

| Feature | Complexity | Estimated Effort | Blockers |
|---------|-----------|------------------|----------|
| Clickable cards | LOW | 2 hours | None - pure CSS + template change |
| Hover/focus states | LOW | 1 hour | None |
| Thumbnail support | MEDIUM | 4-6 hours | Need thumbnail_url column, fallback logic |
| Search with debouncing | LOW | 2-3 hours | None - HTMX built-in |
| Results count | LOW | 1 hour | None |
| Date filter shortcuts | MEDIUM | 4 hours | UI for date picker, SQL date comparison |
| Combined filters | MEDIUM | 3-4 hours | Query building logic, state management |
| View toggle | LOW | 2 hours | Persistence in localStorage |
| Masonry layout (CSS Grid) | MEDIUM | 4-6 hours | Browser testing, responsive breakpoints |
| Lazy loading images | LOW | 1 hour | Native loading="lazy" |

**Total estimated effort for all v1.1 features:** 24-30 hours

**MVP (Phases 1-2 only):** 14-18 hours

## Testing Checklist

### Clickable Cards
- [ ] Entire card area clickable (not just title)
- [ ] Hover state changes on mouse over
- [ ] Focus ring visible on keyboard tab
- [ ] Opens in new tab (target="_blank")
- [ ] Screenreader announces title correctly

### Masonry Layout
- [ ] Columns adapt to viewport (1/2/3 based on width)
- [ ] No horizontal overflow on mobile
- [ ] Cards maintain readable proportions
- [ ] Images load without layout shift (CLS)
- [ ] View toggle persists across page loads

### Search
- [ ] Debounces after 350ms of typing
- [ ] Shows results count
- [ ] Shows "No results" state
- [ ] Clears with X button
- [ ] Works with Enter key
- [ ] Persists when navigating (HTMX state)

### Date Filtering
- [ ] "Last Week" shortcut works correctly
- [ ] "Last Month" shortcut works correctly
- [ ] Custom date range accepts valid dates
- [ ] Invalid dates show error
- [ ] Clear button resets date filter

### Combined Filters
- [ ] All filters apply together (AND logic)
- [ ] Blog + unread + search works
- [ ] Blog + search + date range works
- [ ] All filters + search works
- [ ] Clearing one filter updates results
- [ ] Clearing all filters shows full list

### Thumbnails
- [ ] RSS thumbnail displays if present
- [ ] Falls back to favicon if no thumbnail
- [ ] Lazy loads below fold
- [ ] Loading="lazy" prevents eager fetching
- [ ] Alt text empty (decorative image)

## Sources

### Masonry Layout Research
- [Masonry layout - CSS | MDN](https://developer.mozilla.org/en-US/docs/Web/CSS/Guides/Grid_layout/Masonry_layout)
- [Native CSS Masonry Layout In CSS Grid — Smashing Magazine](https://www.smashingmagazine.com/native-css-masonry-layout-css-grid/)
- [Responsive Grid for FreshRSS Reading View Mode – Alaya·Techne](https://techne.alaya.net/responsive-grid-for-freshrss-reading-view/)
- [Create a Responsive Masonry Layout with HTML, CSS & JavaScript](https://codeshack.io/create-responsive-masonry-layout-html-css-javascript/)
- [CSS - Implementing Responsive Masonry Layouts](https://blog.openreplay.com/css--implementing-responsive-masonry-layouts/)

### RSS Reader UX Patterns
- [The 3 best RSS reader apps in 2026 | Zapier](https://zapier.com/blog/best-rss-feed-reader-apps/)
- [Best RSS Feed Readers 2026: Complete Comparison & Review Guide](https://vpntierlists.com/blog/best-rss-feed-readers-2025-complete-comparison-guide/)
- [RSS Reader User Interface Design Principles](https://www.feedviewer.app/answers/rss-reader-user-interface-design-principles)
- [Are "Read More" Links Important to Help Users Click Blog Posts and Articles?](https://sparkbox.com/foundry/are_read_more_links_necessary_easier_to_use_best_article_listing_layout_first_click_test_usibility_ux_research)

### Search & Filtering
- [Filtering articles · FreshRSS](https://freshrss.github.io/FreshRSS/en/users/10_filter.html)
- [Filter UX Design Patterns & Best Practices - Pencil & Paper](https://www.pencilandpaper.io/articles/ux-pattern-analysis-enterprise-filtering)
- [Filtering UX — Smart Interface Design Patterns](https://smart-interface-design-patterns.com/articles/filtering-ux/)
- [What is a Good Debounce Time for Search?](https://www.byteplus.com/en/topic/498848)
- [Master Search UX in 2026: Best Practices, UI Tips & Design Patterns](https://www.designmonks.co/blog/search-ux-best-practices)

### Thumbnail Implementation
- [How to Set Up the Image Thumbnail Options for E&T - WP RSS Aggregator Knowledge Base](https://kb.wprssaggregator.com/article/341-how-to-set-up-the-image-thumbnail-options-for-excerpts-thumbnails)
- [Fallback Image in Feedzy RSS Feeds - Themeisle Docs](https://docs.themeisle.com/article/1964-fallback-image-in-feedzy-rss-feeds)

### Accessibility
- [Accessible card UI component patterns | Digital Accessibility](https://dap.berkeley.edu/web-a11y-basics/accessible-card-ui-component-patterns)
- [Cards - Inclusive Components](https://inclusive-components.design/cards/)
- [ARIA - Accessibility | MDN](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA)

### Performance
- [Brick by brick: Help us build CSS Masonry | Blog | Chrome for Developers](https://developer.chrome.com/blog/masonry-update)
- [The Evolution of Web Layout: A Look at CSS Masonry and the Future of the Web | MiniFyn Blog](https://www.minifyn.com/blog/the-evolution-of-web-layout-a-look-at-css-masonry-and-the-future-of-the-web)
