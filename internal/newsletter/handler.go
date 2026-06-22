// ABOUTME: Parses inbound RFC 822 emails and ingests them as newsletter articles.
// ABOUTME: Uses stdlib net/mail and mime/multipart — no external dependencies.
package newsletter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
)

// Handler ingests raw RFC 822 emails as newsletter articles.
type Handler struct {
	db *storage.Database
}

// NewHandler returns a Handler backed by the given database.
func NewHandler(db *storage.Database) *Handler {
	return &Handler{db: db}
}

// HandleInbound parses raw RFC 822 bytes, creates or reuses the sender's blog,
// and inserts the email as an Article. Returns the stored Article.
// Calling it twice with the same raw email is idempotent (same Message-ID → same row).
func (h *Handler) HandleInbound(ctx context.Context, raw []byte) (model.Article, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return model.Article{}, fmt.Errorf("parse email: %w", err)
	}

	// Extract sender address and display name.
	fromHeader := msg.Header.Get("From")
	senderName, senderEmail, err := parseFrom(fromHeader)
	if err != nil {
		return model.Article{}, fmt.Errorf("parse From header: %w", err)
	}

	subject := decodeHeader(msg.Header.Get("Subject"))
	messageID := strings.TrimSpace(msg.Header.Get("Message-ID"))

	// URL is the Message-ID encoded as a stable URI so we can de-duplicate.
	articleURL := "message:" + messageID

	// Get or create the newsletter blog for this sender.
	blog, err := h.db.GetOrCreateNewsletterBlog(senderName, senderEmail)
	if err != nil {
		return model.Article{}, fmt.Errorf("get/create newsletter blog: %w", err)
	}

	// De-duplicate: if we already have an article with this URL, return it.
	existing, err := h.db.GetArticleByURL(articleURL)
	if err != nil {
		return model.Article{}, fmt.Errorf("check existing article: %w", err)
	}
	if existing != nil {
		return *existing, nil
	}

	htmlBody, err := extractHTMLBody(msg)
	if err != nil {
		return model.Article{}, fmt.Errorf("extract body: %w", err)
	}

	article := model.Article{
		BlogID:  blog.ID,
		Title:   subject,
		URL:     articleURL,
		Content: htmlBody,
	}

	inserted, err := h.db.AddArticlesBulk([]model.Article{article})
	if err != nil {
		return model.Article{}, fmt.Errorf("store article: %w", err)
	}
	if inserted == 0 {
		return model.Article{}, fmt.Errorf("article not stored")
	}

	// Fetch back to get the assigned ID.
	stored, err := h.db.GetArticleByURL(articleURL)
	if err != nil || stored == nil {
		return model.Article{}, fmt.Errorf("fetch stored article: %w", err)
	}
	return *stored, nil
}

// parseFrom extracts the display name and email address from a From header value.
// When the address has no display name, the local+domain part is used as the name.
func parseFrom(from string) (name, email string, err error) {
	addr, err := mail.ParseAddress(from)
	if err != nil {
		return "", "", err
	}
	email = addr.Address
	name = addr.Name
	if name == "" {
		name = email
	}
	return name, email, nil
}

// decodeHeader decodes an RFC 2047 encoded header value (e.g. Subject).
func decodeHeader(h string) string {
	dec := new(mime.WordDecoder)
	decoded, err := dec.DecodeHeader(h)
	if err != nil {
		return h
	}
	return decoded
}

// extractHTMLBody returns the HTML part of the email body.
// For multipart/alternative emails it prefers the text/html part.
// For single-part HTML emails it returns the body as-is.
func extractHTMLBody(msg *mail.Message) (string, error) {
	contentType := msg.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		// No valid Content-Type — read body verbatim.
		body, err := io.ReadAll(msg.Body)
		if err != nil {
			return "", err
		}
		return string(body), nil
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		return extractHTMLFromMultipart(msg.Body, params["boundary"])
	}

	body, err := io.ReadAll(msg.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// extractHTMLFromMultipart walks multipart parts and returns the first text/html part.
func extractHTMLFromMultipart(r io.Reader, boundary string) (string, error) {
	mr := multipart.NewReader(r, boundary)
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		ct := part.Header.Get("Content-Type")
		partMedia, _, _ := mime.ParseMediaType(ct)
		if partMedia == "text/html" {
			body, err := io.ReadAll(part)
			if err != nil {
				return "", err
			}
			return string(body), nil
		}
	}
	return "", nil
}
