# Agent Guidelines: Deployment Repository

This document provides essential information for agentic coding agents operating in this repository.

## 1. Project Overview

This repository is a monorepo containing several services:
- **Go App** (`apps/pixelfed2wiki`): specialized tool for Pixelfed integration.
- **TiddlyWiki**: The core wiki instance (deployed via `apps/Tiddlywiki.Dockerfile`).
- **Infrastructure** (`spec/`): Kubernetes-ready deployment specs using Carvel tools.

---

## 2. Build, Lint, and Test Commands

### Go (`apps/pixelfed2wiki`)
- **Build:**
  ```bash
  cd apps/pixelfed2wiki && go build
  ```
- **Test:**
  ```bash
  cd apps/pixelfed2wiki && go test ./...
  ```
- **Formatting:**
  ```bash
  cd apps/pixelfed2wiki && go fmt ./...
  ```

### Deployment & Infrastructure
- **Deploy:**
  ```bash
  ./deploy.sh
  ```
- **Infrastructure Templating:**
  The project uses Carvel tools (`ytt`, `kbld`, `kapp`). Configuration templates are in `spec/`.

---

## 3. Code Style Guidelines

### Go Style
- **Formatting:** Strictly follow `gofmt` standards.
- **Error Handling:** Explicitly check for errors using `if err != nil`. 
- **Startup:** Use `log.Fatal` for unrecoverable errors during initialization.
- **Packages:** Manage dependencies via `go.mod`.
- **Key Libraries:**
  - `minio-go`: S3-compatible storage.
  - `gofeed`: RSS/Atom feed parsing.

---

## 4. Infrastructure & Deployment

### Kubernetes Environment
- **Cluster:** Single-node **k3s** cluster (`deep-space-d6`).
- **Namespace:** Most applications reside in the `default` namespace.
- **Ingress:** Managed by **Traefik v3**. It uses `IngressRoute` (CRD) for routing.
- **Storage:** Uses `local-path` provisioner. Key PVCs: `mywiki-pvc` (10Gi), `blog-pvc` (10Gi).

### Carvel Toolchain
The project uses the Carvel suite for deployment:
- `ytt`: YAML templating (logic in `.star` files).
- `kbld`: Image building and digest resolution (outputs `images.lock`).
- `kapp`: Deployment and resource management.
  - Check app status: `kapp inspect -a tiddlywiki`
  - List apps: `kapp list`

### Secrets & Config
- **Secrets:** We now use `git-crypt` to manage encrypted secrets within the repository. The `.creds.yml` file, which contains sensitive configurations, should be encrypted using `git-crypt` and committed.
  - **Setup:**
    1.  Install `git-crypt` (e.g., `brew install git-crypt` on macOS, `sudo apt-get install git-crypt` on Debian/Ubuntu).
    2.  Initialize `git-crypt` in your repository: `git crypt init`.
    3.  Add GPG keys or a symmetric key for decryption. For GPG: `git crypt add-gpg-user <YOUR_GPG_KEY_ID>`. For a symmetric key: `git crypt export-key /path/to/key`.
    4.  Create a `.gitattributes` file (if it doesn't exist) and add `path/to/.creds.yml filter=git-crypt diff=git-crypt` to it.
    5.  Commit `.gitattributes`.
    6.  Encrypt `.creds.yml`: `echo "mysecret: value"` > `.creds.yml` followed by `git add .creds.yml && git commit -m "Add encrypted credentials"`.
    7.  To decrypt, ensure your GPG key is available or the symmetric key is exported and run `git crypt unlock`. The `deploy.sh` script will attempt to unlock automatically if `git-crypt` is set up.
- **Certificates:** Managed via Traefik's ACME (Let's Encrypt) resolver named `le`.

---

## 5. Service Architecture

- **mywiki:** Node.js based TiddlyWiki instance (the "source of truth").
- **blog:** Static site served by **Caddy**. 
  - Content is generated from `mywiki` by an init-container or the `update-blog` CronJob.
  - Static files are stored in `blog-pvc`.
- **pixelfed-importer:** Go-based CronJob that fetches Atom feeds and updates the wiki.
- **Backups:** Managed via **restic** (CronJobs `wiki-backup` and `backup-forget`) targeting Scaleway S3.

---

## 6. Best Practices for Agents

1. **Context Awareness:** Before modifying a service, check its local `package.json` or `go.mod`.
2. **K8s Verification:** After deployment, verify pods and logs:
   ```bash
   kubectl get pods -l kapp.k14s.io/app=<app-id>
   kubectl logs -l app=<name> --tail=50
   ```
3. **Deployment:** Use `./deploy.sh` to apply changes to the cluster.

---

## 7. Cursor & Copilot Rules

- **Cursor:** Follow general project patterns.
- **Copilot:** Follow general project patterns.

---

*This file is generated for agentic use. Maintain this standard when suggesting changes.*
