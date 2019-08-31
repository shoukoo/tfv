#!/bin/bash

set -exuo pipefail

export GO111MODULE=on

EVENT_DATA=$(cat $GITHUB_EVENT_PATH)
RELEASE_NAME=$(echo $EVENT_DATA | jq -r .release.tag_name)
UPLOAD_URL=$(echo $EVENT_DATA | jq -r .release.upload_url)
UPLOAD_URL=${UPLOAD_URL/\{?name,label\}/}


LINUX_BIN="tfv-linux-amd64-${RELEASE_NAME}"
DARWIN_BIN="tfv-darwin-amd64-${RELEASE_NAME}"
WINDOWS_BIN="tfv-windows-amd64-${RELEASE_NAME}.exe"

GOOS=linux GOARCH=amd64 go build -o $LINUX_BIN
GOOS=darwin GOARCH=amd64 go build -o $DARWIN_BIN
GOOS=windows GOARCH=amd64 go build -o $WINDOWS_BIN

for i in "$LINUX_BIN $DARWIN_BIN $WINDOWS_BIN"; do
  curl \
    -X POST \
    --data-binary @${i}\
    -H 'Content-Type: application/octet-stream' \
    -H "Authorization: Bearer ${GITHUB_TOKEN}" \
    "${UPLOAD_URL}?name=${i}"
done
