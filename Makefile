all: push-docker

VERSION = 0.1
TAG = $(VERSION)
PREFIX = priya7390/pod-monitor
DISTDIR = $(CURDIR)/dist
BINDIR = $(CURDIR)/bin

----------------------------
targets to build/run locally
----------------------------

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


---------------
docker targets
---------------

build-cross: #cross-compilation for linux
	CGO_ENABLED=0 gox -output="$(DISTDIR)/{{.OS}}-{{.Arch}}/{{.Dir}}" -osarch='linux/amd64' $(GOFLAGS) -ldflags '-extldflags -static' ./...

build-docker: build-cross #build docker image
	docker build -t $(PREFIX):$(TAG) .

push-docker: build-docker #push image to docker hub
	docker push $(PREFIX):$(TAG)





