#!/bin/bash
set -e
TEST_EXIT_CODE=1
BASEDIR=$(dirname "$0")
if [[ $1 == "slow" ]];
 then
    echo "Running slow stability tests"
    "${BASEDIR}/run-slow.sh"
    TEST_EXIT_CODE=$?
    echo "Done running slow stability tests"
  else
    echo "Running fast stability tests"
    "${BASEDIR}/run-fast.sh"
    TEST_EXIT_CODE=$?
    echo "Done running fast stability tests"
fi

echo "Exit code: $TEST_EXIT_CODE"
exit $TEST_EXIT_CODE