#!/bin/bash
set -e

BASEDIR=$(dirname "$0")
PROJECT_ROOT=$( cd "${BASEDIR}/.."; pwd)

failedTests=()

# echo "Running application-level-garbage"
# cd "${PROJECT_ROOT}/application-level-garbage/run" && ./run.sh || failedTests+=("application-level-garbage")
# echo "Done running application-level-garbage"

echo "Running infra-level-garbage"
cd "${PROJECT_ROOT}/infra-level-garbage/run" && ./run.sh || failedTests+=("infra-level-garbage")
echo "Done running infra-level-garbage"

echo "Running kaspadsanity"
cd "${PROJECT_ROOT}/kaspadsanity/run" && ./run.sh || failedTests+=("kaspadsanity")
echo "Done running kaspadsanity"

echo "Running rpc-stability"
cd "${PROJECT_ROOT}/rpc-stability/run" && ./run.sh || failedTests+=("rpc-stability")
echo "Done running rpc-stability"

echo "Running rpc-idle-clients"
cd "${PROJECT_ROOT}/rpc-idle-clients/run" && ./run.sh || failedTests+=("rpc-idle-clients")
echo "Done running rpc-idle-clients"

echo "Running simple-sync"
cd "${PROJECT_ROOT}/simple-sync/run" && ./run.sh || failedTests+=("simple-sync")
echo "Done running simple-sync"

echo "Running orphans"
cd "${PROJECT_ROOT}/orphans/run" && ./run.sh || failedTests+=("orphans")
echo "Done running orphans"

echo "Running reorg"
cd "${PROJECT_ROOT}/reorg/run" && ./run-full-finality-window-reorg.sh || failedTests+=("reorg")
echo "Done running reorg"

echo "Running mempool-limits"
cd "${PROJECT_ROOT}/mempool-limits/run" && ./run.sh || failedTests+=("mempool-limits")
echo "Done running mempool-limits"

echo "Running netsync - slow"
cd ${PROJECT_ROOT}/netsync/run"" && ./run.sh || failedTests+=("netsync")
echo "Done running netsync - slow"

EXIT_CODE=0
for t in "${failedTests[@]}"; do
  EXIT_CODE=1
  echo "FAILED: ${t}"
done

echo "Exiting with: ${EXIT_CODE}"
exit $EXIT_CODE