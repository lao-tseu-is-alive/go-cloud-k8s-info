## How to build and deploy a simple Golang app on K8s (without docker)

### Intro :
In this directory we have all the files to compile & deploy a simple golang http server without docker.
+ The Go code is in [server.go](https://github.com/lao-tseu-is-alive/go-cloud-k8s-info/blob/main/server.go).
+ we will use the [Rancher desktop](https://docs.rancherdesktop.io/) that deploy for you the excellent [k3s](https://k3s.io/) cluster on your dev computer.
+ The above product will also allow you to choose to build image with the [nerdctl](https://github.com/containerd/nerdctl) : the  Docker-compatible CLI for [containerd](https://containerd.io/).

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
        "version": "0.2.9",
        "param_name": "_EMPTY_STRING_",
        "remote_addr": "127.0.0.1:54694",
        "goos": "linux",
        "goarch": "amd64",  
        "runtime": "go1.18.2",
        "num_goroutine": "5",
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
  version: "0.2.9",
  param_name: "_EMPTY_STRING_",
  remote_addr: "10.4.0.1:59936",
  goos: "linux",
  goarch: "amd64",
  runtime: "go1.18.2",
  num_goroutine: "5",
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
  "version": "0.2.1",
  "param_name": "gilou",
  "goos": "linux",
  "goarch": "amd64",
  "runtime": "go1.17.7",
  "num_goroutine": "5",
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

#### Tools used :
+ [Rancher Desktop: k3s and container management on your desktop](https://rancherdesktop.io/)
+ [Trivy vulnerabilities scan installation](https://aquasecurity.github.io/trivy/v0.23.0/getting-started/installation/)
+ [nerdctl command reference](https://github.com/containerd/nerdctl#command-reference)
+ [jq a lightweight and flexible command-line JSON processor](https://stedolan.github.io/jq/)
+ [yq a portable command-line YAML processor](https://github.com/mikefarah/yq)

#### more information :
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


### From scratch vs From alpine 
Actual image size (with FROM alpine:3.15)  is 13.3 MiB
by building the image FROM scratch the image size goes just half size 6.0MB

Another important thing is that there is **NO WAY to go "inside" this container with an interactive shell**,
because there is just no shell at all it's just your go statically compiled application. 

if you want to test :
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
