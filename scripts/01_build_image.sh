#!/bin/bash
echo "## will test for ENV variable APP_NAME"
# checks whether APP_NAME has length equal to zero:
if [[ -z "${APP_NAME}" ]]
then
	echo "## ENV variable APP_NAME not found"
      	FILE=getAppInfo.sh
	if test -f "$FILE"; then
		echo "## will execute $FILE"
		# shellcheck disable=SC1090
		source $FILE
	elif test -f "./scripts/${FILE}"; then
		echo "## will execute ./scripts/$FILE"
  		# shellcheck disable=SC1090
  		source ./scripts/$FILE
	else
    		echo "-- ERROR getAppInfo.sh was not found"
		exit 1
	fi
else
	echo "## ENV variable APP_NAME found : ${APP_NAME}"
fi
DOCKER_BIN=docker
echo"## using nerdctl to build image on linux : https://docs.rancherdesktop.io/images ready to be used"
DOCKER_BIN="nerdctl -n k8s.io"
DOCKER_REGISTRY_ID=laotseu
echo "APP: ${APP_NAME}, version: ${APP_VERSION} detected in file server.go"
TMP_Docker_Dir=$(mktemp -d)
cp Dockerfile* "$TMP_Docker_Dir"
cd "$TMP_Docker_Dir" || exit
if trivy config --exit-code 1 --severity MEDIUM,HIGH,CRITICAL . ;
#if [ $? -eq 0 ]
then
  echo "Cool no vulnerabilities found in your Dockerfile will change directory :$OLDPWD"
  cd "$OLDPWD" || exit
  rm -rf "$TMP_Docker_Dir" # cleanup
  echo "will parse the multi-stage Dockerfile in the current directory and build the final image"
  ${DOCKER_BIN} build -t ${DOCKER_REGISTRY_ID}/"${APP_NAME}" .
  echo "will tag this image with version ${APP_VERSION}"
  ${DOCKER_BIN} tag ${DOCKER_REGISTRY_ID}/"${APP_NAME}" ${DOCKER_REGISTRY_ID}/"${APP_NAME}":"${APP_VERSION}"
  echo "listing all images containing : ${APP_NAME}"
  ${DOCKER_BIN} images | grep "${APP_NAME}"
  echo "to try your container image locally :  ${DOCKER_BIN} run -p 8080:8080 ${DOCKER_REGISTRY_ID}/${APP_NAME}"
  echo "to deploy your container image to docker hub :  ${DOCKER_BIN} push ${DOCKER_REGISTRY_ID}/${APP_NAME}"
  echo "to latter remove the images :  ${DOCKER_BIN} rmi ${DOCKER_REGISTRY_ID}/${APP_NAME}"
else
  echo "You must correct the MEDIUM,HIGH,CRITICAL vulnerabilities detected by Trivy, before building your DockerFile" >&2
fi

