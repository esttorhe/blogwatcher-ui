// ABOUTME: Orchestrates blog scanning using RSS feeds or HTML scraping as fallback.
// ABOUTME: Used by the web UI sync feature to discover new articles from tracked blogs.
package scanner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/rss"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/scraper"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/thumbnail"
)

type ScanResult struct {
	BlogName    string
	NewArticles int
	TotalFound  int
	Source      string
	Error       string
}

func ScanBlog(ctx context.Context, db *storage.Database, blog model.Blog) ScanResult {
	var (
		articles []model.Article
		source   = "none"
		errText  string
	)

	feedURL := blog.FeedURL
	if feedURL == "" {
		if discovered, err := rss.DiscoverFeedURL(ctx, blog.URL); err == nil && discovered != "" {
			feedURL = discovered
			blog.FeedURL = discovered
			_ = db.UpdateBlog(blog)
		}
	}

	if feedURL != "" {
		feedArticles, err := rss.ParseFeed(ctx, feedURL)
		if err != nil {
			errText = err.Error()
		} else {
			articles = convertFeedArticles(ctx, blog.ID, feedArticles)
			source = "rss"
		}
	}

	if len(articles) == 0 && blog.ScrapeSelector != "" {
		scrapedArticles, err := scraper.ScrapeBlog(ctx, blog.URL, blog.ScrapeSelector)
		if err != nil {
			if errText != "" {
				errText = fmt.Sprintf("RSS: %s; Scraper: %s", errText, err.Error())
			} else {
				errText = err.Error()
			}
		} else {
			articles = convertScrapedArticles(ctx, blog.ID, scrapedArticles)
			source = "scraper"
			errText = ""
		}
	}

	seenURLs := make(map[string]struct{})
	uniqueArticles := make([]model.Article, 0, len(articles))
	for _, article := range articles {
		if _, exists := seenURLs[article.URL]; exists {
			continue
		}
		seenURLs[article.URL] = struct{}{}
		uniqueArticles = append(uniqueArticles, article)
	}

	urlList := make([]string, 0, len(seenURLs))
	for url := range seenURLs {
		urlList = append(urlList, url)
	}

	existing, err := db.GetExistingArticleURLs(urlList)
	if err != nil {
		errText = err.Error()
	}

	discoveredAt := time.Now()
	newArticles := make([]model.Article, 0, len(uniqueArticles))
	for _, article := range uniqueArticles {
		if _, exists := existing[article.URL]; exists {
			continue
		}
		article.DiscoveredDate = &discoveredAt
		newArticles = append(newArticles, article)
	}

	newCount := 0
	if len(newArticles) > 0 {
		count, err := db.AddArticlesBulk(newArticles)
		if err != nil {
			errText = err.Error()
		} else {
			newCount = count
		}
	}

	_ = db.UpdateBlogLastScanned(blog.ID, time.Now())

	return ScanResult{
		BlogName:    blog.Name,
		NewArticles: newCount,
		TotalFound:  len(seenURLs),
		Source:      source,
		Error:       errText,
	}
}

// ScanAllBlogs scans all blogs concurrently using goroutines and channels.
// Each blog gets its own goroutine for network I/O, but database writes are
// serialized through the single db connection to avoid SQLite write conflicts.
func ScanAllBlogs(ctx context.Context, db *storage.Database) ([]ScanResult, error) {
	blogs, err := db.ListBlogs()
	if err != nil {
		return nil, err
	}

	if len(blogs) == 0 {
		return nil, nil
	}

	results := make([]ScanResult, len(blogs))
	resultCh := make(chan struct {
		Index  int
		Result ScanResult
	}, len(blogs))

	var wg sync.WaitGroup
	for i, blog := range blogs {
		wg.Add(1)
		go func(index int, b model.Blog) {
			defer wg.Done()
			result := ScanBlog(ctx, db, b)
			resultCh <- struct {
				Index  int
				Result ScanResult
			}{Index: index, Result: result}
		}(i, blog)
	}

	// Close the channel once all goroutines complete
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for item := range resultCh {
		results[item.Index] = item.Result
	}

	return results, nil
}

func ScanBlogByName(ctx context.Context, db *storage.Database, name string) (*ScanResult, error) {
	blog, err := db.GetBlogByName(name)
	if err != nil {
		return nil, err
	}
	if blog == nil {
		return nil, nil
	}
	result := ScanBlog(ctx, db, *blog)
	return &result, nil
}

func convertFeedArticles(ctx context.Context, blogID int64, articles []rss.FeedArticle) []model.Article {
	result := make([]model.Article, 0, len(articles))
	for _, article := range articles {
		thumbnailURL := article.ThumbnailURL
		// Open Graph fallback if RSS didn't provide thumbnail
		if thumbnailURL == "" {
			thumbnailURL = thumbnail.ExtractFromOpenGraph(ctx, article.URL)
		}
		result = append(result, model.Article{
			BlogID:        blogID,
			Title:         article.Title,
			URL:           article.URL,
			ThumbnailURL:  thumbnailURL,
			PublishedDate: article.PublishedDate,
			IsRead:        false,
		})
	}
	return result
}

func convertScrapedArticles(ctx context.Context, blogID int64, articles []scraper.ScrapedArticle) []model.Article {
	result := make([]model.Article, 0, len(articles))
	for _, article := range articles {
		// Scraped articles don't have thumbnails, try Open Graph
		thumbnailURL := thumbnail.ExtractFromOpenGraph(ctx, article.URL)
		result = append(result, model.Article{
			BlogID:        blogID,
			Title:         article.Title,
			URL:           article.URL,
			ThumbnailURL:  thumbnailURL,
			PublishedDate: article.PublishedDate,
			IsRead:        false,
		})
	}
	return result
}

// ThumbnailSyncResult holds the result of a thumbnail sync operation.
type ThumbnailSyncResult struct {
	Total   int
	Updated int
	Errors  int
}

// SyncThumbnails re-fetches thumbnails for articles that have empty thumbnail_url.
// For each article, it re-parses the RSS feed to find the matching item and extract thumbnail.
// Falls back to Open Graph if RSS doesn't provide a thumbnail.
func SyncThumbnails(ctx context.Context, db *storage.Database) (ThumbnailSyncResult, error) {
	articles, err := db.GetArticlesMissingThumbnails()
	if err != nil {
		return ThumbnailSyncResult{}, err
	}

	result := ThumbnailSyncResult{Total: len(articles)}

	// Group articles by feed URL to avoid re-fetching the same feed multiple times
	feedArticles := make(map[string][]storage.ArticleForThumbnailSync)
	for _, a := range articles {
		feedArticles[a.FeedURL] = append(feedArticles[a.FeedURL], a)
	}

	// Process each feed
	for feedURL, articleList := range feedArticles {
		feedItems, err := rss.ParseFeed(ctx, feedURL)
		if err != nil {
			result.Errors += len(articleList)
			continue
		}

		// Build URL to thumbnail map from feed
		feedThumbnails := make(map[string]string)
		for _, item := range feedItems {
			if item.ThumbnailURL != "" {
				feedThumbnails[item.URL] = item.ThumbnailURL
			}
		}

		// Update each article
		for _, article := range articleList {
			thumbnailURL := feedThumbnails[article.URL]

			// Fallback to Open Graph if RSS didn't provide thumbnail
			if thumbnailURL == "" {
				thumbnailURL = thumbnail.ExtractFromOpenGraph(ctx, article.URL)
			}

			if thumbnailURL != "" {
				if err := db.UpdateArticleThumbnail(article.ID, thumbnailURL); err != nil {
					result.Errors++
				} else {
					result.Updated++
				}
			}
		}
	}

	return result, nil
}
