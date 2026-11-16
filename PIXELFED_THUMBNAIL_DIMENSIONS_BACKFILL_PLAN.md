# Pixelfed thumbnail-dimensions backfill plan

## Goal and safety boundary

Backfill actual resized-thumbnail dimensions on existing Pixelfed picture tiddlers so gallery images can retain their natural proportions while the browser reserves space before lazy loading. **No live tiddlers have been changed.** A read-only inventory and dry run completed; applying the reported updates still requires operator approval.

The importer now writes these TiddlyWiki string fields for newly created posts. The one-shot maintenance command is implemented at `apps/pixelfed2wiki/cmd/pixelfed-thumbnail-backfill`:

- `thumbnail-width`, `thumbnail-height`: dimensions of the primary uploaded thumbnail.
- `additional-thumbnail-widths`, `additional-thumbnail-heights`: space-separated dimensions aligned by index with `additional-thumbnails`.

Use dimensions of the thumbnail bytes stored under `thumbs/`, not dimensions of the original image. The importer resizes thumbnails to fit within 360 × 360 while preserving aspect ratio.

## Read-only dry-run result

- Exported 998 non-system tiddlers and found 166 eligible published Pixelfed tiddlers.
- Measured 170 primary/additional thumbnail references across the allowed image hosts.
- 166 tiddlers need dimension fields; 0 measurement or validation errors were reported.
- The pre-migration export was loaded into an isolated TiddlyWiki copy. The dry-run changes were applied to all 166 Pixelfed tiddlers in that copy, then the static blog was rebuilt; all 177 gallery image elements had matching width/height attributes.
- A private pre-migration export and JSON report are stored under `/tmp` in the current workstation session. They are not committed.
- No production writes or deployment have occurred. Review the dry-run report and confirm the production write window before applying updates.

## Backfill procedure

1. **Maintenance command implemented.** `cmd/pixelfed-thumbnail-backfill` defaults to read-only reporting, requires `--apply --confirm-live-write` for updates, restricts thumbnail requests to explicit HTTPS host allowlists, and enforces timeouts and response-size limits.
2. **Back up and inventory the real wiki.** A private full non-system-tiddler export including text was saved before the dry run; it is outside the repository. Retain it securely and confirm the restore check before production writes. The inventory found 166 eligible published `pixelfed` tiddlers. Record the eligible-tiddler count, existing size fields, and number of primary/additional thumbnail URLs. Keep backups and reports securely; never include credentials.
3. **Run a read-only dry run.** For each tiddler, inspect its `thumbnail` and each URL in `additional-thumbnails` using image decoding/config inspection. Produce a report keyed by title with proposed dimensions, skipped fields, and decode/URL errors. Do not overwrite non-empty values silently.
4. **Validated against a copy.** The export was restored to an isolated TiddlyWiki, all 166 proposed updates were applied there, and the static blog was rebuilt. All 177 image elements in the staging gallery had width/height attributes; the primary and additional thumbnail dimensions matched the report.
5. **Apply to production in checkpoints.** After reviewing the report, backup, and rollback point, update only missing dimension fields in small batches (for example, 50 tiddlers per batch). Record completed titles so the operation is resumable. Do not use `CreateIfNew` for this step; it intentionally skips existing tiddlers.
6. **Verify and close out.** Re-export affected tiddlers; check positive dimensions and exact alignment between additional-thumbnail URLs and their width/height lists. Rebuild the static blog and inspect representative pages before considering the migration complete.

## Rollback

The backfill only adds metadata and does not change image URLs. If verification fails, restore the pre-migration wiki backup or remove the four dimension fields from the recorded affected titles, then rebuild and verify the prior gallery output.

## Go/no-go checks

- Backup restore and the complete dimension migration have been tested on a non-production copy.
- Dry-run has no unexplained URL-host or image-decode errors.
- Existing non-empty dimensions are not silently replaced.
- Additional-thumbnail width/height counts match the additional-thumbnail URL count.
- An operator explicitly approves the real-instance write window and rollback point.
