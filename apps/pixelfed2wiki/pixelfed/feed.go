package pixelfed

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

type Post struct {
	ID        string
	URL       string
	Title     string
	Caption   string
	Published time.Time
	Images    []string
}

type Client struct {
	HTTP          *http.Client
	MaxMediaBytes int64
}

func Parse(r io.Reader) ([]Post, error) {
	feed, err := gofeed.NewParser().Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse Atom feed: %w", err)
	}
	posts := make([]Post, 0, len(feed.Items))
	for _, item := range feed.Items {
		when := item.PublishedParsed
		if when == nil {
			when = item.UpdatedParsed
		}
		if when == nil {
			return nil, fmt.Errorf("entry %q has no RFC3339 timestamp", item.GUID)
		}
		link := item.Link
		if link == "" {
			link = item.GUID
		}
		caption := item.Description
		if caption == "" {
			caption = item.Content
		}
		caption = textCaption(caption)
		images := imageURLs(item)
		posts = append(posts, Post{ID: item.GUID, URL: link, Title: item.Title, Caption: caption, Published: when.UTC(), Images: images})
	}
	return posts, nil
}

func ParseURL(ctx context.Context, feedURL string, client *http.Client) ([]Post, error) {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch feed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch feed: %s", resp.Status)
	}
	return Parse(io.LimitReader(resp.Body, 8<<20))
}

func (c Client) Download(ctx context.Context, imageURL string) ([]byte, string, error) {
	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 45 * time.Second}
	}
	limit := c.MaxMediaBytes
	if limit <= 0 {
		limit = 16 << 20
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("download %s: %w", imageURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("download %s: %s", imageURL, resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, "", err
	}
	if int64(len(data)) > limit {
		return nil, "", fmt.Errorf("download %s exceeds %d bytes", imageURL, limit)
	}
	return data, resp.Header.Get("Content-Type"), nil
}

func CanonicalURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return "", fmt.Errorf("invalid Pixelfed URL %q", raw)
	}
	u.Fragment = ""
	u.RawQuery = ""
	u.Host = strings.ToLower(u.Host)
	return strings.TrimRight(u.String(), "/"), nil
}

func textCaption(raw string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(raw))
	if err != nil {
		return strings.TrimSpace(raw)
	}
	return strings.TrimSpace(doc.Text())
}

func imageURLs(item *gofeed.Item) []string {
	seen := map[string]bool{}
	out := []string{}
	add := func(s string) {
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	if item.Content != "" {
		if doc, err := goquery.NewDocumentFromReader(strings.NewReader(item.Content)); err == nil {
			doc.Find("img").Each(func(_ int, s *goquery.Selection) { v, _ := s.Attr("src"); add(v) })
		}
	}
	for _, media := range item.Extensions["media"]["content"] {
		add(media.Attrs["url"])
	}
	return out
}
