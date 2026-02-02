# Phase 3 Plan 02: Article Card Templates Summary

**One-liner:** Article cards with favicons, titles, blog names, and relative timestamps using ListArticlesWithBlog

## Metadata

| Field | Value |
|-------|-------|
| Phase | 03-article-display |
| Plan | 02 |
| Duration | 2 minutes |
| Completed | 2026-02-02 |
| Status | Complete |

## What Was Built

### Handlers Updated (internal/server/handlers.go)

Both `handleIndex` and `handleArticleList` now use `ListArticlesWithBlog` instead of `ListArticlesByReadStatus`:

```go
// Before
var articles []model.Article
articles, err = s.db.ListArticlesByReadStatus(true, blogID)

// After
var articles []model.ArticleWithBlog
articles, err = s.db.ListArticlesWithBlog(true, blogID)
```

This provides templates with `BlogName` and `BlogURL` fields from the JOIN query.

### Article Card Template (templates/partials/article-list.gohtml)

New card structure with rich metadata:

```gohtml
<article class="article-card">
    <img class="article-favicon" src="{{faviconURL .BlogURL}}" ...>
    <div class="article-content">
        <a href="{{.URL}}" target="_blank" rel="noopener noreferrer" class="article-title">
            {{.Title}}
        </a>
        <div class="article-meta">
            <span class="article-source">{{.BlogName}}</span>
            <span class="article-time">{{timeAgo .PublishedDate}}</span>
        </div>
    </div>
</article>
```

Key features:
- `faviconURL` function generates Google S2 favicon URLs
- `timeAgo` function renders relative timestamps ("1 week ago")
- `rel="noopener noreferrer"` on all external links
- `onerror` handler gracefully hides broken favicons

### Article Card CSS (static/styles.css)

Flexbox layout with proper structure:

```css
.article-card {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  /* ... */
}

.article-favicon {
  flex-shrink: 0;
  width: 32px;
  height: 32px;
}

.article-content {
  flex: 1;
  min-width: 0;  /* Enables text truncation */
}

.article-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
```

## Commits

| Hash | Type | Description |
|------|------|-------------|
| 231bddc | feat | Update handlers to use ListArticlesWithBlog |
| aafd760 | feat | Update article-list template with card layout |
| cc1df00 | style | Add article card CSS with flexbox layout |

## Files Changed

### Created
- None

### Modified
- `internal/server/handlers.go` - Use ListArticlesWithBlog for rich article data
- `templates/partials/article-list.gohtml` - Article card template with metadata
- `static/styles.css` - Article card CSS with flexbox layout

## Verification Results

| Check | Result |
|-------|--------|
| `go build ./...` | Pass |
| `go vet ./...` | Pass |
| Server starts | Pass |
| Article cards render | Pass |
| Favicons display | Pass |
| Relative times shown | Pass |
| Inbox/Archived filter | Pass |
| Blog filter | Pass |

## Screenshot

Visual verification captured at `.claude-visual/03-02-article-cards.png` showing:
- Article cards with favicons from Google S2
- Titles displayed prominently
- Blog name ("Maggie Appleton") below title
- Relative timestamps ("1 week ago", "4 weeks ago", etc.) with dot separator
- Dark theme styling

## Deviations from Plan

None - plan executed exactly as written.

## Success Criteria Met

- [x] User sees article cards with favicon, title, blog name, and relative time
- [x] Clicking article opens original URL in new browser tab
- [x] Articles display correctly in both Inbox and Archived views
- [x] All external links include rel="noopener noreferrer" for security

## Phase 3 Status

| Plan | Name | Status |
|------|------|--------|
| 03-01 | Template Functions | Complete |
| 03-02 | Article Card Templates | Complete |

**Phase 3 Complete** - Article display with rich metadata fully implemented.

## Next Steps

Phase 4: Article Management (mark read/unread functionality)
