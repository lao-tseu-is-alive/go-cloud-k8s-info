#!/bin/bash
echo "## Extracting app name and version from source"
DEPLOYMENT=k8s-deployment.yml
VERSION=`grep -E 'VERSION\s+=' server.go| awk '{ print $3 }'  | tr -d '"'`
APPNAME=`grep -E 'APP\s+=' server.go| awk '{ print $3 }'  | tr -d '"'`
echo "## APP: ${APPNAME}, version: ${VERSION} detected in file server.go"
echo "## Listing relevant images in k8s namespace"
nerdctl -n k8s.io images | grep ${APPNAME}
TMP_K8S_CONFIG=$(mktemp -d)
echo "## Copying ${DEPLOYMENT IN directory ${TMP_K8S_CONFIG}}"
cp $DEPLOYMENT  $TMP_K8S_CONFIG/
cd $TMP_K8S_CONFIG/
echo "## Checking for vulnerabilities in ${DEPLOYMENT}"
trivy config --exit-code 1 --severity MEDIUM,HIGH,CRITICAL .
if [ $? -eq 0 ]
then
  echo "## Cool no vulnerabilities was found in your ${DEPLOYMENT}"
  cd "$OLDPWD"
  rm -rf $TMP_K8S_CONFIG/ # cleanup
  echo "## Deploying ${DEPLOYMENT} in the K8S cluster"
  kubectl apply -f $DEPLOYMENT
  # Check deployment rollout status every 5 seconds (max 1 minutes) until complete.
  ATTEMPTS=0
  ROLLOUT_STATUS_CMD="kubectl rollout status deployment ${APPNAME}"
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
  curl http://go-info-server.rancher.localhost:8000
  # echo "Pods are allocated a private IP address by default and cannot be reached outside of the cluster unless you have a corresponding service."
  # echo "You can also use the kubectl port-forward command to map a local port to a port inside the pod like this : (ctrl+c to terminate)"
  # kubectl port-forward go-info-server-766947b78b-64f7j 8080:8080
else
  echo "## You must correct the MEDIUM,HIGH,CRITICAL vulnerabilities detected by Trivy, before building your DockerFile" >&2
fi

