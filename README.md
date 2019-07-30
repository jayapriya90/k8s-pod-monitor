# k8s-pod-monitor

Travis CI Build Status [![Build Status](https://travis-ci.com/jayapriya90/k8s-pod-monitor.svg?branch=master)](https://travis-ci.com/jayapriya90/k8s-pod-monitor)

K8s pod monitor keeps track of the following two metrics 
1. the total number of pods that have been created, not necessarily scheduled, from the
moment the k8s-pod-monitor has been deployed and until the moment of request
2. the number of pods that are running at the moment of request

## Usage
1. Log in to kubernetes cluster
2. Clone K8s-pod-monitor repository
```
git clone git@github.com:jayapriya90/k8s-pod-monitor.git
```
3. cd into the repository
```
k8s-pod-monitor
```
4. Apply pod monitor deployment
```
kubectl apply -f pod-monitor-deployment.yaml
```
5. List pods from all the namespaces
```
kubectl get pods --all-namespaces
```
6. Get `pod-monitor` custom resource.`podRunningCount` in the pod-monitor status reflects the number of running pods in the cluster at the moment of request
```
kubectl get pm pod-monitor -o yaml
```

![Alt text](images/pod_monitor_crd_1.png?raw=true "Pod Monitor CRD")

## Requirements to build/run/test locally
- Go 1.10
- Docker
- Kubernetes cluster
- Kubectl cli client
- dep
- gox (https://github.com/mitchellh/gox)




