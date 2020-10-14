#!/bin/bash

threads=1

# rate-multipleier != cps (if custom.py not used)
usage ()
{
  printf "$0 usage: <duration> <rate-multiplier>\n"
  exit 1
}

if [ "$#" -ne 2 ]; then
  usage
fi

d=$1
m=$2

# -d Duration of test in sec (default is 3600).
# -m Rate multiplier. Multiply basic rate of templates by this number.
# cps - connections per second (v astf/yaml) - multiplier multiplies cps by m
# cps * m = (number of coonections)
# -c Number of hardware threads to allocate for each port pair.
# -k Run 'warm up' traffic for num before starting the test.
# cache reuse before sending packets (sends packets during cache but doesn't measure)
#  - if we need exact metrics we can't use this !!
# --astf  Enable advanced stateful mode. (profile should be in py format)

docker cp ./tests/trex_test.pcap trex-run:/opt/trex-core/scripts/avl
docker cp ./tests/trex_test.py trex-run:/opt/trex-core/scripts/astf
docker cp ./conf/trex_cfg.yaml trex-run:/etc

read -n 1 -p "Hit key..."
./scripts/t-rex-64 -f astf/trex_test.py -d $d -m $m -c $threads \
 --astf --cfg /etc/trex_cfg.yaml
