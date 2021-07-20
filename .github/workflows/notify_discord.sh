#!/bin/sh

MESSAGE="**${MESSAGE_TITLE}**:
Job **FAILED** in ${REPOSITORY_NAME} for branch '${BRANCH_NAME}' authored by ${COMMITER}"

echo "${MESSAGE}"
