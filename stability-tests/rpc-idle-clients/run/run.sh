#!/bin/bash
rm -rf /tmp/c4exd-temp

NUM_CLIENTS=128
c4exd --devnet --appdir=/tmp/c4exd-temp --profile=6061 --rpcmaxwebsockets=$NUM_CLIENTS &
C4exD_PID=$!
C4exD_KILLED=0
function killC4exdIfNotKilled() {
  if [ $C4exD_KILLED -eq 0 ]; then
    kill $C4exD_PID
  fi
}
trap "killC4exdIfNotKilled" EXIT

sleep 1

rpc-idle-clients --devnet --profile=7000 -n=$NUM_CLIENTS
TEST_EXIT_CODE=$?

kill $C4exD_PID

wait $C4exD_PID
C4exD_EXIT_CODE=$?
C4exD_KILLED=1

echo "Exit code: $TEST_EXIT_CODE"
echo "C4exd exit code: $C4exD_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $C4exD_EXIT_CODE -eq 0 ]; then
  echo "rpc-idle-clients test: PASSED"
  exit 0
fi
echo "rpc-idle-clients test: FAILED"
exit 1
