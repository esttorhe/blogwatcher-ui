// ABOUTME: HTTP request handlers with HTMX detection support
// ABOUTME: Handlers return full pages or partial fragments based on HX-Request header
package server

import (
	"log"
	"net/http"
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
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	blogs, err := s.db.ListBlogs()
	if err != nil {
		log.Printf("Error fetching blogs: %v", err)
		blogs = nil
	}

	articles, err := s.db.ListArticles(false, nil)
	if err != nil {
		log.Printf("Error fetching articles: %v", err)
		articles = nil
	}

	data := map[string]interface{}{
		"Title":    "BlogWatcher",
		"Blogs":    blogs,
		"Articles": articles,
	}
	s.renderTemplate(w, "index.gohtml", data)
}

// handleArticleList serves the article list
// Returns partial fragment for HTMX requests, full page otherwise
func (s *Server) handleArticleList(w http.ResponseWriter, r *http.Request) {
	articles, err := s.db.ListArticles(false, nil)
	if err != nil {
		log.Printf("Error fetching articles: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Articles": articles,
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
