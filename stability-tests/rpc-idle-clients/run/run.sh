#!/bin/bash
rm -rf /tmp/kaspad-temp

NUM_CLIENTS=128
kaspad --devnet --appdir=/tmp/kaspad-temp --profile=6061 --rpcmaxwebsockets=$NUM_CLIENTS &
KASPAD_PID=$!
KASPAD_KILLED=0
function killKaspadIfNotKilled() {
  if [ $KASPAD_KILLED -eq 0 ]; then
    kill $KASPAD_PID
  fi
}
trap "killKaspadIfNotKilled" EXIT

sleep 1

rpc-idle-clients --devnet --profile=7000 -n=$NUM_CLIENTS
TEST_EXIT_CODE=$?

kill $KASPAD_PID

wait $KASPAD_PID
KASPAD_EXIT_CODE=$?
KASPAD_KILLED=1

echo "Exit code: $TEST_EXIT_CODE"
echo "Kaspad exit code: $KASPAD_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $KASPAD_EXIT_CODE -eq 0 ]; then
  echo "rpc-idle-clients test: PASSED"
  exit 0
fi
echo "rpc-idle-clients test: FAILED"
exit 1
