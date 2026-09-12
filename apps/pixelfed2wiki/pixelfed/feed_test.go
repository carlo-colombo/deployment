package pixelfed

import (
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

func TestCanonicalURL(t *testing.T) {
	got, err := CanonicalURL("HTTPS://Pixelfed.social/p/u/1/?x=1#y")
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://pixelfed.social/p/u/1" {
		t.Fatal(got)
	}
}
