package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/carlo-colombo/pixelfed2wiki/pixelfed"
)

const marker = "Original post: "

type imageEmbed struct {
	Image *util.LexBlob `json:"image"`
	Alt   string        `json:"alt"`
}
type embed struct {
	Type   string       `json:"$type"`
	Images []imageEmbed `json:"images"`
}
type feedPost struct {
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
	Embed     *embed `json:"embed,omitempty"`
}

func main() {
	feedURL := os.Getenv("PIXELFED_FEED_URL")
	cutoff, err := time.Parse(time.RFC3339, os.Getenv("BRIDGE_AFTER"))
	if feedURL == "" || err != nil {
		log.Fatal("PIXELFED_FEED_URL and valid BRIDGE_AFTER are required")
	}
	maxBytes := int64(4 << 20)
	if value := os.Getenv("MAX_IMAGE_BYTES"); value != "" {
		maxBytes, err = strconv.ParseInt(value, 10, 64)
		if err != nil || maxBytes <= 0 {
			log.Fatal("invalid MAX_IMAGE_BYTES")
		}
	}
	dryRun := strings.EqualFold(os.Getenv("DRY_RUN"), "true")
	ctx := context.Background()
	posts, err := pixelfed.ParseURL(ctx, feedURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	pds := os.Getenv("BLUESKY_PDS")
	if pds == "" {
		pds = "https://bsky.social"
	}
	var client *xrpc.Client
	if !dryRun {
		client, _, err = authenticate(ctx, pds, os.Getenv("BLUESKY_HANDLE"), os.Getenv("BLUESKY_PASSWORD"))
		if err != nil {
			log.Fatal(err)
		}
	}
	for _, post := range posts {
		if !post.Published.After(cutoff) {
			continue
		}
		canonical, err := pixelfed.CanonicalURL(post.URL)
		if err != nil {
			log.Printf("skip %q: %v", post.ID, err)
			continue
		}
		if dryRun {
			log.Printf("would bridge %s (%d images)", canonical, len(post.Images))
			continue
		}
		exists, err := alreadyBridged(ctx, client, client.Auth.Did, canonical)
		if err != nil {
			log.Printf("check %s: %v", canonical, err)
			continue
		}
		if exists {
			log.Printf("skip existing %s", canonical)
			continue
		}
		if err := publish(ctx, client, post, canonical, maxBytes); err != nil {
			log.Printf("bridge %s: %v", canonical, err)
		}
	}
}

func authenticate(ctx context.Context, pds, handle, password string) (*xrpc.Client, *atproto.ServerCreateSession_Output, error) {
	if handle == "" || password == "" {
		return nil, nil, fmt.Errorf("BLUESKY_HANDLE and BLUESKY_PASSWORD are required")
	}
	c := &xrpc.Client{Host: pds}
	session, err := atproto.ServerCreateSession(ctx, c, &atproto.ServerCreateSession_Input{Identifier: handle, Password: password})
	if err != nil {
		return nil, nil, err
	}
	c.Auth = &xrpc.AuthInfo{AccessJwt: session.AccessJwt, RefreshJwt: session.RefreshJwt, Did: session.Did, Handle: session.Handle}
	return c, session, nil
}

func alreadyBridged(ctx context.Context, c *xrpc.Client, did, canonical string) (bool, error) {
	cursor := ""
	for {
		result, err := atproto.RepoListRecords(ctx, c, "app.bsky.feed.post", cursor, 100, did, false)
		if err != nil {
			return false, err
		}
		for _, record := range result.Records {
			if record.Value == nil {
				continue
			}
			data, err := json.Marshal(record.Value)
			if err == nil && strings.Contains(string(data), marker+canonical) {
				return true, nil
			}
		}
		if result.Cursor == nil || *result.Cursor == "" {
			return false, nil
		}
		cursor = *result.Cursor
	}
}

func publish(ctx context.Context, c *xrpc.Client, post pixelfed.Post, canonical string, maxBytes int64) error {
	images := post.Images
	if len(images) > 4 {
		images = images[:4]
	}
	if len(images) == 0 {
		return fmt.Errorf("no images")
	}
	refs := make([]imageEmbed, 0, len(images))
	downloader := pixelfed.Client{MaxMediaBytes: 32 << 20}
	for _, imageURL := range images {
		data, contentType, err := downloader.Download(ctx, imageURL)
		if err != nil {
			return err
		}
		data, contentType, err = prepare(data, contentType, maxBytes)
		if err != nil {
			return fmt.Errorf("prepare %s: %w", imageURL, err)
		}
		uploaded, err := atproto.RepoUploadBlob(ctx, c, bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("upload %s (%d bytes): %w", imageURL, len(data), err)
		}
		refs = append(refs, imageEmbed{Image: uploaded.Blob, Alt: post.Title})
	}
	text := buildText(post.Caption, canonical)
	created := post.Published.UTC().Format(time.RFC3339Nano)
	record := &bsky.FeedPost{Text: text, CreatedAt: created, Embed: &bsky.FeedPost_Embed{EmbedImages: &bsky.EmbedImages{Images: func() []*bsky.EmbedImages_Image {
		out := make([]*bsky.EmbedImages_Image, len(refs))
		for i, ref := range refs {
			out[i] = &bsky.EmbedImages_Image{Image: ref.Image, Alt: ref.Alt}
		}
		return out
	}()}}}
	result, err := atproto.RepoCreateRecord(ctx, c, &atproto.RepoCreateRecord_Input{Repo: c.Auth.Did, Collection: "app.bsky.feed.post", Record: &util.LexiconTypeDecoder{Val: record}})
	if err != nil {
		return err
	}
	log.Printf("published Bluesky post for %s: uri=%s cid=%s images=%d", canonical, result.Uri, result.Cid, len(refs))
	return nil
}

func buildText(caption, canonical string) string {
	suffix := "\n\n" + marker + canonical
	caption = strings.TrimSpace(caption)
	limit := 300 - len([]byte(suffix))
	if limit < 0 {
		return string([]rune(suffix))
	}
	b := []byte(caption)
	if len(b) > limit {
		b = b[:limit]
		for !utf8Safe(b) {
			b = b[:len(b)-1]
		}
	}
	if len(b) == 0 {
		return suffix
	}
	return string(b) + suffix
}
func utf8Safe(b []byte) bool { return len(string(b)) == len([]rune(string(b))) || len(b) == 0 }

func prepare(data []byte, contentType string, max int64) ([]byte, string, error) {
	detected := http.DetectContentType(data)
	if !strings.HasPrefix(detected, "image/") {
		return nil, "", fmt.Errorf("unsupported content type %s", detected)
	}
	if int64(len(data)) <= max && (detected == "image/jpeg" || detected == "image/png" || detected == "image/gif") {
		return data, detected, nil
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", err
	}
	for quality := 90; quality >= 20; quality -= 10 {
		scale := 1.0
		if int64(len(data)) > max*2 {
			scale = 0.75
		}
		if scale < 1 {
			img = resizeImage(img, scale)
		}
		var out bytes.Buffer
		if err := jpeg.Encode(&out, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, "", err
		}
		if int64(out.Len()) <= max {
			return out.Bytes(), "image/jpeg", nil
		}
	}
	return nil, "", fmt.Errorf("cannot reduce image from %d bytes below %d", len(data), max)
}
func resizeImage(img image.Image, scale float64) image.Image {
	b := img.Bounds()
	return resizeNearest(img, int(float64(b.Dx())*scale), int(float64(b.Dy())*scale))
}
func resizeNearest(src image.Image, w, h int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(x, y, src.At(x*src.Bounds().Dx()/w, y*src.Bounds().Dy()/h))
		}
	}
	return dst
}
