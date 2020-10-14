#!/bin/bash

set -x
ip link del dev host0 &> /dev/null
ip link del dev host1 &> /dev/null
