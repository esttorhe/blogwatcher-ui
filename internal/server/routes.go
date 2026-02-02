// ABOUTME: Route registration for the HTTP server
// ABOUTME: Uses Go 1.22+ method routing with http.ServeMux
package server

import (
	"net/http"
)

// registerRoutes sets up all HTTP routes for the server
func (s *Server) registerRoutes() {
	// Static files
	fs := http.FileServer(http.Dir("static"))
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	// Pages
	s.mux.HandleFunc("GET /", s.handleIndex)
	s.mux.HandleFunc("GET /articles", s.handleArticleList)
	s.mux.HandleFunc("GET /blogs", s.handleBlogList)
}
