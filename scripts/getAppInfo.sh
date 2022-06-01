#!/bin/bash
echo "will extract app name and version from source"
APP_NAME=$(grep -E 'APP\s+=' server.go| awk '{ print $3 }'  | tr -d '"')
APP_VERSION=$(grep -E 'VERSION\s+=' server.go| awk '{ print $3 }'  | tr -d '"')
echo "APP: ${APP_NAME}, VERSION: ${APP_VERSION} detected in file server.go"
export APP_VERSION APP_NAME
