# Plan 11-02 Summary: Schema Migration for Nullable blog_id

**Status:** COMPLETE
**Verified:** 2026-02-09

## What Was Built

Schema migration to make `articles.blog_id` nullable, enabling blog deletion without orphaning articles constraint violations.

### Migration Logic (`internal/storage/database.go`)

Added to `ensureMigrations()`:

1. `columnIsNotNull(table, column)` - Check if column has NOT NULL constraint
2. `migrateBlogIDToNullable()` - Recreate articles table with nullable blog_id

### Migration Process

SQLite does not support ALTER COLUMN, so the migration:
1. Creates `articles_new` table with nullable `blog_id`
2. Copies all data from `articles` to `articles_new`
3. Drops FTS5 triggers (articles_ai, articles_au, articles_ad)
4. Drops FTS5 table (articles_fts)
5. Drops old `articles` table
6. Renames `articles_new` to `articles`
7. FTS5 table and triggers are recreated by subsequent migration step

### Idempotency

- Migration only runs if `blog_id` column currently has NOT NULL constraint
- Safe to run multiple times (checks before executing)
- All existing data preserved

## Schema Change

Before:
```sql
blog_id INTEGER NOT NULL
```

After:
```sql
blog_id INTEGER
```

## Files Modified

1. `internal/storage/database.go` - Added migration functions

## Verification

```bash
sqlite3 ~/.blogwatcher/blogwatcher.db "PRAGMA table_info(articles);" | grep blog_id
# Output: 1|blog_id|INTEGER|0||0
# The 4th field (0) indicates nullable
```
