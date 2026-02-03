// ABOUTME: Route registration for the HTTP server
// ABOUTME: Uses Go 1.22+ method routing with http.ServeMux
package server

import (
	"net/http"
)

// registerRoutes sets up all HTTP routes for the server
func (s *Server) registerRoutes() {
	// Static files from embedded filesystem (already extracted in main.go)
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(s.staticFS))))

	// Pages
	s.mux.HandleFunc("GET /", s.handleIndex)
	s.mux.HandleFunc("GET /articles", s.handleArticleList)
	s.mux.HandleFunc("GET /blogs", s.handleBlogList)

	// Article management actions
	s.mux.HandleFunc("POST /articles/{id}/read", s.handleMarkRead)
	s.mux.HandleFunc("POST /articles/{id}/unread", s.handleMarkUnread)
	s.mux.HandleFunc("POST /articles/mark-all-read", s.handleMarkAllRead)

	// Sync
	s.mux.HandleFunc("POST /sync", s.handleSync)
}
