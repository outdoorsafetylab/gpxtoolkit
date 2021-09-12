PROTOS := $(wildcard *.proto) $(wildcard */*.proto) $(wildcard */*/*.proto)
PBGO := $(PROTOS:.proto=.pb.go)

IMAGE_NAME := outdoorsafetylab/gpxtoolkit

all: $(PBGO)
	go build -o gpxtoolkitd ./cmd/gpxtoolkitd
	go build -o gpxmarks ./cmd/gpxmarks
	go build -o gpx2svg ./cmd/gpx2svg

test:
	go test ./gpx
	go test ./milestone

include .make/golangci-lint.mk
include .make/protoc.mk
include .make/protoc-gen-go.mk
include .make/watcher.mk

watch: $(WATCHER)
	$(realpath $(WATCHER)) -D

lint: $(GOLANGCI_LINT)
	$(realpath $(GOLANGCI_LINT)) run

docker/build:
	docker build --network=host --force-rm \
		$(if $(call eq,$(no-cache),yes),--no-cache --pull,) \
		-t $(IMAGE_NAME) \
		-f Dockerfile \
		.

docker/run:
	docker run -it --rm \
		-p 8080:8080 \
		$(IMAGE_NAME)
