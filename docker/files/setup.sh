#!/bin/bash

set -eux

git pull
git checkout $CID -b build

UNATTENDED=y apt-get update && apt-get upgrade -y
UNATTENDED=y make install-dep
UNATTENDED=y make install-ext-deps
rm -rf /var/lib/apt/lists
