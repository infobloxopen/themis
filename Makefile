SRCROOT := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
BUILDPATH = $(SRCROOT)/build
COVEROUT=$(SRCROOT)/cover.out
COVERTMP=/tmp/cover.out

AT = cd $(SRCROOT)
RM = rm -fv
GOBUILD = go build -v
GOFMTCHECK = test -z `gofmt -s -d *.go | tee /dev/stderr`
GOTEST = go test -v
COVER = $(GOTEST) -coverprofile=$(COVERTMP) -covermode=atomic
JOINCOVER = cat $(COVERTMP) >> $(COVEROUT)
GOTESTRACE = $(COVER) -race && $(JOINCOVER)
GOTESTINTEGRACE = $(COVER) -race -coverpkg github.com/infobloxopen/themis/pdp,github.com/infobloxopen/themis/pdp/ast,github.com/infobloxopen/themis/pdp/ast/yast,github.com/infobloxopen/themis/pdp/ast/jast && $(JOINCOVER)
GOBENCH = $(GOTEST) -run=\^\$$ -bench=
GOBENCHALL = $(GOBENCH).

.PHONY: all
all: fmt build test bench

.PHONY: build-dir
build-dir:
	mkdir -p $(BUILDPATH)

.PHONY: bootstrap
bootstrap: build-dir

.PHONY: vendor
vendor:
	GO111MODULE=on go mod vendor
	GO111MODULE=on go mod tidy

.PHONY: clean
clean:
	@$(RM) $(COVEROUT)
	@$(RM) $(BUILDPATH)

.PHONY: fmt
fmt: fmt-pdp fmt-pdp-yast fmt-pdp-jast fmt-pdp-jcon fmt-pdp-itests fmt-local-selector fmt-pip-selector fmt-pdpctrl-client fmt-papcli fmt-pep fmt-pepcli fmt-pepcli-requests fmt-pepcli-test fmt-pepcli-perf fmt-pdpserver-pkg fmt-pdpserver fmt-pip-server fmt-pip-client fmt-pip-gen fmt-pip-genpkg fmt-pipjcon fmt-pipcli fmt-pipcli-global fmt-pipcli-subflags fmt-pipcli-test fmt-pipcli-perf fmt-egen

.PHONY: build
build: build-dir build-pepcli build-papcli build-pdpserver build-egen build-pip-gen build-pipjcon build-pipcli

.PHONY: test
test: cover-out test-pdp test-pdp-integration test-pdp-yast test-pdp-jast test-pdp-jcon test-local-selector test-pip-selector test-pep test-pip-server test-pip-client test-pip-genpkg test-plugin

.PHONY: bench
bench: bench-pep bench-pip-server bench-pip-client bench-pdpserver-pkg bench-plugin

.PHONY: cover-out
cover-out:
	echo > $(COVEROUT)

# Per package format targets
.PHONY: fmt-pdp
fmt-pdp:
	@echo "Checking PDP format..."
	@$(AT)/pdp && $(GOFMTCHECK)

.PHONY: fmt-pdp-yast
fmt-pdp-yast:
	@echo "Checking PDP YAST format..."
	@$(AT)/pdp/ast/yast && $(GOFMTCHECK)

.PHONY: fmt-pdp-jast
fmt-pdp-jast:
	@echo "Checking PDP JAST format..."
	@$(AT)/pdp/ast/jast && $(GOFMTCHECK)

.PHONY: fmt-pdp-jcon
fmt-pdp-jcon:
	@echo "Checking PDP JCon format..."
	@$(AT)/pdp/jcon && $(GOFMTCHECK)

.PHONY: fmt-pdp-itests
fmt-pdp-itests:
	@echo "Checking PDP integration tests format..."
	@$(AT)/pdp/integration_tests && $(GOFMTCHECK)

.PHONY: fmt-local-selector
fmt-local-selector:
	@echo "Checking PDP local selector format..."
	@$(AT)/pdp/selector/local && $(GOFMTCHECK)

.PHONY: fmt-pip-selector
fmt-pip-selector:
	@echo "Checking PDP PIP selector format..."
	@$(AT)/pdp/selector/pip && $(GOFMTCHECK)

.PHONY: fmt-pdpctrl-client
fmt-pdpctrl-client:
	@echo "Checking PDP control client library format..."
	@$(AT)/pdpctrl-client && $(GOFMTCHECK)

.PHONY: fmt-papcli
fmt-papcli:
	@echo "Checking PAP CLI format..."
	@$(AT)/papcli && $(GOFMTCHECK)

.PHONY: fmt-pep
fmt-pep:
	@echo "Checking PEP client library format..."
	@$(AT)/pep && $(GOFMTCHECK)

.PHONY: fmt-pepcli
fmt-pepcli:
	@echo "Checking PEP CLI format..."
	@$(AT)/pepcli && $(GOFMTCHECK)

.PHONY: fmt-pepcli-requests
fmt-pepcli-requests:
	@echo "Checking PEP CLI requests package format..."
	@$(AT)/pepcli/requests && $(GOFMTCHECK)

.PHONY: fmt-pepcli-test
fmt-pepcli-test:
	@echo "Checking PEP CLI test package format..."
	@$(AT)/pepcli/test && $(GOFMTCHECK)

.PHONY: fmt-pepcli-perf
fmt-pepcli-perf:
	@echo "Checking PEP CLI perf package format..."
	@$(AT)/pepcli/perf && $(GOFMTCHECK)

.PHONY: fmt-pdpserver-pkg
fmt-pdpserver-pkg:
	@echo "Checking PDP server package format..."
	@$(AT)/pdpserver/server && $(GOFMTCHECK)

.PHONY: fmt-pdpserver
fmt-pdpserver:
	@echo "Checking PDP server format..."
	@$(AT)/pdpserver && $(GOFMTCHECK)

.PHONY: fmt-pip-server
fmt-pip-server:
	@echo "Checking PIP server package format..."
	@$(AT)/pip/server && $(GOFMTCHECK)

.PHONY: fmt-pip-client
fmt-pip-client:
	@echo "Checking PIP client package format..."
	@$(AT)/pip/client && $(GOFMTCHECK)

.PHONY: fmt-pip-gen
fmt-pip-gen:
	@echo "Checking PIP handler generator format..."
	@$(AT)/pip/mkpiphandler && $(GOFMTCHECK)

.PHONY: fmt-pip-genpkg
fmt-pip-genpkg:
	@echo "Checking PIP handler generator package format..."
	@$(AT)/pip/mkpiphandler/pkg && $(GOFMTCHECK)

.PHONY: fmt-pipjcon
fmt-pipjcon:
	@echo "Checking PIP JCon server format..."
	@$(AT)/pip/pipjcon && $(GOFMTCHECK)

.PHONY: fmt-pipcli
fmt-pipcli:
	@echo "Checking PIP CLI format..."
	@$(AT)/pip/pipcli && $(GOFMTCHECK)

.PHONY: fmt-pipcli-global
fmt-pipcli-global:
	@echo "Checking PIP CLI global options package format..."
	@$(AT)/pip/pipcli/global && $(GOFMTCHECK)

.PHONY: fmt-pipcli-subflags
fmt-pipcli-subflags:
	@echo "Checking PIP CLI command flags package format..."
	@$(AT)/pip/pipcli/subflags && $(GOFMTCHECK)

.PHONY: fmt-pipcli-test
fmt-pipcli-test:
	@echo "Checking PIP CLI test command package format..."
	@$(AT)/pip/pipcli/test && $(GOFMTCHECK)

.PHONY: fmt-pipcli-perf
fmt-pipcli-perf:
	@echo "Checking PIP CLI perf command package format..."
	@$(AT)/pip/pipcli/perf && $(GOFMTCHECK)

.PHONY: fmt-egen
fmt-egen:
	@echo "Checking EGen format..."
	@$(AT)/egen && $(GOFMTCHECK)

.PHONY: linter
linter:
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install golint
	gometalinter --deadline=1m --disable-all --enable=gofmt --enable=golint --enable=vet --exclude=^vendor/ --exclude=^pb/ ./...

# Per package build targets
.PHONY: build-pepcli
build-pepcli: build-dir
	$(AT)/pepcli && $(GOBUILD) -o $(BUILDPATH)/pepcli

.PHONY: build-papcli
build-papcli: build-dir
	$(AT)/papcli && $(GOBUILD) -o $(BUILDPATH)/papcli

.PHONY: build-pdpserver
build-pdpserver: build-dir
	$(AT)/pdpserver && $(GOBUILD) -o $(BUILDPATH)/pdpserver

.PHONY: build-egen
build-egen: build-dir
	$(AT)/egen && $(GOBUILD) -o $(BUILDPATH)/egen

.PHONY: build-pip-gen
build-pip-gen: build-dir
	$(AT)/pip/mkpiphandler && $(GOBUILD) -o $(BUILDPATH)/mkpiphandler

.PHONY: build-pipjcon
build-pipjcon: build-dir
	$(AT)/pip/pipjcon && $(GOBUILD) -o $(BUILDPATH)/pipjcon

.PHONY: build-pipcli
build-pipcli: build-dir
	$(AT)/pip/pipcli && $(GOBUILD) -o $(BUILDPATH)/pipcli

.PHONY: test-pdp
test-pdp: cover-out
	$(AT)/pdp && $(GOTESTRACE)

.PHONY: test-pdp-integration
test-pdp-integration: cover-out
	$(AT)/pdp/integration_tests && $(GOTESTINTEGRACE)

.PHONY: test-pdp-yast
test-pdp-yast: cover-out
	$(AT)/pdp/ast/yast && $(GOTESTRACE)

.PHONY: test-pdp-jast
test-pdp-jast: cover-out
	$(AT)/pdp/ast/jast && $(GOTESTRACE)

.PHONY: test-pdp-jcon
test-pdp-jcon: cover-out
	$(AT)/pdp/jcon && $(GOTESTRACE)

.PHONY: test-local-selector
test-local-selector: cover-out
	$(AT)/pdp/selector/local && $(GOTESTRACE)

.PHONY: test-pip-selector
test-pip-selector: cover-out
	$(AT)/pdp/selector/pip && $(GOTESTRACE)

.PHONY: test-pep
test-pep: build-pdpserver cover-out
	$(AT)/pep && $(GOTESTRACE)

.PHONY: test-pip-server
test-pip-server:
	$(AT)/pip/server && $(GOTESTRACE)

.PHONY: test-pip-client
test-pip-client:
	$(AT)/pip/client && $(GOTESTRACE)

.PHONY: test-pip-genpkg
test-pip-genpkg:
	$(AT)/pip/mkpiphandler/pkg && $(GOTESTRACE)

.PHONY: test-plugin
test-plugin: cover-out
	$(AT)/contrib/coredns/policy && $(GOTESTRACE)

.PHONY: bench-pep
bench-pep: build-pdpserver
	$(AT)/pep && $(GOBENCHALL)

.PHONY: bench-pip-server
bench-pip-server:
	$(AT)/pip/server && $(GOBENCHALL)

.PHONY: bench-pip-client
bench-pip-client:
	$(AT)/pip/client && $(GOBENCHALL)

.PHONY: bench-pdpserver-pkg
bench-pdpserver-pkg:
	$(AT)/pdpserver/server && $(GOBENCHALL)

.PHONY: bench-plugin
bench-plugin:
	$(AT)/contrib/coredns/policy && $(GOBENCHALL)
