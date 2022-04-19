kubectl create deployment hello-node --image=k8s.gcr.io/echoserver:1.4
kubectl get deployments
kubectl get pods
kubectl get pods -o wide
k get events
k config view
kubectl expose deployment hello-node --type=LoadBalancer --port=8080
k get services
k get pod,svc
k get pod,svc -n kube-system
k delete service hello-node
k delete deployment hello-node
