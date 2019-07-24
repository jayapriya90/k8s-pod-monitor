all: push-docker

VERSION = 0.1
TAG = $(VERSION)
PREFIX = priya7390/pod-monitor
DISTDIR = $(CURDIR)/dist
BINDIR = $(CURDIR)/bin

build:
	dep ensure
	go build

run: build
	./k8s-pod-monitor

test:
	go test

clean:
	rm -rf dist
	rm -rf vendor

build-cross:
	CGO_ENABLED=0 gox -output="$(DISTDIR)/{{.OS}}-{{.Arch}}/{{.Dir}}" -osarch='linux/amd64' $(GOFLAGS) -ldflags '-extldflags -static' ./...

build-docker: build-cross
	docker build -t $(PREFIX):$(TAG) .

push-docker: build-docker
	docker push $(PREFIX):$(TAG)





