#!/bin/sh

echo "${GITHUB_JSON}"
echo "aaa"

EVENT_NAME=$(jq .event_name <<< "${GITHUB_JSON}")
echo ${EVENT_NAME}

MESSAGE="**${MESSAGE_TITLE}**:"
echo "${MESSAGE}"
