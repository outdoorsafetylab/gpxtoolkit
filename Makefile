PROTOS := $(wildcard *.proto) $(wildcard */*.proto) $(wildcard */*/*.proto)
PBGO := $(PROTOS:.proto=.pb.go)
GIT_HASH ?= $(shell git rev-parse --short HEAD)
GIT_TAG ?= $(shell git describe --tags --exact-match 2>/dev/null || echo "")

IMAGE_NAME := outdoorsafetylab/gpxtoolkit

.env:
	@echo "GIT_HASH=$(GIT_HASH)" > $@
	@echo "GIT_TAG=$(GIT_TAG)" >> $@

all: $(PBGO) .env
	go build -o gpxtoolkit .

test:
	go test ./...

include scripts/staticcheck.mk
include scripts/protoc.mk
include scripts/protoc-gen-go.mk

serve:
	go run . serve -d

watch: # To install 'air': go install github.com/cosmtrek/air@latest
	air

frontend:
	cd webroot && npm install && npm run serve

lint: $(STATICCHECK)
	$(realpath $(STATICCHECK)) ./...

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
