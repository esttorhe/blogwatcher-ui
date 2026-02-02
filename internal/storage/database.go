// ABOUTME: Provides database connection and query methods for reading blog and article data.
// ABOUTME: Read-only access to existing blogwatcher database - blog management happens via CLI.
package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/esttorhe/blogwatcher-ui/internal/model"
)

const sqliteTimeLayout = time.RFC3339Nano

func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".blogwatcher", "blogwatcher.db"), nil
}

type Database struct {
	path string
	conn *sql.DB
}

func OpenDatabase(path string) (*Database, error) {
	if path == "" {
		var err error
		path, err = DefaultDBPath()
		if err != nil {
			return nil, err
		}
	}

	// Check if database file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("database not found at %s - run blogwatcher CLI to initialize", path)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", path)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// Set SQLite single-writer constraint
	conn.SetMaxOpenConns(1)

	db := &Database{path: path, conn: conn}

	// Verify connection works (don't create schema)
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return db, nil
}

func (db *Database) Path() string {
	return db.path
}

func (db *Database) Close() error {
	if db.conn == nil {
		return nil
	}
	return db.conn.Close()
}

func (db *Database) ListBlogs() ([]model.Blog, error) {
	rows, err := db.conn.Query(`SELECT id, name, url, feed_url, scrape_selector, last_scanned FROM blogs ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []model.Blog
	for rows.Next() {
		blog, err := scanBlog(rows)
		if err != nil {
			return nil, err
		}
		if blog != nil {
			blogs = append(blogs, *blog)
		}
	}
	return blogs, rows.Err()
}

func (db *Database) ListArticles(unreadOnly bool, blogID *int64) ([]model.Article, error) {
	query := `SELECT id, blog_id, title, url, published_date, discovered_date, is_read FROM articles WHERE 1=1`
	var args []interface{}
	if unreadOnly {
		query += " AND is_read = 0"
	}
	if blogID != nil {
		query += " AND blog_id = ?"
		args = append(args, *blogID)
	}
	query += " ORDER BY discovered_date DESC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []model.Article
	for rows.Next() {
		article, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		if article != nil {
			articles = append(articles, *article)
		}
	}
	return articles, rows.Err()
}

func (db *Database) MarkArticleRead(id int64) (bool, error) {
	result, err := db.conn.Exec(`UPDATE articles SET is_read = 1 WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (db *Database) MarkArticleUnread(id int64) (bool, error) {
	result, err := db.conn.Exec(`UPDATE articles SET is_read = 0 WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func scanBlog(scanner interface{ Scan(dest ...any) error }) (*model.Blog, error) {
	var (
		id             int64
		name           string
		url            string
		feedURL        sql.NullString
		scrapeSelector sql.NullString
		lastScanned    sql.NullString
	)
	if err := scanner.Scan(&id, &name, &url, &feedURL, &scrapeSelector, &lastScanned); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	blog := &model.Blog{
		ID:             id,
		Name:           name,
		URL:            url,
		FeedURL:        feedURL.String,
		ScrapeSelector: scrapeSelector.String,
	}
	if lastScanned.Valid {
		if parsed, err := parseTime(lastScanned.String); err == nil {
			blog.LastScanned = &parsed
		}
	}
	return blog, nil
}

func scanArticle(scanner interface{ Scan(dest ...any) error }) (*model.Article, error) {
	var (
		id            int64
		blogID        int64
		title         string
		url           string
		publishedDate sql.NullString
		discovered    sql.NullString
		isRead        bool
	)
	if err := scanner.Scan(&id, &blogID, &title, &url, &publishedDate, &discovered, &isRead); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	article := &model.Article{
		ID:     id,
		BlogID: blogID,
		Title:  title,
		URL:    url,
		IsRead: isRead,
	}
	if publishedDate.Valid {
		if parsed, err := parseTime(publishedDate.String); err == nil {
			article.PublishedDate = &parsed
		}
	}
	if discovered.Valid {
		if parsed, err := parseTime(discovered.String); err == nil {
			article.DiscoveredDate = &parsed
		}
	}

	return article, nil
}

func parseTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("empty time")
	}
	parsed, err := time.Parse(sqliteTimeLayout, value)
	if err == nil {
		return parsed, nil
	}
	return time.Parse("2006-01-02 15:04:05", value)
}
