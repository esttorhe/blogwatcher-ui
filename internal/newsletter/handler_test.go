// ABOUTME: Tests for the newsletter email ingestion handler.
// ABOUTME: Uses real RFC 822 fixture files in testdata/ — no mocks.
package newsletter_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/newsletter"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
)

func openTestDB(t *testing.T) *storage.Database {
	t.Helper()
	db, err := storage.OpenDatabase(filepath.Join(t.TempDir(), "bw.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return data
}

func TestHandleInboundHTMLOnly(t *testing.T) {
	db := openTestDB(t)
	h := newsletter.NewHandler(db)

	raw := readFixture(t, "html_only.eml")
	article, err := h.HandleInbound(context.Background(), raw)
	if err != nil {
		t.Fatalf("HandleInbound: %v", err)
	}

	if article.Title != "Issue 42 - Big News" {
		t.Errorf("Title = %q, want %q", article.Title, "Issue 42 - Big News")
	}
	if !strings.Contains(article.Content, "<p>Welcome to issue 42!</p>") {
		t.Errorf("Content does not contain expected HTML; got: %s", article.Content)
	}
	if article.URL != "message:<issue42@acme.com>" {
		t.Errorf("URL = %q, want %q", article.URL, "message:<issue42@acme.com>")
	}
	if article.BlogID == 0 {
		t.Error("BlogID must be set after ingestion")
	}
}

func TestHandleInboundMultipart(t *testing.T) {
	db := openTestDB(t)
	h := newsletter.NewHandler(db)

	raw := readFixture(t, "multipart.eml")
	article, err := h.HandleInbound(context.Background(), raw)
	if err != nil {
		t.Fatalf("HandleInbound: %v", err)
	}

	if article.Title != "Issue 43 - Multipart" {
		t.Errorf("Title = %q, want %q", article.Title, "Issue 43 - Multipart")
	}
	if !strings.Contains(article.Content, "<p>HTML content for issue 43.</p>") {
		t.Errorf("Content should prefer HTML part; got: %s", article.Content)
	}
}

func TestHandleInboundNoDisplayName(t *testing.T) {
	db := openTestDB(t)
	h := newsletter.NewHandler(db)

	raw := readFixture(t, "no_display_name.eml")
	article, err := h.HandleInbound(context.Background(), raw)
	if err != nil {
		t.Fatalf("HandleInbound: %v", err)
	}

	// When From has no display name, use the email address as the blog name.
	if article.BlogID == 0 {
		t.Error("BlogID must be set")
	}
}

func TestHandleInboundIdempotent(t *testing.T) {
	db := openTestDB(t)
	h := newsletter.NewHandler(db)

	raw := readFixture(t, "html_only.eml")
	first, err := h.HandleInbound(context.Background(), raw)
	if err != nil {
		t.Fatalf("first HandleInbound: %v", err)
	}
	// Second call with the same Message-ID should return same article without error.
	second, err := h.HandleInbound(context.Background(), raw)
	if err != nil {
		t.Fatalf("second HandleInbound: %v", err)
	}
	if first.ID != second.ID {
		t.Errorf("expected same article ID on repeated ingestion: first=%d second=%d", first.ID, second.ID)
	}
}

func TestHandleInboundSameSenderSameBlog(t *testing.T) {
	db := openTestDB(t)
	h := newsletter.NewHandler(db)

	raw1 := readFixture(t, "html_only.eml")
	raw2 := readFixture(t, "multipart.eml")

	a1, err := h.HandleInbound(context.Background(), raw1)
	if err != nil {
		t.Fatalf("first ingest: %v", err)
	}
	a2, err := h.HandleInbound(context.Background(), raw2)
	if err != nil {
		t.Fatalf("second ingest: %v", err)
	}

	// Both emails are from news@acme.com — they must share the same blog.
	if a1.BlogID != a2.BlogID {
		t.Errorf("expected same BlogID for same sender: %d vs %d", a1.BlogID, a2.BlogID)
	}
}
