// ABOUTME: Version package for storing the application version.
// ABOUTME: Version is set at build time via ldflags for release builds.
package version

// Version is set via ldflags during build. Defaults to "dev" for development.
var Version = "dev"
