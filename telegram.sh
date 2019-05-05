#!/bin/sh

echo "hello world!!!" > aaa.txt

API_TOKEN=$1
CHAT_ID=$2
BUILD_URL=$3
PR_AUTHOR=$4
PR_TITLE=$5
PR_LINK=$6

# Start atd
service atd start

echo "hello!!! ${API_TOKEN} ${CHAT_ID} ${BUILD_URL} ${PR_AUTHOR} ${PR_TITLE} ${PR_LINK}" > aaab.txt

# Build the failure message
MESSAGE="*${PR_AUTHOR}*:
Build *FAILED* for pull request '${PR_TITLE}'
[Github](${PR_LINK})        [Jenkins](${BUILD_URL}console)"

echo "b" >> aaa.txt

# Send the failure message
curl -s \
  -X POST \
  "https://api.telegram.org/bot${API_TOKEN}/sendMessage" \
  -d chat_id="${CHAT_ID}" \
  -d parse_mode=markdown \
  -d disable_web_page_preview=true \
  -d text="${MESSAGE}"

echo "C" >> aaa.txt

# Retrieve the build log
LOG=$(curl ${BUILD_URL}consoleText)

echo "d" >> aaa.txt

# Send the build log
printf "$LOG" | curl \
  "https://api.telegram.org/bot${API_TOKEN}/sendDocument" \
  -F chat_id="${CHAT_ID}" \
  -F document="@-;filename=build.log"

echo "e" >> aaa.txt