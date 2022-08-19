#!/bin/bash
echo "about to enter an interactive session that will end after you exit the bash terminal"

# to get a list of pods via API
# curl -s -k -H "Authorization: Bearer $TOKEN" -H 'Accept: application/json' $APISERVER/api/v1/pods |jq '.items[] | .metadata.namespace +":" + .metadata.name + ", ["  + .status.podIP + "]" + ", image:"  + .spec.containers[0].image + ", startTime:"  + .status.startTime '
kubectl -n default run -i --tty bash --image=ubuntu --restart=Never -- bash