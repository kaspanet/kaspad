#!/bin/sh

echo "${GITHUB_JSON}"
echo "aaa"

ACTOR_NAME=$(jq -r .actor <<< "${GITHUB_JSON}")
echo ${ACTOR_NAME}

JOB_NAME=$(jq -r .job <<< "${GITHUB_JSON}")
echo ${JOB_NAME}

REPOSITORY_NAME=$(jq -r .repository <<< "${GITHUB_JSON}")
echo ${REPOSITORY_NAME}

MESSAGE="**${ACTOR_NAME}**:
Job '${JOB_NAME}' failed in ${REPOSITORY_NAME}"

EVENT_NAME=$(jq -r .event_name <<< "${GITHUB_JSON}")
echo ${EVENT_NAME}

if [ "${EVENT_NAME}" = "pull_request" ]; then
  PULL_REQUEST_TITLE=$(jq -r .event.pull_request.title <<< "${GITHUB_JSON}")
  echo ${PULL_REQUEST_TITLE}

  MESSAGE="${MESSAGE} for pull request '${PULL_REQUEST_TITLE}'"
fi

echo "${MESSAGE}"
