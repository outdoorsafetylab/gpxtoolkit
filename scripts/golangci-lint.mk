TOOLCHAIN ?= .tool
GOLANGCI_LINT := $(TOOLCHAIN)/bin/golangci-lint
GOLANGCI_LINT_VERSION ?= 1.49.0

$(GOLANGCI_LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLCHAIN)/bin v$(GOLANGCI_LINT_VERSION)

clean/golangci-lint:
	rm -f $(GOLANGCI_LINT)

.PHONY: clean/golangci-lint
