# Plan 10-01 Summary: Add Blog Handler, Form, and FAB

**Status:** COMPLETE
**Executed:** 2026-02-09

## What Was Built

### 1. Add Blog Handler (`internal/server/handlers.go`)

- `handleAddBlog` handler with CLI integration
- Uses `exec.CommandContext` with 30-second timeout
- Executes `blogwatcher add <name> <url>` command
- Captures stdout/stderr for error handling
- Queries database for discovered feed URL after successful add
- Auto-syncs new blog using `scanner.ScanBlogByName` (efficient single-blog sync)
- Helper methods: `autoSyncNewBlog`, `renderAddBlogError`, `renderAddBlogSuccess`

### 2. Route Registration (`internal/server/routes.go`)

- Added `POST /blogs/add` route mapped to `handleAddBlog`

### 3. Add Blog Form Template (`assets/templates/partials/add-blog-form.gohtml`)

- HTMX form with `hx-post="/blogs/add"`
- Form fields: name (text), URL (url input)
- Loading indicator with `hx-indicator`
- Success state shows blog name and discovered feed URL
- Error state pre-populates form values
- "Back to Settings" button on success

### 4. Settings Page Integration (`assets/templates/partials/settings-page.gohtml`)

- Added `{{template "add-blog-form.gohtml" .}}` at top of settings page
- Form appears above "Tracked Blogs" section

### 5. FAB Component (`assets/templates/pages/index.gohtml`)

- Floating action button with plus icon SVG
- Fixed position bottom-right
- HTMX navigation to settings page (`hx-get="/settings"`)
- Accessible with `title` and `aria-label` attributes

### 6. CSS Styles (`assets/static/styles.css`)

- `.add-blog-section` - Form container with max-width and border
- `.form-group` - Label and input styling with focus states
- `.error-message` - Red alert box for errors
- `.success-message` - Green alert box for success
- `.add-blog-spinner` - Loading indicator toggle
- `.fab` - Floating action button (56px desktop, 48px mobile)
- Mobile responsive styles (@media max-width: 768px)

## Requirements Satisfied

| Requirement | Description | Status |
|-------------|-------------|--------|
| ADD-01 | User can enter blog URL in add form | Done |
| ADD-02 | System auto-discovers RSS/Atom feed via CLI | Done |
| ADD-03 | User sees success/error feedback | Done |
| ADD-04 | System displays discovered feed URL | Done |
| ADD-05 | System auto-syncs newly added blog | Done |
| ADD-06 | FAB quick access on article list | Done |

## Key Implementation Details

- **CLI Integration:** Uses `exec.LookPath` to find blogwatcher binary, `exec.CommandContext` for timeout
- **Error Handling:** Parses stderr from CLI, strips "Error: " prefix
- **Feed URL Display:** Queries `db.GetBlogByName(name)` after CLI success to get feed URL
- **Auto-sync:** Uses `scanner.ScanBlogByName` (not ScanAllBlogs) for efficient single-blog sync
- **Form Pre-population:** Error state includes `Name` and `URL` in template data

## Verification

- Build succeeds: `go build -o blogwatcher-ui ./cmd/server`
- Handler exists at line 434 of handlers.go
- Route exists at line 30 of routes.go
- Template created with HTMX form
- FAB added to index.gohtml with SVG icon
- CSS styles added for form and FAB

## Files Modified

1. `internal/server/handlers.go` - Added handleAddBlog + helpers
2. `internal/server/routes.go` - Added POST /blogs/add route
3. `assets/templates/partials/add-blog-form.gohtml` - New file
4. `assets/templates/partials/settings-page.gohtml` - Added form include
5. `assets/templates/pages/index.gohtml` - Added FAB component
6. `assets/static/styles.css` - Added form and FAB styles
