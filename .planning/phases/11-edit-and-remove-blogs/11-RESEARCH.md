# Phase 11 Research: Edit and Remove Blogs

**Researched:** 2026-02-09
**Domain:** Blog Management UI with Inline Editing and Confirmation Dialogs
**Overall confidence:** HIGH

## Executive Summary

This phase adds edit and delete functionality to the settings page blog list. The research establishes patterns for inline editing using HTMX's click-to-edit pattern, confirmation dialogs using native HTML `<dialog>` elements, and database operations with proper foreign key handling for blog deletion with/without cascading article deletion.

**Key architectural decisions:**
1. **Inline editing over modals** - HTMX click-to-edit pattern swaps display/edit states without page navigation
2. **Native `<dialog>` element** - Provides built-in accessibility, focus management, and keyboard support
3. **Explicit foreign key handling** - SQLite foreign keys disabled in schema; manual deletion with user choice for articles
4. **Two-step confirmation** - User chooses deletion scope (blog only vs blog + articles) before execution

## Key Findings

### Current State Analysis

**Existing infrastructure (HIGH confidence):**
- Settings page exists at `/settings` with blog list display (Phase 9-10)
- `ListBlogsWithCounts()` provides blog data with article counts
- `blog-settings-card` UI component displays name, URL, article count
- Error/success message patterns established in add-blog-form
- HTMX 2.0.4 integrated with swap patterns for dynamic updates
- Database methods: `UpdateBlog()`, `GetBlogByName()` exist but unused

**Database schema (HIGH confidence):**
```sql
CREATE TABLE blogs (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    feed_url TEXT,
    scrape_selector TEXT,
    last_scanned TIMESTAMP
);

CREATE TABLE articles (
    id INTEGER PRIMARY KEY,
    blog_id INTEGER NOT NULL,
    ...
    FOREIGN KEY (blog_id) REFERENCES blogs(id)
);
```

**Critical finding:** Foreign key constraint exists but **no CASCADE or SET NULL action specified**. Default is "NO ACTION" which will **block blog deletion** if articles exist. This must be handled explicitly in code.

### Technology Stack

**Confirmed stack (HIGH confidence):**
- Go 1.22+ with http.ServeMux pattern routing (`POST /blogs/{id}/edit`, `DELETE /blogs/{id}`)
- Go templates (`.gohtml`) for server-side rendering
- HTMX 2.0.4 for progressive enhancement
- SQLite with modernc.org/sqlite driver
- CSS custom properties for theming

**No additional dependencies needed.** All functionality achievable with existing stack.

## Domain Landscape

### 1. Inline Editing Pattern (HIGH confidence)

**HTMX Click-to-Edit Pattern** is the canonical approach for inline editing without page refresh.

**Pattern structure:**
```html
<!-- Display state -->
<div hx-target="this" hx-swap="outerHTML">
    <div><label>Blog Name</label>: My Blog</div>
    <button hx-get="/blogs/123/edit">Edit</button>
</div>

<!-- Edit state (server returns this) -->
<form hx-put="/blogs/123" hx-target="this" hx-swap="outerHTML">
    <input type="text" name="name" value="My Blog">
    <button type="submit">Save</button>
    <button hx-get="/blogs/123">Cancel</button>
</form>
```

**Key benefits:**
- `hx-target="this"` + `hx-swap="outerHTML"` creates self-contained swap behavior
- Cancel button restores original display by fetching display template
- No JavaScript required for state management
- Works naturally with Go template partials

**Codebase fit:** Aligns perfectly with existing patterns in `settings-page.gohtml` where each blog is rendered in a `.blog-settings-card`. Each card becomes a swap container.

**Source:** [HTMX Click-to-Edit Example](https://htmx.org/examples/click-to-edit/)

### 2. Confirmation Dialogs (HIGH confidence)

**Two viable approaches identified:**

#### Option A: Native `<dialog>` Element (RECOMMENDED)

**Rationale:**
- Built-in accessibility (focus trap, Esc key handling, ARIA roles)
- No JavaScript library dependencies
- Modern browser support excellent (96%+ as of 2025)
- Aligns with "no mock mode" principle - real browser functionality

**Pattern:**
```html
<button onclick="document.getElementById('confirm-delete-123').showModal()">Remove</button>

<dialog id="confirm-delete-123" aria-labelledby="dialog-title-123">
    <h2 id="dialog-title-123">Delete Blog?</h2>
    <p>This blog has 42 articles.</p>
    <form method="dialog">
        <button value="cancel">Cancel</button>
        <button hx-delete="/blogs/123?mode=blog-only"
                hx-target="#blog-settings-list"
                hx-swap="outerHTML">
            Delete blog only
        </button>
        <button hx-delete="/blogs/123?mode=with-articles"
                hx-target="#blog-settings-list"
                hx-swap="outerHTML"
                class="btn-danger">
            Delete blog + articles
        </button>
    </form>
</dialog>
```

**Built-in features:**
- `showModal()` creates modal backdrop and focus trap automatically
- Esc key closes dialog by default
- `aria-labelledby` links title for screen readers (no manual ARIA needed)
- `autofocus` attribute moves focus to appropriate button
- `method="dialog"` closes dialog on form submission

**Remaining considerations:**
- Clicking backdrop does NOT close dialog in Chrome - must provide explicit Cancel button (already required by UX anyway)
- Must use `showModal()` (not `show()`) to get modal behavior with backdrop

**Sources:**
- [MDN: dialog Element](https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/dialog)
- [HTML Dialog: Getting Accessibility Right](https://jaredcunha.com/blog/html-dialog-getting-accessibility-and-ux-right)
- [Native Dialog Element Benefits](https://www.oidaisdes.org/native-dialog-element.en/)

#### Option B: hx-confirm Attribute (NOT RECOMMENDED for this use case)

**Why not:**
- Uses browser's native `window.confirm()` - ugly, can't show article count
- Can't provide two-choice selection (blog-only vs with-articles)
- Would require SweetAlert2 or similar library for custom styling
- Violates "no additional dependencies" principle

**Only mention for completeness.** Native `<dialog>` is superior for this use case.

### 3. Database Deletion Strategy (HIGH confidence)

**Foreign Key Constraint Analysis:**

Current schema: `FOREIGN KEY (blog_id) REFERENCES blogs(id)` with **no ON DELETE action**.

**SQLite default behavior:** "NO ACTION" - deletion **fails** if foreign key references exist.

**Options:**

| Strategy | Implementation | Use Case |
|----------|---------------|----------|
| CASCADE | `ON DELETE CASCADE` | Child data disposable (comments, logs) |
| SET NULL | `ON DELETE SET NULL` | Child data outlives parent (posts, products) |
| RESTRICT | `ON DELETE RESTRICT` | Financial/audit records |
| Manual | Code-based deletion | User choice needed (this phase) |

**DECISION: Manual deletion** because requirement REM-02 explicitly requires **user choice** between:
1. Delete blog only (keep articles orphaned with NULL blog_id)
2. Delete blog + articles (cascade delete)

**Implementation approach:**

```go
// Option 1: Delete blog only (SET NULL pattern)
func (db *Database) DeleteBlogOnly(blogID int64) error {
    tx, err := db.conn.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback() // Rollback if not committed

    // Set articles' blog_id to NULL
    _, err = tx.Exec(`UPDATE articles SET blog_id = NULL WHERE blog_id = ?`, blogID)
    if err != nil {
        return err
    }

    // Delete blog
    _, err = tx.Exec(`DELETE FROM blogs WHERE id = ?`, blogID)
    if err != nil {
        return err
    }

    return tx.Commit()
}

// Option 2: Delete blog + articles (CASCADE pattern)
func (db *Database) DeleteBlogWithArticles(blogID int64) error {
    tx, err := db.conn.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Delete articles first (foreign key constraint)
    _, err = tx.Exec(`DELETE FROM articles WHERE blog_id = ?`, blogID)
    if err != nil {
        return err
    }

    // Delete blog
    _, err = tx.Exec(`DELETE FROM blogs WHERE id = ?`, blogID)
    if err != nil {
        return err
    }

    return tx.Commit()
}
```

**Critical consideration:** Articles table has `blog_id INTEGER NOT NULL`. To support "delete blog only", **must alter schema** to allow NULL:

```sql
-- Migration needed
ALTER TABLE articles ALTER COLUMN blog_id DROP NOT NULL;
-- SQLite doesn't support ALTER COLUMN, so requires table recreation
```

**Alternative:** Keep NOT NULL constraint and "delete blog only" actually **reassigns articles to a special "Orphaned" blog**. This avoids schema migration but adds complexity.

**Recommendation:** Allow NULL blog_id for true orphaning. Simpler model, aligns with SET NULL semantics.

**Sources:**
- [SQLite Foreign Keys](https://sqlite.org/foreignkeys.html)
- [CASCADE vs SET NULL Best Practices](https://medium.com/@sunnywilson.veshapogu/restrict-vs-cascade-vs-set-null-in-sql-choosing-the-right-foreign-key-rule-6d7c98484710)

### 4. Go Transaction Patterns (HIGH confidence)

**Best practices for database transactions with validation:**

```go
func (s *Server) handleDeleteBlog(w http.ResponseWriter, r *http.Request) {
    // 1. Extract and validate ID
    idStr := r.PathValue("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid blog ID", http.StatusBadRequest)
        return
    }

    // 2. Parse deletion mode
    mode := r.URL.Query().Get("mode") // "blog-only" or "with-articles"

    // 3. Execute appropriate deletion
    var deleteErr error
    switch mode {
    case "blog-only":
        deleteErr = s.db.DeleteBlogOnly(id)
    case "with-articles":
        deleteErr = s.db.DeleteBlogWithArticles(id)
    default:
        http.Error(w, "Invalid deletion mode", http.StatusBadRequest)
        return
    }

    if deleteErr != nil {
        log.Printf("Error deleting blog %d: %v", id, deleteErr)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    // 4. Return refreshed blog list (HTMX swap)
    s.handleSettingsBlogList(w, r) // Renders updated list
}
```

**Transaction best practices:**
- Pass `context.Context` for timeout/cancellation (optional for simple cases)
- Defer `tx.Rollback()` - no-op if already committed, safety net if panic
- Validate inputs BEFORE beginning transaction (reduce lock time)
- Use explicit error handling - don't silent fail

**Source:** [Go Official Docs: Execute Transactions](https://go.dev/doc/database/execute-transactions)

### 5. UI/UX Patterns (MEDIUM confidence)

**Established patterns from Phase 10:**

**Error handling:**
- `.error-message` class: red background, border, padding
- Server returns template fragment with `{{.Error}}` on failure
- Form pre-populated with submitted values
- Pattern: `s.renderError(w, templateName, errorMessage, formData)`

**Success feedback:**
- `.success-message` class: green background
- For deletions: blog disappears from list (HTMX swap removes it)
- No explicit success message needed (visual feedback via removal)

**HTMX refresh patterns:**
- Target `#blog-list` in sidebar for subscription list refresh
- Target `.blog-settings-list` for settings page list refresh
- Use `hx-swap="outerHTML"` to replace entire container

**Confirmation dialog UX:**
- Show article count in dialog body
- Two clearly labeled buttons with different visual weights
- Danger action (delete with articles) uses `.btn-danger` class
- Cancel button easily accessible
- Esc key closes dialog (free with `<dialog>`)

### 6. Accessibility Considerations (HIGH confidence)

**Native `<dialog>` provides:**
- Automatic `role="dialog"` and `aria-modal="true"`
- Focus trap - Tab cycles within dialog
- Esc key to close (must be preserved, don't override)
- Return focus to trigger element on close

**Required additions:**
- `aria-labelledby` pointing to dialog title ID
- `autofocus` on appropriate button (close/cancel for confirmation dialogs)
- Semantic HTML: `<h2>` for title, `<button>` elements (not divs)

**Inline editing accessibility:**
- Edit button must be keyboard accessible (`<button>`, not `<span onclick>`)
- Form inputs must have labels (use `<label>` or `aria-label`)
- Focus moves to first input on edit mode swap
- Cancel button restores original state (undo edit)

**Sources:**
- [HTML Dialog Accessibility](https://schalkneethling.com/posts/html-dialog-native-solution-for-accessible-modal-interactions/)
- [Dialog UX Best Practices](https://jaredcunha.com/blog/html-dialog-getting-accessibility-and-ux-right)

## Architecture Patterns

### Component Structure

```
settings-page.gohtml (self-contained page template)
├── add-blog-form.gohtml (existing, untouched)
└── blog-settings-list (NEW, extractable partial)
    └── blog-settings-card (MODIFIED, becomes swap container)
        ├── blog-display (NEW partial for display state)
        │   ├── Blog info (name, URL, count)
        │   ├── Edit button (triggers swap to edit state)
        │   └── Remove button (triggers dialog)
        └── blog-edit-form (NEW partial for edit state)
            ├── Name input
            ├── Save button (PUT request)
            └── Cancel button (restores display)
```

**Swap boundaries:**
- **Edit action:** Swap `.blog-settings-card` (outerHTML)
- **Delete action:** Swap `.blog-settings-list` (innerHTML, replaces all cards)
- **Cancel edit:** Swap `.blog-settings-card` (outerHTML, restore display)

### Handler Structure

```
internal/server/handlers.go
├── handleSettings (existing, renders full settings page)
├── handleBlogEdit (NEW, GET /blogs/{id}/edit -> returns edit form)
├── handleBlogUpdate (NEW, PUT /blogs/{id} -> saves, returns display)
├── handleBlogDelete (NEW, DELETE /blogs/{id}?mode=X -> deletes, returns list)
└── handleSettingsBlogList (NEW helper, renders just blog list partial)
```

### Database Methods (storage/database.go)

```go
// Existing (already present):
- UpdateBlog(blog model.Blog) error
- GetBlogByName(name string) (*model.Blog, error)
- ListBlogsWithCounts() ([]BlogWithCount, error)

// New methods needed:
- GetBlogByID(id int64) (*model.Blog, error)
- DeleteBlogOnly(id int64) error
- DeleteBlogWithArticles(id int64) error
- UpdateBlogName(id int64, name string) error // Simpler than UpdateBlog
```

### Routes (internal/server/routes.go)

```go
// Add to registerRoutes():
s.mux.HandleFunc("GET /blogs/{id}/edit", s.handleBlogEdit)
s.mux.HandleFunc("PUT /blogs/{id}", s.handleBlogUpdate)
s.mux.HandleFunc("DELETE /blogs/{id}", s.handleBlogDelete)
```

## Pitfalls and Risks

### Critical Pitfall: Foreign Key Constraint Blocking Deletion

**What goes wrong:** Attempting `DELETE FROM blogs WHERE id = ?` fails with foreign key constraint error if articles exist.

**Why it happens:** SQLite default is NO ACTION (RESTRICT behavior) when no ON DELETE clause specified.

**Prevention:**
1. Always wrap deletion in transaction
2. Handle articles BEFORE attempting blog deletion
3. Test both deletion modes with blogs that have articles
4. Add error handling for constraint violations

**Detection:** Database error on blog deletion despite valid blog ID.

### Moderate Pitfall: Dialog Not Closing After Deletion

**What goes wrong:** HTMX request succeeds, blog list updates, but dialog remains open.

**Why it happens:** HTMX doesn't automatically close dialog after successful request.

**Prevention:** Add `hx-on::after-request="this.closest('dialog').close()"` to delete buttons.

**Alternative:** Return `HX-Trigger: closeDialog` header from server, listen for event.

### Moderate Pitfall: Stale Article Count in Dialog

**What goes wrong:** User edits blog name, article count in confirmation dialog is stale.

**Why it happens:** Dialog is rendered once on page load with static article count.

**Prevention:**
- Option A: Re-render dialog content on edit (fetch `/blogs/{id}/delete-confirm` endpoint)
- Option B: Include article count in dialog ID and regenerate card on edit
- Option C: Accept stale count (simplest, low-risk since it's just display)

**Recommendation:** Option C for MVP, upgrade to A if users report confusion.

### Minor Pitfall: Edit Form Submit on Enter Key

**What goes wrong:** User presses Enter in name input, form submits unexpectedly.

**Why it happens:** Browser default form submission on Enter in text input.

**Prevention:** This is actually DESIRED behavior - Enter should save edit.

**Not a pitfall, just confirming expected behavior.**

### Minor Pitfall: Concurrent Edit Attempts

**What goes wrong:** Two users edit same blog simultaneously, last write wins.

**Why it happens:** No optimistic locking or version checking.

**Prevention:**
- Phase 11: Accept last-write-wins (simple, good enough for single-user app)
- Future phase: Add `version` column and check on update

**Risk level:** LOW - BlogWatcher is single-user desktop app, concurrency unlikely.

## Implementation Recommendations

### Phase Ordering

1. **Edit functionality first** (simpler, no schema changes)
   - Add edit button to blog-settings-card
   - Implement click-to-edit swap pattern
   - Add `GET /blogs/{id}/edit` and `PUT /blogs/{id}` handlers
   - Test edit flow end-to-end

2. **Remove functionality second** (requires schema migration)
   - Modify articles.blog_id to allow NULL (schema migration)
   - Add confirmation dialog to each card
   - Implement `DELETE /blogs/{id}` handler with mode parameter
   - Add deletion database methods
   - Test both deletion modes

**Rationale:** Edit is isolated, no schema changes. Remove requires schema migration which affects articles table - riskier change, benefit from having edit working first to validate patterns.

### Database Schema Migration

**Required migration:**
```sql
-- SQLite doesn't support ALTER COLUMN, must recreate table
BEGIN TRANSACTION;

-- 1. Create new articles table with blog_id allowing NULL
CREATE TABLE articles_new (
    id INTEGER PRIMARY KEY,
    blog_id INTEGER,  -- Removed NOT NULL constraint
    title TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    published_date TIMESTAMP,
    discovered_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_read BOOLEAN DEFAULT FALSE,
    thumbnail_url TEXT,
    FOREIGN KEY (blog_id) REFERENCES blogs(id)
);

-- 2. Copy data
INSERT INTO articles_new SELECT * FROM articles;

-- 3. Drop old table
DROP TABLE articles;

-- 4. Rename new table
ALTER TABLE articles_new RENAME TO articles;

-- 5. Recreate FTS triggers (required after table recreation)
CREATE TRIGGER articles_ai AFTER INSERT ON articles BEGIN
    INSERT INTO articles_fts(rowid, title) VALUES (new.id, new.title);
END;

CREATE TRIGGER articles_au AFTER UPDATE ON articles BEGIN
    INSERT INTO articles_fts(articles_fts, rowid, title) VALUES('delete', old.id, old.title);
    INSERT INTO articles_fts(rowid, title) VALUES (new.id, new.title);
END;

CREATE TRIGGER articles_ad AFTER DELETE ON articles BEGIN
    INSERT INTO articles_fts(articles_fts, rowid, title) VALUES('delete', old.id, old.title);
END;

COMMIT;
```

**Add to `database.go` `ensureMigrations()`:**
```go
// Add after FTS5 migration
if !db.columnAllowsNull("articles", "blog_id") {
    if err := db.migrateBlogIDToNullable(); err != nil {
        return fmt.Errorf("failed to migrate blog_id to nullable: %w", err)
    }
}
```

### Testing Strategy

**Edit flow test cases:**
1. Edit blog name, save, verify name updated in list and sidebar
2. Edit blog name, cancel, verify name unchanged
3. Edit blog name to empty string, verify validation error
4. Edit blog name to existing name, verify unique constraint error
5. Edit blog while page loading, verify no race condition

**Delete flow test cases:**
1. Delete blog with zero articles (blog-only mode)
2. Delete blog with articles (blog-only mode), verify articles persist with NULL blog_id
3. Delete blog with articles (with-articles mode), verify articles deleted
4. Cancel delete dialog, verify blog remains
5. Delete blog, verify sidebar subscription list updates

**Accessibility test cases:**
1. Keyboard navigation: Tab to edit, press Enter to activate
2. Keyboard navigation: Esc in dialog closes without deleting
3. Screen reader: Dialog announces title and article count
4. Focus management: Focus returns to blog card after dialog closes

### CSS Requirements

**New classes needed:**
```css
/* Edit form within blog-settings-card */
.blog-settings-edit-form {
    /* Similar to .add-blog-section form styles */
}

/* Confirmation dialog */
dialog {
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 1.5rem;
    max-width: 500px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

dialog::backdrop {
    background-color: rgba(0, 0, 0, 0.5);
}

/* Danger button for delete with articles */
.btn-danger {
    background-color: #DC2626;
    color: white;
}

.btn-danger:hover {
    background-color: #B91C1C;
}

/* Dialog buttons layout */
.dialog-actions {
    display: flex;
    gap: 0.5rem;
    justify-content: flex-end;
    margin-top: 1rem;
}
```

**Existing classes to reuse:**
- `.btn-action` - Edit and Cancel buttons
- `.error-message` - Validation errors in edit form
- `.form-group` - Input field styling in edit form
- `.blog-settings-card` - Container for display/edit swap

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| HTMX Patterns | HIGH | Official examples, established codebase patterns |
| Dialog Accessibility | HIGH | Native element with well-documented behavior |
| Database Operations | HIGH | Standard Go SQL patterns, clear constraint behavior |
| Schema Migration | MEDIUM | SQLite table recreation is well-documented but fiddly |
| Overall Implementation | HIGH | All components proven, low risk |

## Open Questions

None for planning phase. Implementation questions:

1. **Dialog position:** Should dialog be at end of `<body>` or within blog-settings-card?
   - **Answer:** Within card is simpler (card-scoped ID, auto-removal on delete), but end of body is more standard. Either works.

2. **Edit validation:** Should URL be editable or name-only?
   - **Answer:** Roadmap says "display name" - implies name only. Safer too (URL changes could break articles).

3. **Orphaned articles UI:** How to display articles with NULL blog_id?
   - **Answer:** Out of scope for Phase 11. Future phase handles orphaned article display. For now, they're hidden from blog filter.

## Sources

### HTMX Patterns
- [HTMX Click-to-Edit Example](https://htmx.org/examples/click-to-edit/)
- [Hypermedia Systems: HTMX Patterns](https://hypermedia.systems/htmx-patterns/)
- [HTMX Confirmation UI](https://htmx.org/examples/confirm/)

### HTML Dialog Element
- [MDN: dialog Element](https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/dialog)
- [HTML Dialog: Getting Accessibility Right](https://jaredcunha.com/blog/html-dialog-getting-accessibility-and-ux-right)
- [Native Dialog Element Benefits](https://www.oidaisdes.org/native-dialog-element.en/)
- [This Dot Labs: Dialog Accessibility](https://www.thisdot.co/blog/the-html-dialog-element-enhancing-accessibility-and-ease-of-use)

### Database Patterns
- [SQLite Foreign Keys](https://sqlite.org/foreignkeys.html)
- [CASCADE vs SET NULL Best Practices](https://medium.com/@sunnywilson.veshapogu/restrict-vs-cascade-vs-set-null-in-sql-choosing-the-right-foreign-key-rule-6d7c98484710)
- [Go Official: Execute Transactions](https://go.dev/doc/database/execute-transactions)
- [Three Dots Labs: Database Transactions in Go](https://threedots.tech/post/database-transactions-in-go/)

---

## Ready for Planning

Research complete. All patterns validated against existing codebase. Schema migration path identified. Next step: Create detailed implementation plan (11-01-PLAN.md).
