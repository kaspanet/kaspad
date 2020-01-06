#!/bin/sh

# This file is part of Continuous Integration. When ran by
# the CI agent, it sends a some details about the build failure
# to a Discord channel.

echo "HELLO I AM DISCORD DOT SH" > kaka.txt

CLIENT_ID="$1"
API_TOKEN="$2"
BUILD_URL="$3"
PR_AUTHOR="$4"
PR_TITLE="$5"
PR_LINK="$6"

# Build the failure message
MESSAGE="*${PR_AUTHOR}*:
Build *FAILED* for pull request '${PR_TITLE}'
[Github](${PR_LINK})        [Jenkins](${BUILD_URL}console)"

# Retrieve the build log
LOG=$(curl ${BUILD_URL}consoleText)

# Send the build log
printf "$LOG" | curl \
  "https://discordapp.com/api/webhooks/${CLIENT_ID}/${API_TOKEN}" \
  -F content="${MESSAGE}" \
  -F document="@-;filename=build.log"