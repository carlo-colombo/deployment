#! /bin/bash

set -euo pipefail

function random_string {
  head -c 50 /dev/urandom | base32 --wrap=0
}

: "${BITWARDEN_CREDS_ITEM_ID:?Set BITWARDEN_CREDS_ITEM_ID to the Bitwarden Secure Note item ID}"
: "${BW_SESSION:?Unlock Bitwarden first and export BW_SESSION}"
command -v bw >/dev/null || { echo "Bitwarden CLI (bw) is required." >&2; exit 1; }
command -v jq >/dev/null || { echo "jq is required to read the Bitwarden note." >&2; exit 1; }

bw get item "$BITWARDEN_CREDS_ITEM_ID" \
  | jq -er 'if .type == 2 and (.notes | type == "string" and length > 0) then .notes else error("Bitwarden item must be a Secure Note with YAML in its Notes field") end' \
  | ytt -f spec/ -f - \
      --data-value release.cookie="$(random_string)" \
      --data-value release.secret_key_base="$(random_string)" \
  | kbld -f - \
      --lock-output "images.lock" \
  | kapp deploy --app tiddlywiki \
      -f - \
      --diff-changes \
      --yes
