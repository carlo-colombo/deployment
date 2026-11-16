package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestInspectTiddlerPlansPrimaryAndAdditionalDimensions(t *testing.T) {
	primary := pngBytes(t, 800, 400)
	additional := pngBytes(t, 120, 240)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		if r.URL.Path == "/primary.png" {
			_, _ = w.Write(primary)
			return
		}
		if r.URL.Path == "/additional.png" {
			_, _ = w.Write(additional)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	wikiURL, err := url.Parse("https://wiki.example.test")
	if err != nil {
		t.Fatal(err)
	}
	backfiller := &backfiller{
		wikiURL:     wikiURL,
		client:      server.Client(),
		allowedHost: map[string]bool{"127.0.0.1": true},
		maxImage:    1 << 20,
		cache:       make(map[string]dimensions),
	}
	item := tiddler{
		"title":                 jsonString("Test photo"),
		"pixelfed":              jsonString("https://pixelfed.example/p/1"),
		"published":             jsonString("20260519090000000"),
		"thumbnail":             jsonString(server.URL + "/primary.png"),
		"additional-thumbnails": jsonString(server.URL + "/additional.png"),
	}

	row, changes := backfiller.inspectTiddler(item)
	if row.Status != "would-update" {
		t.Fatalf("status = %q, want would-update (errors: %v)", row.Status, row.Errors)
	}
	for field, want := range map[string]string{
		"thumbnail-width":              "800",
		"thumbnail-height":             "400",
		"additional-thumbnail-widths":  "120",
		"additional-thumbnail-heights": "240",
	} {
		if changes[field] != want {
			t.Errorf("change %s = %q, want %q", field, changes[field], want)
		}
	}
}

func TestInspectTiddlerDoesNotOverwriteMismatchedDimensions(t *testing.T) {
	primary := pngBytes(t, 800, 400)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(primary)
	}))
	defer server.Close()
	wikiURL, _ := url.Parse("https://wiki.example.test")
	backfiller := &backfiller{
		wikiURL:     wikiURL,
		client:      server.Client(),
		allowedHost: map[string]bool{"127.0.0.1": true},
		maxImage:    1 << 20,
		cache:       make(map[string]dimensions),
	}
	item := tiddler{
		"title":           jsonString("Test photo"),
		"pixelfed":        jsonString("https://pixelfed.example/p/1"),
		"published":       jsonString("20260519090000000"),
		"thumbnail":       jsonString(server.URL + "/primary.png"),
		"thumbnail-width": jsonString("999"),
	}

	row, changes := backfiller.inspectTiddler(item)
	if row.Status != "error" || len(row.Errors) == 0 {
		t.Fatalf("status = %q, errors = %v; want a mismatch error", row.Status, row.Errors)
	}
	if len(changes) != 0 {
		t.Fatalf("mismatched existing dimensions produced changes: %v", changes)
	}
}

func TestImageURLsUsesImageFallbackAndAdditionalOrder(t *testing.T) {
	item := tiddler{
		"image":                 jsonString("https://images.example.test/primary.jpg"),
		"additional-thumbnails": jsonString("https://images.example.test/one.jpg https://images.example.test/two.jpg"),
	}
	want := []string{
		"https://images.example.test/primary.jpg",
		"https://images.example.test/one.jpg",
		"https://images.example.test/two.jpg",
	}
	got := imageURLs(item)
	if len(got) != len(want) {
		t.Fatalf("got URLs %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Errorf("URL %d = %q, want %q", index, got[index], want[index])
		}
	}
}

func TestValidateBackupPathRequiresOutsideRepository(t *testing.T) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := validateBackupPath(filepath.Join(workingDirectory, "backup.json")); err == nil {
		t.Fatal("expected backup path inside repository to be rejected")
	}
	if _, err := validateBackupPath(filepath.Join(os.TempDir(), "thumbnail-backup-test.json")); err != nil {
		t.Fatalf("temporary backup path should be accepted: %v", err)
	}
}

func pngBytes(t *testing.T, width, height int) []byte {
	t.Helper()
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, width, height))); err != nil {
		t.Fatalf("encode test image: %v", err)
	}
	return buffer.Bytes()
}

func jsonString(value string) json.RawMessage {
	encoded, _ := json.Marshal(value)
	return encoded
}
