---
phase: 04-article-management
plan: 01
subsystem: scanner-infrastructure
tags: [rss, scraper, scanner, database, sync]
dependency-graph:
  requires: [03-article-display]
  provides: [scanner-packages, database-write-methods]
  affects: [04-02-sync-handler, 04-03-mark-read-unread]
tech-stack:
  added: [github.com/mmcdole/gofeed, github.com/PuerkitoBio/goquery]
  patterns: [rss-feed-parsing, html-scraping, bulk-insert, url-deduplication]
key-files:
  created:
    - internal/rss/rss.go
    - internal/scraper/scraper.go
    - internal/scanner/scanner.go
  modified:
    - internal/storage/database.go
    - go.mod
    - go.sum
decisions:
  - Copy and adapt from reference codebase rather than rewrite
  - Update import paths to github.com/esttorhe/blogwatcher-ui
  - Keep parallel worker support in ScanAllBlogs for performance
metrics:
  duration: ~3 minutes
  completed: 2026-02-02
---

# Phase 04 Plan 01: Scanner Infrastructure Summary

Scanner packages copied from reference codebase with updated imports, plus database methods added for write operations.

## What Was Built

### RSS Package (`internal/rss/rss.go`)
- `ParseFeed(feedURL, timeout)` - Parses RSS/Atom feeds and returns articles
- `DiscoverFeedURL(blogURL, timeout)` - Autodiscovers feed URL from blog page
- `IsFeedError(err)` - Checks if error is a feed parsing error
- `FeedArticle` struct for parsed article data

### Scraper Package (`internal/scraper/scraper.go`)
- `ScrapeBlog(blogURL, selector, timeout)` - Scrapes articles using CSS selectors
- `IsScrapeError(err)` - Checks if error is a scraping error
- `ScrapedArticle` struct for scraped article data

### Scanner Package (`internal/scanner/scanner.go`)
- `ScanBlog(db, blog)` - Scans single blog, returns result
- `ScanAllBlogs(db, workers)` - Scans all blogs with parallel workers
- `ScanBlogByName(db, name)` - Scans blog by name lookup
- `ScanResult` struct with BlogName, NewArticles, TotalFound, Source, Error

### Database Methods (`internal/storage/database.go`)
- `GetBlogByName(name)` - Lookup blog by name
- `UpdateBlog(blog)` - Update all blog fields
- `UpdateBlogLastScanned(id, time)` - Update scan timestamp
- `AddArticlesBulk(articles)` - Bulk insert with transaction
- `GetExistingArticleURLs(urls)` - URL deduplication with chunking

### Helper Functions
- `nullIfEmpty(value)` - Returns nil for empty strings
- `formatTimePtr(value)` - Formats time pointer for SQLite
- `interfaceSlice(values)` - Converts string slice to interface slice

## Dependencies Added

```go
github.com/mmcdole/gofeed v1.3.0   // RSS/Atom feed parsing
github.com/PuerkitoBio/goquery v1.10.3  // HTML parsing for scraping
```

## How It Works

1. **Feed Discovery**: Scanner first checks if blog has feed_url; if not, discovers via link tags or common paths
2. **RSS First**: Tries RSS feed parsing if feed URL available
3. **Scraper Fallback**: Falls back to HTML scraping if RSS fails and selector configured
4. **Deduplication**: Checks existing URLs in chunks of 900 to avoid SQLite limits
5. **Bulk Insert**: Inserts new articles in single transaction
6. **Timestamp Update**: Updates last_scanned on blog after scan completes

## Deviations from Plan

None - plan executed exactly as written.

## Next Steps

- 04-02: Add sync handler to web UI for triggering blog synchronization
- 04-03: Add mark read/unread API endpoints
