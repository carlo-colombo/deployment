package pixelfed

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestParsePublishedAndImages(t *testing.T) {
	feed := `<feed xmlns="http://www.w3.org/2005/Atom"><entry><id>id</id><title>caption #tag</title><published>2026-01-02T03:04:05Z</published><link rel="alternate" href="https://pixelfed.social/p/u/1"/><content type="html"><p>Hello</p><img src="https://example.test/1.jpg"/><img src="https://example.test/2.jpg"/></content></entry></feed>`
	posts, err := Parse(strings.NewReader(feed))
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 || posts[0].Published != time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) {
		t.Fatalf("unexpected post: %#v", posts)
	}
	if len(posts[0].Images) != 2 || posts[0].Caption != "Hello" {
		t.Fatalf("unexpected content: %#v", posts[0])
	}
}

func TestFetchAPIPaginates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/accounts/lookup":
			_, _ = io.WriteString(w, `{"id":"42"}`)
		case "/api/v1/accounts/42/statuses":
			if r.URL.Query().Get("max_id") == "10" {
				_, _ = io.WriteString(w, `[{"id":"9","url":"https://pixelfed.social/p/u/9","created_at":"2026-01-01T00:00:00Z","content":"older","media_attachments":[]}]`)
				return
			}
			w.Header().Set("Link", "<"+"http://"+r.Host+"/api/v1/accounts/42/statuses?max_id=10>; rel=\"next\"")
			_, _ = io.WriteString(w, `[{"id":"11","url":"https://pixelfed.social/p/u/11","created_at":"2026-01-02T00:00:00Z","content":"new","media_attachments":[]}]`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	posts, err := FetchAPI(t.Context(), server.URL, "u", "test-token", server.Client())
	if err != nil || len(posts) != 2 {
		t.Fatalf("posts=%#v err=%v", posts, err)
	}
}

func TestMergePosts(t *testing.T) {
	when := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	posts := MergePosts(
		[]Post{{ID: "one", URL: "https://example.test/one", Published: when}},
		[]Post{{ID: "one", URL: "https://example.test/one", Published: when}, {ID: "two", URL: "https://example.test/two", Published: when}},
	)
	if len(posts) != 2 || posts[0].ID != "one" || posts[1].ID != "two" {
		t.Fatalf("unexpected merged posts: %#v", posts)
	}
}

func TestAPIConfigFromFeedURL(t *testing.T) {
	apiURL, username, err := APIConfigFromFeedURL("https://pixelfed.social/users/lit.pictures.atom")
	if err != nil {
		t.Fatal(err)
	}
	if apiURL != "https://pixelfed.social" || username != "lit.pictures" {
		t.Fatalf("got apiURL=%q username=%q", apiURL, username)
	}
}

func TestCanonicalURL(t *testing.T) {
	got, err := CanonicalURL("HTTPS://Pixelfed.social/p/u/1/?x=1#y")
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://pixelfed.social/p/u/1" {
		t.Fatal(got)
	}
}
