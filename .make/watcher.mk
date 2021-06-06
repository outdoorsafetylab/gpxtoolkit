TOOLCHAIN ?= .tool
WATCHER := $(TOOLCHAIN)/bin/watcher

$(WATCHER):
	mkdir -p $(dir $@)
	GOPATH=$(realpath $(TOOLCHAIN)) \
		go get github.com/crosstalkio/go-watcher
	GOPATH=$(realpath $(TOOLCHAIN)) \
		go install github.com/crosstalkio/go-watcher/cmd/watcher

clean/watcher:
	rm -f $(WATCHER)

.PHONY: clean/watcher
