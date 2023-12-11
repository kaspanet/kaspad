#!/bin/bash

RUN_STABILITY_TESTS=true go test ../ -v -timeout 86400s
TEST_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ]; then
  echo "daa test: PASSED"
  exit 0
fi
echo "daa test: FAILED"
exit 1
