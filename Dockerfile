FROM alpine:3.10

MAINTAINER "Jayapriya Surendran" <priya7390@gmail.com>

COPY dist/linux-amd64/k8s-pod-monitor /k8s-pod-monitor

ENTRYPOINT ["/k8s-pod-monitor"]
