#!/bin/bash

APPDIR=/tmp/kaspad-temp
KASPAD_RPC_PORT=29587

rm -rf "${APPDIR}"

kaspad --simnet --appdir="${APPDIR}" --rpclisten=0.0.0.0:"${KASPAD_RPC_PORT}" --profile=6061 &
KASPAD_PID=$!

sleep 1

RUN_STABILITY_TESTS=true go test ../ -v -- --rpc-address=127.0.0.1:"${KASPAD_RPC_PORT}" --profile=7000
TEST_EXIT_CODE=$?

kill $KASPAD_PID

wait $KASPAD_PID
KASPAD_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"
echo "Kaspad exit code: $KASPAD_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $KASPAD_EXIT_CODE -eq 0 ]; then
  echo "mempool-limits test: PASSED"
  exit 0
fi
echo "mempool-limits test: FAILED"
exit 1
