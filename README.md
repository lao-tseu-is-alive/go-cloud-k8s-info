## go-cloud-k8s-info

[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-info&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-info)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-info&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-info)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-info&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-info)
[![cve-trivy-scan](https://github.com/lao-tseu-is-alive/go-cloud-k8s-info/actions/workflows/cve-trivy-scan.yml/badge.svg)](https://github.com/lao-tseu-is-alive/go-cloud-k8s-info/actions/workflows/cve-trivy-scan.yml)
[![codecov](https://codecov.io/gh/lao-tseu-is-alive/go-cloud-k8s-info/branch/main/graph/badge.svg)](https://codecov.io/gh/lao-tseu-is-alive/go-cloud-k8s-info)

_**go-cloud-k8s-info** is a simple "nano"-service written in Golang 
that gives some runtime information. 
This project showcases how to build a container image with nerdctl, in a secured way 
(scan of CVE done with [Trivy](https://aquasecurity.github.io/trivy/v0.18.3/installation/))
and how to deploy it on Kubernetes_
### Introduction :
In this repository you have all you need to compile & deploy 
a simple golang microservice http server without using docker in just **two single steps** :
    
    scripts/01_build_image.sh
    scripts/02_deploy_to_k8s.sh
#### Specifications :
+ All the Go code is in one simple file [server.go](https://github.com/lao-tseu-is-alive/go-cloud-k8s-info/blob/main/server.go).
+ Using [Rancher desktop](https://docs.rancherdesktop.io/) to deploy the excellent [k3s](https://k3s.io/) kubernetes on your development computer.
+ We choose to build container image with [nerdctl](https://github.com/containerd/nerdctl): the  Docker-compatible CLI for [containerd](https://containerd.io/) just to show that you don't need Docker on your Linux box anymore.
+ We will scan for security issues and other vulnerabilities **before** building a container image (using [Trivy](https://aquasecurity.github.io/trivy/)) 
+ We will scan for security issues and other vulnerabilities **before** deploying to kubernetes (using [Trivy](https://aquasecurity.github.io/trivy/))

### 00 : Develop and test your Go code as usual

    $> PORT=7070 go run server.go
    HTTP_SERVER_go-info-server 2022/06/02 10:43:44 INFO: 'Starting go-info-server version:0.2.9 HTTP server on port :7070'
    HTTP_SERVER_go-info-server 2022/06/02 10:43:44 INFO: 'Will start ListenAndServe...'
    HTTP_SERVER_go-info-server 2022/06/02 10:45:45 request: GET '/'	remoteAddr: 127.0.0.1:54694

you can then use another terminal to run a :

    curl http://localhost:7070
    {
        "hostname": "pulsar2021",
        "pid": 208632,
        "ppid": 208422,
        "uid": 1000,
        "appname": "go-info-server",
        "version": "0.3.4",
        "param_name": "_EMPTY_STRING_",
        "remote_addr": "127.0.0.1:54694",
        "goos": "linux",
        "goarch": "amd64",  
        "runtime": "go1.18.2",
        "num_goroutine": "5",
        "os_release_name": "Ubuntu",
        "os_release_version": "20.04.4 LTS (Focal Fossa)",
        "os_release_version_id": "20.04",
        "num_cpu": "36",
        "env_vars": [
            "PORT=7070",
            "SHELL=/bin/bash",
        ...

As you can see you got all the environment variables values.
Take also note of the process id in pid, your userid and the num_cpu...


_Now in just 2 easy steps, you will deploy your first "tiny-service" in 
a local kubernetes in your computer, without using docker at all._

### 01 : Build your container image
in this first step we will use a [Multi-stage build](https://docs.docker.com/language/golang/build-images/#multi-stage-builds)
to have a clean and small final container image of our server. 

you can just use the bash script I have prepared for you :
```bash
scripts/01_build_image.sh
```
or run the commands in this script one by one 
```bash
nerdctl -n k8s.io build -t go-info-server .
#list all images in the kubernetes namespace of containerd
nerdctl -n k8s.io images |grep go-info-server
#optionaly you can run your image to test if wou want
nerdctl -n k8s.io run -p 8080:8080 go-info-server
```
if you did run your image as a container with the above command,
you can check the results of a : **_curl http://localhost:8080/_**
```json
{
  hostname: "48fbdda3e5c2",
  pid: 1,
  ppid: 0,
  uid: 10111,
  appname: "go-info-server",
  version: "0.3.4",
  param_name: "_EMPTY_STRING_",
  remote_addr: "10.4.0.1:59936",
  goos: "linux",
  goarch: "amd64",
  runtime: "go1.18.2",
  num_goroutine: "5",
  os_release_name: "Alpine Linux",
  os_release_version: "_UNKNOWN_",
  os_release_version_id: "3.15.4",
  num_cpu: "4",
  env_vars: [
    "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
    "HOME=/home/gouser"
  ]
}
```
*Did you notice how there is now only two environment variables exposed.
Also note that the process id in pid is just 1, and the parent process id in ppid is zero.
Finally, the userid is the one for the gouser defined in the Dockerfile* :

    RUN addgroup -g 10111 -S gouser && adduser -S -G gouser -H -u 10111 gouser
    USER gouser
 
### 02 : Deploy your container image to k8s
again you can just use the bash script:
```bash
scripts/02_deploy_to_k8s.sh
```
or run the commands in this script one by one
```bash
kubectl apply -f k8s-deployment.yml
#let's check the pods in the cluster
kubectl get pods -o wide
kubectl get services -o wide
curl http://go-info-server.rancher.localhost:8000?name=gilou
```
here is the example output from curl :
```json
{
  "hostname": "go-info-server-6d8c486db8-ftwd6",
  "pid": 1,
  "ppid": 0,
  "uid": 0,
  "appname": "go-info-server",
  "version": "0.3.4",
  "param_name": "gilou",
  "goos": "linux",
  "goarch": "amd64",
  "runtime": "go1.17.7",
  "num_goroutine": "5",
  os_release_name: "Alpine Linux",
  os_release_version: "_UNKNOWN_",
  os_release_version_id: "3.15.4",
  "num_cpu": "4",
  "env_vars": [
    "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
    "HOSTNAME=go-info-server-6d8c486db8-ftwd6",
    "MY_POD_IP=10.42.0.36",
    "MY_POD_SERVICE_ACCOUNT=default",
    "PORT=8000",
    "MY_NODE_NAME=lima-rancher-desktop",
    "MY_POD_NAME=go-info-server-6d8c486db8-ftwd6",
    "MY_POD_NAMESPACE=default",
    "GO_INFO_SERVER_SERVICE_PORT=tcp://10.43.192.1:8000",
    "GO_INFO_SERVER_SERVICE_PORT_8000_TCP=tcp://10.43.192.1:8000",
    "KUBERNETES_SERVICE_HOST=10.43.0.1",
    "KUBERNETES_SERVICE_PORT_HTTPS=443",
    "KUBERNETES_PORT=tcp://10.43.0.1:443",
    "KUBERNETES_PORT_443_TCP_ADDR=10.43.0.1",
    "GO_INFO_SERVER_SERVICE_SERVICE_PORT=8000",
    "GO_INFO_SERVER_SERVICE_PORT_8000_TCP_PORT=8000",
    "KUBERNETES_SERVICE_PORT=443",
    "KUBERNETES_PORT_443_TCP_PORT=443",
    "GO_INFO_SERVER_SERVICE_SERVICE_PORT_HTTP=8000",
    "GO_INFO_SERVER_SERVICE_PORT_8000_TCP_ADDR=10.43.192.1",
    "GO_INFO_SERVER_SERVICE_SERVICE_HOST=10.43.192.1",
    "GO_INFO_SERVER_SERVICE_PORT_8000_TCP_PROTO=tcp",
    "KUBERNETES_PORT_443_TCP=tcp://10.43.0.1:443",
    "KUBERNETES_PORT_443_TCP_PROTO=tcp",
    "HOME=/root"
  ]
}
```

To check for vulnerabilities in your Docker and k8s yaml files in the current directory with :

    trivy config .

### Tools used :
+ [Rancher Desktop: k3s and container management on your desktop](https://rancherdesktop.io/)
+ [Trivy vulnerabilities scan installation](https://aquasecurity.github.io/trivy/v0.23.0/getting-started/installation/)
+ [nerdctl command reference](https://github.com/containerd/nerdctl#command-reference)
+ [jq a lightweight and flexible command-line JSON processor](https://stedolan.github.io/jq/)
+ [yq a portable command-line YAML processor](https://github.com/mikefarah/yq)

### more information :
+ [K3S networking : CoreDNS, Traefik and Klipper Load balancer](https://rancher.com/docs/k3s/latest/en/networking/)
+ [K3S Load Balancing at Funky Penguin's Geek Cookbook](https://geek-cookbook.funkypenguin.co.nz/kubernetes/loadbalancer/k3s/)
+ [K3S at Funky Penguin's Geek Cookbook](https://geek-cookbook.funkypenguin.co.nz/kubernetes/cluster/k3s/)
+ [A Guide to K3s Ingress Using Traefik with NodePort](https://levelup.gitconnected.com/a-guide-to-k3s-ingress-using-traefik-with-nodeport-6eb29add0b4b)
+ [Build and Deploy Containerized Applications with Golang on Kubernetes](http://coding-bootcamps.com/blog/build-containerized-applications-with-golang-on-kubernetes.html)
+ [Rancher Desktop and nerdctl for local K8s dev](https://itnext.io/rancher-desktop-and-nerdctl-for-local-k8s-dev-d1348629932a)
+ [nerdctl: Docker-compatible CLI for containerd (github)](https://github.com/containerd/nerdctl)
+ [Best practices for writing Dockerfiles](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
+ [Continuous Container Vulnerability Testing with Trivy](https://semaphoreci.com/blog/continuous-container-vulnerability-testing-with-trivy)
+ [Kubernetes security overview](https://kubernetes.io/docs/concepts/security/overview/)
+ [Getting Real Client IP with k3s](https://github.com/k3s-io/k3s/discussions/2997)
+ [jq cookbook](https://github.com/stedolan/jq/wiki/Cookbook)

**How to enable Traefik ingress controller dashboard :**
```bash
kubectl port-forward -n kube-system $(kubectl -n kube-system get pods --selector "app.kubernetes.io/name=traefik" --output=name) 9000:9000
```
Visit http://127.0.0.1:9000/dashboard/ in your browser to view the Traefik dashboard.


### From scratch vs From alpine : 
Actual image size (with FROM alpine:3.15)  is 13.3 MiB
by building the image FROM scratch the image size goes just half size 6.0MB

Another important thing is that there is **NO WAY to go "inside" this container with an interactive shell**,
because there is just no shell at all it's just your go statically compiled application.

in a classical FROM alpine or whatever Linux distro you use, 
you will always be able to run  a shell inside the container...

    nerdctl -n k8s.io run -it 59316da2a057 /bin/sh

**Every container in a cluster is populated with a token that can be used 
for authenticating to the API server. 
To verify, Inside the above container shell just run:**

    cat /var/run/secrets/kubernetes.io/serviceaccount/token
    #so let's use this to try listing pods with the api
    # https://kubernetes.io/docs/tasks/run-application/access-api-from-pod/
    
    # Point to the internal API server hostname
    APISERVER=https://kubernetes.default.svc

    # Path to ServiceAccount token
    SERVICEACCOUNT=/var/run/secrets/kubernetes.io/serviceaccount
    
    # Read this Pod's namespace
    NAMESPACE=$(cat ${SERVICEACCOUNT}/namespace)

    # Read the ServiceAccount bearer token
    TOKEN=$(cat ${SERVICEACCOUNT}/token)
    # or use a quicker one :
    TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)

    # Reference the internal certificate authority (CA)
    CACERT=${SERVICEACCOUNT}/ca.crt
    
    wget -O - --no-check-certificate --header "Authorization: Bearer ${TOKEN}"  ${APISERVER}/api/v1

```json
{
  "kind": "APIResourceList",
  "groupVersion": "v1",
  "resources": [
    {
      "name": "bindings",
      "singularName": "",
      "namespaced": true,
      "kind": "Binding",
      "verbs": [
        "create"
      ]
    },
    {
      "name": "componentstatuses",
      "singularName": "",
      "namespaced": false,
      "kind": "ComponentStatus",
      "verbs": [
        "get",
        "list"
      ],
      "shortNames": [
        "cs"
      ]
    },
    {
      "name": "configmaps",
      "singularName": "",
      "namespaced": true,
      "kind": "ConfigMap",
      "verbs": [
        "create",
        "delete",
        "deletecollection",
        "get",
        "list",
        "patch",
        "update",
        "watch"
      ],
      "shortNames": [
        "cm"
      ],
      "storageVersionHash": "qFsyl6wFWjQ="
    },
    {
      "name": "endpoints",
      "singularName": "",
      "namespaced": true,
      "kind": "Endpoints",
      "verbs": [
        "create",
        "delete",
        "deletecollection",
        "get",
        "list",
        "patch",
        "update",
        "watch"
      ],
      "shortNames": [
        "ep"
      ],
      "storageVersionHash": "fWeeMqaN/OA="
    },
    {
      "name": "events",
      "singularName": "",
      "namespaced": true,
      "kind": "Event",
      "verbs": [
        "create",
        "delete",
        "deletecollection",
        "get",
        "list",
        "patch",
        "update",
        "watch"
      ],
      "shortNames": [
        "ev"
      ],
      "storageVersionHash": "r2yiGXH7wu8="
    },
    {
      "name": "limitranges",
      "singularName": "",
      "namespaced": true,
      "kind": "LimitRange",
      "verbs": [
        "create",
        "delete",
        "deletecollection",
        "get",
        "list",
        "patch",
        "update",
        "watch"
      ],
      "shortNames": [
        "limits"
      ],
      "storageVersionHash": "EBKMFVe6cwo="
    },
    {
      "name": "namespaces",
      "singularName": "",
      "namespaced": false,
      "kind": "Namespace",
      "verbs": [
        "create",
        "delete",
        "get",
        "list",
        "patch",
        "update",
        "watch"
      ],
      "shortNames": [
        "ns"
      ],
      "storageVersionHash": "Q3oi5N2YM8M="
    },
    {
      "name": "namespaces/finalize",
      "singularName": "",
      "namespaced": false,
      "kind": "Namespace",
      "verbs": [
        "update"
      ]
    },
    {
      "name": "namespaces/status",
      "singularName": "",
      "namespaced": false,
      "kind": "Namespace",
      "verbs": [
        "get",
        "patch",
        "update"
      ]
    },
    {
      "name": "nodes",
      "singularName": "",
      "namespaced": false,
      "kind": "Node",
      "verbs": [
        "create",
        "delete",
        "deletecollection",
        "get",
        "list",
        "patch",
        "update",
        "watch"
      ],
      "shortNames": [
        "no"
      ],
      "storageVersionHash": "XwShjMxG9Fs="
    },
    {
      "name": "nodes/proxy",
      "singularName": "",
      "namespaced": false,
      "kind": "NodeProxyOptions",
      "verbs": [
        "create",
        "delete",
        "get",
        "patch",
        "update"
      ]
    },
    {
      "name": "nodes/status",
      "singularName": "",
      "namespaced": false,
      "kind": "Node",
      "verbs": [
        "get",
        "patch",
        "update"
      ]
    },
    {
      "name": "persistentvolumeclaims",
      "singularName": "",
      "namespaced": true,
      "kind": "PersistentVolumeClaim",
      "verbs": [
        "create",
        "delete",
        "deletecollection",
        "get",
        "list",
        "patch",
        "update",
        "watch"
      ],
      "shortNames": [
        "pvc"
      ],
      "storageVersionHash": "QWTyNDq0dC4="
    },
    {
      "name": "persistentvolumeclaims/status",
      "singularName": "",
      "namespaced": true,
      "kind": "PersistentVolumeClaim",
      "verbs": [
        "get",
        "patch",
        "update"
      ]
    },
    {
      "name": "persistentvolumes",
      "singularName": "",
      "namespaced": false,
      "kind": "PersistentVolume",
      "verbs": [
        "create",
        "delete",
        "deletecollection",
        "get",
        "list",
        "patch",
        "update",
        "watch"
      ],
      "shortNames": [
        "pv"
      ],
      "storageVersionHash": "HN/zwEC+JgM="
    },
    {
      "name": "persistentvolumes/status",
      "singularName": "",
      "namespaced": false,
      "kind": "PersistentVolume",
      "verbs": [
        "get",
        "patch",
        "update"
      ]
    },
    {
      "name": "pods",
      "singularName": "",
      "namespaced": true,
      "kind": "Pod",
      "verbs": [
        "create",
        "delete",
        "deletecollection",
        "get",
        "list",
        "patch",
        "update",
        "watch"
      ],
      "shortNames": [
        "po"
      ],
      "categories": [
        "all"
      ],
      "storageVersionHash": "xPOwRZ+Yhw8="
    },
    {
      "name": "pods/attach",
      "singularName": "",
      "namespaced": true,
      "kind": "PodAttachOptions",
      "verbs": [
        "create",
        "get"
      ]
    },
    {
      "name": "pods/binding",
      "singularName": "",
      "namespaced": true,
      "kind": "Binding",
      "verbs": [
        "create"
      ]
    },
    {
      "name": "pods/eviction",
      "singularName": "",
      "namespaced": true,
      "group": "policy",
      "version": "v1",
      "kind": "Eviction",
      "verbs": [
        "create"
      ]
    },
    {
      "name": "pods/exec",
      "singularName": "",
      "namespaced": true,
      "kind": "PodExecOptions",
      "verbs": [
        "create",
        "get"
      ]
    },
    {
      "name": "pods/log",
      "singularName": "",
      "namespaced": true,
      "kind": "Pod",
      "verbs": [
        "get"
      ]
    },
    {
      "name": "pods/portforward",
      "singularName": "",
      "namespaced": true,
      "kind": "PodPortForwardOptions",
      "verbs": [
        "create",
        "get"
      ]
    },
    {
      "name": "pods/proxy",
      "singularName": "",
      "namespaced": true,
      "kind": "PodProxyOptions",
      "verbs": [
        "create",
        "delete",
        "get",
        "patch",
        "update"
      ]
    },
    {
      "name": "pods/status",
      "singularName": "",
      "namespaced": true,
      "kind": "Pod",
      "verbs": [
        "get",
        "patch",
        "update"
      ]
    },
    {
      "name": "podtemplates",
      "singularName": "",
      "namespaced": true,
      "kind": "PodTemplate",
      "verbs": [
        "create",
        "delete",
        "deletecollection",
        "get",
        "list",
        "patch",
        "update",
        "watch"
      ],
      "storageVersionHash": "LIXB2x4IFpk="
    },
    {
      "name": "replicationcontrollers",
      "singularName": "",
      "namespaced": true,
      "kind": "ReplicationController",
      "verbs": [
        "create",
        "delete",
        "deletecollection",
        "get",
        "list",
        "patch",
        "update",
        "watch"
      ],
      "shortNames": [
        "rc"
      ],
      "categories": [
        "all"
      ],
      "storageVersionHash": "Jond2If31h0="
    },
    {
      "name": "replicationcontrollers/scale",
      "singularName": "",
      "namespaced": true,
      "group": "autoscaling",
      "version": "v1",
      "kind": "Scale",
      "verbs": [
        "get",
        "patch",
        "update"
      ]
    },
    {
      "name": "replicationcontrollers/status",
      "singularName": "",
      "namespaced": true,
      "kind": "ReplicationController",
      "verbs": [
        "get",
        "patch",
        "update"
      ]
    },
    {
      "name": "resourcequotas",
      "singularName": "",
      "namespaced": true,
      "kind": "ResourceQuota",
      "verbs": [
        "create",
        "delete",
        "deletecollection",
        "get",
        "list",
        "patch",
        "update",
        "watch"
      ],
      "shortNames": [
        "quota"
      ],
      "storageVersionHash": "8uhSgffRX6w="
    },
    {
      "name": "resourcequotas/status",
      "singularName": "",
      "namespaced": true,
      "kind": "ResourceQuota",
      "verbs": [
        "get",
        "patch",
        "update"
      ]
    },
    {
      "name": "secrets",
      "singularName": "",
      "namespaced": true,
      "kind": "Secret",
      "verbs": [
        "create",
        "delete",
        "deletecollection",
        "get",
        "list",
        "patch",
        "update",
        "watch"
      ],
      "storageVersionHash": "S6u1pOWzb84="
    },
    {
      "name": "serviceaccounts",
      "singularName": "",
      "namespaced": true,
      "kind": "ServiceAccount",
      "verbs": [
        "create",
        "delete",
        "deletecollection",
        "get",
        "list",
        "patch",
        "update",
        "watch"
      ],
      "shortNames": [
        "sa"
      ],
      "storageVersionHash": "pbx9ZvyFpBE="
    },
    {
      "name": "serviceaccounts/token",
      "singularName": "",
      "namespaced": true,
      "group": "authentication.k8s.io",
      "version": "v1",
      "kind": "TokenRequest",
      "verbs": [
        "create"
      ]
    },
    {
      "name": "services",
      "singularName": "",
      "namespaced": true,
      "kind": "Service",
      "verbs": [
        "create",
        "delete",
        "get",
        "list",
        "patch",
        "update",
        "watch"
      ],
      "shortNames": [
        "svc"
      ],
      "categories": [
        "all"
      ],
      "storageVersionHash": "0/CO1lhkEBI="
    },
    {
      "name": "services/proxy",
      "singularName": "",
      "namespaced": true,
      "kind": "ServiceProxyOptions",
      "verbs": [
        "create",
        "delete",
        "get",
        "patch",
        "update"
      ]
    },
    {
      "name": "services/status",
      "singularName": "",
      "namespaced": true,
      "kind": "Service",
      "verbs": [
        "get",
        "patch",
        "update"
      ]
    }
  ]

  
}
```

if your kubernetes cluster and your deployment  is correctly configured & secured,
then if you try to list the pods you should get:

    wget -O - --no-check-certificate --header "Authorization: Bearer ${TOKEN}"  ${APISERVER}/api/v1/namespaces/default/pods
    HTTP/1.1 403 Forbidden

but sometimes you will just get the list of pods... 

Yes, inside every alpine you have 
the wget command to download what you want.

So maybe let's download the kubectl command

    # first let's open a shell inside your pod (user your own  
    kubectl exec -it go-info-server-79db446d8-l9tn5 -- /bin/sh
    mkdir /dev/shm/bin
    cd mytools
    

**So again, YES maybe it is safer to use container images that are build from scratch...**

_if you want to build manually your from scratch container  :_
```bash
cd DockerfileFromScratch
cp ../go.* .
cp ../*.go .
nerdctl -n k8s.io build -t go-info-server-from-scratch .
nerdctl -n k8s.io tag go-info-server-from-scratch go-info-server-from-scratch:0.1.1
nerdctl -n k8s.io images | grep go-info
kubectl apply -f k8s-deployment-from-scratch.yml 
curl http://localhost:8000

```


