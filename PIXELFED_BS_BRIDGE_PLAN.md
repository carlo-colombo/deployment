# Plan: Pixelfed to Bluesky Bridge

## Objective

Add a `pixelfed-bs-bridge` command alongside `pixelfed2wiki` in the existing Pixelfed Go module. The bridge periodically reads a Pixelfed feed and publishes new image posts to a Bluesky account without requiring a local state PVC.

## Command layout

Keep both applications side by side as commands in `apps/pixelfed2wiki`:

```text
apps/pixelfed2wiki/
  go.mod
  go.sum
  cmd/
    pixelfed2wiki/
      main.go
    pixelfed-bs-bridge/
      main.go
  pixelfed/
    feed.go
    media.go
    feed_test.go
  tiddlywiki/
  uploader/
```

`pixelfed2wiki` remains the existing importer executable. `pixelfed-bs-bridge` is a separate executable that reuses the shared Pixelfed package.

## Feed source

Use the public Pixelfed Atom feed as the initial source:

- It already exposes post IDs, URLs, timestamps, captions, and image URLs.
- It does not require Pixelfed credentials.
- It is compatible with the existing `pixelfed2wiki` implementation.
- The parser must use `published` when available and fall back to `updated`.
- All timestamps are parsed and compared as RFC3339 values.

The shared feed package should make the source replaceable later. A future API adapter can be added if Atom feeds become unreliable or do not expose sufficient media metadata. The initial implementation should not require the Pixelfed API.

## Shared Pixelfed package

Extract common functionality from the current importer into a package independent of TiddlyWiki and S3:

- Parse an Atom feed from a URL or reader.
- Normalize each entry into a Pixelfed post type.
- Preserve the stable entry ID and Pixelfed URL.
- Extract the caption and title.
- Extract all image URLs in feed order.
- Parse the post timestamp as `time.Time`.
- Download media through an HTTP client with timeout, status validation, and response-size limits.

Keep destination-specific behavior out of the shared package:

- `pixelfed2wiki` retains Tiddler creation and S3 upload behavior.
- `pixelfed-bs-bridge` owns Bluesky authentication, duplicate detection, image preparation, blob uploads, and post creation.

The importer should be adapted to use the shared normalized post/media model while preserving its current behavior of processing every image.

## Bridge behavior

For every feed entry:

1. Parse its RFC3339 publication timestamp.
2. Ignore entries older than `BRIDGE_AFTER`.
3. Canonicalize and validate the Pixelfed URL.
4. Query the Bluesky destination account for existing post records.
5. Skip the entry if its canonical Pixelfed URL is already present.
6. Download the first four images only, preserving order.
7. Resize or recompress each image until it is no larger than `MAX_IMAGE_BYTES`.
8. Upload each prepared image to Bluesky with `com.atproto.repo.uploadBlob`.
9. Create one `app.bsky.feed.post` containing the uploaded images.
10. Include the caption and the original Pixelfed URL in the Bluesky post.

A Pixelfed post containing more than four images is intentionally truncated to its first four images. `pixelfed2wiki` is not subject to this limit.

## Bluesky duplicate detection

Do not use a PVC or local state file.

Use the Bluesky API as the source of truth:

- Authenticate with `BLUESKY_HANDLE` and an app password in `BLUESKY_PASSWORD`.
- List the destination repository's `app.bsky.feed.post` records using the repository API.
- Paginate through records and inspect post text for the stable Pixelfed URL marker.
- Stop scanning once records are older than the relevant cutoff where safe, while retaining pagination support for correctness.
- Treat a matching canonical Pixelfed URL as already bridged.

Every generated post should include a stable marker such as:

```text
Original post: https://pixelfed.social/p/account/post-id
```

The URL must be included even when the Pixelfed caption is empty. Duplicate lookup should match the URL rather than title or caption, since those can change or repeat.

## Bluesky post content

Create one Bluesky post per Pixelfed post, with:

- Original Pixelfed caption as the main text.
- Original Pixelfed URL appended as a stable source reference.
- Image embed containing up to four uploaded images.
- Hashtags represented as Bluesky facets where practical.
- `createdAt` based on the Pixelfed timestamp when accepted by the API; otherwise use the current UTC time and log the fallback.

The implementation must respect Bluesky record and text byte limits. Caption truncation must be UTF-8 safe and must preserve the Pixelfed URL marker.

## Image preparation

For each selected image:

- Download with a bounded HTTP client and context cancellation.
- Validate the response status and detected content type.
- Decode supported raster formats.
- If the original is within the limit, preserve it when Bluesky accepts the MIME type.
- If larger than `MAX_IMAGE_BYTES` (default `4 MiB`), resize dimensions and/or recompress quality.
- Prefer JPEG for photographic content when conversion is necessary.
- Preserve PNG only when it remains within the limit and is suitable for upload.
- Fail the individual Pixelfed post clearly if an image cannot be decoded or reduced safely.

Because PDS limits can be stricter than the requested 4 MiB threshold, make the target configurable with `MAX_IMAGE_BYTES`. The default is 4 MiB, and upload errors should include the image URL and detected size.

## Configuration

Expected bridge configuration:

- `PIXELFED_FEED_URL`: Atom feed URL.
- `BRIDGE_AFTER`: RFC3339 cutoff; only posts newer than this value are eligible.
- `BLUESKY_PDS`: optional PDS URL, defaulting to `https://bsky.social`.
- `BLUESKY_HANDLE`: destination account handle.
- `BLUESKY_PASSWORD`: Bluesky app password.
- `MAX_IMAGE_BYTES`: optional byte limit, defaulting to `4194304`.
- `DRY_RUN`: optional mode that parses and reports eligible posts without publishing.

Credentials must be supplied through Kubernetes Secret references. Do not commit test credentials or plaintext passwords to deployment specs.

## Deployment

Add a new image and CronJob to `spec/`, alongside the existing `pixelfed-importer`:

- Image name: `pixelfed-bs-bridge`.
- Build source: `apps/pixelfed2wiki`.
- Build import path: `github.com/carlo-colombo/pixelfed2wiki/cmd/pixelfed-bs-bridge`.
- Image destination: `rg.nl-ams.scw.cloud/carlo-colombo/pixelfed-bs-bridge`.
- CronJob name: `pixelfed-bs-bridge`.
- Use `concurrencyPolicy: Forbid` to prevent overlapping duplicate scans.
- Use `restartPolicy: Never` and a bounded backoff policy.
- Configure the feed URL and RFC3339 cutoff through deployment values or environment variables.
- Read Bluesky handle and app password from a Kubernetes Secret.
- No PVC is required.

The existing importer image should build the new `cmd/pixelfed2wiki` path after its main package is moved from the module root.

## Tests

Add tests for:

- Atom parsing from `example.feed`.
- `published` and `updated` timestamp fallback.
- RFC3339 cutoff filtering.
- Stable Pixelfed URL extraction.
- Caption and image extraction.
- Multi-image ordering and first-four selection.
- Duplicate URL matching against Bluesky records.
- UTF-8-safe caption and URL construction.
- Image resizing/recompression under the configured byte limit.
- Unsupported formats, HTTP errors, and oversized download responses.
- Dry-run behavior.

Run at minimum:

```bash
cd apps/pixelfed2wiki
go fmt ./...
go test ./...
go vet ./...
go build ./cmd/pixelfed2wiki
go build ./cmd/pixelfed-bs-bridge
```

## Verification

Before deployment:

1. Run the bridge with `DRY_RUN=true` and a test feed/cutoff.
2. Confirm only entries newer than `BRIDGE_AFTER` are selected.
3. Confirm only the first four images are prepared.
4. Test a Bluesky login using the supplied app password.
5. Publish one controlled test post.
6. Run the bridge again and verify the same Pixelfed URL is skipped.
7. Verify the published post contains the Pixelfed URL, caption, and expected images.
8. Build and render the deployment manifests through the existing `ytt | kbld` pipeline.
9. Deploy the CronJob and inspect job logs.

## Open implementation notes

- Confirm the exact Indigo version and generated types available in the module before implementing Bluesky record listing, blob upload, and post creation.
- Use Bluesky repository records rather than only the account's public feed for duplicate detection.
- Avoid copying credentials from the reference `bs-reposter-liker` spec; use Secret references for this application.
- Preserve the existing importer behavior while refactoring its feed parsing into the shared package.
