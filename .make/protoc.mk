TOOLCHAIN ?= .tool
PROTOC := $(TOOLCHAIN)/bin/protoc
PROTOC_VERSION := 3.12.4

$(PROTOC):
	mkdir -p $(dir $@)
ifeq ($(shell uname -s),Linux)
	curl -L -o $(TOOLCHAIN)/protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-linux-x86_64.zip
else ifeq ($(shell uname -s),Darwin)
	curl -L -o $(TOOLCHAIN)/protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-osx-x86_64.zip
endif
	cd $(TOOLCHAIN) && unzip -o protoc.zip
	chmod +x $@
	rm -f $(TOOLCHAIN)/protoc.zip

clean/protoc:
	rm -f $(PROTOC)

.PHONY: clean/protoc
