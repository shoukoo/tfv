#!/bin/bash

set -exuo pipefail

export GO111MODULE=on

EVENT_DATA=$(cat $GITHUB_EVENT_PATH)
RELEASE_NAME=$(echo $EVENT_DATA | jq -r .release.tag_name)
UPLOAD_URL=$(echo $EVENT_DATA | jq -r .release.upload_url)
UPLOAD_URL=${UPLOAD_URL/\{?name,label\}/}

GOOS=linux GOARCH=amd64 go build -o tfv-linux-amd64-${RELEASE_NAME}
GOOS=darwin GOARCH=amd64 go build -o tfv-darwin-amd64-${RELEASE_NAME}
GOOS=windows GOARCH=amd64 go build -o tfv-windows-amd64-${RELEASE_NAME}.exe

ls -al
