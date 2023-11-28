#! /bin/bash

set -exuo pipefail

function random_string {
  base32 --wrap=0 /dev/urandom | head -c 80
}

lpass sync

lpass show "deployments/k3s/tiddlywiki" --notes \
    | ytt -f spec/  -f - \
    --data-value release.cookie="$(random_string)" \
    --data-value release.secret_key_base="$(random_string)" \
    | kbld -f - \
        --lock-output "images.lock"  \
    | kapp deploy --app tiddlywiki  \
    -f - \
    --diff-changes \
    --yes
