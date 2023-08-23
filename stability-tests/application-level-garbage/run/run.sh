#!/bin/bash
rm -rf /tmp/c4exd-temp

c4exd --devnet --appdir=/tmp/c4exd-temp --profile=6061 --loglevel=debug &
C4exD_PID=$!
C4exD_KILLED=0
function killC4exdIfNotKilled() {
    if [ $C4exD_KILLED -eq 0 ]; then
      kill $C4exD_PID
    fi
}
trap "killC4exdIfNotKilled" EXIT

sleep 1

application-level-garbage --devnet -alocalhost:16611 -b blocks.dat --profile=7000
TEST_EXIT_CODE=$?

kill $C4exD_PID

wait $C4exD_PID
C4exD_KILLED=1
C4exD_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"
echo "C4exd exit code: $C4exD_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $C4exD_EXIT_CODE -eq 0 ]; then
  echo "application-level-garbage test: PASSED"
  exit 0
fi
echo "application-level-garbage test: FAILED"
exit 1
