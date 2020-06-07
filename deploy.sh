#! /bin/bash

set -exuo pipefail

lpass sync

lpass show "deployments/k3s/tiddlywiki" --notes \
  | ytt -f spec/  -f - \
    | kbld -f - 2> /dev/null \
    | kapp deploy --app tiddlywiki  \
           -f - \
           --diff-changes \
           --yes

