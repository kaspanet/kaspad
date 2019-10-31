#!/bin/bash

# This is a short script that compiles the code inside a docker container, and runs two instances connected to each other
# docker defenition is in docker/Dockerfile.dev
# instances defenition is in docker/docker-compose.yml

set -e

if [[ $* == *--help* ]]
then
	echo "Usage:"
	echo -e "\t./run-dev.sh [--rm] [--debug]"
	echo ""
	echo -e "\t--rm\t\tRemove dockers prior to running them, to clear data"
	echo -e "\t--debug\t\tEnable debugging on second server. Server will not start until debugger is attached"
	echo -e "\t--no-build\t\tRun without building docker images"
	echo -e "\t--no-run\t\tBuild docker images without running"
	echo -e "\t--cleanup\t\tAlias for --rm --no-run --no-build. Overrides all other flags"
	exit
fi

if [[ $* == *--cleanup* ]]
then
  set -- "--rm --no-run --no-build"
fi

export SERVICE_NAME=btcd
export GIT_COMMIT=$(git rev-parse --short=12 HEAD)

if [[ $* != *--no-build* ]]
then
  docker build -t "${SERVICE_NAME}:${GIT_COMMIT}" . \
    -f docker/Dockerfile.dev \
    || fatal 'Failed to build the docker image'
  docker tag "${SERVICE_NAME}:${GIT_COMMIT}" "${SERVICE_NAME}:latest"
fi

cd docker

if [[ $* == *--rm* ]]
then
  docker-compose rm -f -s -v
fi

if [[ $* != *--no-run* ]]
then
  if [[ $* == *--debug* ]]
  then
    docker-compose up first second-debug
  else
    docker-compose up first second
  fi
fi
