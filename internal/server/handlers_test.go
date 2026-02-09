// ABOUTME: Tests for HTTP handler functions.
// ABOUTME: Covers blog addition, validation, and error handling via HTTP endpoints.
package server

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/esttorhe/blogwatcher-ui/assets"
	"github.com/esttorhe/blogwatcher-ui/internal/storage"
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
