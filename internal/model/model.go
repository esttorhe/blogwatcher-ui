// ABOUTME: Defines Blog and Article data models matching the database schema.
// ABOUTME: These structs represent the core entities for tracking blogs and their articles.
package model

import "time"

type Blog struct {
	ID             int64
	Name           string
	URL            string
	FeedURL        string
	ScrapeSelector string
	LastScanned    *time.Time
}

type Article struct {
	ID             int64
	BlogID         int64
	Title          string
	URL            string
	PublishedDate  *time.Time
	DiscoveredDate *time.Time
	IsRead         bool
}

// ArticleWithBlog extends Article with blog metadata for display in article cards.
// Used when rendering article lists where blog name and favicon are needed.
type ArticleWithBlog struct {
	ID             int64
	BlogID         int64
	Title          string
	URL            string
	PublishedDate  *time.Time
	DiscoveredDate *time.Time
	IsRead         bool
	BlogName       string
	BlogURL        string
}
