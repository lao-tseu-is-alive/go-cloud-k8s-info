#!/bin/bash
echo "will extract app name and version from source"
VERSION=`grep -E 'VERSION\s+=' server.go| awk '{ print $3 }'  | tr -d '"'`
APPNAME=`grep -E 'APP\s+=' server.go| awk '{ print $3 }'  | tr -d '"'`
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
  # using nerdctl to build image on linux : https://docs.rancherdesktop.io/images ready to be used
  echo "will parse the multi-stage Dockerfile in the current directory and build the final image"
  nerdctl -n k8s.io build -t ${APPNAME} .
  echo "will tag this image with version ${VERSION}"
  nerdctl -n k8s.io tag ${APPNAME} ${APPNAME}:${VERSION}
  echo "listing all images containing : ${APPNAME}"
  nerdctl -n k8s.io images | grep ${APPNAME}
  echo "to latter remove the images :  nerdctl -n k8s.io rmi ${APPNAME}"
else
  echo "You must correct the MEDIUM,HIGH,CRITICAL vulnerabilities detected by Trivy, before building your DockerFile" >&2
fi


