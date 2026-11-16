# Pixelfed thumbnail-dimensions backfill plan

## Goal and safety boundary

Backfill actual resized-thumbnail dimensions on existing Pixelfed picture tiddlers so gallery images can retain their natural proportions while the browser reserves space before lazy loading. Live changes were applied only after the operator approved the dry-run scope and confirmed other backups were available.

The importer now writes these TiddlyWiki string fields for newly created posts. The one-shot maintenance command is implemented at `apps/pixelfed2wiki/cmd/pixelfed-thumbnail-backfill`:

- `thumbnail-width`, `thumbnail-height`: dimensions of the primary uploaded thumbnail.
- `additional-thumbnail-widths`, `additional-thumbnail-heights`: space-separated dimensions aligned by index with `additional-thumbnails`.

Use dimensions of the thumbnail bytes stored under `thumbs/`, not dimensions of the original image. The importer resizes thumbnails to fit within 360 × 360 while preserving aspect ratio.

## Run results

- Exported 998 non-system tiddlers and found 166 eligible published Pixelfed tiddlers.
- Measured 170 primary/additional thumbnail references across five allowlisted image hosts. The dry run reported 166 updates and 0 errors.
- Restored the pre-migration export into an isolated TiddlyWiki copy, applied all 166 updates there, and rebuilt the static blog; all 177 image elements in that gallery had width/height attributes.
- After operator approval, applied the production updates in four batches (50, 50, 50, 16). No batch errors occurred.
- A final read-only verification found 166 complete tiddlers, 0 pending updates, and 0 errors.
- Updated the five live gallery/home/page templates and regenerated the static blog. The live gallery now has dimensions on 167/167 image tags; the photography tag page has dimensions on 105/105.
- Private pre-migration and per-batch exports plus the JSON dry-run report are stored under `/tmp` with mode `0600`. No credentials are included.

## Backfill procedure

1. **Maintenance command implemented.** `cmd/pixelfed-thumbnail-backfill` defaults to read-only reporting, requires `--apply --confirm-live-write` for updates, restricts thumbnail requests to explicit HTTPS host allowlists, and enforces timeouts and response-size limits.
2. **Back up and inventory the real wiki.** A private full non-system-tiddler export including text was saved before the dry run; it is outside the repository. Retain it securely and confirm the restore check before production writes. The inventory found 166 eligible published `pixelfed` tiddlers. Record the eligible-tiddler count, existing size fields, and number of primary/additional thumbnail URLs. Keep backups and reports securely; never include credentials.
3. **Run a read-only dry run.** For each tiddler, inspect its `thumbnail` and each URL in `additional-thumbnails` using image decoding/config inspection. Produce a report keyed by title with proposed dimensions, skipped fields, and decode/URL errors. Do not overwrite non-empty values silently.
4. **Validated against a copy.** The export was restored to an isolated TiddlyWiki, all 166 proposed updates were applied there, and the static blog was rebuilt. All 177 image elements in the staging gallery had width/height attributes; the primary and additional thumbnail dimensions matched the report.
5. **Apply to production in checkpoints.** After reviewing the report, backup, and rollback point, update only missing dimension fields in small batches (for example, 50 tiddlers per batch). Record completed titles so the operation is resumable. Do not use `CreateIfNew` for this step; it intentionally skips existing tiddlers.
6. **Verify and close out.** Re-export affected tiddlers; check positive dimensions and exact alignment between additional-thumbnail URLs and their width/height lists. Rebuild the static blog and inspect representative pages before considering the migration complete.

## Rollback

The backfill only adds metadata and does not change image URLs. If verification fails, restore the pre-migration wiki backup or remove the four dimension fields from the recorded affected titles, then rebuild and verify the prior gallery output.

## Completed safety checks

- The backup export restored into a non-production TiddlyWiki copy; the dimension changes applied and the static rendering check passed.
- The dry run reported no URL-host or image-decode errors.
- Existing non-empty dimensions were not silently replaced.
- Additional-thumbnail dimension counts matched their URL counts.
- The operator approved the production batch writes; the rollback exports were retained outside the repository.
