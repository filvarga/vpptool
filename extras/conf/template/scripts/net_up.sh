#!/bin/bash
./scripts/net_down.sh

set -x
# veth pair 1
ip link add host0 type veth peer name vpp0
ip link set dev host0 up
ip addr add 20.0.0.1/24 dev host0
# veth pair 2
ip link add host1 type veth peer name vpp1
ip link set dev host1 up
ip addr add 20.0.1.1/24 dev host1
