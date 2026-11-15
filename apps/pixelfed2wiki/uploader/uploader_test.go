package uploader

import "testing"

func TestPublicBaseURLOverridesS3EndpointAddress(t *testing.T) {
	t.Setenv("S3_PUBLIC_BASE_URL", "http://pixelfed-import.web.garage.localhost:3902/")

	got := (Uploader{}).getPublicBucketUrl()
	want := "http://pixelfed-import.web.garage.localhost:3902"
	if got != want {
		t.Fatalf("getPublicBucketUrl() = %q, want %q", got, want)
	}
}
