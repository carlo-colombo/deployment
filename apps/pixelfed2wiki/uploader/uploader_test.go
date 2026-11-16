package uploader

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestMakeThumbnailReturnsActualResizedDimensions(t *testing.T) {
	original := image.NewRGBA(image.Rect(0, 0, 800, 400))
	var source bytes.Buffer
	if err := png.Encode(&source, original); err != nil {
		t.Fatalf("encode source image: %v", err)
	}

	thumbnailBytes, width, height, err := makeThumbnail(source.Bytes(), "image/png")
	if err != nil {
		t.Fatalf("makeThumbnail() error: %v", err)
	}
	if width != 360 || height != 180 {
		t.Fatalf("thumbnail dimensions = %dx%d, want 360x180", width, height)
	}

	thumbnail, err := jpeg.Decode(bytes.NewReader(thumbnailBytes))
	if err != nil {
		t.Fatalf("decode generated thumbnail: %v", err)
	}
	if got := thumbnail.Bounds().Size(); got.X != width || got.Y != height {
		t.Fatalf("encoded thumbnail dimensions = %v, want %dx%d", got, width, height)
	}
}

func TestPublicBaseURLOverridesS3EndpointAddress(t *testing.T) {
	t.Setenv("S3_PUBLIC_BASE_URL", "http://pixelfed-import.web.garage.localhost:3902/")

	got := (Uploader{}).getPublicBucketUrl()
	want := "http://pixelfed-import.web.garage.localhost:3902"
	if got != want {
		t.Fatalf("getPublicBucketUrl() = %q, want %q", got, want)
	}
}
