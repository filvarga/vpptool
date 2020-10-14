#!/bin/bash
docker cp test_api.py vpp-run:/tmp
docker exec -it vpp-run /scripts/test-api /tmp/test_api.py
