// ABOUTME: Version package for storing the application version.
// ABOUTME: Uses runtime/debug to read module version from go install builds.
package version

import "runtime/debug"

// Version is the application version. It reads from Go module build info
// (populated by go install), falls back to ldflags, or defaults to "dev".
var Version = func() string {
	// First check if set via ldflags (for manual builds)
	if version != "" && version != "dev" {
		return version
	}

	// Read from Go module build info (works with go install)
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	return "dev"
}()

// version can be set via ldflags: -X github.com/esttorhe/blogwatcher-ui/internal/version.version=v1.0.0
var version = "dev"
