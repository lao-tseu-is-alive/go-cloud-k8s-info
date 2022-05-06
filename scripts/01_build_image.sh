#!/bin/bash
echo "will extract app name and version from source"
VERSION=`grep -E 'VERSION\s+=' server.go| awk '{ print $3 }'  | tr -d '"'`
APP_NAME=`grep -E 'APP\s+=' server.go| awk '{ print $3 }'  | tr -d '"'`
DOCKER_BIN=docker
# using nerdctl to build image on linux : https://docs.rancherdesktop.io/images ready to be used
DOCKER_BIN="nerdctl -n k8s.io"
DOCKER_REGISTRY_ID=laotseu
echo "APP: ${APP_NAME}, version: ${VERSION} detected in file server.go"
TMP_Docker_Dir=$(mktemp -d)
cp Dockerfile* $TMP_Docker_Dir
cd $TMP_Docker_Dir
trivy config --exit-code 1 --severity MEDIUM,HIGH,CRITICAL .
if [ $? -eq 0 ]
then
  echo "Cool no vulnerabilities found in your Dockerfile"
  cd "$OLDPWD"
  rm -rf $TMP_Docker_Dir # cleanup
  echo "will parse the multi-stage Dockerfile in the current directory and build the final image"
  ${DOCKER_BIN} build -t ${DOCKER_REGISTRY_ID}/${APP_NAME} .
  echo "will tag this image with version ${VERSION}"
  ${DOCKER_BIN} tag ${DOCKER_REGISTRY_ID}/${APP_NAME} ${DOCKER_REGISTRY_ID}/${APP_NAME}:${VERSION}
  echo "listing all images containing : ${APP_NAME}"
  ${DOCKER_BIN} images | grep ${APP_NAME}
  echo "to try your container image localy :  ${DOCKER_BIN} run -p 8080:8080 ${DOCKER_REGISTRY_ID}/${APP_NAME}"
  echo "to deploy your container image to docker hub :  ${DOCKER_BIN} push ${DOCKER_REGISTRY_ID}/${APP_NAME}"
  echo "to latter remove the images :  ${DOCKER_BIN} rmi ${DOCKER_REGISTRY_ID}/${APP_NAME}"
else
  echo "You must correct the MEDIUM,HIGH,CRITICAL vulnerabilities detected by Trivy, before building your DockerFile" >&2
fi


