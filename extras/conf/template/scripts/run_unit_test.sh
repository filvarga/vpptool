#!/bin/bash
docker cp ./tests/test_unit.py vpp-run:/opt/vpp/test
docker exec -it vpp-run make test-debug TEST=test_unit
