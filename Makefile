# Copyright (c) The Kowabunga Project
# Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
# SPDX-License-Identifier: Apache-2.0

BINDIR = bin
BIN = terraform-provider-kowabunga
LDFLAGS += -X main.version=$$(git describe --always --abbrev=40 --dirty)

GOVULNCHECK = $(BINDIR)/govulncheck
GOVULNCHECK_VERSION = v1.1.4

GOLINT = $(BINDIR)/golangci-lint
GOLINT_VERSION = v2.0.2

GOSEC = $(BINDIR)/gosec
GOSEC_VERSION = v2.22.3

V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")

.PHONY: all
all: mod fmt lint vet $(BIN) ; @

# Updates all go modules
update: ; $(info $(M) updating modules…) @
	$Q go get -u ./...
	$Q go mod tidy

.PHONY: mod
mod: ; $(info $(M) collecting modules…) @
	$Q go mod download
	$Q go mod tidy

.PHONY: fmt
fmt: ; $(info $(M) formatting code…) @
	$Q go fmt ./internal/provider .

.PHONY: get-lint
get-lint: ; $(info $(M) downloading go-lint…) @
	$Q test -x $(GOLINT) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $(GOLINT_VERSION)

.PHONY: lint
lint: get-lint ; $(info $(M) running linter…) @
	$Q $(GOLINT) run ./... ; exit 0

.PHONY: get-govulncheck
get-govulncheck: ; $(info $(M) downloading govulncheck…) @
	$Q test -x $(GOVULNCHECK) || GOBIN="$(PWD)/$(BINDIR)/" go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

.PHONY: vuln
vuln: get-govulncheck ; $(info $(M) running govulncheck…) @ ## Check for known vulnerabilities
	$Q $(GOVULNCHECK) ./... ; exit 0

.PHONY: get-gosec
get-gosec: ; $(info $(M) downloading gosec…) @
	$Q test -x $(GOSEC) || GOBIN="$(PWD)/$(BINDIR)/" go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)

.PHONY: sec
sec: get-gosec ; $(info $(M) running gosec…) @ ## AST / SSA code checks
	$Q $(GOSEC) -terse -exclude=G101,G115 ./... ; exit 0

.PHONY: vet
vet: ; $(info $(M) running vetter…) @
	$Q go vet ./internal/provider .

.PHONY: doc
doc: ; $(info $(M) generating documentation…) @
	$Q go generate ./...

.PHONY: $(BIN)
$(BIN): ; $(info $(M) building terraform provider plugin…) @
	$Q go build -ldflags "${LDFLAGS}"

.PHONY: install
install: ; $(info $(M) installing terraform provider plugin…) @
	$Q go install -ldflags "${LDFLAGS}"

.PHONY: clean
clean: ; $(info $(M) cleanup…) @
	$Q rm -f $(BIN)
	$Q rm -rf $(BINDIR)
