#!/bin/sh

echo "${JOB_JSON}"
echo "aaaaa"

ACTOR_NAME=$(jq -r .actor <<< "${GITHUB_JSON}")
JOB_NAME=$(jq -r .job <<< "${GITHUB_JSON}")
REPOSITORY_NAME=$(jq -r .repository <<< "${GITHUB_JSON}")
EVENT_NAME=$(jq -r .event_name <<< "${GITHUB_JSON}")

MESSAGE="**${ACTOR_NAME}**:
Job '${JOB_NAME}' failed in ${REPOSITORY_NAME}"

if [ "${EVENT_NAME}" = "pull_request" ]; then
  PULL_REQUEST_TITLE=$(jq -r .event.pull_request.title <<< "${GITHUB_JSON}")
  MESSAGE="${MESSAGE} for pull request '${PULL_REQUEST_TITLE}'"
fi

echo "${MESSAGE}"
