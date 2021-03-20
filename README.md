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
cd k8s-pod-monitor
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

![Alt text](images/pod_monitor_crd_1.png?raw=true "Pod Monitor CRD - PodRunningCount")

7. cd into test_k8s_manifests repository
```
cd test_k8s_manifests
```
8. Apply nginx deployment (replicas = 2. this'll create 2 nginx pods)
```
kubectl apply -f nginx-deployment.yaml
```
9. Get `pod-monitor` custom resource. `PodCreatedCount` in the pod-monitor status reflects the number of pods created from the
moment the k8s-pod-monitor has been deployed and until the moment of request. `podCreatedCount` is `2` as we created 2 pods via the
previous nginx deployment. The pods are currently in running state and hence the `podRunningCount` is updated from 9 to 11 
```
kubectl get pm pod-monitor -o yaml
```

![Alt text](images/pod_monitor_crd_2.png?raw=true "Pod Monitor CRD - PodCreatedCount")


10. Delete nginx test deployment
```
kubectl delete deployment nginx-deployment-test
```

11. Get `pod-monitor` custom resource. `podRunningCount` in the pod-monitor status is updated from 11 to 9 as the 2 nginx test pods are not running anymore (since deletion in the previous step)

![Alt text](images/pod_monitor_crd_3.png?raw=true "Pod Monitor CRD")




## Requirements to build/run/test locally
- Go 1.10+
- Docker
- Kubernetes cluster
- Kubectl cli client
- dep
- gox (https://github.com/mitchellh/gox)
### Steps
1. cd into the repository
```
cd k8s-pod-monitor
```
2. Build and Run. This will fetch all the dependencies, build the binary and start the controller
```
make run
```

![Alt text](images/pod_monitor_controller_1.png?raw=true "Pod Monitor Controller")

3. Run test. Open a new terminal and cd into `k8s-pod-monitor` directory. 
```
make test
```

![Alt text](images/pod_monitor_test.png?raw=true "Pod Monitor Test Result")

4. Observe the controller log after running `make test`. The test auto-deploys test nginx deployment and checks if the `podRunningCount` is updated accordingly in the `pod-monitor` CRD. At the end of the test, `kubectl delete -f nginx-deployment.yaml` is automatically run to clean up the resources created for test
 
![Alt text](images/pod_monitor_controller_log_for_test.png?raw=true "Pod Monitor Controller Log For Test")

5. Dockerize the solution and push to public docker hub so that this image can be referenced in the pod-monitor deployment yaml (`pod-monitor-deployment.yaml`). This is optional if there are no new enhancements/changes as we already have the image in docker hub.

```
make push-docker
```

## Future Enhancements
- To make pod-monitor controller (responsible for metrics acquisition and storage into etcd) highly available. `leader-election` package in the `client-go` can be used to implement HA controllers (https://github.com/kubernetes/client-go/tree/066127c6df69d5945429ced9754841765b5a0310/tools/leaderelection)

## References
- https://github.com/kubernetes/client-go
- https://github.com/kubernetes/sample-controller
- https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/
- https://medium.com/velotio-perspectives/extending-kubernetes-apis-with-custom-resource-definitions-crds-139c99ed3477
- https://medium.com/@trstringer/create-kubernetes-controllers-for-core-and-custom-resources-62fc35ad64a3
- https://itnext.io/how-to-create-a-kubernetes-custom-controller-using-client-go-f36a7a7536cc
- https://engineering.bitnami.com/articles/kubewatch-an-example-of-kubernetes-custom-controller.html
- https://github.com/kubernetes-sigs/kubebuilder

