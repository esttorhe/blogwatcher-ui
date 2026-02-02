// ABOUTME: HTTP request handlers with HTMX detection support
// ABOUTME: Handlers return full pages or partial fragments based on HX-Request header
package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/esttorhe/blogwatcher-ui/internal/model"
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

	// Fetch articles based on filter
	var articles []model.Article
	switch filter {
	case "read":
		articles, err = s.db.ListArticlesByReadStatus(true, blogID)
	case "unread", "":
		// Default to unread (inbox view)
		articles, err = s.db.ListArticlesByReadStatus(false, blogID)
		if filter == "" {
			filter = "unread" // Set default for template active state
		}
	default:
		// Unknown filter, default to unread
		articles, err = s.db.ListArticlesByReadStatus(false, blogID)
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

	// Fetch articles based on filter
	var articles []model.Article
	var err error
	switch filter {
	case "read":
		articles, err = s.db.ListArticlesByReadStatus(true, blogID)
	case "unread", "":
		// Default to unread (inbox view)
		articles, err = s.db.ListArticlesByReadStatus(false, blogID)
		if filter == "" {
			filter = "unread" // Set default for template active state
		}
	default:
		// Unknown filter, default to unread
		articles, err = s.db.ListArticlesByReadStatus(false, blogID)
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
