// ABOUTME: Business logic layer for blog operations with validation and feed discovery.
// ABOUTME: Provides clean error types and orchestrates storage and RSS operations.
package service

import (
	"context"
	"fmt"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/rss"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
)

// BlogAlreadyExistsError indicates a blog with the same name or URL already exists.
type BlogAlreadyExistsError struct {
	Field string // "name" or "URL"
	Value string
}

func (e BlogAlreadyExistsError) Error() string {
	return fmt.Sprintf("blog with %s '%s' already exists", e.Field, e.Value)
}

// BlogService provides business logic for blog operations.
type BlogService struct {
	db *storage.Database
}

// NewBlogService creates a new BlogService with the given database.
func NewBlogService(db *storage.Database) *BlogService {
	return &BlogService{db: db}
}

// AddBlogInput contains the parameters for adding a new blog.
type AddBlogInput struct {
	Name           string
	URL            string
	FeedURL        string // Optional, will be auto-discovered if empty
	ScrapeSelector string // Optional
}

// AddBlogResult contains the result of adding a blog.
type AddBlogResult struct {
	Blog           model.Blog
	DiscoveredFeed string // The feed URL if it was auto-discovered (empty if provided or not found)
}

// AddBlog validates input, checks for duplicates, discovers feed URL if needed,
// and creates the blog. Returns BlogAlreadyExistsError for duplicates.
func (s *BlogService) AddBlog(ctx context.Context, input AddBlogInput) (AddBlogResult, error) {
	var result AddBlogResult

	// Check for duplicate name
	existing, err := s.db.GetBlogByName(input.Name)
	if err != nil {
		return result, fmt.Errorf("failed to check blog name: %w", err)
	}
	if existing != nil {
		return result, BlogAlreadyExistsError{Field: "name", Value: input.Name}
	}

	// Check for duplicate URL
	existing, err = s.db.GetBlogByURL(input.URL)
	if err != nil {
		return result, fmt.Errorf("failed to check blog URL: %w", err)
	}
	if existing != nil {
		return result, BlogAlreadyExistsError{Field: "URL", Value: input.URL}
	}

	// Discover feed URL if not provided
	feedURL := input.FeedURL
	if feedURL == "" {
		discovered, _ := rss.DiscoverFeedURL(ctx, input.URL)
		if discovered != "" {
			feedURL = discovered
			result.DiscoveredFeed = discovered
		}
	}

	// Create the blog
	blog := model.Blog{
		Name:           input.Name,
		URL:            input.URL,
		FeedURL:        feedURL,
		ScrapeSelector: input.ScrapeSelector,
	}

	blog, err = s.db.AddBlog(blog)
	if err != nil {
		return result, fmt.Errorf("failed to save blog: %w", err)
	}

	result.Blog = blog
	return result, nil
}
