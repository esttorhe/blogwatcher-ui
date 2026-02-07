// ABOUTME: Tests for thumbnail extraction from RSS items.
// ABOUTME: Covers media:content, Item.Image, and Enclosure extraction methods.
package thumbnail

import (
	"testing"

	ext "github.com/mmcdole/gofeed/extensions"

	"github.com/mmcdole/gofeed"
)

func TestExtractFromRSS_MediaContent(t *testing.T) {
	// Test that media:content is checked first
	item := &gofeed.Item{
		Extensions: ext.Extensions{
			"media": {
				"content": []ext.Extension{
					{
						Name: "content",
						Attrs: map[string]string{
							"url":    "https://example.com/media-content.jpg",
							"medium": "image",
						},
					},
				},
			},
		},
	}

	result := ExtractFromRSS(item)
	expected := "https://example.com/media-content.jpg"
	if result != expected {
		t.Errorf("ExtractFromRSS() = %q, want %q", result, expected)
	}
}

func TestExtractFromRSS_MediaContentFirst(t *testing.T) {
	// Test that media:content takes precedence over Item.Image
	item := &gofeed.Item{
		Image: &gofeed.Image{
			URL: "https://example.com/item-image.jpg",
		},
		Extensions: ext.Extensions{
			"media": {
				"content": []ext.Extension{
					{
						Name: "content",
						Attrs: map[string]string{
							"url": "https://example.com/media-content.jpg",
						},
					},
				},
			},
		},
	}

	result := ExtractFromRSS(item)
	expected := "https://example.com/media-content.jpg"
	if result != expected {
		t.Errorf("ExtractFromRSS() = %q, want %q (media:content should take precedence)", result, expected)
	}
}

func TestExtractFromRSS_FallbackToItemImage(t *testing.T) {
	// Test fallback to Item.Image when no media:content
	item := &gofeed.Item{
		Image: &gofeed.Image{
			URL: "https://example.com/item-image.jpg",
		},
	}

	result := ExtractFromRSS(item)
	expected := "https://example.com/item-image.jpg"
	if result != expected {
		t.Errorf("ExtractFromRSS() = %q, want %q", result, expected)
	}
}

func TestExtractFromRSS_FallbackToEnclosure(t *testing.T) {
	// Test fallback to Enclosure when no media:content or Item.Image
	item := &gofeed.Item{
		Enclosures: []*gofeed.Enclosure{
			{
				URL:  "https://example.com/enclosure.jpg",
				Type: "image/jpeg",
			},
		},
	}

	result := ExtractFromRSS(item)
	expected := "https://example.com/enclosure.jpg"
	if result != expected {
		t.Errorf("ExtractFromRSS() = %q, want %q", result, expected)
	}
}

func TestExtractFromRSS_SkipsNonImageEnclosure(t *testing.T) {
	// Test that non-image enclosures are skipped
	item := &gofeed.Item{
		Enclosures: []*gofeed.Enclosure{
			{
				URL:  "https://example.com/audio.mp3",
				Type: "audio/mpeg",
			},
			{
				URL:  "https://example.com/image.png",
				Type: "image/png",
			},
		},
	}

	result := ExtractFromRSS(item)
	expected := "https://example.com/image.png"
	if result != expected {
		t.Errorf("ExtractFromRSS() = %q, want %q", result, expected)
	}
}

func TestExtractFromRSS_NilItem(t *testing.T) {
	result := ExtractFromRSS(nil)
	if result != "" {
		t.Errorf("ExtractFromRSS(nil) = %q, want empty string", result)
	}
}

func TestExtractFromRSS_EmptyItem(t *testing.T) {
	item := &gofeed.Item{}
	result := ExtractFromRSS(item)
	if result != "" {
		t.Errorf("ExtractFromRSS(empty item) = %q, want empty string", result)
	}
}

func TestExtractFromRSS_MediaContentWithoutURL(t *testing.T) {
	// Test that media:content without URL attribute falls back
	item := &gofeed.Item{
		Image: &gofeed.Image{
			URL: "https://example.com/fallback.jpg",
		},
		Extensions: ext.Extensions{
			"media": {
				"content": []ext.Extension{
					{
						Name:  "content",
						Attrs: map[string]string{},
					},
				},
			},
		},
	}

	result := ExtractFromRSS(item)
	expected := "https://example.com/fallback.jpg"
	if result != expected {
		t.Errorf("ExtractFromRSS() = %q, want %q (should fallback when media:content has no URL)", result, expected)
	}
}

func TestExtractFromRSS_MediaThumbnail(t *testing.T) {
	// Test that media:thumbnail is also checked
	item := &gofeed.Item{
		Extensions: ext.Extensions{
			"media": {
				"thumbnail": []ext.Extension{
					{
						Name: "thumbnail",
						Attrs: map[string]string{
							"url": "https://example.com/media-thumbnail.jpg",
						},
					},
				},
			},
		},
	}

	result := ExtractFromRSS(item)
	expected := "https://example.com/media-thumbnail.jpg"
	if result != expected {
		t.Errorf("ExtractFromRSS() = %q, want %q", result, expected)
	}
}

func TestExtractFromRSS_MediaContentPrecedenceOverThumbnail(t *testing.T) {
	// Test that media:content takes precedence over media:thumbnail
	item := &gofeed.Item{
		Extensions: ext.Extensions{
			"media": {
				"content": []ext.Extension{
					{
						Name: "content",
						Attrs: map[string]string{
							"url": "https://example.com/media-content.jpg",
						},
					},
				},
				"thumbnail": []ext.Extension{
					{
						Name: "thumbnail",
						Attrs: map[string]string{
							"url": "https://example.com/media-thumbnail.jpg",
						},
					},
				},
			},
		},
	}

	result := ExtractFromRSS(item)
	expected := "https://example.com/media-content.jpg"
	if result != expected {
		t.Errorf("ExtractFromRSS() = %q, want %q (media:content should take precedence over media:thumbnail)", result, expected)
	}
}
