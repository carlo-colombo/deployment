# Pixelfed thumbnail-dimensions backfill plan

## Goal and safety boundary

Backfill actual resized-thumbnail dimensions on existing Pixelfed picture tiddlers so gallery images can retain their natural proportions while the browser reserves space before lazy loading. This is a plan only: **no real wiki instance was queried or changed**, and no deployment is authorized by this document.

The importer now writes these TiddlyWiki string fields for newly created posts:

- `thumbnail-width`, `thumbnail-height`: dimensions of the primary uploaded thumbnail.
- `additional-thumbnail-widths`, `additional-thumbnail-heights`: space-separated dimensions aligned by index with `additional-thumbnails`.

Use dimensions of the thumbnail bytes stored under `thumbs/`, not dimensions of the original image. The importer resizes thumbnails to fit within 360 × 360 while preserving aspect ratio.

## Backfill procedure

1. **Prepare a maintenance command.** Implement a one-shot tool separate from the scheduled importer. Make dry-run the default; require an explicit `--apply` flag to write. Restrict image downloads to the configured public object-storage host, enforce timeouts and response-size limits, and use low bounded concurrency.
2. **Back up and inventory the real wiki.** Export eligible published `pixelfed` tiddlers and create a restorable wiki backup. Record the eligible-tiddler count, existing size fields, and number of primary/additional thumbnail URLs. Keep backups and reports securely; never include credentials.
3. **Run a read-only dry run.** For each tiddler, inspect its `thumbnail` and each URL in `additional-thumbnails` using image decoding/config inspection. Produce a report keyed by title with proposed dimensions, skipped fields, and decode/URL errors. Do not overwrite non-empty values silently.
4. **Validate against a copy first.** Apply the report to a restored test copy. Build the static blog and inspect landscape, portrait, square, and multi-photo pages. Confirm each emitted `width`/`height` pair matches the corresponding thumbnail and that the gallery keeps natural masonry proportions.
5. **Apply to production in checkpoints.** After reviewing the report, backup, and rollback point, update only missing dimension fields in small batches (for example, 50 tiddlers per batch). Record completed titles so the operation is resumable. Do not use `CreateIfNew` for this step; it intentionally skips existing tiddlers.
6. **Verify and close out.** Re-export affected tiddlers; check positive dimensions and exact alignment between additional-thumbnail URLs and their width/height lists. Rebuild the static blog and inspect representative pages before considering the migration complete.

## Rollback

The backfill only adds metadata and does not change image URLs. If verification fails, restore the pre-migration wiki backup or remove the four dimension fields from the recorded affected titles, then rebuild and verify the prior gallery output.

## Go/no-go checks

- Backup restore has been tested on a non-production copy.
- Dry-run has no unexplained URL-host or image-decode errors.
- Existing non-empty dimensions are not silently replaced.
- Additional-thumbnail width/height counts match the additional-thumbnail URL count.
- An operator explicitly approves the real-instance write window and rollback point.
