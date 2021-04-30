PROTOS := $(wildcard *.proto) $(wildcard */*.proto) $(wildcard */*/*.proto)
PBGO := $(PROTOS:.proto=.pb.go)

all: $(PBGO)
	go build -o gpxtoolkit .

test:
	go test ./gpx

include .make/golangci-lint.mk
include .make/protoc.mk
include .make/protoc-gen-go.mk

lint: $(GOLANGCI_LINT)
	$(realpath $(GOLANGCI_LINT)) run
