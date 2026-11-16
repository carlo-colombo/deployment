package tiddlywiki

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mmcdole/gofeed"
)

func imageServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/fail/") {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, "image-bytes-", strings.TrimPrefix(r.URL.Path, "/"))
	}))
}

func fixtureItem(title, content string) gofeed.Item {
	return gofeed.Item{
		Title:     title,
		Content:   content,
		Link:      "https://pixelfed.example/p/account/123",
		Published: "2026-01-02T03:04:05Z",
	}
}

func TestThumbnailDimensionsUseTiddlyWikiFieldNames(t *testing.T) {
	encoded, err := json.Marshal(Tiddler{
		ThumbnailWidth:             "360",
		ThumbnailHeight:            "240",
		AdditionalThumbnailWidths:  "360 180",
		AdditionalThumbnailHeights: "240 360",
	})
	if err != nil {
		t.Fatalf("marshal tiddler: %v", err)
	}

	var fields map[string]string
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("unmarshal tiddler fields: %v", err)
	}
	for field, want := range map[string]string{
		"thumbnail-width":              "360",
		"thumbnail-height":             "240",
		"additional-thumbnail-widths":  "360 180",
		"additional-thumbnail-heights": "240 360",
	} {
		if fields[field] != want {
			t.Errorf("%s = %q, want %q", field, fields[field], want)
		}
	}
}

func TestNewTiddlersUnsplitMultiPhotoPreservesOrder(t *testing.T) {
	server := imageServer(t)
	defer server.Close()
	urls := []string{server.URL + "/first.jpg", server.URL + "/second.jpg", server.URL + "/third.jpg"}
	content := fmt.Sprintf(`<img src="%s"><img src="%s"><img src="%s"><p>Caption #birds</p>`, urls[0], urls[1], urls[2])

	tiddlers := NewTiddlers(fixtureItem("Birds #birds", content))
	if len(tiddlers) != 1 {
		t.Fatalf("got %d tiddlers, want one canonical post tiddler", len(tiddlers))
	}
	tiddler := tiddlers[0]
	if tiddler.Err != nil {
		t.Fatalf("NewTiddlers returned error: %v", tiddler.Err)
	}
	if tiddler.Title != "Birds #birds" {
		t.Errorf("title = %q, want %q", tiddler.Title, "Birds #birds")
	}
	if tiddler.Image != urls[0] {
		t.Errorf("primary image = %q, want %q", tiddler.Image, urls[0])
	}
	if tiddler.AdditionalImages != strings.Join(urls[1:], " ") {
		t.Errorf("additional images = %q, want %q", tiddler.AdditionalImages, strings.Join(urls[1:], " "))
	}
	if strings.Contains(tiddler.Text, "<img src={{!!image}}") {
		t.Errorf("multi-photo text should leave image rendering to the wiki ViewTemplate: %q", tiddler.Text)
	}
	if !strings.Contains(tiddler.Text, "Caption #birds") {
		t.Errorf("multi-photo text should retain its caption: %q", tiddler.Text)
	}
	if len(tiddler.media) != len(urls) {
		t.Fatalf("got %d downloaded photos, want %d", len(tiddler.media), len(urls))
	}
	for i, want := range urls {
		if tiddler.media[i].source != want {
			t.Errorf("photo %d source = %q, want %q", i+1, tiddler.media[i].source, want)
		}
	}
}

func TestNewTiddlersSplitPostKeepsOneTiddlerPerPhoto(t *testing.T) {
	server := imageServer(t)
	defer server.Close()
	urls := []string{server.URL + "/first.jpg", server.URL + "/second.jpg"}
	content := fmt.Sprintf(`<img src="%s"><img src="%s"><p>Caption</p>`, urls[0], urls[1])

	tiddlers := NewTiddlers(fixtureItem("A walk #split #birds", content))
	if len(tiddlers) != 2 {
		t.Fatalf("got %d tiddlers, want two split photo tiddlers", len(tiddlers))
	}
	wantTitles := []string{"A walk #split #birds", "A walk #split #birds /2"}
	for i, tiddler := range tiddlers {
		if tiddler.Err != nil {
			t.Fatalf("tiddler %d returned error: %v", i+1, tiddler.Err)
		}
		if tiddler.Title != wantTitles[i] {
			t.Errorf("tiddler %d title = %q, want %q", i+1, tiddler.Title, wantTitles[i])
		}
		if tiddler.Image != urls[i] {
			t.Errorf("tiddler %d image = %q, want %q", i+1, tiddler.Image, urls[i])
		}
		if tiddler.AdditionalImages != "" || len(tiddler.media) != 1 {
			t.Errorf("split tiddler %d should contain only its own photo", i+1)
		}
		if strings.Contains(tiddler.Tags, "pf:split") {
			t.Errorf("split control tag leaked into tags: %q", tiddler.Tags)
		}
	}
}

func TestNewTiddlersSplitCanBeInCaption(t *testing.T) {
	server := imageServer(t)
	defer server.Close()
	content := fmt.Sprintf(`<img src="%s/first.jpg"><img src="%s/second.jpg"><p>Caption with #split</p>`, server.URL, server.URL)

	if got := len(NewTiddlers(fixtureItem("A walk #birds", content))); got != 2 {
		t.Fatalf("got %d tiddlers for caption #split, want two", got)
	}
}

func TestNewTiddlersSinglePhotoCompatibility(t *testing.T) {
	server := imageServer(t)
	defer server.Close()
	imageURL := server.URL + "/single.jpg"
	content := fmt.Sprintf(`<img src="%s"><p>Caption</p>`, imageURL)

	tiddlers := NewTiddlers(fixtureItem("Single #photo", content))
	if len(tiddlers) != 1 {
		t.Fatalf("got %d tiddlers, want one", len(tiddlers))
	}
	tiddler := tiddlers[0]
	if tiddler.Err != nil {
		t.Fatalf("NewTiddlers returned error: %v", tiddler.Err)
	}
	if tiddler.Title != "Single #photo" || tiddler.Image != imageURL {
		t.Errorf("single photo title/image changed: title=%q image=%q", tiddler.Title, tiddler.Image)
	}
	if tiddler.AdditionalImages != "" || len(tiddler.media) != 1 {
		t.Errorf("single photo should not have additional images")
	}
	if !strings.Contains(tiddler.Text, "<img src={{!!image}}") {
		t.Errorf("single-photo text should keep its existing inline image")
	}
}

func TestNewTiddlersImageErrorIdentifiesPhotoAndPost(t *testing.T) {
	server := imageServer(t)
	defer server.Close()
	imageURL := server.URL + "/fail/photo.jpg"
	content := fmt.Sprintf(`<img src="%s"><p>Caption</p>`, imageURL)

	tiddlers := NewTiddlers(fixtureItem("Failed post", content))
	if len(tiddlers) != 1 || tiddlers[0].Err == nil {
		t.Fatalf("expected one tiddler with a download error")
	}
	for _, expected := range []string{"Failed post", imageURL, "photo 1"} {
		if !strings.Contains(tiddlers[0].Err.Error(), expected) {
			t.Errorf("error %q does not identify %q", tiddlers[0].Err, expected)
		}
	}
}
