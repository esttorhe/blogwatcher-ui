// ABOUTME: HTTP server package with dependency injection pattern
// ABOUTME: Manages server lifecycle, template parsing, and request routing
package server

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/esttorhe/blogwatcher-ui/internal/storage"
)

//go:embed templates/*.gohtml templates/pages/*.gohtml templates/partials/*.gohtml
var templateFS embed.FS

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

	// Parse all templates once at startup from embedded filesystem
	tmpl := template.New("").Funcs(funcMap)

	// Walk the embedded template filesystem and parse all templates
	err := fs.WalkDir(templateFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := templateFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", path, err)
		}
		_, err = tmpl.Parse(string(data))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", path, err)
		}
		return nil
	})
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
