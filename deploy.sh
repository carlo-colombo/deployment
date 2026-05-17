#! /bin/bash

set -exuo pipefail

function random_string {
  base32 --wrap=0 /dev/urandom | head -c 80
}


function unlock_credentials {
  local key_file="/tmp/downloads/deployment.git-crypt-key"
  if command -v git-crypt &> /dev/null; then
    if git-crypt status &> /dev/null; then
      if [ -f "$key_file" ]; then
        echo "git-crypt detected. Attempting to unlock credentials using key file: $key_file"
        git-crypt unlock "$key_file" || { echo "Failed to unlock credentials with key file."; exit 1; }
      else
        echo "git-crypt detected. Attempting to unlock credentials using default method (environment or local key)..."
        git-crypt unlock || { echo "Failed to unlock credentials. Ensure GIT_CRYPT_KEY is set or key file is present."; exit 1; }
      fi
    else
      echo "git-crypt detected but not initialized in this repository. Skipping unlock."
    fi
  else
    echo "git-crypt not found. Skipping credential unlock. Please install git-crypt if .creds.yml is encrypted."
  fi
}

unlock_credentials

cat .creds.yml \
    | ytt -f spec/  -f - \
    --data-value release.cookie="$(random_string)" \
    --data-value release.secret_key_base="$(random_string)" \
    | kbld -f - \
        --lock-output "images.lock"  \
    | kapp deploy --app tiddlywiki  \
    -f - \
    --diff-changes \
    --yes
