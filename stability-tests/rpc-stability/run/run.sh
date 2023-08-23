#!/bin/bash
rm -rf /tmp/c4exd-temp

c4exd --devnet --appdir=/tmp/c4exd-temp --profile=6061 --loglevel=debug &
C4exD_PID=$!

sleep 1

rpc-stability --devnet -p commands.json --profile=7000
TEST_EXIT_CODE=$?

kill $C4exD_PID

wait $C4exD_PID
C4exD_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"
echo "C4exd exit code: $C4exD_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $C4exD_EXIT_CODE -eq 0 ]; then
  echo "rpc-stability test: PASSED"
  exit 0
fi
echo "rpc-stability test: FAILED"
exit 1
