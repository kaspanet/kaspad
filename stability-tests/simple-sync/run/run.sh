#!/bin/bash
set -e
simple-sync --simnet -n=1000 --profile=7000
TEST_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ]; then
  echo "simple-sync test: PASSED"
  exit 0
fi
echo "simple-sync test: FAILED"
exit 1
