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
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "BlogWatcher",
	}
	s.renderTemplate(w, "index.gohtml", data)
}

// handleArticleList serves the article list
// Returns partial fragment for HTMX requests, full page otherwise
func (s *Server) handleArticleList(w http.ResponseWriter, r *http.Request) {
	// Placeholder data - will be replaced with real database queries
	data := map[string]interface{}{
		"Articles": []interface{}{},
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Return partial fragment for HTMX
		s.renderTemplate(w, "article-list.gohtml", data)
	} else {
		// Return full page for direct navigation
		s.renderTemplate(w, "index.gohtml", data)
	}
}

// handleBlogList serves the blog list
// Returns partial fragment for HTMX requests, full page otherwise
func (s *Server) handleBlogList(w http.ResponseWriter, r *http.Request) {
	// Placeholder data - will be replaced with real database queries
	data := map[string]interface{}{
		"Blogs": []interface{}{},
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Return partial fragment for HTMX
		s.renderTemplate(w, "blog-list.gohtml", data)
	} else {
		// Return full page for direct navigation
		s.renderTemplate(w, "index.gohtml", data)
	}
}
