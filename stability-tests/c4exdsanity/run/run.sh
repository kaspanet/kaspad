#!/bin/bash
c4exdsanity --command-list-file ./commands-list --profile=7000
TEST_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ]; then
  echo "c4exdsanity test: PASSED"
  exit 0
fi
echo "c4exdsanity test: FAILED"
exit 1
