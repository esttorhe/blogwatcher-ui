// ABOUTME: HTTP request handlers with HTMX detection support
// ABOUTME: Handlers return full pages or partial fragments based on HX-Request header
package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/esttorhe/blogwatcher-ui/internal/model"
	"github.com/esttorhe/blogwatcher-ui/internal/scanner"
)

// renderTemplate executes a named template with the given data
// Logs errors and returns 500 status on failure
func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	err := s.templates.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Printf("Error rendering template %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleIndex serves the main index page
// Fetches both blogs and articles for initial render
// Supports filter and blog query params for direct URL access
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	blogs, err := s.db.ListBlogs()
	if err != nil {
		log.Printf("Error fetching blogs: %v", err)
		blogs = nil
	}

	// Parse query parameters for filter and blog
	filter := r.URL.Query().Get("filter")
	blogParam := r.URL.Query().Get("blog")

	// Parse blogID if provided (0 means no filter, DB IDs start at 1)
	var blogID *int64
	var currentBlogID int64
	if blogParam != "" {
		if id, err := strconv.ParseInt(blogParam, 10, 64); err == nil {
			blogID = &id
			currentBlogID = id
		}
	}

	// Fetch articles based on filter (using ListArticlesWithBlog for rich metadata)
	var articles []model.ArticleWithBlog
	switch filter {
	case "read":
		articles, err = s.db.ListArticlesWithBlog(true, blogID)
	case "unread", "":
		// Default to unread (inbox view)
		articles, err = s.db.ListArticlesWithBlog(false, blogID)
		if filter == "" {
			filter = "unread" // Set default for template active state
		}
	default:
		// Unknown filter, default to unread
		articles, err = s.db.ListArticlesWithBlog(false, blogID)
		filter = "unread"
	}
	if err != nil {
		log.Printf("Error fetching articles: %v", err)
		articles = nil
	}

	data := map[string]interface{}{
		"Title":         "BlogWatcher",
		"Blogs":         blogs,
		"Articles":      articles,
		"CurrentFilter": filter,
		"CurrentBlogID": currentBlogID, // 0 means no blog filter active
	}
	s.renderTemplate(w, "index.gohtml", data)
}

// handleArticleList serves the article list
// Returns partial fragment for HTMX requests, full page otherwise
// Supports filter (unread/read) and blog query parameters
func (s *Server) handleArticleList(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filter := r.URL.Query().Get("filter")
	blogParam := r.URL.Query().Get("blog")

	// Parse blogID if provided (0 means no filter, DB IDs start at 1)
	var blogID *int64
	var currentBlogID int64
	if blogParam != "" {
		if id, err := strconv.ParseInt(blogParam, 10, 64); err == nil {
			blogID = &id
			currentBlogID = id
		}
	}

	// Fetch articles based on filter (using ListArticlesWithBlog for rich metadata)
	var articles []model.ArticleWithBlog
	var err error
	switch filter {
	case "read":
		articles, err = s.db.ListArticlesWithBlog(true, blogID)
	case "unread", "":
		// Default to unread (inbox view)
		articles, err = s.db.ListArticlesWithBlog(false, blogID)
		if filter == "" {
			filter = "unread" // Set default for template active state
		}
	default:
		// Unknown filter, default to unread
		articles, err = s.db.ListArticlesWithBlog(false, blogID)
		filter = "unread"
	}
	if err != nil {
		log.Printf("Error fetching articles: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Articles":      articles,
		"CurrentFilter": filter,
		"CurrentBlogID": currentBlogID, // 0 means no blog filter active
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Return partial fragment for HTMX
		s.renderTemplate(w, "article-list.gohtml", data)
	} else {
		// Return full page for direct navigation
		data["Title"] = "BlogWatcher"
		blogs, err := s.db.ListBlogs()
		if err != nil {
			log.Printf("Error fetching blogs: %v", err)
		} else {
			data["Blogs"] = blogs
		}
		s.renderTemplate(w, "index.gohtml", data)
	}
}

// handleBlogList serves the blog list
// Returns partial fragment for HTMX requests, full page otherwise
func (s *Server) handleBlogList(w http.ResponseWriter, r *http.Request) {
	blogs, err := s.db.ListBlogs()
	if err != nil {
		log.Printf("Error fetching blogs: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Blogs": blogs,
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Return partial fragment for HTMX
		s.renderTemplate(w, "blog-list.gohtml", data)
	} else {
		// Return full page for direct navigation
		data["Title"] = "BlogWatcher"
		articles, err := s.db.ListArticles(false, nil)
		if err != nil {
			log.Printf("Error fetching articles: %v", err)
		} else {
			data["Articles"] = articles
		}
		s.renderTemplate(w, "index.gohtml", data)
	}
}

// handleMarkRead marks an article as read and returns empty response for HTMX swap
func (s *Server) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	found, err := s.db.MarkArticleRead(id)
	if err != nil {
		log.Printf("Error marking article %d as read: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	// Return 200 OK with empty body - HTMX outerHTML swap will remove the card
	w.WriteHeader(http.StatusOK)
}

// handleMarkUnread marks an article as unread and returns empty response for HTMX swap
func (s *Server) handleMarkUnread(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	found, err := s.db.MarkArticleUnread(id)
	if err != nil {
		log.Printf("Error marking article %d as unread: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	// Return 200 OK with empty body - HTMX outerHTML swap will remove the card
	w.WriteHeader(http.StatusOK)
}

// handleMarkAllRead marks all unread articles as read and returns refreshed article list
func (s *Server) handleMarkAllRead(w http.ResponseWriter, r *http.Request) {
	// Parse optional blog filter from query params
	blogParam := r.URL.Query().Get("blog")
	var blogID *int64
	var currentBlogID int64
	if blogParam != "" && blogParam != "0" {
		if id, err := strconv.ParseInt(blogParam, 10, 64); err == nil {
			blogID = &id
			currentBlogID = id
		}
	}

	// Mark all unread as read
	if err := s.db.MarkAllUnreadArticlesRead(blogID); err != nil {
		log.Printf("Error marking all articles as read: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Return refreshed article list (will be empty for inbox view)
	articles, err := s.db.ListArticlesWithBlog(false, blogID) // unread = false is inbox
	if err != nil {
		log.Printf("Error fetching articles: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Articles":      articles,
		"CurrentFilter": "unread",
		"CurrentBlogID": currentBlogID,
	}
	s.renderTemplate(w, "article-list.gohtml", data)
}

// handleSync triggers a scan of all blogs and returns refreshed article list
func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	// Run scanner with single worker to avoid SQLite write conflicts
	results, err := scanner.ScanAllBlogs(s.db, 1)
	if err != nil {
		log.Printf("Sync failed: %v", err)
		http.Error(w, "Sync failed", http.StatusInternalServerError)
		return
	}

	// Log results
	totalNew := 0
	for _, result := range results {
		if result.Error != "" {
			log.Printf("Sync error for %s: %s", result.BlogName, result.Error)
		} else {
			log.Printf("Synced %s: %d new articles (source: %s)", result.BlogName, result.NewArticles, result.Source)
			totalNew += result.NewArticles
		}
	}
	log.Printf("Sync complete: %d blogs scanned, %d new articles total", len(results), totalNew)

	// Parse current filter from query params to return appropriate view
	filter := r.URL.Query().Get("filter")
	blogParam := r.URL.Query().Get("blog")
	var blogID *int64
	var currentBlogID int64
	if blogParam != "" && blogParam != "0" {
		if id, err := strconv.ParseInt(blogParam, 10, 64); err == nil {
			blogID = &id
			currentBlogID = id
		}
	}

	// Determine read status based on filter
	isRead := false
	if filter == "read" {
		isRead = true
	} else {
		filter = "unread"
	}

	// Return refreshed article list
	articles, err := s.db.ListArticlesWithBlog(isRead, blogID)
	if err != nil {
		log.Printf("Error fetching articles: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Articles":      articles,
		"CurrentFilter": filter,
		"CurrentBlogID": currentBlogID,
	}
	s.renderTemplate(w, "article-list.gohtml", data)
}
