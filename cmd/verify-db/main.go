// ABOUTME: Temporary verification program to test database connection.
// ABOUTME: Opens the blogwatcher database and reports blog/article counts.
package main

import (
	"fmt"
	"log"

	"github.com/esttorhe/blogwatcher-ui/internal/storage"
)

func main() {
	db, err := storage.OpenDatabase("")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	blogs, err := db.ListBlogs()
	if err != nil {
		log.Fatalf("Failed to list blogs: %v", err)
	}

	articles, err := db.ListArticles(false, nil)
	if err != nil {
		log.Fatalf("Failed to list articles: %v", err)
	}

	fmt.Printf("✓ Database connection successful\n")
	fmt.Printf("  Database path: %s\n", db.Path())
	fmt.Printf("  Blogs: %d\n", len(blogs))
	fmt.Printf("  Articles: %d\n", len(articles))
}
