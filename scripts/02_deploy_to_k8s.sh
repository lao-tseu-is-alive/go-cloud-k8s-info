#!/bin/bash
DOCKER_BIN=docker
## Using nerdctl instead of docker on Linux, check: https://docs.rancherdesktop.io/images it's cool & ready to be used
DOCKER_BIN="nerdctl -n k8s.io"
# you obviously will need to adjust next line to your own favorite value :-)
CONTAINER_REGISTRY_ID=laotseu
K8s_NAMESPACE=dev
DEPLOYMENT_TEMPLATE="scripts/k8s-deployment_template.yml"
K8S_DEPLOYMENT=deployment.yml
echo "## Checking if ENV variable APP_NAME is already defined..."
# checks whether APP_NAME has length equal to zero:
if [[ -z "${APP_NAME}" ]]
then
	echo "## ENV variable APP_NAME not found"
      	FILE=getAppInfo.sh
	if test -f "$FILE"; then
		echo "## Sourcing $FILE"
		# shellcheck disable=SC1090
		source $FILE
	elif test -f "./scripts/${FILE}"; then
		echo "## Sourcing ./scripts/$FILE"
  		# shellcheck disable=SC1090
  	source ./scripts/$FILE
	else
	  echo "## 💥💥 ERROR: getAppInfo.sh was not found"
		exit 1
	fi
else
	echo "## ENV variable APP_NAME is defined to : ${APP_NAME} . So we will use this one !"
fi
echo "## USING APP_NAME: \"${APP_NAME}\", APP_VERSION: \"${APP_VERSION}\""
IMAGE_FILTER="${CONTAINER_REGISTRY_ID}/${APP_NAME}"
echo "## Checking if image:tag was build  in k8s namespace ${IMAGE_FILTER} tag:${APP_VERSION}"
JSON_APP=$(${DOCKER_BIN} images --format '{{json .}}' | jq ".| select(.Repository | contains(\"${IMAGE_FILTER}\")) |select(.Tag | contains(\"${APP_VERSION}\"))")
APP_ID=$(echo "${JSON_APP}" | jq '.|.ID')
# checks whether APP_ID has length equal to zero --> meaning this image:version is not present and was probably not already build
if [[ -z "${APP_ID}" ]]
then
	echo "## 💥💥 ERROR: ${IMAGE_FILTER}:${APP_VERSION} image was not found ! May be you need to build it first ?"
	exit
else
  echo "## OK : \"${IMAGE_FILTER}:${APP_VERSION}\" image was found"
  echo "${JSON_APP}" | jq '.'
fi
echo "## Generating a deployment based on template : ${DEPLOYMENT_TEMPLATE}"
DEPLOYMENT_DIRECTORY="deployments/${K8s_NAMESPACE}"
sed s/APP_NAME/"${APP_NAME}"/g  ${DEPLOYMENT_TEMPLATE} > ${DEPLOYMENT_DIRECTORY}/${K8S_DEPLOYMENT}

sed -i s/APP_VERSION/"${APP_VERSION}"/g  ${DEPLOYMENT_DIRECTORY}/${K8S_DEPLOYMENT}
sed -i s/GO_CONTAINER_REGISTRY_PREFIX/${CONTAINER_REGISTRY_ID}/g  ${DEPLOYMENT_DIRECTORY}/${K8S_DEPLOYMENT}
echo "## Checking result of substitution in image name :"
yq  ".spec.template.spec.containers[0].image" ${DEPLOYMENT_DIRECTORY}/${K8S_DEPLOYMENT}
#yq -i ".spec.template.spec.containers[0].image=\"${IMAGE_FILTER}:${APP_VERSION}\"" deployments/dev/deployment.yml
echo "## Checking for vulnerabilities with trivy"
cd "${DEPLOYMENT_DIRECTORY}" || exit
echo "## Checking for vulnerabilities in ${K8S_DEPLOYMENT}"
if trivy config --exit-code 1 --severity MEDIUM,HIGH,CRITICAL . ;
then
  echo "## Cool no vulnerabilities was found in your ${DEPLOYMENT}, will now change directory :$OLDPWD"
  cd "$OLDPWD" || exit
  echo "## Deploying ${K8S_DEPLOYMENT} in the K8S cluster"
  kubectl apply -f ${DEPLOYMENT_DIRECTORY}/${K8S_DEPLOYMENT}
  # Check deployment rollout status every 5 seconds (max 1 minutes) until complete.
  ATTEMPTS=0
  ROLLOUT_STATUS_CMD="kubectl rollout status deployment ${APP_NAME}"
  until $ROLLOUT_STATUS_CMD || [ $ATTEMPTS -eq 12 ]; do
    echo "## doing rollout status attempt num: ${ATTEMPTS} ..."
    $ROLLOUT_STATUS_CMD
    ATTEMPTS=$((ATTEMPTS + 1))
    sleep 5
  done
  echo "## Listing  pods in the cluster "
  kubectl get pods -o wide
  echo "## Listing  services in the cluster "
  kubectl get service -o wide
  #echo "## Listing  ingress in the cluster "
  #kubectl get ingress -o wide
  sleep 2
  echo "## Running a curl on new service at cluster http://localhost:8000"
  curl http://localhost:8000
  echo "## Running a curl on new service at cluster http://go-info-server.rancher.localhost:8000"
  curl http://go-info-server.rancher.localhost:8000
  # echo "Pods are allocated a private IP address by default and cannot be reached outside of the cluster unless you have a corresponding service."
  # echo "You can also use the kubectl port-forward command to map a local port to a port inside the pod like this : (ctrl+c to terminate)"
  # kubectl port-forward go-info-server-766947b78b-64f7j 8080:8080
else
  echo "## You must correct the MEDIUM,HIGH,CRITICAL vulnerabilities detected by Trivy, before building your DockerFile" >&2
fi

