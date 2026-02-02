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
	// Register template functions BEFORE parsing templates
	funcMap := template.FuncMap{
		"timeAgo":    timeAgo,
		"faviconURL": faviconURL,
	}

	// Parse all templates once at startup
	tmpl := template.New("").Funcs(funcMap)

	// Parse base template
	tmpl, err := tmpl.ParseGlob("templates/*.gohtml")
	if err != nil {
		return nil, fmt.Errorf("failed to parse base templates: %w", err)
	}

	// Parse page templates
	tmpl, err = tmpl.ParseGlob("templates/pages/*.gohtml")
	if err != nil {
		return nil, fmt.Errorf("failed to parse page templates: %w", err)
	}

	// Parse partial templates
	tmpl, err = tmpl.ParseGlob("templates/partials/*.gohtml")
	if err != nil {
		return nil, fmt.Errorf("failed to parse partial templates: %w", err)
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
