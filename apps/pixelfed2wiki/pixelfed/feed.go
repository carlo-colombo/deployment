package pixelfed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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

type apiStatus struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	Content   string    `json:"content"`
	Account   struct {
		Username string `json:"username"`
	} `json:"account"`
	Media []struct {
		URL     string `json:"url"`
		Preview string `json:"preview_url"`
		Type    string `json:"type"`
	} `json:"media_attachments"`
}

type apiAccount struct {
	ID string `json:"id"`
}

func APIConfigFromFeedURL(feedURL string) (string, string, error) {
	u, err := url.Parse(strings.TrimSpace(feedURL))
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return "", "", fmt.Errorf("invalid Pixelfed feed URL %q", feedURL)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 2 || parts[0] != "users" || !strings.HasSuffix(parts[1], ".atom") {
		return "", "", fmt.Errorf("Pixelfed feed URL must match /users/{username}.atom: %q", feedURL)
	}
	username := strings.TrimSuffix(parts[1], ".atom")
	if username == "" {
		return "", "", fmt.Errorf("Pixelfed feed URL has no username: %q", feedURL)
	}
	return u.Scheme + "://" + u.Host, username, nil
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

func FetchAPI(ctx context.Context, apiURL, username, accessToken string, client *http.Client) ([]Post, error) {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	base := strings.TrimRight(apiURL, "/")
	accountURL := base + "/api/v1/accounts/lookup?acct=" + url.QueryEscape(username)
	var account apiAccount
	if err := getJSON(ctx, client, accountURL, accessToken, &account); err != nil {
		return nil, fmt.Errorf("look up Pixelfed account: %w", err)
	}
	posts := make([]Post, 0)
	endpoint := base + "/api/v1/accounts/" + url.PathEscape(account.ID) + "/statuses?exclude_reblogs=true&exclude_replies=true&limit=40"
	for endpoint != "" {
		var statuses []apiStatus
		next, err := getJSONWithNext(ctx, client, endpoint, accessToken, &statuses)
		if err != nil {
			return nil, fmt.Errorf("fetch Pixelfed API: %w", err)
		}
		for _, status := range statuses {
			images := make([]string, 0, len(status.Media))
			for _, media := range status.Media {
				if media.Type == "image" && media.URL != "" {
					images = append(images, media.URL)
				}
			}
			if status.URL == "" || status.CreatedAt.IsZero() {
				continue
			}
			posts = append(posts, Post{ID: status.ID, URL: status.URL, Caption: textCaption(status.Content), Published: status.CreatedAt.UTC(), Images: images})
		}
		endpoint = next
	}
	return posts, nil
}

func getJSON(ctx context.Context, client *http.Client, endpoint, accessToken string, target any) error {
	_, err := getJSONWithNext(ctx, client, endpoint, accessToken, target)
	return err
}

func getJSONWithNext(ctx context.Context, client *http.Client, endpoint, accessToken string, target any) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("%s", resp.Status)
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 16<<20)).Decode(target); err != nil {
		return "", err
	}
	return nextLink(resp.Header.Get("Link")), nil
}

func nextLink(header string) string {
	for _, part := range strings.Split(header, ",") {
		fields := strings.Split(strings.TrimSpace(part), ";")
		if len(fields) < 2 || !strings.Contains(fields[1], `rel="next"`) {
			continue
		}
		link := strings.TrimSpace(fields[0])
		return strings.Trim(link, "<>")
	}
	return ""
}

func MergePosts(feeds ...[]Post) []Post {
	seen := make(map[string]bool)
	merged := make([]Post, 0)
	for _, posts := range feeds {
		for _, post := range posts {
			key := post.ID
			if key == "" {
				key = post.URL
			}
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			merged = append(merged, post)
		}
	}
	return merged
}

func ParseFile(path string) ([]Post, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open feed: %w", err)
	}
	defer file.Close()
	return Parse(io.LimitReader(file, 64<<20))
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
