# Phase 7: Search & Date Filtering - Research

**Researched:** 2026-02-03
**Domain:** Full-text search (SQLite FTS5) + Date filtering + HTMX active search
**Confidence:** HIGH

## Summary

Phase 7 implements title search and date filtering for articles. The standard approach uses SQLite's built-in FTS5 (Full-Text Search 5) extension for search, combined with date range filtering in SQL WHERE clauses. HTMX provides native debounce support via `hx-trigger` with delay modifiers, eliminating need for custom JavaScript. HTML5's native `<input type="date">` provides adequate date picking for simple ranges without dependencies.

FTS5 is already available in modernc.org/sqlite (the project's current SQLite driver) without additional setup. The recommended pattern uses "external content tables" where FTS5 maintains a separate index synchronized via triggers, avoiding data duplication while enabling efficient full-text search.

**Primary recommendation:** Use FTS5 external content table with triggers for title search, HTML5 date inputs for date filtering, combine filters via JOIN pattern, implement results count with single query using window functions.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| SQLite FTS5 | Built-in | Full-text search | Native SQLite extension, zero dependencies, excellent performance |
| modernc.org/sqlite | 1.44.3 | SQLite driver | Already in project, includes FTS5 support |
| HTMX hx-trigger | 2.0.4 | Debounce pattern | Built-in delay modifier, no custom JS needed |
| HTML5 input[type=date] | Native | Date picker | Universal browser support (April 2021+), no library needed |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| None needed | - | - | Native features sufficient |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| FTS5 | LIKE '%term%' | LIKE is 10-100x slower, no ranking, no multi-word search |
| HTML5 date | JavaScript library (flatpickr, etc.) | Adds dependency for marginal UX gain, HTML5 is "good enough" |
| Trigger sync | Manual INSERT to FTS | Error-prone, easy to desync, triggers ensure consistency |

**Installation:**
```bash
# No additional dependencies needed - FTS5 built into SQLite
# Already have: modernc.org/sqlite v1.44.3
```

## Architecture Patterns

### Recommended Database Structure
```
Database Schema:
├── articles                # Existing table (blog_id, title, url, published_date, discovered_date, is_read)
├── articles_fts            # FTS5 virtual table (title only)
└── Triggers                # INSERT/UPDATE/DELETE to keep FTS5 in sync
```

### Pattern 1: External Content Table with Triggers
**What:** FTS5 virtual table references main table, triggers maintain sync
**When to use:** Always - avoids data duplication, single source of truth
**Example:**
```sql
-- Source: https://sqlite.org/fts5.html
-- FTS5 virtual table (external content pattern)
CREATE VIRTUAL TABLE articles_fts USING fts5(
    title,
    content='articles',
    content_rowid='id'
);

-- Trigger: sync on INSERT
CREATE TRIGGER articles_ai AFTER INSERT ON articles BEGIN
    INSERT INTO articles_fts(rowid, title) VALUES (new.id, new.title);
END;

-- Trigger: sync on UPDATE
CREATE TRIGGER articles_au AFTER UPDATE ON articles BEGIN
    INSERT INTO articles_fts(articles_fts, rowid, title) VALUES('delete', old.id, old.title);
    INSERT INTO articles_fts(rowid, title) VALUES (new.id, new.title);
END;

-- Trigger: sync on DELETE
CREATE TRIGGER articles_ad AFTER DELETE ON articles BEGIN
    INSERT INTO articles_fts(articles_fts, rowid, title) VALUES('delete', old.id, old.title);
END;
```

### Pattern 2: Combined Search + Date + Status Filters
**What:** JOIN FTS5 table with main table, apply multiple WHERE conditions
**When to use:** When combining full-text search with other filters
**Example:**
```sql
-- Source: https://www.sqlitetutorial.net/sqlite-full-text-search/
SELECT a.id, a.title, a.url, a.published_date, a.is_read, b.name, b.url
FROM articles a
JOIN articles_fts fts ON a.id = fts.rowid
INNER JOIN blogs b ON a.blog_id = b.id
WHERE fts MATCH ?                                    -- Search query
  AND a.is_read = ?                                  -- Status filter
  AND a.blog_id = COALESCE(?, a.blog_id)            -- Blog filter (NULL = all)
  AND a.published_date >= COALESCE(?, '1970-01-01') -- Date from
  AND a.published_date <= COALESCE(?, '9999-12-31') -- Date to
ORDER BY a.discovered_date DESC;
```

### Pattern 3: Results Count with Same Filters
**What:** Use COUNT(*) with identical WHERE clause as main query
**When to use:** Displaying "X results" message
**Example:**
```sql
-- Source: https://www.sqlitetutorial.net/sqlite-count-function/
-- Option 1: Separate COUNT query
SELECT COUNT(*) FROM articles a
JOIN articles_fts fts ON a.id = fts.rowid
WHERE fts MATCH ? AND a.is_read = ? AND ...;

-- Option 2: Single query with window function (more efficient)
SELECT
    a.*,
    COUNT(*) OVER() as total_count
FROM articles a
JOIN articles_fts fts ON a.id = fts.rowid
WHERE fts MATCH ? AND ...;
```

### Pattern 4: HTMX Debounced Search Input
**What:** Use `hx-trigger` with `changed delay:300ms` for active search
**When to use:** Search inputs that query on keystroke
**Example:**
```html
<!-- Source: https://htmx.org/attributes/hx-trigger/ -->
<input
    type="text"
    name="search"
    placeholder="Search articles..."
    hx-get="/articles"
    hx-trigger="keyup changed delay:300ms, search"
    hx-target="#article-list"
    hx-include="[name='status'],[name='blog'],[name='date_from'],[name='date_to']">
```

### Pattern 5: Date Range Shortcuts
**What:** Buttons that calculate and populate date ranges
**When to use:** Common date ranges (Last Week, Last Month)
**Example:**
```html
<!-- Source: https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/input/date -->
<input type="date" name="date_from" id="date_from">
<input type="date" name="date_to" id="date_to">

<button type="button" onclick="setLastWeek()">Last Week</button>
<button type="button" onclick="setLastMonth()">Last Month</button>

<script>
function setLastWeek() {
    const to = new Date();
    const from = new Date(to.getTime() - 7 * 24 * 60 * 60 * 1000);
    document.getElementById('date_from').value = from.toISOString().split('T')[0];
    document.getElementById('date_to').value = to.toISOString().split('T')[0];
    document.getElementById('date_from').dispatchEvent(new Event('change'));
}
</script>
```

### Pattern 6: Dynamic SQL Query Building in Go
**What:** Build WHERE clause conditionally based on provided filters
**When to use:** Optional filters (search may be empty, dates may be null)
**Example:**
```go
// Source: https://medium.com/tokopedia-engineering/dynamic-sql-query-builder-in-golang-2c71e2c21ff8
func (db *Database) SearchArticles(opts SearchOptions) ([]Article, error) {
    query := `SELECT a.id, a.title, a.url, a.published_date, b.name, b.url
              FROM articles a INNER JOIN blogs b ON a.blog_id = b.id`
    var conditions []string
    var args []interface{}

    // Add FTS5 search if query provided
    if opts.SearchQuery != "" {
        query = `SELECT a.id, a.title, a.url, a.published_date, b.name, b.url
                 FROM articles a
                 JOIN articles_fts fts ON a.id = fts.rowid
                 INNER JOIN blogs b ON a.blog_id = b.id`
        conditions = append(conditions, "fts MATCH ?")
        args = append(args, opts.SearchQuery)
    }

    // Add status filter
    conditions = append(conditions, "a.is_read = ?")
    args = append(args, opts.IsRead)

    // Add blog filter if provided
    if opts.BlogID != nil {
        conditions = append(conditions, "a.blog_id = ?")
        args = append(args, *opts.BlogID)
    }

    // Add date range if provided
    if opts.DateFrom != nil {
        conditions = append(conditions, "a.published_date >= ?")
        args = append(args, opts.DateFrom.Format("2006-01-02"))
    }
    if opts.DateTo != nil {
        conditions = append(conditions, "a.published_date <= ?")
        args = append(args, opts.DateTo.Format("2006-01-02"))
    }

    if len(conditions) > 0 {
        query += " WHERE " + strings.Join(conditions, " AND ")
    }
    query += " ORDER BY a.discovered_date DESC"

    rows, err := db.conn.Query(query, args...)
    // ... scan and return
}
```

### Anti-Patterns to Avoid
- **Contentless FTS5 tables** - Don't use `content=''` (duplicates data, wastes space)
- **Manual FTS sync** - Don't INSERT to FTS5 in application code (use triggers)
- **LIKE for search** - Don't use `WHERE title LIKE '%search%'` (slow, no ranking)
- **Separate COUNT query** - Don't run two queries for count + results (use window function)
- **String concatenation for SQL** - Don't build SQL with string concat (SQL injection risk)
- **Updating FTS5 with wrong values** - Never update FTS5 delete with values different from original INSERT (causes corruption)

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Full-text search | Custom LIKE queries, prefix matching | FTS5 virtual table | FTS5 handles tokenization, ranking, multi-word search, prefix queries - reimplementing is complex and slow |
| Input debouncing | Custom setTimeout/clearTimeout JS | HTMX `delay:300ms` | HTMX built-in, fewer lines, consistent pattern |
| Date picker | Custom calendar UI | HTML5 `<input type="date">` | Native, accessible, mobile-friendly, zero bytes |
| SQL query builder | String concatenation | Parameterized queries with slice building | Prevents SQL injection, cleaner code |
| FTS5 synchronization | Manual INSERT to FTS in app code | Database triggers | Triggers are atomic, can't forget to sync, survives crashes |

**Key insight:** SQLite FTS5 is battle-tested for 10+ years with edge cases (Unicode, stemming, ranking) already solved. Custom search implementations miss these and perform poorly at scale.

## Common Pitfalls

### Pitfall 1: FTS5 Trigger Ordering Mistake
**What goes wrong:** UPDATE trigger that inserts new row before deleting old row causes duplicate FTS entries
**Why it happens:** SQLite processes triggers in order; wrong order leaves old entry
**How to avoid:** Always DELETE from FTS5 first, then INSERT new entry in UPDATE triggers
**Warning signs:** Search results show same article multiple times

### Pitfall 2: Desynchronized FTS Index
**What goes wrong:** FTS5 table out of sync with main table, search returns wrong results
**Why it happens:** Forgot to add trigger for one operation (UPDATE/DELETE), or trigger has wrong column names
**How to avoid:** Test all three triggers (INSERT/UPDATE/DELETE), verify with `SELECT * FROM articles_fts` after each operation
**Warning signs:** Search returns deleted articles, doesn't find existing articles, or returns wrong rowids

### Pitfall 3: FTS5 DELETE Syntax Confusion
**What goes wrong:** Using `DELETE FROM articles_fts WHERE rowid = ?` corrupts index
**Why it happens:** FTS5 has special syntax: `INSERT INTO articles_fts(articles_fts, rowid, title) VALUES('delete', ?, ?)`
**How to avoid:** Never use standard DELETE syntax on FTS5 tables, always use special INSERT syntax
**Warning signs:** "FTS5 extension malfunction detected" error, corrupted index

### Pitfall 4: iOS Safari Date Picker Min/Max
**What goes wrong:** User can select dates outside min/max range in iOS Safari date picker
**Why it happens:** Safari on iOS ignores min/max attributes in picker UI (validates on submit only)
**How to avoid:** Always validate dates server-side, don't rely on browser validation
**Warning signs:** Server receives dates outside expected range from iOS users

### Pitfall 5: Empty Search Query Matches Everything
**What goes wrong:** Empty string to `MATCH` returns all rows or errors
**Why it happens:** FTS5 MATCH requires non-empty query
**How to avoid:** Check if search query is empty before adding FTS5 JOIN/MATCH to SQL
**Warning signs:** All articles returned when search box is empty, or SQL error

### Pitfall 6: Date Format Mismatch
**What goes wrong:** Date comparisons fail because stored format differs from query format
**Why it happens:** SQLite stores dates as TEXT, comparison is lexicographic
**How to avoid:** Always use ISO 8601 format (`YYYY-MM-DD` or `YYYY-MM-DD HH:MM:SS`) for dates
**Warning signs:** Date filters return wrong results, seemingly random filtering behavior

### Pitfall 7: HTMX Changed Modifier Misunderstanding
**What goes wrong:** Search triggers on every keystroke despite `changed` modifier
**Why it happens:** Confused `change` event with `changed` modifier - need both `keyup` and `changed`
**How to avoid:** Use `hx-trigger="keyup changed delay:300ms"` (event + modifier + timing)
**Warning signs:** Network tab shows request every keystroke, server logs show excessive queries

## Code Examples

Verified patterns from official sources:

### Creating FTS5 Table with Triggers
```sql
-- Source: https://sqlite.org/fts5.html
-- External content table (recommended pattern)
CREATE VIRTUAL TABLE articles_fts USING fts5(
    title,
    content='articles',
    content_rowid='id'
);

-- Sync trigger: INSERT
CREATE TRIGGER articles_ai AFTER INSERT ON articles BEGIN
    INSERT INTO articles_fts(rowid, title) VALUES (new.id, new.title);
END;

-- Sync trigger: UPDATE (delete old + insert new)
CREATE TRIGGER articles_au AFTER UPDATE ON articles BEGIN
    INSERT INTO articles_fts(articles_fts, rowid, title) VALUES('delete', old.id, old.title);
    INSERT INTO articles_fts(rowid, title) VALUES (new.id, new.title);
END;

-- Sync trigger: DELETE
CREATE TRIGGER articles_ad AFTER DELETE ON articles BEGIN
    INSERT INTO articles_fts(articles_fts, rowid, title) VALUES('delete', old.id, old.title);
END;
```

### Querying with Combined Filters
```sql
-- Source: https://www.sqlitetutorial.net/sqlite-full-text-search/
-- Combine search + status + blog + date filters
SELECT
    a.id,
    a.title,
    a.url,
    a.published_date,
    a.is_read,
    b.name as blog_name,
    b.url as blog_url,
    COUNT(*) OVER() as total_count
FROM articles a
JOIN articles_fts fts ON a.id = fts.rowid
INNER JOIN blogs b ON a.blog_id = b.id
WHERE fts MATCH 'search terms'
  AND a.is_read = 0
  AND a.blog_id = 5
  AND a.published_date >= '2026-01-01'
  AND a.published_date <= '2026-01-31'
ORDER BY a.discovered_date DESC;
```

### HTMX Debounced Search Input
```html
<!-- Source: https://htmx.org/attributes/hx-trigger/ -->
<input
    type="search"
    name="search"
    placeholder="Search articles by title..."
    hx-get="/articles"
    hx-trigger="keyup changed delay:300ms, search"
    hx-target="#article-list"
    hx-include="[name='status'],[name='blog'],[name='date_from'],[name='date_to']"
    hx-push-url="true">
```

### Date Range Shortcuts (Vanilla JS)
```html
<!-- Source: https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/input/date -->
<div class="date-filters">
    <button type="button" onclick="setDateRange('week')">Last Week</button>
    <button type="button" onclick="setDateRange('month')">Last Month</button>

    <input type="date"
           name="date_from"
           id="date_from"
           hx-get="/articles"
           hx-trigger="change"
           hx-include="[name='status'],[name='blog'],[name='search'],[name='date_to']"
           hx-target="#article-list">

    <input type="date"
           name="date_to"
           id="date_to"
           hx-get="/articles"
           hx-trigger="change"
           hx-include="[name='status'],[name='blog'],[name='search'],[name='date_from']"
           hx-target="#article-list">
</div>

<script>
function setDateRange(range) {
    const to = new Date();
    const from = new Date();

    if (range === 'week') {
        from.setDate(to.getDate() - 7);
    } else if (range === 'month') {
        from.setMonth(to.getMonth() - 1);
    }

    document.getElementById('date_from').value = from.toISOString().split('T')[0];
    document.getElementById('date_to').value = to.toISOString().split('T')[0];

    // Trigger HTMX update
    document.getElementById('date_from').dispatchEvent(new Event('change', {bubbles: true}));
}
</script>
```

### Go Dynamic Query Building
```go
// Source: https://medium.com/tokopedia-engineering/dynamic-sql-query-builder-in-golang-2c71e2c21ff8
type SearchOptions struct {
    SearchQuery string
    IsRead      bool
    BlogID      *int64
    DateFrom    *time.Time
    DateTo      *time.Time
}

func (db *Database) SearchArticles(opts SearchOptions) ([]ArticleWithBlog, int, error) {
    baseQuery := `SELECT a.id, a.title, a.url, a.thumbnail_url,
                         a.published_date, a.is_read,
                         b.name as blog_name, b.url as blog_url,
                         COUNT(*) OVER() as total_count
                  FROM articles a`

    var conditions []string
    var args []interface{}

    // Add FTS5 JOIN only if search query provided
    if opts.SearchQuery != "" {
        baseQuery += " JOIN articles_fts fts ON a.id = fts.rowid"
        conditions = append(conditions, "fts MATCH ?")
        args = append(args, opts.SearchQuery)
    }

    baseQuery += " INNER JOIN blogs b ON a.blog_id = b.id"

    // Status filter (always present)
    conditions = append(conditions, "a.is_read = ?")
    args = append(args, opts.IsRead)

    // Optional blog filter
    if opts.BlogID != nil {
        conditions = append(conditions, "a.blog_id = ?")
        args = append(args, *opts.BlogID)
    }

    // Optional date range
    if opts.DateFrom != nil {
        conditions = append(conditions, "a.published_date >= ?")
        args = append(args, opts.DateFrom.Format("2006-01-02"))
    }
    if opts.DateTo != nil {
        // Include entire end date (23:59:59)
        conditions = append(conditions, "a.published_date < ?")
        endDate := opts.DateTo.AddDate(0, 0, 1) // Next day
        args = append(args, endDate.Format("2006-01-02"))
    }

    if len(conditions) > 0 {
        baseQuery += " WHERE " + strings.Join(conditions, " AND ")
    }

    baseQuery += " ORDER BY a.discovered_date DESC"

    rows, err := db.conn.Query(baseQuery, args...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var articles []ArticleWithBlog
    var totalCount int
    for rows.Next() {
        var article ArticleWithBlog
        err := rows.Scan(
            &article.ID, &article.Title, &article.URL, &article.ThumbnailURL,
            &article.PublishedDate, &article.IsRead,
            &article.BlogName, &article.BlogURL,
            &totalCount,
        )
        if err != nil {
            return nil, 0, err
        }
        articles = append(articles, article)
    }

    return articles, totalCount, rows.Err()
}
```

### Rebuilding FTS5 Index (Recovery)
```sql
-- Source: https://sqlite.org/fts5.html
-- If FTS5 gets out of sync, rebuild from content table
INSERT INTO articles_fts(articles_fts) VALUES('rebuild');

-- Optimize FTS5 index (merge b-trees for faster queries)
INSERT INTO articles_fts(articles_fts) VALUES('optimize');
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| FTS3/FTS4 | FTS5 | SQLite 3.9.0 (2015) | Better performance, simpler API, better ranking |
| Custom debounce JS | HTMX delay modifier | HTMX 1.0 (2020) | Fewer lines, declarative, consistent |
| jQuery Datepicker | HTML5 `<input type="date">` | Baseline support (April 2021) | Zero dependencies, native UX, accessible |
| LIKE queries | FTS5 MATCH | Available since 2015 | 10-100x faster, ranking, multi-word support |

**Deprecated/outdated:**
- FTS3/FTS4: Use FTS5 instead (better performance, active development)
- JavaScript debounce libraries (Lodash, etc.): Use HTMX built-in delay
- Heavy date picker libraries (jQuery UI, bootstrap-datepicker): Use HTML5 native input

## Open Questions

1. **Should search query be sanitized beyond parameterization?**
   - What we know: Parameterized queries prevent SQL injection
   - What's unclear: FTS5 has query syntax (AND, OR, NOT, quotes) - should we escape these or allow them?
   - Recommendation: Allow FTS5 syntax initially (power users benefit), add client-side help text explaining syntax

2. **Should date filtering use published_date or discovered_date?**
   - What we know: Both dates available, published_date may be NULL for some articles
   - What's unclear: User expectation - "articles from last week" means discovered or published?
   - Recommendation: Use published_date with fallback to discovered_date: `COALESCE(a.published_date, a.discovered_date)`

3. **Should empty search box bypass FTS5 JOIN?**
   - What we know: Joining FTS5 table with no MATCH is unnecessary overhead
   - What's unclear: Performance impact in practice for small datasets (<10k articles)
   - Recommendation: Conditionally add JOIN only when search query non-empty (shown in code examples)

## Sources

### Primary (HIGH confidence)
- [SQLite FTS5 Extension](https://sqlite.org/fts5.html) - Official FTS5 documentation
- [HTMX hx-trigger Attribute](https://htmx.org/attributes/hx-trigger/) - Official HTMX delay modifier docs
- [MDN: input type="date"](https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/input/date) - HTML5 date input specification

### Secondary (MEDIUM confidence)
- [SQLite Full-Text Search Tutorial](https://www.sqlitetutorial.net/sqlite-full-text-search/) - FTS5 query patterns
- [SQLite COUNT Function](https://www.sqlitetutorial.net/sqlite-count-function/) - COUNT with window functions
- [Dynamic SQL Query Builder in Golang](https://medium.com/tokopedia-engineering/dynamic-sql-query-builder-in-golang-2c71e2c21ff8) - Go query building patterns
- [Can I use: Date input types](https://caniuse.com/input-datetime) - Browser support data (baseline April 2021)

### Tertiary (LOW confidence)
- [FTS5 Trigger Synchronization Gotchas](https://sqlite.org/forum/info/da59bf102d7a7951740bd01c4942b1119512a86bfa1b11d4f762056c8eb7fc4e) - SQLite forum discussion on trigger pitfalls
- [iOS Safari Date Input Limitations](https://adactio.com/journal/21050) - iOS min/max attribute behavior
- [SQLite Date Filtering Patterns](https://www.slingacademy.com/article/filtering-data-in-sqlite-with-advanced-conditions/) - Date WHERE clause examples

## Metadata

**Confidence breakdown:**
- FTS5 setup and usage: HIGH - Official SQLite documentation, well-established patterns
- HTMX debounce: HIGH - Official HTMX documentation, simple feature
- Date filtering SQL: HIGH - Standard SQLite date operations, ISO 8601 format
- HTML5 date input: HIGH - MDN documentation, baseline browser support verified
- Combined filters: MEDIUM - Pattern is logical combination of verified components
- Go query building: MEDIUM - Common pattern, multiple sources agree
- iOS Safari gotchas: MEDIUM - Community-reported, consistent across sources
- Performance implications: LOW - No benchmarks run on actual project data

**Research date:** 2026-02-03
**Valid until:** ~30 days (stable technologies, but FTS5 best practices may evolve)
