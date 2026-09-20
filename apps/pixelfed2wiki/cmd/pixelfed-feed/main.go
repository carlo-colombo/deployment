package main

import (
	"bufio"
	"context"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/carlo-colombo/pixelfed2wiki/pixelfed"
)

type application struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Xmlns   string      `xml:"xmlns,attr"`
	ID      string      `xml:"id"`
	Title   string      `xml:"title"`
	Updated string      `xml:"updated"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	ID        string      `xml:"id"`
	Title     string      `xml:"title"`
	Updated   string      `xml:"updated"`
	Published string      `xml:"published"`
	Link      atomLink    `xml:"link"`
	Content   atomContent `xml:"content"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
}

type atomContent struct {
	Type string `xml:"type,attr"`
	HTML string `xml:",innerxml"`
}

func main() {
	instance := flag.String("instance", "https://pixelfed.social", "Pixelfed instance URL")
	username := flag.String("username", "", "Pixelfed username")
	output := flag.String("output", "pixelfed.feed", "output Atom feed path")
	flag.Parse()
	if *username == "" {
		fatal("-username is required")
	}
	base := strings.TrimRight(*instance, "/")
	ctx := context.Background()
	client := &http.Client{Timeout: 30 * time.Second}
	accessToken, err := authenticate(ctx, client, base)
	if err != nil {
		fatal(err.Error())
	}
	posts, err := pixelfed.FetchAPI(ctx, base, *username, accessToken, client)
	if err != nil {
		fatal(err.Error())
	}
	if err := writeFeed(*output, base, *username, posts); err != nil {
		fatal(err.Error())
	}
	fmt.Printf("wrote %d posts to %s\n", len(posts), *output)
}

func authenticate(ctx context.Context, client *http.Client, base string) (string, error) {
	values := url.Values{"client_name": {"pixelfed2wiki feed export"}, "redirect_uris": {"urn:ietf:wg:oauth:2.0:oob"}, "scopes": {"read"}}
	var app application
	if err := requestJSON(ctx, client, http.MethodPost, base+"/api/v1/apps", strings.NewReader(values.Encode()), "application/x-www-form-urlencoded", &app); err != nil {
		return "", fmt.Errorf("register OAuth application: %w", err)
	}
	if app.ClientID == "" || app.ClientSecret == "" {
		return "", fmt.Errorf("Pixelfed returned incomplete OAuth application credentials")
	}
	authURL := base + "/oauth/authorize?" + url.Values{"client_id": {app.ClientID}, "response_type": {"code"}, "redirect_uri": {"urn:ietf:wg:oauth:2.0:oob"}, "scope": {"read"}}.Encode()
	fmt.Printf("Open this URL in a browser:\n%s\n\nPaste the authorization code: ", authURL)
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	values = url.Values{"grant_type": {"authorization_code"}, "client_id": {app.ClientID}, "client_secret": {app.ClientSecret}, "redirect_uri": {"urn:ietf:wg:oauth:2.0:oob"}, "code": {strings.TrimSpace(code)}}
	var token tokenResponse
	if err := requestJSON(ctx, client, http.MethodPost, base+"/oauth/token", strings.NewReader(values.Encode()), "application/x-www-form-urlencoded", &token); err != nil {
		return "", fmt.Errorf("exchange authorization code: %w", err)
	}
	if token.AccessToken == "" {
		return "", fmt.Errorf("Pixelfed returned an empty access token")
	}
	return token.AccessToken, nil
}

func requestJSON(ctx context.Context, client *http.Client, method, endpoint string, body io.Reader, contentType string, target any) error {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(string(data)))
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func writeFeed(path, base, username string, posts []pixelfed.Post) error {
	feed := atomFeed{Xmlns: "http://www.w3.org/2005/Atom", ID: base + "/users/" + username + ".atom", Title: username + " on Pixelfed", Updated: time.Now().UTC().Format(time.RFC3339)}
	for _, post := range posts {
		content := strings.Builder{}
		content.WriteString("<p>")
		content.WriteString(xmlEscape(post.Caption))
		content.WriteString("</p>")
		for _, image := range post.Images {
			content.WriteString(`<img src="`)
			content.WriteString(xmlEscape(image))
			content.WriteString(`"/>`)
		}
		when := post.Published.UTC().Format(time.RFC3339)
		feed.Entries = append(feed.Entries, atomEntry{ID: post.ID, Title: post.Caption, Updated: when, Published: when, Link: atomLink{Href: post.URL}, Content: atomContent{Type: "html", HTML: content.String()}})
	}
	data, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return err
	}
	data = append([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"), data...)
	return os.WriteFile(path, data, 0600)
}

func xmlEscape(value string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(value))
	return b.String()
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
