#!/bin/sh

# This is a short script that compiles the code inside a docker container, and runs two instances connected to each other
# docker defenition is in docker/Dockerfile.dev
# instances defenition is in docker/docker-compose.yml

export SERVICE_NAME=btcd
export GIT_COMMIT=$(git rev-parse --short=12 HEAD)

docker build -t "${SERVICE_NAME}:${GIT_COMMIT}" . \
  -f docker/Dockerfile.dev \
  || fatal 'Failed to build the docker image'
docker tag "${SERVICE_NAME}:${GIT_COMMIT}" "${SERVICE_NAME}:latest"

cd docker

docker-compose up
