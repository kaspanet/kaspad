#!/bin/bash
kaspadsanity --command-list-file ./commands-list --profile=7000
TEST_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ]; then
  echo "kaspadsanity test: PASSED"
  exit 0
fi
echo "kaspadsanity test: FAILED"
exit 1
