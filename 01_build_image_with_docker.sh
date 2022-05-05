#!/bin/bash
echo "will extract app name and version from source"
VERSION=`grep -E 'VERSION\s+=' server.go| awk '{ print $3 }'  | tr -d '"'`
APPNAME=`grep -E 'APP\s+=' server.go| awk '{ print $3 }'  | tr -d '"'`
DOCKER_REGISTRY_ID=laotseu
echo "APP: ${APPNAME}, version: ${VERSION} detected in file server.go"
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
  docker build -t ${DOCKER_REGISTRY_ID}/${APPNAME} .
  echo "will tag this image with version ${VERSION}"
  docker tag ${DOCKER_REGISTRY_ID}/${APPNAME} ${DOCKER_REGISTRY_ID}/${APPNAME}:${VERSION}
  echo "listing all images containing : ${APPNAME}"
  docker images | grep ${APPNAME}
  echo "to try your docker image localy :  docker run -p 8080:8080 ${DOCKER_REGISTRY_ID}/${APPNAME}"
  echo "to deploy your docker image to docker hub :  docker push ${DOCKER_REGISTRY_ID}/${APPNAME}"
  echo "to latter remove the images :  docker rmi ${DOCKER_REGISTRY_ID}/${APPNAME}"
else
  echo "You must correct the MEDIUM,HIGH,CRITICAL vulnerabilities detected by Trivy, before building your DockerFile" >&2
fi


