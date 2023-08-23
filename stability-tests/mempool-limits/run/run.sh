#!/bin/bash

APPDIR=/tmp/c4exd-temp
C4exD_RPC_PORT=29587

rm -rf "${APPDIR}"

c4exd --simnet --appdir="${APPDIR}" --rpclisten=0.0.0.0:"${C4exD_RPC_PORT}" --profile=6061 &
C4exD_PID=$!

sleep 1

RUN_STABILITY_TESTS=true go test ../ -v -timeout 86400s -- --rpc-address=127.0.0.1:"${C4exD_RPC_PORT}" --profile=7000
TEST_EXIT_CODE=$?

kill $C4exD_PID

wait $C4exD_PID
C4exD_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"
echo "C4exd exit code: $C4exD_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $C4exD_EXIT_CODE -eq 0 ]; then
  echo "mempool-limits test: PASSED"
  exit 0
fi
echo "mempool-limits test: FAILED"
exit 1
