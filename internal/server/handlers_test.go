// ABOUTME: Tests for HTTP handler functions.
// ABOUTME: Covers blog addition, validation, and error handling via HTTP endpoints.
package server

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/esttorhe/blogwatcher-ui/v2/assets"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
)

func TestHandleAddBlogSuccess(t *testing.T) {
	srv := createTestServer(t)

	form := url.Values{}
	form.Set("name", "Test Blog")
	form.Set("url", "https://example.com")

	req := httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Check response contains success indicator
	body := rec.Body.String()
	if !strings.Contains(body, "Test Blog") {
		t.Errorf("response should contain blog name, got: %s", body)
	}
}

func TestHandleAddBlogDuplicateName(t *testing.T) {
	srv := createTestServer(t)

	// Add first blog
	form := url.Values{}
	form.Set("name", "Duplicate")
	form.Set("url", "https://first.example.com")

	req := httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("first add: status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Try to add blog with same name
	form = url.Values{}
	form.Set("name", "Duplicate")
	form.Set("url", "https://second.example.com")

	req = httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("duplicate add: status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Should contain error message about duplicate
	body := rec.Body.String()
	if !strings.Contains(body, "already exists") {
		t.Errorf("response should contain 'already exists', got: %s", body)
	}
}

func TestHandleAddBlogDuplicateURL(t *testing.T) {
	srv := createTestServer(t)

	// Add first blog
	form := url.Values{}
	form.Set("name", "First Blog")
	form.Set("url", "https://example.com")

	req := httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("first add: status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Try to add blog with same URL
	form = url.Values{}
	form.Set("name", "Second Blog")
	form.Set("url", "https://example.com")

	req = httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("duplicate add: status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Should contain error message about duplicate
	body := rec.Body.String()
	if !strings.Contains(body, "already exists") {
		t.Errorf("response should contain 'already exists', got: %s", body)
	}
}

func TestHandleAddBlogValidationEmptyName(t *testing.T) {
	srv := createTestServer(t)

	form := url.Values{}
	form.Set("name", "")
	form.Set("url", "https://example.com")

	req := httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "required") {
		t.Errorf("response should contain 'required', got: %s", body)
	}
}

func TestHandleAddBlogValidationEmptyURL(t *testing.T) {
	srv := createTestServer(t)

	form := url.Values{}
	form.Set("name", "Test Blog")
	form.Set("url", "")

	req := httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "required") {
		t.Errorf("response should contain 'required', got: %s", body)
	}
}

func TestHandleAddBlogValidationBothEmpty(t *testing.T) {
	srv := createTestServer(t)

	form := url.Values{}
	form.Set("name", "   ") // Whitespace only
	form.Set("url", "   ")  // Whitespace only

	req := httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "required") {
		t.Errorf("response should contain 'required', got: %s", body)
	}
}

func TestArticleListHeaderShowsBlogName(t *testing.T) {
	srv := createTestServer(t)

	// Add a blog first
	form := url.Values{}
	form.Set("name", "My Cool Blog")
	form.Set("url", "https://coolblog.example.com")

	req := httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("add blog: status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Request articles filtered by blog=1 via HTMX
	req = httptest.NewRequest(http.MethodGet, "/articles?blog=1", nil)
	req.Header.Set("HX-Request", "true")
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("articles: status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "My Cool Blog") {
		t.Errorf("header should contain blog name 'My Cool Blog', got: %s", body)
	}
}

func TestArticleListHeaderShowsInboxWithoutBlogFilter(t *testing.T) {
	srv := createTestServer(t)

	// Request articles without blog filter via HTMX
	req := httptest.NewRequest(http.MethodGet, "/articles?filter=unread", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("articles: status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Inbox") {
		t.Errorf("header should contain 'Inbox' when no blog filter, got: %s", body)
	}
}

func TestHandleAPISync_Success(t *testing.T) {
	srv := createTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/sync", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Verify Content-Type is JSON
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	// Verify response is valid JSON with expected fields
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	// Check required top-level fields exist
	for _, field := range []string{"blogs_scanned", "new_articles"} {
		if _, ok := resp[field]; !ok {
			t.Errorf("response missing field %q", field)
		}
	}
}

func TestHandleAPISync_GetDoesNotReturnJSON(t *testing.T) {
	srv := createTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sync", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	// GET /api/sync should not hit the API handler (falls through to index)
	ct := rec.Header().Get("Content-Type")
	if ct == "application/json" {
		t.Errorf("GET /api/sync should not return JSON, got Content-Type = %q", ct)
	}
}

func createTestServer(t *testing.T) http.Handler {
	t.Helper()

	// Create temp database
	path := filepath.Join(t.TempDir(), "blogwatcher.db")
	db, err := storage.OpenDatabase(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Extract embedded filesystems
	staticFiles, err := fs.Sub(assets.StaticFS, "static")
	if err != nil {
		t.Fatalf("extract static: %v", err)
	}
	templateFiles, err := fs.Sub(assets.TemplateFS, "templates")
	if err != nil {
		t.Fatalf("extract templates: %v", err)
	}

	// Create server
	srv, err := NewServerWithFS(db, templateFiles, staticFiles, "test")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	return srv
}
