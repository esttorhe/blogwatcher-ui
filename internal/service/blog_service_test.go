// ABOUTME: Tests for blog service business logic.
// ABOUTME: Covers validation, duplicate detection, and error handling.
package service

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
)

func TestAddBlogSuccess(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	svc := NewBlogService(db)

	result, err := svc.AddBlog(context.Background(), AddBlogInput{
		Name: "Test Blog",
		URL:  "https://example.com",
	})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}
	if result.Blog.ID == 0 {
		t.Fatal("expected blog ID")
	}
	if result.Blog.Name != "Test Blog" {
		t.Errorf("expected name 'Test Blog', got %q", result.Blog.Name)
	}
	if result.Blog.URL != "https://example.com" {
		t.Errorf("expected URL 'https://example.com', got %q", result.Blog.URL)
	}
}

func TestAddBlogDuplicateName(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	svc := NewBlogService(db)

	_, err := svc.AddBlog(context.Background(), AddBlogInput{Name: "Test", URL: "https://one.example.com"})
	if err != nil {
		t.Fatalf("first add: %v", err)
	}

	_, err = svc.AddBlog(context.Background(), AddBlogInput{Name: "Test", URL: "https://two.example.com"})
	if err == nil {
		t.Fatal("expected duplicate name error")
	}

	var dupErr BlogAlreadyExistsError
	if !errors.As(err, &dupErr) {
		t.Fatalf("expected BlogAlreadyExistsError, got %T: %v", err, err)
	}
	if dupErr.Field != "name" {
		t.Errorf("expected field 'name', got %q", dupErr.Field)
	}
	if dupErr.Value != "Test" {
		t.Errorf("expected value 'Test', got %q", dupErr.Value)
	}
}

func TestAddBlogDuplicateURL(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	svc := NewBlogService(db)

	_, err := svc.AddBlog(context.Background(), AddBlogInput{Name: "First", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("first add: %v", err)
	}

	_, err = svc.AddBlog(context.Background(), AddBlogInput{Name: "Second", URL: "https://example.com"})
	if err == nil {
		t.Fatal("expected duplicate URL error")
	}

	var dupErr BlogAlreadyExistsError
	if !errors.As(err, &dupErr) {
		t.Fatalf("expected BlogAlreadyExistsError, got %T: %v", err, err)
	}
	if dupErr.Field != "URL" {
		t.Errorf("expected field 'URL', got %q", dupErr.Field)
	}
	if dupErr.Value != "https://example.com" {
		t.Errorf("expected value 'https://example.com', got %q", dupErr.Value)
	}
}

func TestBlogAlreadyExistsErrorMessage(t *testing.T) {
	err := BlogAlreadyExistsError{Field: "name", Value: "TestBlog"}
	expected := "blog with name 'TestBlog' already exists"
	if err.Error() != expected {
		t.Errorf("error message = %q, want %q", err.Error(), expected)
	}

	err = BlogAlreadyExistsError{Field: "URL", Value: "https://example.com"}
	expected = "blog with URL 'https://example.com' already exists"
	if err.Error() != expected {
		t.Errorf("error message = %q, want %q", err.Error(), expected)
	}
}

func TestAddBlogWithProvidedFeedURL(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	svc := NewBlogService(db)

	result, err := svc.AddBlog(context.Background(), AddBlogInput{
		Name:    "Blog with Feed",
		URL:     "https://example.com",
		FeedURL: "https://example.com/custom-feed",
	})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}

	if result.Blog.FeedURL != "https://example.com/custom-feed" {
		t.Errorf("FeedURL = %q, want 'https://example.com/custom-feed'", result.Blog.FeedURL)
	}
	// DiscoveredFeed should be empty since we provided a feed URL
	if result.DiscoveredFeed != "" {
		t.Errorf("DiscoveredFeed = %q, want empty (feed was provided, not discovered)", result.DiscoveredFeed)
	}
}

func TestAddBlogWithScrapeSelector(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	svc := NewBlogService(db)

	result, err := svc.AddBlog(context.Background(), AddBlogInput{
		Name:           "Blog with Selector",
		URL:            "https://example.com",
		ScrapeSelector: "article.post",
	})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}

	if result.Blog.ScrapeSelector != "article.post" {
		t.Errorf("ScrapeSelector = %q, want 'article.post'", result.Blog.ScrapeSelector)
	}
}

func TestAddMultipleBlogsSuccess(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	svc := NewBlogService(db)

	blogs := []AddBlogInput{
		{Name: "Blog A", URL: "https://a.example.com"},
		{Name: "Blog B", URL: "https://b.example.com"},
		{Name: "Blog C", URL: "https://c.example.com"},
	}

	for _, input := range blogs {
		_, err := svc.AddBlog(context.Background(), input)
		if err != nil {
			t.Fatalf("add blog %q: %v", input.Name, err)
		}
	}

	// Verify all blogs were added
	allBlogs, err := db.ListBlogs()
	if err != nil {
		t.Fatalf("list blogs: %v", err)
	}
	if len(allBlogs) != 3 {
		t.Errorf("expected 3 blogs, got %d", len(allBlogs))
	}
}

func openTestDB(t *testing.T) *storage.Database {
	t.Helper()
	path := filepath.Join(t.TempDir(), "blogwatcher.db")
	db, err := storage.OpenDatabase(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	return db
}
