// ABOUTME: HTTP server package with dependency injection pattern
// ABOUTME: Manages server lifecycle, template parsing, and request routing
package server

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/esttorhe/blogwatcher-ui/internal/storage"
)

// Server represents the HTTP server with all dependencies
type Server struct {
	db        *storage.Database
	templates *template.Template
	mux       *http.ServeMux
}

// NewServer creates a new HTTP server with dependency injection
// Parses all templates at startup and registers routes
func NewServer(db *storage.Database) (http.Handler, error) {
	// Parse all templates once at startup
	tmpl, err := template.ParseGlob("templates/**/*.gohtml")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	// Create server with dependencies
	s := &Server{
		db:        db,
		templates: tmpl,
		mux:       http.NewServeMux(),
	}

	// Register all routes
	s.registerRoutes()

	return s, nil
}

// ServeHTTP implements http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
