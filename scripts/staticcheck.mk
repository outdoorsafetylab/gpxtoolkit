TOOLCHAIN ?= .tool
STATICCHECK := $(TOOLCHAIN)/bin/staticcheck

$(STATICCHECK):
	GOBIN=$(realpath $(TOOLCHAIN)/bin) go install honnef.co/go/tools/cmd/staticcheck@latest

clean/staticcheck:
	rm -f $(STATICCHECK)

.PHONY: clean/staticcheck
