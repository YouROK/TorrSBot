#!/bin/bash

PLATFORMS_ARM="linux"

type setopt >/dev/null 2>&1

SCRIPT_NAME=`basename "$0"`
FAILURES=""
SOURCE_FILE="dist/torrsbot"
CURRENT_DIRECTORY=${PWD##*/}
OUTPUT=${SOURCE_FILE:-$CURRENT_DIRECTORY}
LDFLAGS="-s -w"

#Clean
go clean -i -r -cache

GOARCH="amd64"
GOOS="linux"

BIN_FILENAME="${OUTPUT}-${GOOS}-${GOARCH}${GOARM}"
CMD="GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags='${LDFLAGS}' -o ${OUTPUT}_${GOOS}_${GOARCH} ./cmd"
echo "${CMD}"
eval "${CMD}" || FAILURES="${FAILURES} ${GOOS}/${GOARCH}${GOARM}"

# eval errors
if [[ "${FAILURES}" != "" ]]; then
  echo ""
  echo "${SCRIPT_NAME} failed on: ${FAILURES}"
  exit 1
fi
