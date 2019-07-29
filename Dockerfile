FROM alpine:3.10

COPY dist/linux-amd64/k8s-pod-monitor /k8s-pod-monitor

ENTRYPOINT ["/k8s-pod-monitor"]
