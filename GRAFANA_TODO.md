# Grafana + SQLite Blog Stats — Session TODO

Track progress on deploying the Grafana/SQLite blog stats stack (see `GRAFANA_PLAN.md`).

## Completed
- [x] Diagnosed empty-result issue: `frser-sqlite-datasource` reads the query from
      the **`queryText`** JSON field, NOT `rawSql`. The provisioned dashboard (and my
      API tests) used `rawSql`, so the plugin received an empty query and returned
      empty frames. Verified `queryText` returns data for all panel queries.
- [x] Fixed dashboard JSON in `spec/blog.yml`: the plugin frontend derives
      `queryText` from **`rawQueryText`** (`applyTemplateVariables`), so provisioned
      dashboard targets must use `rawQueryText` (not `queryText`/`rawSql`). Missing it
      made the frontend send empty `queryText` -> "no data". All 8 panels now use
      `rawQueryText`. Applied to cluster; Grafana file provider reloaded it.
- [x] Fixed "Requests per day" timeseries: added `"timeColumns": ["time"]` so the
      plugin marks the `time` column as a real `time.Time` (without it, Grafana got
      a numeric `time` field and the timeseries panel showed "no data").
- [x] Bumped plugin to `4.0.6` (init-container download URL).
- [x] Set Grafana image to `grafana/grafana-oss:latest` (was `11.6.5`), resolved to
      digest in `images.lock`.
- [x] Fixed `load-stats` CronJob: removed duplicated
      `UPDATE meta SET v=? WHERE k='offset'` block.
- [x] Redeployed via the ytt|kbld|kapp pipeline (deploy.sh's git-crypt unlock step
      fails on a dirty tree, so it was run directly; `.creds.yml` was already unlocked).
- [x] Verified end-to-end: datasource queries return data; Grafana serves through
      Caddy basicauth at `stats.blog.litapp.ovh`; dashboard "Blog Stats" (uid
      `blog-stats`) renders.
- [x] Cleaned up leftover test datasources and stray debug db files in `/data/stats`.

## Active / notes
- **deploy.sh git-crypt unlock**: fails with "Working directory not clean" whenever
  there are uncommitted changes. The pipeline must be run directly (bypassing the
  unlock) or the tree must be committed/stashed first. Not addressed in this session.
- **Grafana version**: running `grafana/grafana-oss:latest` (13.0.2) per decision to
  stay on `latest`. Spec updated to match.

## Verification checklist (all passing)
- [x] Blog pod rolls with Caddy + Grafana sidecar (`2/2` Running).
- [x] Caddy writes JSON lines to `/data/logs/access.log`.
- [x] `load-stats` CronJob ingests rows into `access.db` (offset tracked in `meta`).
- [x] Grafana logs show plugin + datasource (`sqlite`) + dashboard provisioned.
- [x] Dashboard panels return data (Total Requests: 30,013; Top pages populated).
- [x] `https://stats.blog.litapp.ovh` -> basicauth -> read-only Grafana dashboard.

## Follow-ups
- [ ] Run a full `./deploy.sh` once the tree is clean (committed) to keep the
      git-crypt unlock path working and confirm no drift.
- [ ] Commit the spec/`images.lock` changes and the `GRAFANA_PLAN.md`/`GRAFANA_TODO.md`
      notes when ready.
