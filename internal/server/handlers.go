// ABOUTME: HTTP request handlers with HTMX detection support
// ABOUTME: Handlers return full pages or partial fragments based on HX-Request header
package server

import (
	"log"
	"net/http"
	"strconv"
	"time"

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
// Supports filter, blog, search, and date query params for direct URL access
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	blogs, err := s.db.ListBlogs()
	if err != nil {
		log.Printf("Error fetching blogs: %v", err)
		blogs = nil
	}

	// Build search options from query parameters
	opts, filter, currentBlogID := parseSearchOptions(r)

	// Fetch articles using SearchArticles for all filter combinations
	articles, articleCount, err := s.db.SearchArticles(opts)
	if err != nil {
		log.Printf("Error fetching articles: %v", err)
		articles = nil
		articleCount = 0
	}

	// Calculate if there are more articles
	pageSize := opts.Limit
	if pageSize <= 0 {
		pageSize = model.DefaultPageSize
	}
	hasMore := len(articles) == pageSize && opts.Offset+len(articles) < articleCount
	nextOffset := opts.Offset + len(articles)
	displayedCount := opts.Offset + len(articles)

	data := map[string]interface{}{
		"Title":          "BlogWatcher",
		"Blogs":          blogs,
		"Articles":       articles,
		"ArticleCount":   articleCount,
		"DisplayedCount": displayedCount,
		"CurrentFilter":  filter,
		"CurrentBlogID":  currentBlogID, // 0 means no blog filter active
		"SearchQuery":    opts.SearchQuery,
		"DateFrom":       r.URL.Query().Get("date_from"),
		"DateTo":         r.URL.Query().Get("date_to"),
		"Version":        s.version,
		"HasMore":        hasMore,
		"NextOffset":     nextOffset,
	}
	s.renderTemplate(w, "index.gohtml", data)
}

// handleArticleList serves the article list
// Returns partial fragment for HTMX requests, full page otherwise
// Supports filter, blog, search, and date query parameters
func (s *Server) handleArticleList(w http.ResponseWriter, r *http.Request) {
	// Build search options from query parameters
	opts, filter, currentBlogID := parseSearchOptions(r)

	// Fetch articles using SearchArticles for all filter combinations
	articles, articleCount, err := s.db.SearchArticles(opts)
	if err != nil {
		log.Printf("Error fetching articles: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Calculate if there are more articles
	pageSize := opts.Limit
	if pageSize <= 0 {
		pageSize = model.DefaultPageSize
	}
	hasMore := len(articles) == pageSize && opts.Offset+len(articles) < articleCount
	nextOffset := opts.Offset + len(articles)
	displayedCount := opts.Offset + len(articles)

	data := map[string]interface{}{
		"Articles":       articles,
		"ArticleCount":   articleCount,
		"DisplayedCount": displayedCount,
		"CurrentFilter":  filter,
		"CurrentBlogID":  currentBlogID, // 0 means no blog filter active
		"SearchQuery":    opts.SearchQuery,
		"DateFrom":       r.URL.Query().Get("date_from"),
		"DateTo":         r.URL.Query().Get("date_to"),
		"HasMore":        hasMore,
		"NextOffset":     nextOffset,
		"IsLoadMore":     opts.Offset > 0,
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// If this is a "load more" request (offset > 0), return just the articles
		if opts.Offset > 0 {
			s.renderTemplate(w, "article-items.gohtml", data)
			return
		}
		// Return partial fragment for HTMX
		s.renderTemplate(w, "article-list.gohtml", data)
		return
	}

	// Return full page for direct navigation
	data["Title"] = "BlogWatcher"
	data["Version"] = s.version
	blogs, err := s.db.ListBlogs()
	if err != nil {
		log.Printf("Error fetching blogs: %v", err)
	} else {
		data["Blogs"] = blogs
	}
	s.renderTemplate(w, "index.gohtml", data)
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
		data["Version"] = s.version
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
	if blogParam != "" && blogParam != "0" {
		if id, err := strconv.ParseInt(blogParam, 10, 64); err == nil {
			blogID = &id
		}
	}

	// Mark all unread as read
	if err := s.db.MarkAllUnreadArticlesRead(blogID); err != nil {
		log.Printf("Error marking all articles as read: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Build search options from query parameters (preserves search/date filters)
	opts, filter, currentBlogID := parseSearchOptions(r)

	// Return refreshed article list with current filters
	articles, articleCount, err := s.db.SearchArticles(opts)
	if err != nil {
		log.Printf("Error fetching articles: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Articles":      articles,
		"ArticleCount":  articleCount,
		"CurrentFilter": filter,
		"CurrentBlogID": currentBlogID,
		"SearchQuery":   opts.SearchQuery,
		"DateFrom":      r.URL.Query().Get("date_from"),
		"DateTo":        r.URL.Query().Get("date_to"),
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

	// Build search options from query parameters (preserves all filters)
	opts, filter, currentBlogID := parseSearchOptions(r)

	// Return refreshed article list with current filters
	articles, articleCount, err := s.db.SearchArticles(opts)
	if err != nil {
		log.Printf("Error fetching articles: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Articles":      articles,
		"ArticleCount":  articleCount,
		"CurrentFilter": filter,
		"CurrentBlogID": currentBlogID,
		"SearchQuery":   opts.SearchQuery,
		"DateFrom":      r.URL.Query().Get("date_from"),
		"DateTo":        r.URL.Query().Get("date_to"),
	}
	s.renderTemplate(w, "article-list.gohtml", data)
}

// handleSyncThumbnails re-fetches thumbnails for articles missing them
func (s *Server) handleSyncThumbnails(w http.ResponseWriter, r *http.Request) {
	result, err := scanner.SyncThumbnails(s.db)
	if err != nil {
		log.Printf("Thumbnail sync failed: %v", err)
		http.Error(w, "Thumbnail sync failed", http.StatusInternalServerError)
		return
	}

	log.Printf("Thumbnail sync complete: %d total, %d updated, %d errors", result.Total, result.Updated, result.Errors)

	// Build search options from query parameters (preserves all filters)
	opts, filter, currentBlogID := parseSearchOptions(r)

	// Return refreshed article list with current filters
	articles, articleCount, err := s.db.SearchArticles(opts)
	if err != nil {
		log.Printf("Error fetching articles: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Articles":      articles,
		"ArticleCount":  articleCount,
		"CurrentFilter": filter,
		"CurrentBlogID": currentBlogID,
		"SearchQuery":   opts.SearchQuery,
		"DateFrom":      r.URL.Query().Get("date_from"),
		"DateTo":        r.URL.Query().Get("date_to"),
	}
	s.renderTemplate(w, "article-list.gohtml", data)
}

// parseSearchOptions extracts all search and filter parameters from the request.
// Returns SearchOptions, the filter string (for template), and currentBlogID.
func parseSearchOptions(r *http.Request) (model.SearchOptions, string, int64) {
	opts := model.SearchOptions{
		SearchQuery: r.URL.Query().Get("search"),
	}

	// Parse status filter
	filter := r.URL.Query().Get("filter")
	switch filter {
	case "read":
		isRead := true
		opts.IsRead = &isRead
	case "unread", "":
		isRead := false
		opts.IsRead = &isRead
		filter = "unread" // Default
	default:
		isRead := false
		opts.IsRead = &isRead
		filter = "unread"
	}

	// Parse blog filter
	var currentBlogID int64
	if blogParam := r.URL.Query().Get("blog"); blogParam != "" && blogParam != "0" {
		if id, err := strconv.ParseInt(blogParam, 10, 64); err == nil {
			opts.BlogID = &id
			currentBlogID = id
		}
	}

	// Parse date filters (format: 2006-01-02)
	if dateFrom := r.URL.Query().Get("date_from"); dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			opts.DateFrom = &t
		}
	}
	if dateTo := r.URL.Query().Get("date_to"); dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			opts.DateTo = &t
		}
	}

	// Parse pagination
	if offsetParam := r.URL.Query().Get("offset"); offsetParam != "" {
		if offset, err := strconv.Atoi(offsetParam); err == nil && offset >= 0 {
			opts.Offset = offset
		}
	}
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if limit, err := strconv.Atoi(limitParam); err == nil && limit > 0 {
			opts.Limit = limit
		}
	}

	return opts, filter, currentBlogID
}
