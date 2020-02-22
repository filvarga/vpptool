#!/bin/bash
set -eux
git pull
git checkout $CID -b build

UNATTENDED=y apt-get update && apt-get upgrade -y
# INFO: missing dependencie in make install-dep !!!
# required by install-ext-deps
UNATTENDED=y apt install libssl-dev -y
#
UNATTENDED=y make install-dep
UNATTENDED=y make install-ext-deps
rm -rf /var/lib/apt/lists
