#!/bin/bash
set -e
many-tips --devnet -n=1000 --profile=7000
TEST_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ]; then
  echo "many-tips test: PASSED"
  exit 0
fi
echo "many-tips test: FAILED"
exit 1
