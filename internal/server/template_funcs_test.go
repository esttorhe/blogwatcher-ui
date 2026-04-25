// ABOUTME: Tests for custom template functions used in HTML rendering.
// ABOUTME: Covers smryURL for generating smry.ai links from article URLs.
package server

import "testing"

func TestSmryURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "strips https protocol",
			input:    "https://example.com/blog/post",
			expected: "https://smry.ai/example.com/blog/post",
		},
		{
			name:     "strips http protocol",
			input:    "http://example.com/blog/post",
			expected: "https://smry.ai/example.com/blog/post",
		},
		{
			name:     "handles URL without protocol",
			input:    "example.com/blog/post",
			expected: "https://smry.ai/example.com/blog/post",
		},
		{
			name:     "handles URL with query parameters",
			input:    "https://example.com/post?id=123&ref=rss",
			expected: "https://smry.ai/example.com/post?id=123&ref=rss",
		},
		{
			name:     "handles URL with fragment",
			input:    "https://example.com/post#section",
			expected: "https://smry.ai/example.com/post#section",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "https://smry.ai/",
		},
		{
			name:     "handles domain only",
			input:    "https://example.com",
			expected: "https://smry.ai/example.com",
		},
		{
			name:     "handles trailing slash",
			input:    "https://example.com/",
			expected: "https://smry.ai/example.com/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := smryURL(tt.input)
			if result != tt.expected {
				t.Errorf("smryURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
