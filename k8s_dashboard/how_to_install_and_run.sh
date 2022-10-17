#!/bin/bash
echo "install k8s dashboard in k3s : https://docs.k3s.io/installation/kube-dashboard"
export GITHUB_URL=https://github.com/kubernetes/dashboard/releases
export VERSION_KUBE_DASHBOARD=$(curl -w '%{url_effective}' -I -L -s -S ${GITHUB_URL}/latest -o /dev/null | sed -e 's|.*/||')
kubectl create -f https://raw.githubusercontent.com/kubernetes/dashboard/${VERSION_KUBE_DASHBOARD}/aio/deploy/recommended.yaml
cat dashboard.admin-user.yml
cat dashboard.admin-user-role.yml
echo "creating dashboard admin user"
kubectl create -f dashboard.admin-user.yml -f dashboard.admin-user-role.yml
kubectl -n kubernetes-dashboard describe secret admin-user-token | grep '^token'
kubectl proxy
