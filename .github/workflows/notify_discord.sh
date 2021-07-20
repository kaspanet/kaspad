#!/bin/sh

ACTOR_NAME=$(jq -r .actor <<< "${GITHUB_JSON}")
JOB_NAME=$(jq -r .job <<< "${GITHUB_JSON}")
REPOSITORY_NAME=$(jq -r .repository <<< "${GITHUB_JSON}")
EVENT_NAME=$(jq -r .event_name <<< "${GITHUB_JSON}")

MESSAGE="**${ACTOR_NAME}**:
Job '${JOB_NAME}' failed in ${REPOSITORY_NAME}"

if [ "${EVENT_NAME}" = "pull_request" ]; then
  PULL_REQUEST_TITLE=$(jq -r .event.pull_request.title <<< "${GITHUB_JSON}")
  PULL_REQUEST_URL=$(jq -r .event.pull_request._links.html.href <<< "${GITHUB_JSON}")
  RUN_ID=$(jq -r .run_id <<< "${GITHUB_JSON}")
  RUN_URL="${PULL_REQUEST_URL}/checks?check_run_id=${RUN_ID}"

  MESSAGE="${MESSAGE} for pull request '${PULL_REQUEST_TITLE}'
  [Pull Request](${PULL_REQUEST_URL})  [JOB](${RUN_URL})"
fi

echo "${MESSAGE}"
