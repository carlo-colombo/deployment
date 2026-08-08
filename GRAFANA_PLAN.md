# Plan: Grafana + SQLite for Blog Log Analysis

Replaces the previous GoAccess-based stats setup. This document describes the Grafana-based stack used to analyze Caddy access logs for the blog, integrated with `ytt`, `kbld`, and `kapp`.

## Objective
Replace GoAccess with a self-hosted, low-maintenance Grafana stack that reads Caddy access logs stored in a local SQLite file and serves a single-user, read-only dashboard.

## Architecture
- **Log Source**: Caddy writes JSON access logs to `/data/logs/access.log` on the shared `blog-pvc`.
- **Ingestion**: A `load-stats` CronJob (Python, stdlib only) parses new JSON log lines into `/data/stats/access.db` (SQLite), tracking an ingestion offset so stats accumulate cumulatively across rotations/restarts.
- **Query**: A Grafana sidecar runs in the `blog` pod and reads `access.db` via the `frser-sqlite-datasource` plugin.
- **Serving**: Caddy serves Grafana at `stats.blog.litapp.ovh` behind `basicauth`; Grafana is anonymous read-only (Viewer), so basicauth is the single-user gate.
- **Storage**: One SQLite file + Grafana's metadata dir, both on `blog-pvc`. No TSDB or extra PVCs.

## Design decisions
- **JSON logs**: Caddy's built-in `format json` encoder (stock image) is used because the `transform` encoder needed for CLF-with-headers requires the `transform-encoder` plugin and a custom `xcaddy` build. JSON includes `method`, `uri`, `status`, `size`, `ts`, `remote_ip`, and request headers (User-Agent, Referer) without any extra image.
- **SQLite datasource**: Grafana has no built-in SQLite datasource; the community `frser-sqlite-datasource` plugin is auto-installed via `GF_INSTALL_PLUGINS` and provisioned with `path: /data/stats/access.db`.
- **Single user / static dashboard**: basicauth in front + Grafana anonymous Viewer role; the dashboard is provisioned from a versioned ConfigMap and set read-only.

## Components in `spec/blog.yml`
1. **Caddyfile**: blog site `log` block now uses `format json`.
2. **Stats site**: basicauth + `reverse_proxy 127.0.0.1:3000` to the Grafana sidecar (replaces static `/data/stats` file serving).
3. **Grafana sidecar** in the `blog` Deployment:
   - `grafana/grafana-oss` image, port 3000, no Service needed (localhost within the pod).
   - Env: `GF_INSTALL_PLUGINS=frser-sqlite-datasource`, `GF_PATHS_DATA=/data/grafana`, `GF_SERVER_ROOT_URL=https://stats.blog.litapp.ovh`, anonymous Viewer auth, admin password from `values`.
4. **`load-stats` CronJob** (replaces `generate-stats`): `python:3-alpine`, inline script that ingests JSON lines into `access.db` (table `access`; columns `ts, remote_ip, method, uri, status, size, user_agent, referer`).
5. **ConfigMaps**: `grafana-datasources` (provisioned SQLite datasource, uid `sqlite`), `grafana-dashboards` (file provider), `grafana-dashboard-json` (the read-only dashboard).

## Config
`spec/values.yml` gained `blog.stats.grafana_password`; the real value is supplied via `.creds.yml` (git-crypt encrypted).

## Verification
1. `./deploy.sh` -> blog pod rolls with Caddy + Grafana sidecar; `load-stats` CronJob created.
2. `kubectl logs -l app=blog -c caddy` -> JSON access lines in `/data/logs/access.log`.
3. `kubectl create job --from=cronjob/load-stats initial-load` -> rows with `user_agent`/`referer` inserted into `access.db`.
4. `kubectl logs -l app=blog -c grafana` -> plugin + datasource + dashboard provisioned.
5. Open `https://stats.blog.litapp.ovh` -> basicauth -> read-only Grafana dashboard.

## Notes / caveats
- The Grafana report is a live web app, not a fully static HTML file; the dashboard definition is static/versioned and read-only.
- Log format migration: `access.log` will briefly contain mixed CLF (old) and JSON lines. The `load-stats` script skips non-JSON lines, so ingestion starts cleanly from the offset at the JSON transition.
- SQLite is fine here: single-node k3s, low write volume; WAL allows the CronJob to write while Grafana reads.
