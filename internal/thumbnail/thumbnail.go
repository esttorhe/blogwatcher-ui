// ABOUTME: Provides thumbnail URL extraction from RSS items and Open Graph meta tags.
// ABOUTME: Used during sync to populate article thumbnails with fallback chain.
package thumbnail

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/otiai10/opengraph/v2"
)

// ExtractFromRSS attempts to extract thumbnail URL from gofeed.Item.
// Checks Item.Image first, then Enclosures for image MIME types.
// Returns empty string if no thumbnail found.
func ExtractFromRSS(item *gofeed.Item) string {
	if item == nil {
		return ""
	}

	// Try Item.Image first (channel-level image reference)
	if item.Image != nil && item.Image.URL != "" {
		return item.Image.URL
	}

	// Try Enclosures (common in RSS 2.0)
	for _, enc := range item.Enclosures {
		if isImageMIMEType(enc.Type) && enc.URL != "" {
			return enc.URL
		}
	}

	return ""
}

// ExtractFromOpenGraph fetches og:image from article page.
// Uses context with timeout to prevent hanging on slow sites.
// Returns empty string on any error (thumbnail is optional).
func ExtractFromOpenGraph(articleURL string, timeout time.Duration) string {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	intent := opengraph.Intent{
		Context:    ctx,
		Strict:     true, // Only parse <meta> tags
		HTTPClient: &http.Client{Timeout: timeout},
	}

	ogp, err := opengraph.Fetch(articleURL, intent)
	if err != nil {
		return "" // Fail silently - thumbnail is optional
	}

	// Convert relative URLs to absolute
	if err := ogp.ToAbs(); err != nil {
		return ""
	}

	// Return first image if available
	if len(ogp.Image) > 0 && ogp.Image[0].URL != "" {
		return ogp.Image[0].URL
	}

	return ""
}

func isImageMIMEType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}
