#!/bin/bash
set -e
KASPAD_VERSION=$1
SLOW=$2

# Prune dockers so that we don't run out of space while running tests
docker system prune

$(aws ecr get-login --region=eu-central-1 --no-include-email) &&
  docker build -f docker/Dockerfile . --build-arg "KASPAD_VERSION=${KASPAD_VERSION}" \
    --build-arg "KASPAROV_VERSION=${KASPAROV_VERSION}" -t stability-tests

docker run -p 7000:7000 -p 6061:6061 -p 6062:6062 stability-tests:latest $SLOW
