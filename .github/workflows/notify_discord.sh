#!/bin/sh

echo "${EVENT}"

JOB_TYPE="commit"
if [ "${EVENT_NAME}" = "pull_request" ]; then
  JOB_TYPE="pull request"
fi

MESSAGE="**${MESSAGE_TITLE}**:
Job **FAILED** in ${REPOSITORY_NAME} for ${JOB_TYPE} authored by ${COMMITER}"

echo "${MESSAGE}"
