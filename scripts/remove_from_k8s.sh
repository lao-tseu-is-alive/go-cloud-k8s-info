#!/bin/bash
echo "check the pods in the cluster "
kubectl get pods
echo "deleting go-info-server deployment"
kubectl delete deployment  go-info-server
echo "deleting go-info-server service"

kubectl get service
kubectl delete service go-info-server-service
echo "deleting go-info-server ingress"
kubectl delete ingress go-info-server-ingress