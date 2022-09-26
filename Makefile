PROTOS := $(wildcard *.proto) $(wildcard */*.proto) $(wildcard */*/*.proto)
PBGO := $(PROTOS:.proto=.pb.go)
GIT_HASH ?= $(shell git rev-parse --short HEAD)
GIT_TAG ?= $(shell git describe --tags --exact-match 2>/dev/null || echo "")

IMAGE_NAME := outdoorsafetylab/gpxtoolkit

all: $(PBGO)
	go build -ldflags="-X gpxtoolkit/version.GitHash=$(GIT_HASH) -X gpxtoolkit/version.GitTag=$(GIT_TAG)" -o gpxtoolkit .

test:
	go test ./gpx
	go test ./gpxutil

include .make/golangci-lint.mk
include .make/protoc.mk
include .make/protoc-gen-go.mk

serve:
	go run . serve -d

watch: $(WATCHER) # To install 'nodemon': npm install -g nodemon
	nodemon -e go --signal SIGTERM --exec 'make serve'

lint: $(GOLANGCI_LINT)
	$(realpath $(GOLANGCI_LINT)) run

docker/build:
	docker build --network=host --force-rm \
		$(if $(call eq,$(no-cache),yes),--no-cache --pull,) \
		--build-arg GIT_HASH=$(GIT_HASH) \
		--build-arg GIT_TAG=$(GIT_TAG) \
		-t $(IMAGE_NAME) \
		-f Dockerfile \
		.

docker/run:
	docker run -it --rm \
		-p 8080:8080 \
		$(IMAGE_NAME)
