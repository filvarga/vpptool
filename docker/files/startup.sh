#!/bin/sh
cfg="/etc/startup.conf"
vpp="/opt/vpp/build-root/install-vpp_debug-native/vpp/bin/vpp"

rm -rf /run/vpp/cli.sock &> /dev/null
until $vpp -c $cfg; do
  rm -rf /run/vpp/cli.sock &> /dev/null
  echo "Restarting 'vpp'"
  sleep 1
done
